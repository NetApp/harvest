package clustersoftware

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
)

const (
	clusterSoftware  = "cluster_software"
	updateMatrix     = "update"
	statusMatrix     = "status"
	validationMatrix = "validation"
)

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
	globalLabels := dataMap[c.Object].GetGlobalLabels()

	for _, instance := range dataMap[c.Object].GetInstances() {
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
	mat := matrix.New(c.Parent+"."+updateMatrix, clusterSoftware, clusterSoftware)
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "phase")
	instanceKeys.NewChildS("", "state")
	instanceKeys.NewChildS("", "node")
	instanceKeys.NewChildS("", "elapsed_duration")

	mat.SetExportOptions(exportOptions)

	if _, err := mat.NewMetricFloat64(updateMatrix); err != nil {
		c.SLogger.Error("Failed to create metric", slogx.Err(err), slog.String("metric", updateMatrix))
		return err
	}

	c.data[updateMatrix] = mat
	return nil
}

func (c *ClusterSoftware) createStatusMetrics() error {
	mat := matrix.New(c.Parent+"."+statusMatrix, clusterSoftware, clusterSoftware)
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "state")
	instanceKeys.NewChildS("", "node")
	instanceKeys.NewChildS("", "name")
	instanceKeys.NewChildS("", "startTime")
	instanceKeys.NewChildS("", "endTime")

	mat.SetExportOptions(exportOptions)

	if _, err := mat.NewMetricFloat64(statusMatrix); err != nil {
		c.SLogger.Error("Failed to create metric", slogx.Err(err), slog.String("metric", statusMatrix))
		return err
	}

	c.data[statusMatrix] = mat
	return nil
}

func (c *ClusterSoftware) createValidationMetrics() error {
	mat := matrix.New(c.Parent+"."+validationMatrix, clusterSoftware, clusterSoftware)
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "status")
	instanceKeys.NewChildS("", "update_check")

	mat.SetExportOptions(exportOptions)

	if _, err := mat.NewMetricFloat64(validationMatrix); err != nil {
		c.SLogger.Error("Failed to create metric", slogx.Err(err), slog.String("metric", validationMatrix))
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
		phase := updateDetail.Get("phase").ClonedString()
		state := updateDetail.Get("state").ClonedString()
		elapsedDuration := updateDetail.Get("elapsed_duration").ClonedString()
		nodeName := updateDetail.Get("node.name").ClonedString()

		// If nodeName is empty then skip further processing
		if nodeName == "" {
			continue
		}

		key = phase + state + nodeName

		if clusterUpdateInstance, err = c.data[updateMatrix].NewInstance(key); err != nil {
			c.SLogger.Error("Failed to create instance", slogx.Err(err), slog.String("key", key))
			continue
		}
		clusterUpdateInstance.SetLabel("node", nodeName)
		clusterUpdateInstance.SetLabel("state", state)
		clusterUpdateInstance.SetLabel("phase", phase)
		clusterUpdateInstance.SetLabel("elapsed_duration", elapsedDuration)

		// populate numeric data
		value := 0.0
		if state == "completed" {
			value = 1.0
		}

		met := c.data[updateMatrix].GetMetric(updateMatrix)
		met.SetValueFloat64(clusterUpdateInstance, value)
		c.SLogger.Debug("added value", slog.Float64("value", value))
	}
}

func (c *ClusterSoftware) handleStatusDetails(statusDetailsJSON gjson.Result, globalLabels map[string]string) {
	var (
		clusterStatusInstance *matrix.Instance
		key                   string
		err                   error
	)
	// Purge and reset data
	c.data[statusMatrix].PurgeInstances()
	c.data[statusMatrix].Reset()

	// Set all global labels
	c.data[statusMatrix].SetGlobalLabels(globalLabels)

	for _, statusDetail := range statusDetailsJSON.Array() {
		name := statusDetail.Get("name").ClonedString()
		state := statusDetail.Get("state").ClonedString()
		nodeName := statusDetail.Get("node.name").ClonedString()
		startTime := statusDetail.Get("start_time").ClonedString()
		endTime := statusDetail.Get("end_time").ClonedString()
		key = name + state + nodeName + startTime

		if clusterStatusInstance, err = c.data[statusMatrix].NewInstance(key); err != nil {
			c.SLogger.Error("Failed to create instance", slogx.Err(err), slog.String("key", key))
			continue
		}
		clusterStatusInstance.SetLabel("node", nodeName)
		clusterStatusInstance.SetLabel("state", state)
		clusterStatusInstance.SetLabel("name", name)
		clusterStatusInstance.SetLabel("startTime", startTime)
		clusterStatusInstance.SetLabel("endTime", endTime)

		// populate numeric data
		value := 0.0
		if state == "completed" {
			value = 1.0
		}

		met := c.data[statusMatrix].GetMetric(statusMatrix)
		met.SetValueFloat64(clusterStatusInstance, value)
		c.SLogger.Debug("added value", slog.Float64("value", value))
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

	for _, validationDetail := range validationDetailsJSON.Array() {
		updateCheck := validationDetail.Get("update_check").ClonedString()
		status := validationDetail.Get("status").ClonedString()
		key = updateCheck + status

		if clusterValidationInstance, err = c.data[validationMatrix].NewInstance(key); err != nil {
			c.SLogger.Error("Failed to create instance", slogx.Err(err), slog.String("key", key))
			continue
		}
		clusterValidationInstance.SetLabel("update_check", updateCheck)
		clusterValidationInstance.SetLabel("status", status)

		// ignore all the validation result which are not in warning status
		if status != "warning" {
			continue
		}

		// populate numeric data
		value := 1.0
		met := c.data[validationMatrix].GetMetric(validationMatrix)
		met.SetValueFloat64(clusterValidationInstance, value)
	}
}
