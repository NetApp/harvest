package metroclustercheck

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"strings"
)

type MetroclusterCheck struct {
	*plugin.AbstractPlugin
	data *matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &MetroclusterCheck{AbstractPlugin: p}
}

func (m *MetroclusterCheck) Init() error {

	var err error
	pluginMetrics := []string{"cluster_status", "node_status", "aggr_status", "volume_status"}
	pluginLabels := []string{"result", "name", "node", "aggregate", "volume", "object"}

	if err = m.InitAbc(); err != nil {
		return err
	}

	m.data = matrix.New(m.Parent+".Metrocluster", "metrocluster_check", "metrocluster_check")
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	for _, label := range pluginLabels {
		instanceKeys.NewChildS("", label)
	}
	m.data.SetExportOptions(exportOptions)

	for _, metric := range pluginMetrics {
		if err = m.createMetric(metric); err != nil {
			return err
		}
	}
	return nil
}

func (m *MetroclusterCheck) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	// Purge and reset data
	m.data.PurgeInstances()
	m.data.Reset()

	// Set all global labels
	data := dataMap[m.Object]
	m.data.SetGlobalLabels(data.GetGlobalLabels())

	for _, instance := range data.GetInstances() {
		instance.SetExportable(false)
		m.update(instance.GetLabel("cluster_detail"), "cluster")
		m.update(instance.GetLabel("node_detail"), "node")
		m.update(instance.GetLabel("aggregate_detail"), "aggregate")
		m.update(instance.GetLabel("volume_detail"), "volume")
	}

	return []*matrix.Matrix{m.data}, nil
}

func (m *MetroclusterCheck) update(objectDetail string, object string) {
	var (
		newDetailInstance *matrix.Instance
		key               string
		err               error
	)

	if objectDetail == "" {
		return
	}

	detailSlice := strings.Split(objectDetail, "},{")

	for _, detail := range detailSlice {
		if !strings.HasSuffix(detail, "}") {
			detail = detail + "}"
		}
		if !strings.HasPrefix(detail, "{") {
			detail = "{" + detail
		}
		detailJSON := gjson.Result{Type: gjson.JSON, Raw: detail}
		clusterName := detailJSON.Get("cluster.name").String()
		nodeName := detailJSON.Get("node.name")
		aggregateName := detailJSON.Get("aggregate.name")
		volumeName := detailJSON.Get("volume.name")
		for _, check := range detailJSON.Get("checks").Array() {
			name := check.Get("name").String()
			result := check.Get("result").String()
			switch object {
			case "volume":
				key = clusterName + nodeName.String() + aggregateName.String() + volumeName.String() + name
			case "aggregate":
				key = clusterName + nodeName.String() + aggregateName.String() + name
			case "node":
				key = clusterName + nodeName.String() + name
			case "cluster":
				key = clusterName + name
			}

			if newDetailInstance, err = m.data.NewInstance(key); err != nil {
				m.Logger.Error().Err(err).Str("arwInstanceKey", key).Msg("Failed to create arw instance")
				continue
			}
			newDetailInstance.SetLabel("name", name)
			newDetailInstance.SetLabel("result", result)
			newDetailInstance.SetLabel("object", object)
			newDetailInstance.SetLabel("volume", volumeName.String())
			newDetailInstance.SetLabel("aggregate", aggregateName.String())
			newDetailInstance.SetLabel("node", nodeName.String())

			switch object {
			case "volume":
				m.setValue("volume_status", newDetailInstance, result)
			case "aggregate":
				m.setValue("aggr_status", newDetailInstance, result)
			case "node":
				m.setValue("node_status", newDetailInstance, result)
			case "cluster":
				m.setValue("cluster_status", newDetailInstance, result)
			}
		}
	}
}

func (m *MetroclusterCheck) createMetric(metricName string) error {
	if _, err := m.data.NewMetricFloat64(metricName, metricName); err != nil {
		m.Logger.Error().Stack().Err(err).Msg("add metric")
		return err
	}
	return nil
}
func (m *MetroclusterCheck) setValue(metricName string, newDetailInstance *matrix.Instance, result string) {
	// populate numeric data
	value := 0.0
	if result == "ok" {
		value = 1.0
	}

	met := m.data.GetMetric(metricName)
	if err := met.SetValueFloat64(newDetailInstance, value); err != nil {
		m.Logger.Error().Stack().Err(err).Float64("value", value).Msg("Failed to parse value")
	} else {
		m.Logger.Debug().Float64("value", value).Msg("added value")
	}
}
