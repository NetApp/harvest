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

func metricValueDiff(metricName string) map[string][]string {
	timeNow := time.Now().Unix()
	queryUrl := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		PrometheusUrl, metricName, timeNow)
	data, _ := getResponse(queryUrl)
	replacer := strings.NewReplacer("[", "", "]", "", "\"", "")
	zapiMetric := make(map[string]float64)
	restMetric := make(map[string]float64)
	results := make([]gjson.Result, 0)
	if strings.HasPrefix(metricName, "disk_") {
		results = gjson.GetMany(data, "data.result.#.metric.disk", "data.result.#.value.1", "data.result.#.metric.datacenter")
	}
	if strings.HasPrefix(metricName, "aggr_") {
		results = gjson.GetMany(data, "data.result.#.metric.aggr", "data.result.#.value.1", "data.result.#.metric.datacenter")
	}
	if strings.HasPrefix(metricName, "lun_") {
		results = gjson.GetMany(data, "data.result.#.metric.lun", "data.result.#.value.1", "data.result.#.metric.datacenter")
	}
	if strings.HasPrefix(metricName, "node_") {
		results = gjson.GetMany(data, "data.result.#.metric.node", "data.result.#.value.1", "data.result.#.metric.datacenter")
	}
	// ignore qtree for now as it is under development
	//if strings.HasPrefix(metricName, "qtree_") && !strings.HasPrefix(metricName, "qtree_id") {
	//	results = gjson.GetMany(data, "data.result.#.metric.qtree", "data.result.#.value.1", "data.result.#.metric.datacenter")
	//}
	if strings.HasPrefix(metricName, "environment_sensor_") {
		results = gjson.GetMany(data, "data.result.#.metric.sensor", "data.result.#.value.1", "data.result.#.metric.datacenter")
	}
	if strings.HasPrefix(metricName, "shelf_") {
		results = gjson.GetMany(data, "data.result.#.metric.shelf", "data.result.#.value.1", "data.result.#.metric.datacenter")
	}
	if strings.HasPrefix(metricName, "snapmirror_") {
		results = gjson.GetMany(data, "data.result.#.metric.relationship_id", "data.result.#.value.1", "data.result.#.metric.datacenter")
	}
	if strings.HasPrefix(metricName, "snapshot_") {
		results = gjson.GetMany(data, "data.result.#.metric.snapshot_policy", "data.result.#.value.1", "data.result.#.metric.datacenter")
	}
	if strings.HasPrefix(metricName, "cluster_subsystem_") {
		results = gjson.GetMany(data, "data.result.#.metric.subsystem", "data.result.#.value.1", "data.result.#.metric.datacenter")
	}
	if strings.HasPrefix(metricName, "volume_") {
		results = gjson.GetMany(data, "data.result.#.metric.volume", "data.result.#.value.1", "data.result.#.metric.datacenter")
	}
	if len(results) > 0 {
		metric := strings.Split(replacer.Replace(results[0].String()), ",")
		value := strings.Split(replacer.Replace(results[1].String()), ",")
		dc := strings.Split(replacer.Replace(results[2].String()), ",")
		for i, v := range value {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				fmt.Println(err)
			}
			if dc[i] == "Zapi" {
				zapiMetric[metric[i]] = f
			}
			if dc[i] == "Rest" {
				restMetric[metric[i]] = f
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

	return nil
}

func labelDiff(zapiDataCenterName string, restDataCenterName string) map[string][]string {
	queryZapi := "match[]={datacenter=\"" + zapiDataCenterName + "\"}"
	queryRest := "match[]={datacenter=\"" + restDataCenterName + "\"}"

	zapiMap := getLabelNames(queryZapi)

	restMap := getLabelNames(queryRest)

	diffMap := make(map[string][]string)
	for zk, zv := range zapiMap {
		if rv, ok := restMap[zk]; ok {
			diff, _ := difference(zv, rv)
			diffMap[zk] = diff
		} else {
			diffMap[zk] = zv
		}
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
