package cmperf

import (
	"bytes"
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/cmmetrics"
	rest2 "github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const cmperfRetainFilesEnv = "HARVEST_CMPERF_RETAIN_FILES"

// retainCmperfFiles returns how many CM2 pb files to keep for debug. Unset, empty,
// invalid, or negative values mean 0 (delete after parse / wipe before download).
func retainCmperfFiles() int {
	raw := os.Getenv(cmperfRetainFilesEnv)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// pruneCmperfTempDir removes non-directory entries from dir. When retain <= 0 all
// files are removed. When retain > 0 the newest retain files are kept (sorted by
// {unixMilli}_ filename prefix, falling back to ModTime).
func pruneCmperfTempDir(dir string, retain int, logger *slog.Logger) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	type tempFile struct {
		path    string
		sortKey int64
	}
	files := make([]tempFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		files = append(files, tempFile{
			path:    filepath.Clean(filepath.Join(dir, name)),
			sortKey: cmperfTempFileSortKey(name, entry),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].sortKey > files[j].sortKey
	})

	start := 0
	if retain > 0 {
		if retain >= len(files) {
			return nil
		}
		start = retain
	}
	for _, f := range files[start:] {
		if removeErr := os.Remove(f.path); removeErr != nil && logger != nil {
			logger.Warn("failed to remove CM2 pb file",
				slog.String("file", filepath.Base(f.path)), slogx.Err(removeErr))
		}
	}
	return nil
}

func cmperfTempFileSortKey(name string, entry os.DirEntry) int64 {
	prefix, _, ok := strings.Cut(name, "_")
	if ok {
		if ms, err := strconv.ParseInt(prefix, 10, 64); err == nil {
			return ms
		}
	}
	info, err := entry.Info()
	if err != nil {
		return 0
	}
	return info.ModTime().UnixMilli()
}

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

func hasHistogramSuffix(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, "_hist") ||
		strings.HasSuffix(lower, "_histo") ||
		strings.HasSuffix(lower, "_histogram")
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

		// The CM2 protobuf schema has no counter description (unlike ZapiPerf/RestPerf), so a
		// counter is treated as a histogram if its name matches the known heuristic, or the
		// template's `histograms:` section lists it.
		isHisto := len(cs.LabelsX) > 0 && (hasHistogramSuffix(name) || c.perfProp.histogramCounters[name])

		c.perfProp.counterInfo[name] = &counter{
			counterType: ctrType,
			denominator: denominator,
			isHistogram: isHisto,
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
	path := "api/cluster/counter-cache/files?fields=path,checksum_path,node,timestamp,object&object=" +
		url.QueryEscape(query) + "&sample_period=" + url.QueryEscape(c.perfProp.samplePeriod) + "&order_by=timestamp+desc&max_records=1"
	if !c.lastTimestamp.IsZero() {
		path += "&timestamp=>" + url.QueryEscape(c.lastTimestamp.UTC().Format(time.RFC3339))
	}
	c.Logger.Debug("polling CM2 files endpoint", slog.String("path", path))

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
	if pruneErr := pruneCmperfTempDir(dir, retainCmperfFiles(), c.Logger); pruneErr != nil {
		c.Logger.Debug("could not read CM2 temp dir for cleanup", slog.String("dir", dir), slogx.Err(pruneErr))
	}
	fname := fmt.Sprintf("%d_%s.pb", time.Now().UnixMilli(), c.Prop.Query)
	path := filepath.Join(dir, fname)
	if err := c.downloadSPIFileONTAP(rec.SpiURL, path); err != nil {
		return "", err
	}
	if err := c.verifyChecksum(path, rec.ChecksumURL); err != nil {
		_ = os.Remove(filepath.Clean(path))
		return "", err
	}
	return path, nil
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

	return nil
}

func (c *CmPerf) verifyChecksum(filePath, checksumURL string) error {
	if checksumURL == "" {
		return nil
	}

	restPath := "spi" + checksumURL
	data, err := c.Client.GetPlainRest(&c.RequestMetadata, restPath, false, map[string]string{
		"Accept": "text/plain",
	})
	if err != nil {
		return fmt.Errorf("verifyChecksum fetch %s: %w", restPath, err)
	}

	line := strings.TrimSpace(string(data))
	hexStr := strings.SplitN(line, " ", 2)[0]
	if len(hexStr) != 32 {
		return fmt.Errorf("verifyChecksum: unexpected format %q", line)
	}

	pb, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("verifyChecksum: read file: %w", err)
	}

	// #nosec G401 -- MD5 used for data integrity verification, not cryptography
	actual := md5.Sum(pb)
	actualHex := hex.EncodeToString(actual[:])
	if actualHex != hexStr {
		return fmt.Errorf("verifyChecksum: mismatch for %s: expected %s, got %s",
			filepath.Base(filePath), hexStr, actualHex)
	}

	return nil
}

// isCompleteCollection reports true when the only status codes present are CompleteCollection and optionally NoAdditionalStatus (any order; may repeat).
func isCompleteCollection(statuses []cmmetrics.StatusCode) bool {
	hasComplete := false
	for _, sc := range statuses {
		switch sc.Code {
		case cmmetrics.CompleteCollection:
			hasComplete = true
		case cmmetrics.NoAdditionalStatus:
			// ONTAP always appends this alongside other codes
		default:
			return false
		}
	}
	return hasComplete
}

func (c *CmPerf) pollCM2Files(path string, curMat *matrix.Matrix, prevMat *matrix.Matrix) (uint64, uint64, error) {
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
			c.Logger.Error("error reading CM2 pb file",
				slog.String("file", filepath.Base(path)), slogx.Err(msgErr))
			readErr = msgErr
			break
		}
		if rec.Schema != nil {
			fileSchema = rec.Schema
		}
		if rec.Batch != nil {
			if !schemaLoaded && fileSchema != nil && len(fileSchema.CounterSchema) > 0 {
				c.buildCountersFromSchema(*fileSchema, curMat)
				schemaLoaded = true
			}
			metricCount += c.populateMatrix(rec.Batch, curMat, prevMat)
		}
		if rec.Summary != nil {
			fileSummary = rec.Summary
		}
	}

	if fileSummary == nil {
		c.Logger.Warn("no collection status summary in CM2 file — treating as incomplete",
			slog.String("file", filepath.Base(path)))
	}
	isComplete := fileSummary != nil && isCompleteCollection(fileSummary.Statuses)
	if fileSummary != nil && !isComplete {
		c.Logger.Debug("collection status is not complete",
			slog.String("file", filepath.Base(path)),
			slog.Any("statuses", fileSummary.Statuses))
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
		c.Logger.Debug("no schema loaded from CM2 pb file — file may be empty or corrupt",
			slog.String("file", filepath.Base(path)))
	}

	if retain := retainCmperfFiles(); retain > 0 {
		if pruneErr := pruneCmperfTempDir(filepath.Dir(path), retain, c.Logger); pruneErr != nil {
			c.Logger.Debug("could not prune CM2 temp dir",
				slog.String("dir", filepath.Dir(path)), slogx.Err(pruneErr))
		}
	} else if removeErr := os.Remove(filepath.Clean(path)); removeErr != nil {
		c.Logger.Warn("failed to remove CM2 pb file",
			slog.String("file", filepath.Base(path)), slogx.Err(removeErr))
	}

	return metricCount, numPartials, readErr
}

func (c *CmPerf) populateMatrix(oc *cmmetrics.ObjectCollection, curMat *matrix.Matrix, prevMat *matrix.Matrix) uint64 {
	schemaMap := c.perfProp.schemaMap

	ts := float64(oc.Timestamp) / 1000.0

	var metricCount uint64

	tsMetric := curMat.MustGetMetric(timestampMetricName)

	for _, inst := range oc.Data.Instances {
		stringVals := make(map[string]string)

		if inst.Name != "" {
			stringVals["instance_name"] = inst.Name
		}
		if inst.UUID != "" {
			stringVals["instance_uuid"] = inst.UUID
		}

		if oc.Node != "" {
			stringVals["node_name"] = oc.Node
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
			if isWorkloadObject(c.Prop.Query) {
				// Workload instances are created exclusively by PollInstance. Skipping here mirrors RestPerf behavior and prevents
				// exporting new volumes with empty svm/volume labels before PollInstance runs.
				c.Logger.Debug("skip unknown workload instance in PollData, defer to PollInstance",
					slog.String("key", instanceKey))
				continue
			}
			var newErr error
			matInst, newErr = curMat.NewInstance(instanceKey)
			if newErr != nil {
				c.Logger.Warn("failed to create instance",
					slog.String("key", instanceKey), slogx.Err(newErr))
				continue
			}
		}

		tsMetric.SetValueFloat64(matInst, ts)

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
				metricCount += c.populateArrayCounter(curMat, prevMat, matInst, cs, vals64)
				continue
			}
			if vals32, ok := ct.List32(); ok {
				u64 := make([]uint64, len(vals32))
				for i, v := range vals32 {
					u64[i] = uint64(v)
				}
				metricCount += c.populateArrayCounter(curMat, prevMat, matInst, cs, u64)
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
	prevMat *matrix.Matrix,
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
				metr, ok := curMat.GetMetrics()[k]
				if !ok {
					var err error
					if metr, err = collectors.GetMetric(curMat, prevMat, k, propMetric.Label); err != nil {
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
	isHisto := false
	if co := c.perfProp.counterInfo[cs.Name]; co != nil {
		isHisto = co.isHistogram
	}

	if isHisto {
		bucketKey := cs.Name + ".bucket"
		if curMat.GetMetric(bucketKey) == nil {
			bm, err := collectors.GetMetric(curMat, prevMat, bucketKey, propMetric.Label)
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
		metr, ok := curMat.GetMetrics()[k]
		if !ok {
			var err error
			if metr, err = collectors.GetMetric(curMat, prevMat, k, propMetric.Label); err != nil {
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
