package changelog

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree/yaml"
)

/*The changelog feature is only applicable to labels and requires a UUID for the label name to exist.
A default configuration for volume, SVM, and node is available, but the DSL can be overwritten as needed.
The shape of the change_log is specific to each label change and is only applicable to matrix collected by the collector.
*/

// Constants for ChangeLog metrics and labels
const (
	changeLog     = "change"
	objectLabel   = "object"
	opLabel       = "op"
	create        = "create"
	update        = "update"
	del           = "delete"
	track         = "track"
	oldLabelValue = "old_value"
	newLabelValue = "new_value"
)

// Metrics to be used in ChangeLog
var metrics = []string{
	"log",
}

// ChangeLog represents the main structure of the ChangeLog plugin
type ChangeLog struct {
	*plugin.AbstractPlugin
	matrixName      string
	previousData    *matrix.Matrix
	changeLogMap    map[string]*matrix.Matrix
	changeLogConfig Entry
}

// Change represents a single change entry in the ChangeLog
type Change struct {
	key      string
	object   string
	op       string
	labels   map[string]string
	track    string
	oldValue string
	newValue string
}

// New initializes a new instance of the ChangeLog plugin
func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &ChangeLog{AbstractPlugin: p}
}

// Init initializes the ChangeLog plugin
func (c *ChangeLog) Init() error {

	// Initialize the abstract plugin
	if err := c.AbstractPlugin.Init(); err != nil {
		return err
	}

	// Initialize the changeLogMap
	c.changeLogMap = make(map[string]*matrix.Matrix)

	object := c.ParentParams.GetChildS("object")
	c.matrixName = object.GetContentS() + "_" + changeLog

	// Initialize the changeLogMatrix
	if err := c.initMatrix(); err != nil {
		return err
	}

	// Populate the ChangeLog configuration
	if err := c.populateChangeLogConfig(); err != nil {
		return err
	}

	return nil
}

// populateChangeLogConfig populates the ChangeLog configuration from the plugin parameters
func (c *ChangeLog) populateChangeLogConfig() error {
	var err error
	changeLogYaml, err := yaml.Dump(c.Params)
	if err != nil {
		return err
	}

	c.changeLogConfig, err = getChangeLogConfig(c.ParentParams, string(changeLogYaml), c.Logger)
	if err != nil {
		return err
	}
	return nil
}

// initMatrix initializes a new matrix with the given name
func (c *ChangeLog) initMatrix() error {
	c.changeLogMap[c.matrixName] = matrix.New(c.Parent+c.matrixName, changeLog, c.matrixName)
	for _, changeLogMatrix := range c.changeLogMap {
		changeLogMatrix.SetExportOptions(matrix.DefaultExportOptions())
	}
	for _, k := range metrics {
		err := matrix.CreateMetric(k, c.changeLogMap[c.matrixName])
		if err != nil {
			c.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
			return err
		}
	}
	return nil
}

// Run processes the data and generates ChangeLog instances
func (c *ChangeLog) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {

	data := dataMap[c.Object]
	// Purge and reset data
	// remove all metrics as analytics label may change over time
	err := c.initMatrix()
	if err != nil {
		c.Logger.Warn().Err(err).Msg("error while init matrix")
		return nil, err
	}

	// if this is first poll
	if c.previousData == nil {
		c.copyPreviousData(data)
		return nil, nil
	}

	changeMat := c.changeLogMap[c.matrixName]

	changeMat.SetGlobalLabels(data.GetGlobalLabels())
	object := data.Object
	if c.changeLogConfig.Object == "" {
		c.Logger.Warn().Str("object", object).Msg("ChangeLog is not supported. Missing correct configuration")
		return nil, nil
	}

	prevMat := c.previousData
	oldInstances := set.New()
	prevInstancesUUIDKey := make(map[string]string)
	for key, prevInstance := range prevMat.GetInstances() {
		uuid := prevInstance.GetLabel("uuid")
		if uuid == "" {
			c.Logger.Warn().Str("object", object).Str("key", key).Msg("missing uuid")
			continue
		}
		prevInstancesUUIDKey[uuid] = key
		oldInstances.Add(key)
	}
	// check if prev exists
	if c.previousData != nil {
		// loop over current instances
		for key, instance := range data.GetInstances() {
			uuid := instance.GetLabel("uuid")
			if uuid == "" {
				c.Logger.Warn().Str("object", object).Str("key", key).Msg("missing uuid. ChangeLog is not supported")
				continue
			}

			prevKey := prevInstancesUUIDKey[uuid]
			if prevKey != "" {
				// instance already in cache
				oldInstances.Remove(prevKey)
			}

			prevInstance := c.previousData.GetInstance(prevKey)

			if prevInstance == nil {
				//instance created
				change := &Change{
					key:    uuid + "_" + object,
					object: object,
					op:     create,
					labels: make(map[string]string),
				}
				c.updateChangeLogLabels(object, instance, change)
				c.createChangeLogInstance(changeMat, change)
			} else {
				// check for any modification
				cur, old := instance.GetLabels().CompareLabels(prevInstance.GetLabels(), c.changeLogConfig.Track)
				if !cur.IsEmpty() {
					for currentLabel, newLabel := range cur.Iter() {
						change := &Change{
							key:      uuid + "_" + object + "_" + currentLabel,
							object:   object,
							op:       update,
							labels:   make(map[string]string),
							track:    currentLabel,
							oldValue: old.Get(currentLabel),
							newValue: newLabel,
						}
						c.updateChangeLogLabels(object, instance, change)
						// add changed track and its old, new value
						change.labels[track] = currentLabel
						change.labels[oldLabelValue] = old.Get(currentLabel)
						change.labels[newLabelValue] = newLabel
						c.createChangeLogInstance(changeMat, change)
					}
				}
			}
		}
		// create deleted instances change_log
		for key := range oldInstances.Iter() {
			prevInstance := prevMat.GetInstance(key)
			uuid := prevInstance.GetLabel("uuid")
			if uuid == "" {
				c.Logger.Warn().Str("object", object).Str("key", key).Msg("missing uuid. ChangeLog is not supported")
				continue
			}
			if prevInstance != nil {
				change := &Change{
					key:    uuid + "_" + object,
					object: object,
					op:     del,
					labels: make(map[string]string),
				}
				c.updateChangeLogLabels(object, prevInstance, change)
				c.createChangeLogInstance(changeMat, change)
			} else {
				c.Logger.Warn().Str("object", object).Str("key", key).Msg("missing instance")
			}
		}
	}

	var matricesArray []*matrix.Matrix
	matricesArray = append(matricesArray, changeMat)

	c.copyPreviousData(data)

	return matricesArray, nil
}

// copyPreviousData creates a copy of the previous data for comparison
func (c *ChangeLog) copyPreviousData(cur *matrix.Matrix) {
	labels := append(c.changeLogConfig.PublishLabels, c.changeLogConfig.Track...)
	labels = append(labels, "uuid")
	c.previousData = cur.Clone(matrix.With{Data: true, Metrics: false, Instances: true, ExportInstances: false, Labels: labels})
}

// createChangeLogInstance creates a new ChangeLog instance with the given change data
func (c *ChangeLog) createChangeLogInstance(mat *matrix.Matrix, change *Change) {
	cInstance, err := mat.NewInstance(change.key)
	if err != nil {
		c.Logger.Warn().Str("object", change.object).Str("key", change.key).Msg("error while creating instance")
		return
	}
	// copy keys
	cInstance.SetLabel(objectLabel, change.object)
	cInstance.SetLabel(opLabel, change.op)
	for k, v := range change.labels {
		cInstance.SetLabel(k, v)
	}
	m := mat.GetMetric("log")
	if m == nil {
		if m, err = mat.NewMetricFloat64("log"); err != nil {
			c.Logger.Warn().Err(err).Str("key", "alerts").Msg("error while creating metric")
			return
		}
	}
	if err = m.SetValueFloat64(cInstance, 1); err != nil {
		c.Logger.Error().Err(err).Str("metric", "alerts").Msg("Unable to set value on metric")
		return
	}
}

// updateChangeLogLabels populates change log labels
func (c *ChangeLog) updateChangeLogLabels(object string, instance *matrix.Instance, change *Change) {
	cl := c.changeLogConfig
	if len(cl.PublishLabels) > 0 {
		for _, l := range cl.PublishLabels {
			labelValue := instance.GetLabel(l)
			if labelValue == "" {
				c.Logger.Warn().Str("object", object).Str("label", l).Msg("Missing label")
			} else {
				change.labels[l] = instance.GetLabel(l)
			}
		}
	} else if cl.includeAll {
		for k := range instance.GetLabels().Iter() {
			change.labels[k] = instance.GetLabel(k)
		}
	} else {
		c.Logger.Warn().Str("object", object).Msg("missing publish labels")
	}
}
