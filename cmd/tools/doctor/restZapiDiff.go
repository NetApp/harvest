package doctor

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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

	fmt.Println("################## Missing Metrics by Object in prometheus ##############")
	metricDiffMap := metricDiff(zapiDataCenterName, restDataCenterName)
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

	metricDiff := difference(zapiKeys, restKeys)
	// group diff by template type
	x := make(map[string][]string)
	for _, label := range metricDiff {
		k := strings.Split(label, "_")[0]
		x[k] = append(x[k], label)
	}
	return x
}

func labelDiff(zapiDataCenterName string, restDataCenterName string) map[string][]string {
	queryZapi := "match[]={datacenter=\"" + zapiDataCenterName + "\"}"
	queryRest := "match[]={datacenter=\"" + restDataCenterName + "\"}"

	zapiMap := getLabelNames(queryZapi)

	restMap := getLabelNames(queryRest)

	diffMap := make(map[string][]string)
	for zk, zv := range zapiMap {
		if rv, ok := restMap[zk]; ok {
			diff := difference(zv, rv)
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
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
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
