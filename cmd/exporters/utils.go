package exporters

import (
	"bytes"
	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/changelog"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

var numAndUnitRe = regexp.MustCompile(`(\d+)\s*(\w+)`)

type Histogram struct {
	Metric *matrix.Metric
	Values []string
}

func Escape(replacer *strings.Replacer, key string, value string) string {
	// See https://prometheus.io/docs/instrumenting/exposition_formats/#comments-help-text-and-type-information
	// label_value can be any sequence of UTF-8 characters, but the backslash (\), double-quote ("),
	// and line feed (\n) characters have to be escaped as \\, \", and \n, respectively.

	return key + "=" + strconv.Quote(replacer.Replace(value))
}

func NewReplacer() *strings.Replacer {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", "\\n")
}

func HistogramFromBucket(histograms map[string]*Histogram, metric *matrix.Metric) *Histogram {
	h, ok := histograms[metric.GetName()]
	if ok {
		return h
	}
	buckets := metric.Buckets()
	var capacity int
	if buckets != nil {
		capacity = len(*buckets)
	}
	h = &Histogram{
		Metric: metric,
		Values: make([]string, capacity),
	}
	histograms[metric.GetName()] = h
	return h
}

func (h *Histogram) ComputeCountAndSum(normalizedNames []string) (string, int) {
	// If the buckets are normalizable, iterate through the values to:
	// 1) calculate Prometheus's cumulative buckets
	// 2) add _count metric
	// 3) calculate and add _sum metric
	cumValues := make([]string, len(h.Values))
	runningTotal := 0
	sum := 0
	for i, value := range h.Values {
		num, _ := strconv.Atoi(value)
		runningTotal += num
		cumValues[i] = strconv.Itoa(runningTotal)
		normalName := normalizedNames[i]
		leValue, _ := strconv.Atoi(normalName)
		sum += leValue * num
	}
	h.Values = cumValues
	return cumValues[len(cumValues)-1], sum
}

// NormalizeHistogram tries to normalize ONTAP values by converting units to multiples of the smallest unit.
// When the unit cannot be determined, return an empty string
func NormalizeHistogram(ontap string) string {
	numAndUnit := ontap
	if strings.HasPrefix(ontap, "<") {
		numAndUnit = ontap[1:]
	} else if strings.HasPrefix(ontap, ">") {
		return "+Inf"
	}
	submatch := numAndUnitRe.FindStringSubmatch(numAndUnit)
	if len(submatch) != 3 {
		return ""
	}
	num := submatch[1]
	unit := submatch[2]
	float, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return ""
	}
	var normal float64
	switch unit {
	case "us":
		return num
	case "ms", "msec":
		normal = 1_000 * float
	case "s", "sec":
		normal = 1_000_000 * float
	default:
		return ""
	}
	return strconv.FormatFloat(normal, 'f', -1, 64)
}

// Render metrics and labels into the exposition format, as described in
// https://prometheus.io/docs/instrumenting/exposition_formats/
//
// All metrics are implicitly "Gauge" counters. If requested, we also submit
// HELP and TYPE metadata (see add_meta_tags in config).
//
// Metric name is concatenation of the collector object (e.g. "volume",
// "fcp_lif") + the metric name (e.g. "read_ops" => "volume_read_ops").
// We do this since the same metrics for different objects can have
// different sets of labels, and Prometheus does not allow this.
//
// Example outputs:
//
// volume_read_ops{node="my-node",vol="some_vol"} 2523
// fcp_lif_read_ops{vserver="nas_svm",port_id="e02"} 771

func Render(data *matrix.Matrix, addMetaTags bool, sortLabels bool, globalPrefix string, logger *slog.Logger, timestamp string) ([][]byte, exporter.Stats) {
	var (
		rendered          [][]byte
		tagged            *set.Set
		labelsToInclude   []string
		keysToInclude     []string
		prefix            string
		err               error
		joinedKeys        string
		histograms        map[string]*Histogram
		normalizedLabels  map[string][]string // cache of histogram normalized labels
		instancesExported uint64
		renderedBytes     uint64
		instanceKeysOk    bool
		buf               bytes.Buffer // shared buffer for rendering
	)

	buf.Grow(4096)
	globalLabels := make([]string, 0, len(data.GetGlobalLabels()))
	normalizedLabels = make(map[string][]string)

	replacer := NewReplacer()

	if addMetaTags {
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
			logger.Error("parameter: include_all_labels", slogx.Err(err))
		}
	}

	if x := options.GetChildContentS("require_instance_keys"); x != "" {
		if requireInstanceKeys, err = strconv.ParseBool(x); err != nil {
			logger.Error("parameter: require_instance_keys", slogx.Err(err))
		}
	}

	if data.Object == "" {
		prefix = strings.TrimSuffix(globalPrefix, "_")
	} else {
		prefix = globalPrefix + data.Object
	}

	for key, value := range data.GetGlobalLabels() {
		globalLabels = append(globalLabels, Escape(replacer, key, value))
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
	if addMetaTags {
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
		if globalPrefix != "" && data.Object == changelog.ObjectChangeLog {
			if categoryValue, ok := instance.GetLabels()[changelog.Category]; ok {
				if categoryValue == changelog.Metric {
					if tracked, ok := instance.GetLabels()[changelog.Track]; ok {
						instance.GetLabels()[changelog.Track] = globalPrefix + tracked
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
					escaped := Escape(replacer, label, value)
					instanceKeys = append(instanceKeys, escaped)
				}
			}
		} else {
			for _, key := range keysToInclude {
				value := instance.GetLabel(key)
				escaped := Escape(replacer, key, value)
				instanceKeys = append(instanceKeys, escaped)
				if !instanceKeysOk && value != "" {
					instanceKeysOk = true
				}
			}

			for _, label := range labelsToInclude {
				value := instance.GetLabel(label)
				kv := Escape(replacer, label, value)
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
				if sortLabels {
					sort.Strings(allLabels)
				}

				buf.Reset()

				buf.WriteString(prefix)
				buf.WriteString("_labels{")
				buf.WriteString(strings.Join(allLabels, ","))
				buf.WriteString("} 1.0")
				if timestamp != "" {
					buf.WriteString(" ")
					buf.WriteString(timestamp)
				}

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

		if sortLabels {
			sort.Strings(instanceKeys)
		}

		joinedKeys = strings.Join(instanceKeys, ",")
		histograms = make(map[string]*Histogram)

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
							logger.Debug(
								"Unable to find bucket for metric, skip",
								slog.String("metric", metric.GetName()),
							)
							continue
						}
						metricIndex := metric.GetLabel("comment")
						index, err := strconv.Atoi(metricIndex)
						if err != nil {
							logger.Error(
								"Unable to find index of metric, skip",
								slog.String("metric", metric.GetName()),
								slog.String("index", metricIndex),
							)
						}
						histogram := HistogramFromBucket(histograms, bucketMetric)
						histogram.Values[index] = value
						continue
					}
					metricLabels := make([]string, 0, len(metric.GetLabels()))
					for k, l := range metric.GetLabels() {
						metricLabels = append(metricLabels, Escape(replacer, k, l))
					}
					if sortLabels {
						sort.Strings(metricLabels)
					}
					x := prefix + "_" + metric.GetName() + "{" + joinedKeys + "," + strings.Join(metricLabels, ",") + "} " + value
					if timestamp != "" {
						x += " " + timestamp
					}

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
					if timestamp != "" {
						buf.WriteString(" ")
						buf.WriteString(timestamp)
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
					normalized := NormalizeHistogram(bucketName)
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
					if timestamp != "" {
						x += " " + timestamp
					}
				} else {
					x = prefix + "_" + metric.GetName() + "{" + joinedKeys + `,` + Escape(replacer, "metric", bucketName) + "} " + value
					if timestamp != "" {
						x += " " + timestamp
					}
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
