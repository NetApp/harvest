package victoriametrics

import (
	"bytes"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/exporters"
	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/changelog"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort          = 8428
	defaultTimeout       = 5
	defaultAPIVersion    = "1"
	globalPrefix         = ""
	expectedResponseCode = 204
)

type VictoriaMetrics struct {
	*exporter.AbstractExporter
	client       *http.Client
	url          string
	addMetaTags  bool
	globalPrefix string
	replacer     *strings.Replacer
}

func New(abc *exporter.AbstractExporter) exporter.Exporter {
	return &VictoriaMetrics{AbstractExporter: abc}
}

func (v *VictoriaMetrics) Init() error {

	if err := v.InitAbc(); err != nil {
		return err
	}

	var (
		url, addr, version *string
		port               *int
	)

	if instance, err := v.Metadata.NewInstance("http"); err == nil {
		instance.SetLabel("task", "http")
	} else {
		return err
	}

	v.replacer = exporters.NewReplacer()

	if instance, err := v.Metadata.NewInstance("info"); err == nil {
		instance.SetLabel("task", "info")
	} else {
		return err
	}

	if x := v.Params.GlobalPrefix; x != nil {
		v.Logger.Debug("use global prefix", slog.String("prefix", *x))
		v.globalPrefix = *x
		if !strings.HasSuffix(v.globalPrefix, "_") {
			v.globalPrefix += "_"
		}
	} else {
		v.globalPrefix = globalPrefix
	}

	// Checking the required/optional params
	// customer should either provide url or addr
	// url is expected to be the full write URL api/v1/import/prometheus
	// when url is defined, addr and port are ignored

	// addr is expected to include host only (no port)
	// when addr is defined, port is required

	dbEndpoint := "addr"
	if url = v.Params.URL; url != nil {
		v.url = *url
		dbEndpoint = "url"
	} else {
		if addr = v.Params.Addr; addr == nil {
			v.Logger.Error("missing url or addr")
			return errs.New(errs.ErrMissingParam, "url or addr")
		}
		if port = v.Params.Port; port == nil {
			v.Logger.Debug("using default port", slog.Int("default", defaultPort))
			defPort := defaultPort
			port = &defPort
		}
		if version = v.Params.Version; version == nil {
			v := defaultAPIVersion
			version = &v
		}
		v.Logger.Debug("using api version", slog.String("version", *version))

		//goland:noinspection HttpUrlsUsage
		urlToUse := "http://" + *addr + ":" + strconv.Itoa(*port)
		url = &urlToUse
		v.url = fmt.Sprintf("%s/api/v%s/import/prometheus", *url, *version)
	}

	// timeout parameter
	timeout := time.Duration(defaultTimeout) * time.Second
	if ct := v.Params.ClientTimeout; ct != nil {
		if t, err := strconv.Atoi(*ct); err == nil {
			timeout = time.Duration(t) * time.Second
		} else {
			v.Logger.Warn(
				"invalid client_timeout, using default",
				slog.String("client_timeout", *ct),
				slog.Int("default", defaultTimeout),
			)
		}
	} else {
		v.Logger.Debug("using default client_timeout", slog.Int("default", defaultTimeout))
	}

	v.Logger.Debug("initializing exporter", slog.String("endpoint", dbEndpoint), slog.String("url", v.url))

	// construct HTTP client
	v.client = &http.Client{Timeout: timeout}

	return nil
}

func (v *VictoriaMetrics) Export(data *matrix.Matrix) (exporter.Stats, error) {

	var (
		metrics [][]byte
		err     error
		s       time.Time
		stats   exporter.Stats
	)

	v.Lock()
	defer v.Unlock()

	s = time.Now()

	// render metrics into open metrics format
	metrics, stats = v.Render(data)

	// fix render time
	if err = v.Metadata.LazyAddValueInt64("time", "render", time.Since(s).Microseconds()); err != nil {
		v.Logger.Error("metadata render time", slogx.Err(err))
	}
	// in test mode, don't emit metrics
	if v.Options.IsTest {
		return stats, nil
		// otherwise, to the actual export: send to the DB
	} else if err = v.Emit(metrics); err != nil {
		return stats, fmt.Errorf("unable to emit object: %s, uuid: %s, err=%w", data.Object, data.UUID, err)
	}

	v.Logger.Debug(
		"exported",
		slog.String("object", data.Object),
		slog.String("uuid", data.UUID),
		slog.Int("numMetric", len(metrics)),
	)

	// update metadata
	if err = v.Metadata.LazySetValueInt64("time", "export", time.Since(s).Microseconds()); err != nil {
		v.Logger.Error("metadata export time", slogx.Err(err))
	}

	// render metadata metrics into open metrics format
	metrics, stats = v.Render(v.Metadata)
	if err = v.Emit(metrics); err != nil {
		v.Logger.Error("emit metadata", slogx.Err(err))
	}

	return stats, nil
}

func (v *VictoriaMetrics) Emit(data [][]byte) error {
	var buffer *bytes.Buffer
	var request *http.Request
	var response *http.Response
	var err error

	buffer = bytes.NewBuffer(bytes.Join(data, []byte("\n")))

	if request, err = requests.New("POST", v.url, buffer); err != nil {
		return err
	}

	if response, err = v.client.Do(request); err != nil {
		return err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer response.Body.Close()
	if response.StatusCode != expectedResponseCode {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return errs.New(errs.ErrAPIResponse, err.Error())
		}
		return fmt.Errorf("%w: %s", errs.ErrAPIRequestRejected, string(body))
	}
	return nil
}

func (v *VictoriaMetrics) Render(data *matrix.Matrix) ([][]byte, exporter.Stats) {
	var (
		rendered          [][]byte
		tagged            *set.Set
		labelsToInclude   []string
		keysToInclude     []string
		prefix            string
		err               error
		joinedKeys        string
		histograms        map[string]*exporters.Histogram
		normalizedLabels  map[string][]string // cache of histogram normalized labels
		instancesExported uint64
		renderedBytes     uint64
		instanceKeysOk    bool
		buf               bytes.Buffer // shared buffer for rendering
	)

	buf.Grow(4096)
	globalLabels := make([]string, 0, len(data.GetGlobalLabels()))
	normalizedLabels = make(map[string][]string)

	if v.addMetaTags {
		tagged = set.New()
	}

	options := data.GetExportOptions()

	if x := options.GetChildS("instance_labels"); x != nil {
		labelsToInclude = x.GetAllChildContentS()
	}

	if x := options.GetChildS("instance_keys"); x != nil {
		keysToInclude = x.GetAllChildContentS()
	}

	includeAllLabels := false
	requireInstanceKeys := true

	if x := options.GetChildContentS("include_all_labels"); x != "" {
		if includeAllLabels, err = strconv.ParseBool(x); err != nil {
			v.Logger.Error("parameter: include_all_labels", slogx.Err(err))
		}
	}

	if x := options.GetChildContentS("require_instance_keys"); x != "" {
		if requireInstanceKeys, err = strconv.ParseBool(x); err != nil {
			v.Logger.Error("parameter: require_instance_keys", slogx.Err(err))
		}
	}

	if data.Object == "" {
		prefix = strings.TrimSuffix(v.globalPrefix, "_")
	} else {
		prefix = v.globalPrefix + data.Object
	}

	for key, value := range data.GetGlobalLabels() {
		globalLabels = append(globalLabels, exporters.Escape(v.replacer, key, value))
	}

	// Count the number of metrics so the rendered slice can be sized without reallocation
	numMetrics := 0

	exportableInstances := 0
	exportableMetrics := 0

	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		exportableInstances++
	}

	for _, metric := range data.GetMetrics() {
		if !metric.IsExportable() {
			continue
		}
		exportableMetrics++
	}

	numMetrics += exportableInstances * exportableMetrics
	if v.addMetaTags {
		numMetrics += exportableMetrics * 2 // for help and type
	}

	rendered = make([][]byte, 0, numMetrics)

	for _, instance := range data.GetInstances() {

		if !instance.IsExportable() {
			continue
		}
		instancesExported++

		moreKeys := 0
		if includeAllLabels {
			moreKeys = len(instance.GetLabels())
		}

		instanceKeys := make([]string, 0, len(globalLabels)+len(keysToInclude)+moreKeys)
		instanceKeys = append(instanceKeys, globalLabels...)

		instanceLabels := make([]string, 0, len(labelsToInclude))
		instanceLabelsSet := make(map[string]struct{})

		// The ChangeLog plugin tracks metric values and publishes the names of metrics that have changed.
		// For example, it might indicate that 'volume_size_total' has been updated.
		// If a global prefix for the exporter is defined, we need to amend the metric name with this prefix.
		if v.globalPrefix != "" && data.Object == changelog.ObjectChangeLog {
			if categoryValue, ok := instance.GetLabels()[changelog.Category]; ok {
				if categoryValue == changelog.Metric {
					if tracked, ok := instance.GetLabels()[changelog.Track]; ok {
						instance.GetLabels()[changelog.Track] = v.globalPrefix + tracked
					}
				}
			}
		}

		if includeAllLabels {
			for label, value := range instance.GetLabels() {
				// temporary fix for the rarely happening duplicate labels
				// known case is: ZapiPerf -> 7mode -> disk.yaml
				// actual cause is the Aggregator plugin, which is adding node as
				// instance label (even though it's already a global label for 7modes)
				_, ok := data.GetGlobalLabels()[label]
				if !ok {
					escaped := exporters.Escape(v.replacer, label, value)
					instanceKeys = append(instanceKeys, escaped)
				}
			}
		} else {
			for _, key := range keysToInclude {
				value := instance.GetLabel(key)
				escaped := exporters.Escape(v.replacer, key, value)
				instanceKeys = append(instanceKeys, escaped)
				if !instanceKeysOk && value != "" {
					instanceKeysOk = true
				}
			}

			for _, label := range labelsToInclude {
				value := instance.GetLabel(label)
				kv := exporters.Escape(v.replacer, label, value)
				_, ok := instanceLabelsSet[kv]
				if ok {
					continue
				}
				instanceLabelsSet[kv] = struct{}{}
				instanceLabels = append(instanceLabels, kv)
			}

			if !instanceKeysOk && requireInstanceKeys {
				continue
			}

			if len(instanceLabels) != 0 {
				allLabels := make([]string, 0, len(instanceLabels)+len(instanceKeys))
				allLabels = append(allLabels, instanceLabels...)
				// include each instanceKey not already included in the list of labels
				for _, instanceKey := range instanceKeys {
					_, ok := instanceLabelsSet[instanceKey]
					if ok {
						continue
					}
					instanceLabelsSet[instanceKey] = struct{}{}
					allLabels = append(allLabels, instanceKey)
				}
				if v.Params.SortLabels {
					sort.Strings(allLabels)
				}

				buf.Reset()

				buf.WriteString(prefix)
				buf.WriteString("_labels{")
				buf.WriteString(strings.Join(allLabels, ","))
				buf.WriteString("} 1.0")

				xbr := buf.Bytes()
				labelData := make([]byte, len(xbr))
				copy(labelData, xbr)

				prefixed := prefix + "_labels"
				if tagged != nil && !tagged.Has(prefixed) {
					tagged.Add(prefixed)
					help := "# HELP " + prefixed + " Pseudo-metric for " + data.Object + " labels"
					typeT := "# TYPE " + prefixed + " gauge"
					rendered = append(rendered, []byte(help), []byte(typeT))
					renderedBytes += uint64(len(help)) + uint64(len(typeT))
				}
				rendered = append(rendered, labelData)
				renderedBytes += uint64(len(labelData))
			}
		}

		if v.Params.SortLabels {
			sort.Strings(instanceKeys)
		}

		joinedKeys = strings.Join(instanceKeys, ",")
		histograms = make(map[string]*exporters.Histogram)

		for _, metric := range data.GetMetrics() {

			if !metric.IsExportable() {
				continue
			}

			if value, ok := metric.GetValueString(instance); ok {

				// metric is array, determine if this is a plain array or histogram
				if metric.HasLabels() {
					if metric.IsHistogram() {
						// Metric is histogram. Create a new metric to accumulate
						// the flattened metrics and export them in order
						bucketMetric := data.GetMetric(metric.GetLabel("bucket"))
						if bucketMetric == nil {
							v.Logger.Debug(
								"Unable to find bucket for metric, skip",
								slog.String("metric", metric.GetName()),
							)
							continue
						}
						metricIndex := metric.GetLabel("comment")
						index, err := strconv.Atoi(metricIndex)
						if err != nil {
							v.Logger.Error(
								"Unable to find index of metric, skip",
								slog.String("metric", metric.GetName()),
								slog.String("index", metricIndex),
							)
						}
						histogram := exporters.HistogramFromBucket(histograms, bucketMetric)
						histogram.Values[index] = value
						continue
					}
					metricLabels := make([]string, 0, len(metric.GetLabels()))
					for k, l := range metric.GetLabels() {
						metricLabels = append(metricLabels, exporters.Escape(v.replacer, k, l))
					}
					if v.Params.SortLabels {
						sort.Strings(metricLabels)
					}
					x := prefix + "_" + metric.GetName() + "{" + joinedKeys + "," + strings.Join(metricLabels, ",") + "} " + value

					prefixedName := prefix + "_" + metric.GetName()
					if tagged != nil && !tagged.Has(prefixedName) {
						tagged.Add(prefixedName)
						help := "# HELP " + prefixedName + " Metric for " + data.Object
						typeT := "# TYPE " + prefixedName + " gauge"
						rendered = append(rendered, []byte(help), []byte(typeT))
						renderedBytes += uint64(len(help)) + uint64(len(typeT))
					}

					rendered = append(rendered, []byte(x))
					renderedBytes += uint64(len(x))
					// scalar metric
				} else {
					buf.Reset()

					if prefix == "" {
						buf.WriteString(metric.GetName())
						buf.WriteString("{")
						buf.WriteString(joinedKeys)
						buf.WriteString("} ")
						buf.WriteString(value)
					} else {
						buf.WriteString(prefix)
						buf.WriteString("_")
						buf.WriteString(metric.GetName())
						buf.WriteString("{")
						buf.WriteString(joinedKeys)
						buf.WriteString("} ")
						buf.WriteString(value)
					}
					xbr := buf.Bytes()
					scalarMetric := make([]byte, len(xbr))
					copy(scalarMetric, xbr)

					prefixedName := prefix + "_" + metric.GetName()
					if tagged != nil && !tagged.Has(prefixedName) {
						tagged.Add(prefixedName)

						buf.Reset()
						buf.WriteString("# HELP ")
						buf.WriteString(prefixedName)
						buf.WriteString(" Metric for ")
						buf.WriteString(data.Object)

						xbr := buf.Bytes()
						helpB := make([]byte, len(xbr))
						copy(helpB, xbr)

						rendered = append(rendered, helpB)
						renderedBytes += uint64(len(helpB))

						buf.Reset()
						buf.WriteString("# TYPE ")
						buf.WriteString(prefixedName)
						buf.WriteString(" gauge")

						tbr := buf.Bytes()
						typeB := make([]byte, len(tbr))
						copy(typeB, tbr)

						rendered = append(rendered, typeB)
						renderedBytes += uint64(len(typeB))
					}

					rendered = append(rendered, scalarMetric)
					renderedBytes += uint64(len(scalarMetric))
				}
			}
		}

		// All metrics have been processed and flattened metrics accumulated. Determine which histograms can be
		// normalized and exported.
		for _, h := range histograms {
			metric := h.Metric
			bucketNames := metric.Buckets()
			objectMetric := data.Object + "_" + metric.GetName()
			_, ok := normalizedLabels[objectMetric]
			if !ok {
				canNormalize := true
				normalizedNames := make([]string, 0, len(*bucketNames))
				// check if the buckets can be normalized and collect normalized names
				for _, bucketName := range *bucketNames {
					normalized := exporters.NormalizeHistogram(bucketName)
					if normalized == "" {
						canNormalize = false
						break
					}
					normalizedNames = append(normalizedNames, normalized)
				}
				if canNormalize {
					normalizedLabels[objectMetric] = normalizedNames
				}
			}

			// Before writing out the histogram, check that every bucket value is non-empty.
			// Some bucket values may be empty if certain bucket metrics were skipped in the collector while others were not.
			allBucketsHaveValues := true
			if slices.Contains(h.Values, "") {
				allBucketsHaveValues = false
			}
			if !allBucketsHaveValues {
				// Skip rendering this histogram entirely.
				continue
			}

			prefixedName := prefix + "_" + metric.GetName()
			if tagged != nil && !tagged.Has(prefixedName) {
				tagged.Add(prefix + "_" + metric.GetName())

				help := "# HELP " + prefixedName + " Metric for " + data.Object
				typeT := "# TYPE " + prefixedName + " histogram"
				rendered = append(rendered, []byte(help), []byte(typeT))
				renderedBytes += uint64(len(help)) + uint64(len(typeT))
			}

			normalizedNames, canNormalize := normalizedLabels[objectMetric]
			var (
				countMetric string
				sumMetric   string
			)
			if canNormalize {
				count, sum := h.ComputeCountAndSum(normalizedNames)
				countMetric = prefix + "_" + metric.GetName() + "_count{" + joinedKeys + "} " + count
				sumMetric = prefix + "_" + metric.GetName() + "_sum{" + joinedKeys + "} " + strconv.Itoa(sum)
			}
			for i, value := range h.Values {
				bucketName := (*bucketNames)[i]
				var x string
				if canNormalize {
					x = prefix + "_" + metric.GetName() + "_bucket{" + joinedKeys + `,le="` + normalizedNames[i] + `"} ` + value
				} else {
					x = prefix + "_" + metric.GetName() + "{" + joinedKeys + `,` + exporters.Escape(v.replacer, "metric", bucketName) + "} " + value
				}
				rendered = append(rendered, []byte(x))
				renderedBytes += uint64(len(x))
			}
			if canNormalize {
				rendered = append(rendered, []byte(countMetric), []byte(sumMetric))
				renderedBytes += uint64(len(countMetric)) + uint64(len(sumMetric))
			}
		}
	}

	stats := exporter.Stats{
		InstancesExported: instancesExported,
		MetricsExported:   uint64(len(rendered)),
		RenderedBytes:     renderedBytes,
	}

	return rendered, stats
}
