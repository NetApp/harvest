package metroclustercheck

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
)

type MetroclusterCheck struct {
	*plugin.AbstractPlugin
	data *matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &MetroclusterCheck{AbstractPlugin: p}
}

func (m *MetroclusterCheck) Init(conf.Remote) error {

	pluginMetrics := []string{"cluster_status", "node_status", "aggr_status", "volume_status"}
	pluginLabels := []string{"result", "name", "node", "aggregate", "volume", "type"}

	if err := m.InitAbc(); err != nil {
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
		if err := m.createMetric(metric); err != nil {
			return err
		}
	}
	return nil
}

func (m *MetroclusterCheck) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	// Purge and reset data
	m.data.PurgeInstances()
	m.data.Reset()

	// Set all global labels
	data := dataMap[m.Object]
	localClusterName := data.GetGlobalLabels()["cluster"]
	m.data.SetGlobalLabels(data.GetGlobalLabels())

	for _, instance := range data.GetInstances() {
		instance.SetExportable(false)
		m.update(instance.GetLabel("cluster"), "cluster", localClusterName)
		m.update(instance.GetLabel("node"), "node", localClusterName)
		m.update(instance.GetLabel("aggregate"), "aggregate", localClusterName)
		m.update(instance.GetLabel("volume"), "volume", localClusterName)
	}

	return []*matrix.Matrix{m.data}, nil, nil
}

func (m *MetroclusterCheck) update(objectInfo string, object string, localClusterName string) {
	var (
		newDetailInstance *matrix.Instance
		key               string
		err               error
	)

	if objectInfo == "" {
		return
	}

	objectInfoJSON := gjson.Result{Type: gjson.JSON, Raw: objectInfo}
	for _, detail := range objectInfoJSON.Get("details").Array() {
		clusterName := detail.Get("cluster.name").ClonedString()
		nodeName := detail.Get("node.name")
		aggregateName := detail.Get("aggregate.name")
		volumeName := detail.Get("volume.name")
		for _, check := range detail.Get("checks").Array() {
			name := check.Get("name").ClonedString()
			result := check.Get("result").ClonedString()
			switch object {
			case "volume":
				key = clusterName + nodeName.ClonedString() + aggregateName.ClonedString() + volumeName.ClonedString() + name
			case "aggregate":
				key = clusterName + nodeName.ClonedString() + aggregateName.ClonedString() + name
			case "node":
				key = clusterName + nodeName.ClonedString() + name
			case "cluster":
				key = clusterName + name
			}

			if newDetailInstance, err = m.data.NewInstance(key); err != nil {
				m.SLogger.Error("Failed to create instance", slogx.Err(err), slog.String("key", key))
				continue
			}
			newDetailInstance.SetLabel("name", name)
			newDetailInstance.SetLabel("result", result)
			newDetailInstance.SetLabel("volume", volumeName.ClonedString())
			newDetailInstance.SetLabel("aggregate", aggregateName.ClonedString())
			newDetailInstance.SetLabel("node", nodeName.ClonedString())

			switch object {
			case "volume":
				m.setValue("volume_status", newDetailInstance, result)
			case "aggregate":
				m.setValue("aggr_status", newDetailInstance, result)
			case "node":
				m.setValue("node_status", newDetailInstance, result)
			case "cluster":
				if localClusterName == clusterName {
					newDetailInstance.SetLabel("type", "local")
				} else {
					newDetailInstance.SetLabel("type", "remote")
				}
				m.setValue("cluster_status", newDetailInstance, result)
			}
		}
	}
}

func (m *MetroclusterCheck) createMetric(metricName string) error {
	if _, err := m.data.NewMetricFloat64(metricName, metricName); err != nil {
		m.SLogger.Error("Failed to create metric", slogx.Err(err), slog.String("metric", metricName))
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
	met.SetValueFloat64(newDetailInstance, value)
	m.SLogger.Debug("added value", slog.Float64("value", value))
}
