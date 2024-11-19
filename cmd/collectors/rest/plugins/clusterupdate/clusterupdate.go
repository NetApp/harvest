package clusterupdate

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"log/slog"
)

type ClusterUpdate struct {
	*plugin.AbstractPlugin
	data *matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &ClusterUpdate{AbstractPlugin: p}
}

func (c *ClusterUpdate) Init(conf.Remote) error {
	if err := c.InitAbc(); err != nil {
		return err
	}

	c.data = matrix.New(c.Parent+".ClusterUpdate", "cluster_update", "cluster_update")
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "phase")
	instanceKeys.NewChildS("", "state")
	instanceKeys.NewChildS("", "node")
	c.data.SetExportOptions(exportOptions)

	if _, err := c.data.NewMetricFloat64("status", "status"); err != nil {
		c.SLogger.Error("Failed to create metric", slogx.Err(err), slog.String("metric", "status"))
		return err
	}

	return nil
}

func (c *ClusterUpdate) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	var (
		clusterUpdateInstance *matrix.Instance
		key                   string
		err                   error
	)
	// Purge and reset data
	c.data.PurgeInstances()
	c.data.Reset()

	// Set all global labels
	data := dataMap[c.Object]
	c.data.SetGlobalLabels(data.GetGlobalLabels())

	for _, instance := range data.GetInstances() {
		instance.SetExportable(false)
		updateDetails := instance.GetLabel("update_details")
		updateDetailsJSON := gjson.Result{Type: gjson.JSON, Raw: "[" + updateDetails + "]"}
		for _, updateDetail := range updateDetailsJSON.Array() {
			phase := updateDetail.Get("phase").String()
			state := updateDetail.Get("state").String()
			nodeName := updateDetail.Get("node.name").String()
			key = phase + state + nodeName

			if clusterUpdateInstance, err = c.data.NewInstance(key); err != nil {
				c.SLogger.Error("Failed to create instance", slogx.Err(err), slog.String("key", key))
				continue
			}
			clusterUpdateInstance.SetLabel("node", nodeName)
			clusterUpdateInstance.SetLabel("state", state)
			clusterUpdateInstance.SetLabel("phase", phase)

			// populate numeric data
			value := 0.0
			if state == "completed" {
				value = 1.0
			}

			met := c.data.GetMetric("status")
			if err := met.SetValueFloat64(clusterUpdateInstance, value); err != nil {
				c.SLogger.Error("Failed to parse value", slogx.Err(err), slog.Float64("value", value))
			} else {
				c.SLogger.Debug("added value", slog.Float64("value", value))
			}
		}
	}

	return []*matrix.Matrix{c.data}, nil, nil
}
