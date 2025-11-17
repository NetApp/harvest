package exporters

import (
	"github.com/netapp/harvest/v2/pkg/matrix"
	"regexp"
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
