package doctor

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const PrometheusUrl string = "http://localhost:9090"

func DoDiffRestZapi(zapiDataCenterName string, restDataCenterName string) {
	labelDiffMap := labelDiff(zapiDataCenterName, restDataCenterName)
	fmt.Println("################## Missing Labels ##############")
	for k, v := range labelDiffMap {
		fmt.Println(k, v)
	}

	metricDiffMap := metricDiff(zapiDataCenterName, restDataCenterName)
	fmt.Println("################## Missing Metrics by Object in prometheus ##############")
	for k, v := range metricDiffMap {
		fmt.Println(k, v)
	}

	fmt.Println("################## Missing Metrics from dashboard ##############")
	dashboardDiffMap := make(map[string][]string)
	for key, value := range metricDiffMap {
		for _, v := range value {
			filepath.Walk("grafana/dashboards/cmode", getWalkFunc(v, dashboardDiffMap, key))
		}
	}

	for k, v := range dashboardDiffMap {
		fmt.Println(k, v)
	}
	fmt.Println("##################")
}

func getWalkFunc(searchKey string, dashboardDiffMap map[string][]string, key string) filepath.WalkFunc {
	return func(path string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}

			isExist, err := regexp.Match(searchKey, b)
			if err != nil {
				panic(err)
			}
			if isExist {
				dashboardDiffMap[key] = appendIfMissing(dashboardDiffMap[key], searchKey)
			}
		}
		return nil
	}
}

func metricDiff(zapiDataCenterName string, restDataCenterName string) map[string][]string {
	queryZapi := "match[]={datacenter=\"" + zapiDataCenterName + "\"}"
	queryRest := "match[]={datacenter=\"" + restDataCenterName + "\"}"

	zapiMap := getMetricNames(queryZapi)
	zapiKeys := make([]string, 0, len(zapiMap))

	for k := range zapiMap {
		zapiKeys = append(zapiKeys, k)
	}
	restMap := getMetricNames(queryRest)
	restKeys := make([]string, 0, len(restMap))

	for k := range restMap {
		restKeys = append(restKeys, k)
	}

	metricDiff, common := difference(zapiKeys, restKeys)
	// group diff by template type
	x := make(map[string][]string)
	for _, label := range metricDiff {
		k := strings.Split(label, "_")[0]
		x[k] = append(x[k], label)
	}

	fmt.Println("################## Metrics Diffs in prometheus ##############")
	for _, c := range common {
		metricValueDiff(c)
	}
	return x
}

func metricValueDiff(metricName string) {
	if strings.HasSuffix(metricName, "_labels") {
		return
	}

	timeNow := time.Now().Unix()
	queryUrl := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		PrometheusUrl, metricName, timeNow)
	data, _ := getResponse(queryUrl)
	replacer := strings.NewReplacer("[", "", "]", "", "\"", "")
	zapiMetric := make(map[string]float64)
	restMetric := make(map[string]float64)
	results := make([]gjson.Result, 0)
	keyIndexes := make([]int, 0)

	if strings.HasPrefix(metricName, "disk_") {
		results = gjson.GetMany(data, "data.result.#.value.1", "data.result.#.metric.datacenter", "data.result.#.metric.disk")
		keyIndexes = []int{2}
	}
	if strings.HasPrefix(metricName, "aggr_") {
		results = gjson.GetMany(data, "data.result.#.value.1", "data.result.#.metric.datacenter", "data.result.#.metric.aggr", "data.result.#.metric.node")
		keyIndexes = []int{2, 3}
	}
	if strings.HasPrefix(metricName, "lun_") {
		results = gjson.GetMany(data, "data.result.#.value.1", "data.result.#.metric.datacenter", "data.result.#.metric.lun", "data.result.#.metric.svm", "data.result.#.metric.node")
		keyIndexes = []int{2, 3, 4}
	}
	if strings.HasPrefix(metricName, "node_") {
		results = gjson.GetMany(data, "data.result.#.value.1", "data.result.#.metric.datacenter", "data.result.#.metric.node")
		keyIndexes = []int{2}
	}
	if strings.HasPrefix(metricName, "qtree_") && !strings.HasPrefix(metricName, "qtree_id") {
		results = gjson.GetMany(data, "data.result.#.value.1", "data.result.#.metric.datacenter", "data.result.#.metric.qtree")
		keyIndexes = []int{2}
	}

	if strings.HasPrefix(metricName, "shelf_") {
		results = gjson.GetMany(data, "data.result.#.value.1", "data.result.#.metric.datacenter", "data.result.#.metric.shelf")
		keyIndexes = []int{2}
	}
	if strings.HasPrefix(metricName, "snapmirror_") {
		results = gjson.GetMany(data, "data.result.#.value.1", "data.result.#.metric.datacenter", "data.result.#.metric.relationship_id")
		keyIndexes = []int{2}
	}
	if strings.HasPrefix(metricName, "snapshot_") {
		results = gjson.GetMany(data, "data.result.#.value.1", "data.result.#.metric.datacenter", "data.result.#.metric.snapshot_policy", "data.result.#.metric.svm")
		keyIndexes = []int{2, 3}
	}
	if strings.HasPrefix(metricName, "cluster_subsystem_") {
		results = gjson.GetMany(data, "data.result.#.value.1", "data.result.#.metric.datacenter", "data.result.#.metric.subsystem")
		keyIndexes = []int{2}
	}
	if strings.HasPrefix(metricName, "volume_") {
		results = gjson.GetMany(data, "data.result.#.value.1", "data.result.#.metric.datacenter", "data.result.#.metric.volume", "data.result.#.metric.svm")
		keyIndexes = []int{2, 3}
	}

	if len(results) > 0 && len(keyIndexes) > 0 {
		metrics := make([][]string, 0)
		for _, i := range keyIndexes {
			metric := strings.Split(replacer.Replace(results[i].String()), ",")
			metrics = append(metrics, metric)
		}
		value := strings.Split(replacer.Replace(results[0].String()), ",")
		dc := strings.Split(replacer.Replace(results[1].String()), ",")
		for i, v := range value {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				fmt.Println(err)
			}
			key := ""
			for x := range metrics {
				key = key + "_" + metrics[x][i]
			}
			if dc[i] == "Zapi" {
				zapiMetric[key] = f
			}
			if dc[i] == "Rest" {
				restMetric[key] = f
			}
		}
		for k, v := range zapiMetric {
			if v1, ok := restMetric[k]; ok {
				if math.Abs(v-v1) > 0 {
					fmt.Printf("%s %s %v -> %v\n", metricName, k, v, v1)
				}
			}
		}
	}
}

func IndexOf(data []string, search string) int {
	for i, v := range data {
		if v == search {
			return i
		}
	}
	return -1
}

func labelValueDiff(label string, labelNames []string) {
	timeNow := time.Now().Unix()
	queryUrl := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		PrometheusUrl, label, timeNow)
	data, _ := getResponse(queryUrl)
	replacer := strings.NewReplacer("[", "", "]", "", "\"", "")
	zapiMetric := make(map[string]string)
	restMetric := make(map[string]string)
	results := make([]gjson.Result, 0)
	prefixLabelsName := make([]string, 0)
	// remove data from slice
	removeLabels := []string{"__name__", "instance", "job"}
	finalLabelNames, _ := difference(labelNames, removeLabels)
	for _, l := range finalLabelNames {
		l1 := "data.result.#.metric." + l
		prefixLabelsName = append(prefixLabelsName, l1)
	}
	keyIndexes := make([]int, 0)
	dataCenterIndex := -1
	skipMatch := make([]string, 0)
	skipMatch = append(skipMatch, "datacenter")

	// To ignore child labels(ex: shelf_psu_labels) from shelf object, following exact comparison
	if strings.HasPrefix(label, "disk_") || strings.Compare(label, "shelf_labels") == 0 {
		keyIndexes = append(keyIndexes, IndexOf(finalLabelNames, "serial_number"))
		dataCenterIndex = IndexOf(finalLabelNames, "datacenter")
		results = gjson.GetMany(data, prefixLabelsName...)
	}

	if strings.HasPrefix(label, "snapmirror_") {
		keyIndexes = append(keyIndexes, IndexOf(finalLabelNames, "relationship_id"))
		dataCenterIndex = IndexOf(finalLabelNames, "datacenter")
		results = gjson.GetMany(data, prefixLabelsName...)
	}

	if strings.HasPrefix(label, "volume_") {
		keyIndexes = append(keyIndexes, IndexOf(finalLabelNames, "volume"))
		keyIndexes = append(keyIndexes, IndexOf(finalLabelNames, "svm"))
		dataCenterIndex = IndexOf(finalLabelNames, "datacenter")
		results = gjson.GetMany(data, prefixLabelsName...)
	}

	if strings.HasPrefix(label, "aggr_") {
		keyIndexes = append(keyIndexes, IndexOf(finalLabelNames, "aggr"))
		keyIndexes = append(keyIndexes, IndexOf(finalLabelNames, "node"))
		dataCenterIndex = IndexOf(finalLabelNames, "datacenter")
		results = gjson.GetMany(data, prefixLabelsName...)
	}

	if strings.HasPrefix(label, "lun_") {
		keyIndexes = append(keyIndexes, IndexOf(finalLabelNames, "lun"))
		keyIndexes = append(keyIndexes, IndexOf(finalLabelNames, "node"))
		keyIndexes = append(keyIndexes, IndexOf(finalLabelNames, "svm"))
		dataCenterIndex = IndexOf(finalLabelNames, "datacenter")
		results = gjson.GetMany(data, prefixLabelsName...)
	}

	if strings.HasPrefix(label, "node_") {
		keyIndexes = append(keyIndexes, IndexOf(finalLabelNames, "node"))
		dataCenterIndex = IndexOf(finalLabelNames, "datacenter")
		results = gjson.GetMany(data, prefixLabelsName...)
	}

	if strings.HasPrefix(label, "qtree_") {
		keyIndexes = append(keyIndexes, IndexOf(finalLabelNames, "export_policy"))
		keyIndexes = append(keyIndexes, IndexOf(finalLabelNames, "svm"))
		dataCenterIndex = IndexOf(finalLabelNames, "datacenter")
		results = gjson.GetMany(data, prefixLabelsName...)
	}

	if len(results) > 0 && len(keyIndexes) > 0 && dataCenterIndex != -1 {
		metrics := make([][]string, 0)
		for _, i := range keyIndexes {
			metric := strings.Split(replacer.Replace(results[i].String()), ",")
			metrics = append(metrics, metric)
		}
		dc := strings.Split(replacer.Replace(results[dataCenterIndex].String()), ",")
		for i, f := range finalLabelNames {
			if IndexOf(skipMatch, f) == -1 {
				value := strings.Split(replacer.Replace(results[i].String()), ",")
				if len(dc) != len(value) {
					fmt.Printf("******* Mismatch in label length. Check data %s %s\n", label, f)
					continue
				}
				for i, v := range value {
					key := ""
					for x := range metrics {
						key = key + "_" + metrics[x][i]
					}
					if dc[i] == "Zapi" {
						zapiMetric[key] = v
					}
					if dc[i] == "Rest" {
						restMetric[key] = v
					}
				}
				for k, v := range zapiMetric {
					if v1, ok := restMetric[k]; ok {
						if v != v1 {
							fmt.Printf("%s %s %s %v -> %v\n", label, f, k, v, v1)
						}
					}
				}
			}
		}
	}
}

func labelDiff(zapiDataCenterName string, restDataCenterName string) map[string][]string {
	queryZapi := "match[]={datacenter=\"" + zapiDataCenterName + "\"}"
	queryRest := "match[]={datacenter=\"" + restDataCenterName + "\"}"

	zapiMap := getLabelNames(queryZapi)

	restMap := getLabelNames(queryRest)

	diffMap := make(map[string][]string)
	commonMap := make(map[string][]string)
	for zk, zv := range zapiMap {
		if rv, ok := restMap[zk]; ok {
			diff, common := difference(zv, rv)
			diffMap[zk] = diff
			commonMap[zk] = common
		} else {
			diffMap[zk] = zv
		}
	}
	fmt.Println("################## Label value Diffs in prometheus ##############")
	for k, v := range commonMap {
		labelValueDiff(k, v)
	}

	return diffMap
}

func getMetricNames(query string) map[string]string {
	var dataMap = make(map[string]string)
	timeNow := time.Now().Unix()
	queryUrl := fmt.Sprintf("%s/api/v1/series?%s&time=%d",
		PrometheusUrl, query, timeNow)
	data, _ := getResponse(queryUrl)
	result := gjson.Get(data, "data.#.__name__")

	for _, name := range result.Array() {
		dataMap[name.String()] = name.String()
	}
	return dataMap
}

func getLabelNames(query string) map[string][]string {
	var dataMap = make(map[string][]string)
	timeNow := time.Now().Unix()
	queryUrl := fmt.Sprintf("%s/api/v1/series?%s&time=%d",
		PrometheusUrl, query, timeNow)
	data, _ := getResponse(queryUrl)

	result := gjson.Get(data, "data")
	result.ForEach(func(key, value gjson.Result) bool {
		labelName := value.Get("__name__")
		if strings.Contains(labelName.String(), "_labels") {
			v := value.Get("@keys")
			v.ForEach(func(key, value gjson.Result) bool {
				dataMap[labelName.String()] = appendIfMissing(dataMap[labelName.String()], value.String())
				return true // keep iterating
			})

		}
		return true // keep iterating
	})
	return dataMap
}

func appendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) ([]string, []string) {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	var common []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		} else {
			common = append(common, x)
		}
	}
	return diff, common
}

func getResponse(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	return string(body), nil
}
