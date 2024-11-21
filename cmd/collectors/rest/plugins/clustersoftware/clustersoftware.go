package clustersoftware

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

const updateMatrix = "cluster_software_update"
const StatusMatrix = "cluster_software_status"
const validationMatrix = "cluster_software_validation"
const labels = "labels"

type ClusterSoftware struct {
	*plugin.AbstractPlugin
	data map[string]*matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &ClusterSoftware{AbstractPlugin: p}
}

func (c *ClusterSoftware) Init(conf.Remote) error {
	if err := c.InitAbc(); err != nil {
		return err
	}

	c.data = make(map[string]*matrix.Matrix)
	if err := c.createUpdateMetrics(); err != nil {
		return err
	}
	if err := c.createStatusMetrics(); err != nil {
		return err
	}
	if err := c.createValidationMetrics(); err != nil {
		return err
	}

	return nil
}

func (c *ClusterSoftware) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[c.Object]
	globalLabels := data.GetGlobalLabels()

	for _, instance := range data.GetInstances() {
		instance.SetExportable(false)
		// generate update details metrics
		updateDetails := instance.GetLabel("update_details")
		updateDetailsJSON := gjson.Result{Type: gjson.JSON, Raw: "[" + updateDetails + "]"}
		c.handleUpdateDetails(updateDetailsJSON, globalLabels)

		// generate status details metrics
		statusDetails := instance.GetLabel("status_details")
		statusDetailsJSON := gjson.Result{Type: gjson.JSON, Raw: "[" + statusDetails + "]"}
		c.handleStatusDetails(statusDetailsJSON, globalLabels)

		// generate update details metrics
		validationResults := instance.GetLabel("validation_results")
		validationResultsJSON := gjson.Result{Type: gjson.JSON, Raw: "[" + validationResults + "]"}
		c.handleValidationDetails(validationResultsJSON, globalLabels)
	}

	softwareMetrics := make([]*matrix.Matrix, 0, len(c.data))
	for _, val := range c.data {
		softwareMetrics = append(softwareMetrics, val)
	}

	return softwareMetrics, nil, nil
}

func (c *ClusterSoftware) createUpdateMetrics() error {
	mat := matrix.New(c.Parent+".ClusterSoftware", updateMatrix, updateMatrix)
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "phase")
	instanceKeys.NewChildS("", "state")
	instanceKeys.NewChildS("", "node")

	mat.SetExportOptions(exportOptions)

	if _, err := mat.NewMetricFloat64(labels, labels); err != nil {
		c.SLogger.Error("Failed to create metric", slogx.Err(err), slog.String("metric", labels))
		return err
	}

	c.data[updateMatrix] = mat
	return nil
}

func (c *ClusterSoftware) createStatusMetrics() error {
	mat := matrix.New(c.Parent+".ClusterUpdate", StatusMatrix, StatusMatrix)
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "state")
	instanceKeys.NewChildS("", "node")
	instanceKeys.NewChildS("", "name")

	mat.SetExportOptions(exportOptions)

	if _, err := mat.NewMetricFloat64(labels, labels); err != nil {
		c.SLogger.Error("Failed to create metric", slogx.Err(err), slog.String("metric", labels))
		return err
	}

	c.data[StatusMatrix] = mat
	return nil
}

func (c *ClusterSoftware) createValidationMetrics() error {
	mat := matrix.New(c.Parent+".ClusterUpdate", validationMatrix, validationMatrix)
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "status")
	instanceKeys.NewChildS("", "update_check")

	mat.SetExportOptions(exportOptions)

	if _, err := mat.NewMetricFloat64(labels, labels); err != nil {
		c.SLogger.Error("Failed to create metric", slogx.Err(err), slog.String("metric", labels))
		return err
	}

	c.data[validationMatrix] = mat
	return nil
}

func (c *ClusterSoftware) handleUpdateDetails(updateDetailsJSON gjson.Result, globalLabels map[string]string) {
	var (
		clusterUpdateInstance *matrix.Instance
		key                   string
		err                   error
	)
	// Purge and reset data
	c.data[updateMatrix].PurgeInstances()
	c.data[updateMatrix].Reset()

	// Set all global labels
	c.data[updateMatrix].SetGlobalLabels(globalLabels)

	for _, updateDetail := range updateDetailsJSON.Array() {
		phase := updateDetail.Get("phase").String()
		state := updateDetail.Get("state").String()
		nodeName := updateDetail.Get("node.name").String()
		key = phase + state + nodeName

		if clusterUpdateInstance, err = c.data[updateMatrix].NewInstance(key); err != nil {
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

		met := c.data[updateMatrix].GetMetric(labels)
		if err := met.SetValueFloat64(clusterUpdateInstance, value); err != nil {
			c.SLogger.Error("Failed to parse value", slogx.Err(err), slog.Float64("value", value))
		} else {
			c.SLogger.Debug("added value", slog.Float64("value", value))
		}
	}
}

func (c *ClusterSoftware) handleStatusDetails(statusDetailsJSON gjson.Result, globalLabels map[string]string) {
	var (
		clusterStatusInstance *matrix.Instance
		key                   string
		err                   error
	)
	// Purge and reset data
	c.data[StatusMatrix].PurgeInstances()
	c.data[StatusMatrix].Reset()

	// Set all global labels
	c.data[StatusMatrix].SetGlobalLabels(globalLabels)

	for _, updateDetail := range statusDetailsJSON.Array() {
		name := updateDetail.Get("name").String()
		state := updateDetail.Get("state").String()
		nodeName := updateDetail.Get("node.name").String()
		key = name + state + nodeName

		if clusterStatusInstance, err = c.data[StatusMatrix].NewInstance(key); err != nil {
			c.SLogger.Error("Failed to create instance", slogx.Err(err), slog.String("key", key))
			continue
		}
		clusterStatusInstance.SetLabel("node", nodeName)
		clusterStatusInstance.SetLabel("state", state)
		clusterStatusInstance.SetLabel("name", name)

		// populate numeric data
		value := 0.0
		if state == "completed" {
			value = 1.0
		}

		met := c.data[StatusMatrix].GetMetric(labels)
		if err := met.SetValueFloat64(clusterStatusInstance, value); err != nil {
			c.SLogger.Error("Failed to parse value", slogx.Err(err), slog.Float64("value", value))
		} else {
			c.SLogger.Debug("added value", slog.Float64("value", value))
		}
	}
}

func (c *ClusterSoftware) handleValidationDetails(validationDetailsJSON gjson.Result, globalLabels map[string]string) {
	var (
		clusterValidationInstance *matrix.Instance
		key                       string
		err                       error
	)
	// Purge and reset data
	c.data[validationMatrix].PurgeInstances()
	c.data[validationMatrix].Reset()

	// Set all global labels
	c.data[validationMatrix].SetGlobalLabels(globalLabels)

	for _, updateDetail := range validationDetailsJSON.Array() {
		updateCheck := updateDetail.Get("update_check").String()
		status := updateDetail.Get("status").String()
		key = updateCheck + status

		if clusterValidationInstance, err = c.data[validationMatrix].NewInstance(key); err != nil {
			c.SLogger.Error("Failed to create instance", slogx.Err(err), slog.String("key", key))
			continue
		}
		clusterValidationInstance.SetLabel("update_check", updateCheck)
		clusterValidationInstance.SetLabel("status", status)

		// populate numeric data
		value := 0.0
		if status == "warning" {
			value = 1.0
		}

		met := c.data[validationMatrix].GetMetric(labels)
		if err := met.SetValueFloat64(clusterValidationInstance, value); err != nil {
			c.SLogger.Error("Failed to parse value", slogx.Err(err), slog.Float64("value", value))
		} else {
			c.SLogger.Debug("added value", slog.Float64("value", value))
		}
	}
}
