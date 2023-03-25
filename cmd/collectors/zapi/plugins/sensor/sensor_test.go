package sensor

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"os"
	"strings"
	"testing"
)

var testxml = "../../../../../cmd/collectors/zapi/plugins/sensor/testdata/sensor.xml"
var mat *matrix.Matrix
var sensor = &Sensor{AbstractPlugin: plugin.New("sensor", nil, nil, nil, "sensor", nil)}

func init() {
	//setup matrix data
	var err error
	var fetch func(*matrix.Instance, *node.Node, []string)
	dat, _ := os.ReadFile(testxml)
	instanceLabelPaths := map[string]string{"environment-sensors-info.discrete-sensor-state": "discrete_state",
		"environment-sensors-info.sensor-type":            "type",
		"environment-sensors-info.threshold-sensor-state": "threshold_state",
		"environment-sensors-info.warning-high-threshold": "warning_high",
		"environment-sensors-info.discrete-sensor-value":  "discrete_value",
		"environment-sensors-info.node-name":              "node",
		"environment-sensors-info.sensor-name":            "sensor",
		"environment-sensors-info.value-units":            "unit",
		"environment-sensors-info.warning-low-threshold":  "warning_low",
	}

	fetch = func(instance *matrix.Instance, node *node.Node, path []string) {

		newpath := append(path, node.GetNameS())
		key := strings.Join(newpath, ".")
		if value := node.GetContentS(); value != "" {
			if label, has := instanceLabelPaths[key]; has {
				instance.SetLabel(label, value)
			} else if metric := mat.GetMetric(key); metric != nil {
				_ = metric.SetValueString(instance, value)
			}
		}

		for _, child := range node.GetChildren() {
			fetch(instance, child, newpath)
		}
	}

	mat = matrix.New("TestRemoveInstance", "sensor", "test")
	instanceKeyPath := [][]string{{"environment-sensors-info", "node-name"}, {"environment-sensors-info", "sensor-name"}}
	shortestPathPrefix := []string{"environment-sensors-info"}
	_, _ = mat.NewMetricInt64("environment-sensors-info.critical-high-threshold")
	_, _ = mat.NewMetricInt64("environment-sensors-info.critical-low-threshold")
	_, _ = mat.NewMetricInt64("environment-sensors-info.threshold-sensor-value")
	response, _ := tree.LoadXML(dat)
	instances := response.SearchChildren(shortestPathPrefix)
	for _, instanceElem := range instances {
		keys, found := instanceElem.SearchContent(shortestPathPrefix, instanceKeyPath)

		if !found {
			continue
		}

		key := strings.Join(keys, ".")
		instance := mat.GetInstance(key)

		if instance == nil {
			if instance, err = mat.NewInstance(key); err != nil {
				continue
			}
		}
		fetch(instance, instanceElem, make([]string, 0))

	}

	sensor.data = matrix.New("Sensor", "environment_sensor", "environment_sensor")
	sensor.instanceKeys = make(map[string]string)
	sensor.instanceLabels = make(map[string]*dict.Dict)
	sensor.AbstractPlugin.Logger = logging.Get()

	for _, k := range eMetrics {
		_ = matrix.CreateMetric(k, sensor.data)
	}
}

// Verified temperature sensor values by parsing, pivoting, etc. externally via dasel, jq, miller

// average_ambient_temperature is
// cat cmd/collectors/zapi/plugins/sensor/testdata/sensor.xml | dasel -r xml -w json | jq -r '.root."attributes-list"."environment-sensors-info"[] | select(."sensor-type" | test("thermal")) | {node: (."node-name"), name: (."sensor-name"), value: (."threshold-sensor-value")} | [.node, .name, .value] | @csv' | rg "Ambient Temp|Ambient Temp \d|PSU\d AmbTemp|PSU\d Inlet|PSU\d Inlet Temp|In Flow Temp|Front Temp|System Inlet|Bat Ambient \d|Riser Inlet Temp" | rg -v "Fake" | mlr --csv --implicit-csv-header label node,name,value then stats1 -a min,mean,max -f value -g node | mlr --csv --opprint --barred cat

// +------------+-----------+--------------------+-----------+
// | node       | value_min | value_mean         | value_max |
// +------------+-----------+--------------------+-----------+
// | cdot-k3-05 | 21        | 21.666666666666668 | 23        |
// | cdot-k3-06 | 21        | 22                 | 24        |
// | cdot-k3-07 | 21        | 21.666666666666668 | 23        |
// | cdot-k3-08 | 21        | 22.333333333333332 | 24        |
// +------------+-----------+--------------------+-----------+
//
// average_temperature [min, avg, max] is calculated like so
// cat cmd/collectors/zapi/plugins/sensor/testdata/sensor.xml | dasel -r xml -w json | jq -r '.root."attributes-list"."environment-sensors-info"[] | select(."sensor-type" | test("thermal")) | {node: (."node-name"), name: (."sensor-name"), value: (."threshold-sensor-value")} | [.node, .name, .value] | @csv' | rg -v "Ambient Temp|Ambient Temp \d|PSU\d AmbTemp|PSU\d Inlet|PSU\d Inlet Temp|In Flow Temp|Front Temp|System Inlet|Bat Ambient \d|Riser Inlet Temp" | rg -v "Fake" | mlr --csv --implicit-csv-header label node,name,value then stats1 -a min,mean,max -f value -g node | mlr --csv --opprint --barred cat

// +------------+-----------+------------+-----------+
// | node       | value_min | value_mean | value_max |
// +------------+-----------+------------+-----------+
// | cdot-k3-05 | 19        | 27.1875    | 36        |
// | cdot-k3-06 | 19        | 26.6875    | 35        |
// | cdot-k3-07 | 19        | 26.6875    | 35        |
// | cdot-k3-08 | 20        | 27.5       | 36        |
// +------------+-----------+------------+-----------+

func TestSensor_Run(t *testing.T) {

	dataMap := map[string]*matrix.Matrix{
		mat.Object: mat,
	}
	omat, _ := sensor.Run(dataMap)

	expected := map[string]map[string]float64{
		"average_ambient_temperature": {"cdot-k3-05": 21.666666666666668, "cdot-k3-06": 22, "cdot-k3-07": 21.666666666666668, "cdot-k3-08": 22.333333333333332},
		"average_fan_speed":           {"cdot-k3-05": 7030, "cdot-k3-06": 7050, "cdot-k3-07": 7040, "cdot-k3-08": 7050},
		"max_fan_speed":               {"cdot-k3-05": 7700, "cdot-k3-06": 7700, "cdot-k3-07": 7700, "cdot-k3-08": 7700},
		"min_fan_speed":               {"cdot-k3-05": 4600, "cdot-k3-06": 4500, "cdot-k3-07": 4600, "cdot-k3-08": 4500},
		"power":                       {"cdot-k3-05": 383.4, "cdot-k3-06": 347.9, "cdot-k3-07": 340.8, "cdot-k3-08": 362.1},
		"average_temperature":         {"cdot-k3-05": 27.1875, "cdot-k3-06": 26.6875, "cdot-k3-07": 26.6875, "cdot-k3-08": 27.5},
		"max_temperature":             {"cdot-k3-05": 36, "cdot-k3-06": 35, "cdot-k3-07": 35, "cdot-k3-08": 36},
		"min_ambient_temperature":     {"cdot-k3-05": 21, "cdot-k3-06": 21, "cdot-k3-07": 21, "cdot-k3-08": 21},
		"min_temperature":             {"cdot-k3-05": 19, "cdot-k3-06": 19, "cdot-k3-07": 19, "cdot-k3-08": 20},
	}

	for _, k := range eMetrics {
		metrics := omat[0].GetMetrics()
		for iKey, v := range omat[0].GetInstances() {
			got, _ := metrics[k].GetValueFloat64(v)
			exp := expected[k][iKey]
			if got != exp {
				t.Errorf("instance %s metrics %s expected: = %v, got: %v", iKey, k, exp, got)
			}
		}
	}
}
