package doctor

import (
	"encoding/json"
	"errors"
	"fmt"
	tw "github.com/netapp/harvest/v2/third_party/olekukonko/tablewriter"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Response struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

type Data struct {
	ResultType string   `json:"resultType"`
	Result     []Result `json:"result"`
}

type Result struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}

var ignoreMissingMetrics = map[string]struct{}{
	"aggr_hybrid_cache_size_total":     {},
	"aggr_snapshot_inode_used_percent": {},
	"aggr_space_reserved":              {},
	"security_audit_destination_port":  {},
	"wafl_reads_from_pmem":             {},
	"flexcache_":                       {},
	"rw_ctx_":                          {},
}

func fetchMetrics(datacenter string, prometheusURL string) ([]Result, error) {
	if prometheusURL == "" {
		panic("Error: prometheusUrl is not set")
	}
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/query?query={datacenter=%q}", prometheusURL, datacenter))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Data.Result, nil
}

func compareMetrics(restMetrics, zapiMetrics []Result) ([]string, error) {
	restMetricMap := make(map[string]int)
	zapiMetricMap := make(map[string]int)
	var missingMetrics []string

	for _, metric := range restMetrics {
		restMetricMap[metric.Metric["__name__"]]++
	}

	for _, metric := range zapiMetrics {
		zapiMetricMap[metric.Metric["__name__"]]++
	}

	table := tw.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Metric Name", "Rest Instances", "Zapi Instances", "Exists in Dashboard"})

	metricNames := make([]string, 0, len(zapiMetricMap))
	for metricName := range zapiMetricMap {
		metricNames = append(metricNames, metricName)
	}
	sort.Strings(metricNames)

	for _, metricName := range metricNames {
		zapiCount := zapiMetricMap[metricName]
		_, ok := restMetricMap[metricName]
		if !ok {
			existsInDashboard := checkDashboard("grafana/dashboards/cmode", metricName)
			table.Append([]string{metricName, "0", strconv.Itoa(zapiCount), existsInDashboard})
			fmt.Println(metricName)
			if !startsWithAny(metricName, ignoreMissingMetrics) {
				missingMetrics = append(missingMetrics, metricName)
			}
		}
	}

	fmt.Println("Missing Metrics in Rest:")
	table.Render()

	if len(missingMetrics) > 0 {
		return missingMetrics, errors.New("missing metrics detected")
	}
	return nil, nil
}

func checkDashboard(dir string, metricName string) string {
	var existsInDashboard string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(string(data), metricName) {
				existsInDashboard = "Yes"
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error walking the path:", err)
		return "Error"
	}
	if existsInDashboard == "" {
		existsInDashboard = "No"
	}
	return existsInDashboard
}

func DoCompareZapiRest(zapiDatacenter string, restDatacenter string, prometheusURL string) ([]string, error) {
	restMetrics, err := fetchMetrics(restDatacenter, prometheusURL)
	if err != nil {
		fmt.Println("Error fetching metrics for Rest:", err)
		return nil, nil
	}

	zapiMetrics, err := fetchMetrics(zapiDatacenter, prometheusURL)
	if err != nil {
		fmt.Println("Error fetching metrics for Zapi:", err)
		return nil, nil
	}

	return compareMetrics(restMetrics, zapiMetrics)
}

func startsWithAny(s string, prefixes map[string]struct{}) bool {
	for prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
