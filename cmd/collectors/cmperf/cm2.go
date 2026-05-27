package cmperf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/cmmetrics"
	rest2 "github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (c *CmPerf) buildCounters() {
	mat := c.Matrix[c.Object]
	for name, propMetric := range c.Prop.Metrics {
		if mat.GetMetric(name) == nil {
			m, mErr := mat.NewMetricFloat64(name, propMetric.Label)
			if mErr != nil {
				c.Logger.Error("add metric", slogx.Err(mErr), slog.String("name", name))
				continue
			}
			m.SetExportable(propMetric.Exportable)
		}
	}
}

// counterTypeString maps a CounterTypeEnum to its cook-pipeline string name.
func counterTypeString(t cmmetrics.CounterTypeEnum) string {
	switch t {
	case cmmetrics.CookRaw:
		return "raw"
	case cmmetrics.CookRate:
		return "rate"
	case cmmetrics.CookDelta:
		return "delta"
	case cmmetrics.CookAverage:
		return "average"
	case cmmetrics.CookPercent:
		return "percent"
	case cmmetrics.CookString:
		return "string"
	default:
		return "raw"
	}
}

// buildCountersFromSchema populates counterInfo from the embedded schema and registers denominator metrics.
func (c *CmPerf) buildCountersFromSchema(schema cmmetrics.ObjectSchema, curMat *matrix.Matrix) {
	mat := c.Matrix[c.Object]

	indexToName := make(map[uint32]string, len(schema.CounterSchema))
	schemaMap := make(map[uint32]cmmetrics.CounterSchema, len(schema.CounterSchema))
	// TODO cache it?
	for _, cs := range schema.CounterSchema {
		indexToName[cs.Index] = cs.Name
		schemaMap[cs.Index] = cs
	}
	c.perfProp.schemaMap = schemaMap

	for _, cs := range schema.CounterSchema {
		name := cs.Name
		ctrType := counterTypeString(cs.Type)
		denominator := indexToName[cs.BaseIndex]

		if ov := c.GetOverride(name); ov != "" {
			ctrType = ov
		}

		c.perfProp.counterInfo[name] = &counter{
			counterType: ctrType,
			denominator: denominator,
		}

		if propMetric, inTemplate := c.Prop.Metrics[name]; inTemplate {
			if mat.GetMetric(name) == nil {
				m, mErr := mat.NewMetricFloat64(name, propMetric.Label)
				if mErr != nil {
					c.Logger.Error("add metric from schema", slogx.Err(mErr), slog.String("name", name))
				} else {
					m.SetExportable(propMetric.Exportable)
				}
			}
		}

		if denominator != "" {
			if _, inTemplate := c.Prop.Metrics[name]; inTemplate {
				if _, exists := c.Prop.Metrics[denominator]; !exists {
					c.Prop.Metrics[denominator] = &rest2.Metric{
						Label:      denominator,
						Name:       denominator,
						Exportable: false,
					}
				}
				if mat.GetMetric(denominator) == nil {
					m, mErr := mat.NewMetricFloat64(denominator, denominator)
					if mErr != nil {
						c.Logger.Error("add denominator metric from schema", slogx.Err(mErr), slog.String("name", denominator))
					} else {
						m.SetExportable(false)
					}
				}
				if curMat != nil && curMat.GetMetric(denominator) == nil {
					m, mErr := curMat.NewMetricFloat64(denominator, denominator)
					if mErr != nil {
						c.Logger.Error("add denominator metric to curMat from schema", slogx.Err(mErr), slog.String("name", denominator))
					} else {
						m.SetExportable(false)
					}
				}
			}
		}
	}
}

type cm2FileRecord struct {
	Object      string
	SpiURL      string
	ChecksumURL string
	Timestamp   time.Time
	NodeUUID    string
}

func (c *CmPerf) pollONTAPFilesEndpoint(query string) ([]cm2FileRecord, error) {
	path := "api/cluster/counter-cache/files?fields=path,checksum_path,node,timestamp,object&object=" + url.QueryEscape(query) + "&order_by=timestamp+desc&max_records=1"
	if !c.lastTimestamp.IsZero() {
		path += "&timestamp=>" + url.QueryEscape(c.lastTimestamp.UTC().Format(time.RFC3339))
	}

	data, err := c.Client.GetRest(&c.RequestMetadata, path)
	if err != nil {
		return nil, fmt.Errorf("pollONTAPFilesEndpoint %s: %w", path, err)
	}

	return c.parseCM2FilesResponse(bytes.NewReader(data))
}

type cm2FilesAPIResponse struct {
	NumRecords int `json:"num_records"`
	Records    []struct {
		Object       string `json:"object"`
		Path         string `json:"path"`
		ChecksumPath string `json:"checksum_path"`
		Timestamp    string `json:"timestamp"`
		Node         struct {
			UUID string `json:"uuid"`
		} `json:"node"`
	} `json:"records"`
}

func (c *CmPerf) parseCM2FilesResponse(body io.Reader) ([]cm2FileRecord, error) {
	var apiResp cm2FilesAPIResponse
	if err := json.NewDecoder(body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("parseCM2FilesResponse: %w", err)
	}

	records := make([]cm2FileRecord, 0, len(apiResp.Records))
	for _, r := range apiResp.Records {
		if r.Timestamp == "" {
			c.Logger.Warn("skipping CM2 file record with missing timestamp",
				slog.String("object", r.Object),
				slog.String("path", r.Path))
			continue
		}
		ts, parseErr := time.Parse(time.RFC3339, r.Timestamp)
		if parseErr != nil {
			c.Logger.Warn("skipping CM2 file record with unparseable timestamp",
				slog.String("timestamp", r.Timestamp),
				slog.String("object", r.Object),
				slog.String("path", r.Path),
				slogx.Err(parseErr))
			continue
		}
		records = append(records, cm2FileRecord{
			Object:      r.Object,
			SpiURL:      r.Path,
			ChecksumURL: r.ChecksumPath,
			Timestamp:   ts,
			NodeUUID:    r.Node.UUID,
		})
	}
	return records, nil
}

func (c *CmPerf) downloadCM2Files(dir string) (string, time.Time, error) {
	records, err := c.pollONTAPFilesEndpoint(c.Prop.Query)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("downloadCM2Files: poll: %w", err)
	}
	if len(records) == 0 {
		c.Logger.Warn("no CM2 files available", slog.String("query", c.Prop.Query))
		return "", time.Time{}, nil
	}

	rec := records[0]
	if !rec.Timestamp.IsZero() && !rec.Timestamp.After(c.lastTimestamp) {
		c.Logger.Warn("no new CM2 file",
			slog.String("path", rec.SpiURL),
			slog.Time("fileTS", rec.Timestamp),
			slog.Time("lastTS", c.lastTimestamp))
		return "", time.Time{}, nil
	}

	c.Logger.Debug("downloading CM2 file",
		slog.String("file", rec.SpiURL),
		slog.Time("fileTS", rec.Timestamp),
		slog.Time("prevTS", c.lastTimestamp))

	path, dlErr := c.downloadSPIFile(rec, dir)
	if dlErr != nil {
		c.Logger.Warn("downloadSPIFile failed", slog.String("url", rec.SpiURL), slogx.Err(dlErr))
		return "", time.Time{}, dlErr
	}

	// lastTimestamp is advanced by the caller only after the file is successfully parsed.
	return path, rec.Timestamp, nil
}

func (c *CmPerf) downloadSPIFile(rec cm2FileRecord, dir string) (string, error) {
	entries, rdErr := os.ReadDir(dir)
	if rdErr != nil {
		c.Logger.Debug("could not read CM2 temp dir for cleanup", slog.String("dir", dir), slogx.Err(rdErr))
	} else {
		for _, entry := range entries {
			if !entry.IsDir() {
				_ = os.Remove(filepath.Clean(filepath.Join(dir, entry.Name())))
			}
		}
	}
	fname := fmt.Sprintf("%d_%s.pb", time.Now().UnixMilli(), c.Prop.Query)
	path := filepath.Join(dir, fname)
	return path, c.downloadSPIFileONTAP(rec.SpiURL, path)
}

func (c *CmPerf) downloadSPIFileONTAP(spiRelPath, destPath string) error {
	restPath := "spi" + spiRelPath
	destPath = filepath.Clean(destPath)

	data, err := c.Client.GetPlainRest(&c.RequestMetadata, restPath, false, map[string]string{
		"Accept": "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("downloadSPIFileONTAP %s: %w", restPath, err)
	}

	if err := os.WriteFile(destPath, data, 0600); err != nil {
		return fmt.Errorf("downloadSPIFileONTAP write %s: %w", destPath, err)
	}

	// TODO validate checksum? checksum URL doesn't seem to be working.
	return nil
}

func (c *CmPerf) pollCM2Files(path string, curMat *matrix.Matrix) (uint64, uint64, error) {
	var (
		fileSchema   *cmmetrics.ObjectSchema
		fileSummary  *cmmetrics.CollectionStatus
		readErr      error
		schemaLoaded bool
		metricCount  uint64
		numPartials  uint64
	)
	for rec, msgErr := range cmmetrics.Messages(path) {
		if msgErr != nil {
			c.Logger.Warn("error reading CM2 pb file",
				slog.String("file", filepath.Base(path)), slogx.Err(msgErr))
			readErr = msgErr
			break
		}
		if rec.Schema != nil {
			fileSchema = rec.Schema
		}
		if rec.Batch != nil {
			batch := rec.Batch
			schema := batch.Schema
			if len(schema.CounterSchema) == 0 && fileSchema != nil {
				schema = *fileSchema
				batch.Schema = schema
			}
			if !schemaLoaded && len(schema.CounterSchema) > 0 {
				c.buildCountersFromSchema(schema, curMat)
				schemaLoaded = true
			}
			metricCount += c.populateMatrix(batch, curMat)
		}
		if rec.Summary != nil {
			fileSummary = rec.Summary
		}
	}

	if fileSummary == nil {
		c.Logger.Warn("no collection status summary in CM2 file — treating as incomplete",
			slog.String("file", filepath.Base(path)))
	}
	isComplete := false
	if fileSummary != nil {
		for _, sc := range fileSummary.Statuses {
			switch sc.Code {
			case cmmetrics.CompleteCollection:
				isComplete = true
			case cmmetrics.NoAdditionalStatus:
				// ONTAP always appends this alongside other codes
			default:
				c.Logger.Warn("non-complete collection status",
					slog.String("file", filepath.Base(path)),
					slog.Uint64("status", uint64(sc.Code)),
					slog.Any("nodes", sc.Nodes))
			}
		}
	}
	if !isComplete {
		if c.AllowPartialAggregation {
			c.Logger.Debug("incomplete collection allowed by template",
				slog.String("file", filepath.Base(path)))
		} else {
			c.Logger.Debug("incomplete collection — instances will be skipped",
				slog.String("file", filepath.Base(path)))
			for _, inst := range curMat.GetInstances() {
				inst.SetPartial(true)
				inst.SetExportable(false)
			}
			numPartials += uint64(len(curMat.GetInstances()))
			metricCount = 0
		}
	}

	if !schemaLoaded && readErr == nil {
		c.Logger.Warn("no schema loaded from CM2 pb file — file may be empty or corrupt",
			slog.String("file", filepath.Base(path)))
	}

	if removeErr := os.Remove(filepath.Clean(path)); removeErr != nil {
		c.Logger.Warn("failed to remove CM2 pb file",
			slog.String("file", filepath.Base(path)), slogx.Err(removeErr))
	}

	return metricCount, numPartials, readErr
}

func (c *CmPerf) populateMatrix(oc *cmmetrics.ObjectCollection, curMat *matrix.Matrix) uint64 {
	schemaMap := c.perfProp.schemaMap

	ts := float64(oc.Timestamp) / 1000.0

	var metricCount uint64

	for _, inst := range oc.Data.Instances {
		stringVals := make(map[string]string, 8)

		if inst.Name != "" {
			stringVals["instance_name"] = inst.Name
		}
		if inst.UUID != "" {
			stringVals["instance_uuid"] = inst.UUID
		}

		for _, ct := range inst.Counters {
			cs, csOK := schemaMap[ct.Index]
			if !csOK {
				continue
			}
			if sv, strOK := ct.StringValue(); strOK {
				stringVals[cs.Name] = sv
			}
		}

		instanceKey := c.buildInstanceKey(inst, stringVals)
		if instanceKey == "" {
			c.Logger.Debug("skip instance, key is empty",
				slog.String("name", inst.Name),
				slog.String("uuid", inst.UUID))
			continue
		}

		matInst := curMat.GetInstance(instanceKey)
		if matInst == nil {
			var newErr error
			matInst, newErr = curMat.NewInstance(instanceKey)
			if newErr != nil {
				c.Logger.Warn("failed to create instance",
					slog.String("key", instanceKey), slogx.Err(newErr))
				continue
			}
		}

		if tsMetric := curMat.GetMetric(timestampMetricName); tsMetric != nil {
			tsMetric.SetValueFloat64(matInst, ts)
		}

		for name, sv := range stringVals {
			display, ok := c.Prop.InstanceLabels[name]
			if !ok {
				continue
			}
			matInst.SetLabel(display, sv)
		}

		for _, ct := range inst.Counters {
			cs, ok := schemaMap[ct.Index]
			if !ok || ct.IsString() {
				continue
			}
			counterName := cs.Name

			if vals64, ok := ct.List64(); ok {
				metricCount += c.populateArrayCounter(curMat, matInst, cs, vals64)
				continue
			}
			if vals32, ok := ct.List32(); ok {
				u64 := make([]uint64, len(vals32))
				for i, v := range vals32 {
					u64[i] = uint64(v)
				}
				metricCount += c.populateArrayCounter(curMat, matInst, cs, u64)
				continue
			}

			metric := curMat.GetMetric(counterName)
			if metric == nil {
				continue
			}

			if v, ok := ct.Uint64Value(); ok {
				metric.SetValueUint64(matInst, v)
				metricCount++
			} else if v32, ok := ct.Uint32Value(); ok {
				metric.SetValueUint64(matInst, uint64(v32))
				metricCount++
			}
		}
	}

	return metricCount
}

func (c *CmPerf) buildInstanceKey(inst cmmetrics.ObjectInstance, stringVals map[string]string) string {
	if len(c.Prop.InstanceKeys) > 0 {
		var b strings.Builder
		var anyNonEmpty bool
		for i, k := range c.Prop.InstanceKeys {
			v := stringVals[k]
			if v == "" {
				c.Logger.Debug("instance key counter has no value",
					slog.String("counter", k),
					slog.String("uuid", inst.UUID))
			} else {
				anyNonEmpty = true
			}
			if i > 0 {
				b.WriteByte(':')
			}
			b.WriteString(v)
		}
		if !anyNonEmpty {
			return ""
		}
		return b.String()
	}
	if inst.UUID != "" {
		return inst.UUID
	}
	return inst.Name
}

func (c *CmPerf) populateArrayCounter(
	curMat *matrix.Matrix,
	inst *matrix.Instance,
	cs cmmetrics.CounterSchema,
	values []uint64,
) uint64 {
	propMetric := c.Prop.Metrics[cs.Name]
	if propMetric == nil {
		return 0
	}

	// 2D path: cross-product of LabelsX × LabelsY.
	// TODO validate against data. wafl_hya_sizer has this kind of data but vsim has it as zero
	if len(cs.LabelsY) > 0 {
		labelsX, labelsY := cs.LabelsX, cs.LabelsY
		if len(labelsX) == 0 || len(labelsY) == 0 {
			c.Logger.Warn("2D array counter has empty LabelsX or LabelsY, skipping",
				slog.String("counter", cs.Name),
			)
			return 0
		}
		var count uint64
		for i, labelX := range labelsX {
			for j, labelY := range labelsY {
				idx := i*len(labelsY) + j
				if idx >= len(values) {
					c.Logger.Warn("2D array counter: values exhausted before labels",
						slog.String("counter", cs.Name),
						slog.Int("expected", len(labelsX)*len(labelsY)),
						slog.Int("got", len(values)),
					)
					return count
				}
				k := cs.Name + arrayKeyToken + labelX + arrayKeyToken + labelY
				metr := curMat.GetMetric(k)
				if metr == nil {
					var err error
					metr, err = curMat.NewMetricFloat64(k, propMetric.Label)
					if err != nil {
						continue
					}
					metr.SetArray(true)
					metr.SetExportable(propMetric.Exportable)
					metr.SetLabel("metric", labelX)
					metr.SetLabel("submetric", labelY)
				}
				metr.SetValueFloat64(inst, float64(values[idx]))
				count++
			}
		}
		return count
	}

	// 1D path: LabelsX only.
	labels := cs.LabelsX
	// TODO check if this is correct way to identify histograms
	isHisto := len(labels) > 0 && strings.Contains(strings.ToLower(cs.Name), "_hist")

	if isHisto {
		bucketKey := cs.Name + ".bucket"
		if curMat.GetMetric(bucketKey) == nil {
			bm, err := curMat.NewMetricFloat64(bucketKey, propMetric.Label)
			if err == nil {
				bm.SetArray(true)
				bm.SetHistogram(true)
				bm.SetExportable(propMetric.Exportable)
				bm.SetBuckets(&labels)
			}
		}
	}

	var count uint64
	for i, label := range labels {
		if i >= len(values) {
			break
		}
		k := cs.Name + arrayKeyToken + label
		metr := curMat.GetMetric(k)
		if metr == nil {
			var err error
			metr, err = curMat.NewMetricFloat64(k, propMetric.Label)
			if err != nil {
				continue
			}
			metr.SetArray(true)
			metr.SetExportable(propMetric.Exportable)
			metr.SetLabel("metric", label)
			if isHisto {
				metr.SetHistogram(true)
				metr.SetLabel("comment", strconv.Itoa(i))
				metr.SetLabel("bucket", cs.Name+".bucket")
			}
		}
		metr.SetValueFloat64(inst, float64(values[i]))
		count++
	}
	return count
}
