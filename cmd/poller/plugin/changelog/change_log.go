package changelog

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree/yaml"
)

// Constants for ChangeLog metrics and labels
const (
	changeLog       = "change"
	objectTpeLabel  = "object_type"
	changeTypeLabel = "change_type"
	create          = "create"
	modify          = "modify"
	del             = "delete"
	labelName       = "label_name"
	oldLabelValue   = "old_label_value"
	newLabelValue   = "new_label_value"
)

// Metrics to be used in ChangeLog
var metrics = []string{
	"log",
}

// ChangeLog represents the main structure of the ChangeLog plugin
type ChangeLog struct {
	*plugin.AbstractPlugin
	matrixName      string
	previousData    map[string]*matrix.Matrix
	changeLogMap    map[string]*matrix.Matrix
	changeLogConfig map[string]Entry
}

// Change represents a single change entry in the ChangeLog
type Change struct {
	key           string
	object        string
	changeType    string
	labels        map[string]string
	labelName     string
	oldLabelValue string
	newLabelValue string
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
	dump, err := yaml.Dump(c.Params)
	if err != nil {
		return err
	}
	object := c.ParentParams.GetChildS("object")

	c.changeLogConfig, err = getChangeLogConfig(object.GetContentS(), string(dump), c.Logger)
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
func (c *ChangeLog) Run(data map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
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

	changeMat.SetGlobalLabels(data[c.Object].GetGlobalLabels())
	for objectKey, objectMatrix := range data {
		object := objectMatrix.Object
		if _, ok := c.changeLogConfig[object]; !ok {
			c.Logger.Warn().Str("object", object).Msg("ChangeLog is not supported. Missing correct configuration")
			return nil, nil
		}

		prevMat := c.previousData[objectKey]
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
		// check if object exists
		if prev, ok := c.previousData[objectKey]; ok {
			// loop over current instances
			for key, instance := range objectMatrix.GetInstances() {
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

				prevInstance := prev.GetInstance(prevKey)

				if prevInstance == nil {
					//instance created
					change := &Change{
						key:        uuid + "_" + object,
						object:     object,
						changeType: create,
						labels:     make(map[string]string),
					}
					for _, l := range c.changeLogConfig[object].PublishLabels {
						labelValue := instance.GetLabel(l)
						if labelValue == "" {
							c.Logger.Warn().Str("object", object).Str("key", key).Str("label", l).Msg("Missing label")
						} else {
							change.labels[l] = instance.GetLabel(l)
						}
					}
					c.createChangeLogInstance(changeMat, change)
				} else {
					// check for any modification
					cur, old := instance.GetLabels().CompareLabels(prevInstance.GetLabels(), c.changeLogConfig[object].TrackLabels)
					if !cur.IsEmpty() {
						for currentLabel, newLabel := range cur.Iter() {
							change := &Change{
								key:           uuid + "_" + object + "_" + currentLabel,
								object:        object,
								changeType:    modify,
								labels:        make(map[string]string),
								labelName:     currentLabel,
								oldLabelValue: old.Get(currentLabel),
								newLabelValue: newLabel,
							}
							for _, l := range c.changeLogConfig[object].PublishLabels {
								labelValue := instance.GetLabel(l)
								if labelValue == "" {
									c.Logger.Warn().Str("object", object).Str("key", key).Str("label", l).Msg("Missing label")
								} else {
									change.labels[l] = instance.GetLabel(l)
								}
							}
							// add changed labelname and its old, new value
							change.labels[labelName] = currentLabel
							change.labels[oldLabelValue] = old.Get(currentLabel)
							change.labels[newLabelValue] = newLabelValue
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
						key:        uuid + "_" + object,
						object:     object,
						changeType: del,
						labels:     make(map[string]string),
					}
					for _, l := range c.changeLogConfig[object].PublishLabels {
						labelValue := prevInstance.GetLabel(l)
						if labelValue == "" {
							c.Logger.Warn().Str("object", object).Str("key", key).Str("label", l).Msg("Missing label")
						} else {
							change.labels[l] = prevInstance.GetLabel(l)
						}
					}
					c.createChangeLogInstance(changeMat, change)
				} else {
					c.Logger.Warn().Str("object", object).Str("key", key).Msg("missing instance")
				}
			}
		}
	}

	var matricesArray []*matrix.Matrix
	matricesArray = append(matricesArray, changeMat)

	c.copyPreviousData(data)

	return matricesArray, nil
}

// copyPreviousData creates a copy of the previous data for comparison
func (c *ChangeLog) copyPreviousData(data map[string]*matrix.Matrix) {
	c.previousData = make(map[string]*matrix.Matrix)
	for k, v := range data {
		c.previousData[k] = v.Clone(matrix.With{Data: true, Metrics: false, Instances: true, ExportInstances: false})
	}
}

// createChangeLogInstance creates a new ChangeLog instance with the given change data
func (c *ChangeLog) createChangeLogInstance(mat *matrix.Matrix, change *Change) {
	cInstance, err := mat.NewInstance(change.key)
	if err != nil {
		c.Logger.Warn().Str("object", change.object).Str("key", change.key).Msg("error while creating instance")
		return
	}
	// copy keys
	cInstance.SetLabel(objectTpeLabel, change.object)
	cInstance.SetLabel(changeTypeLabel, change.changeType)
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
