package changelog

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree/yaml"
	"maps"
	"strconv"
	"time"
)

/*The changelog feature is only applicable to labels and requires a UUID for the label name to exist.
A default configuration for volume, SVM, and node is available, but the DSL can be overwritten as needed.
The shape of the change_log is specific to each label change and is only applicable to matrix collected by the collector.
*/

// Constants for ChangeLog metrics and labels
const (
	changeLog   = "change"
	objectLabel = "object"
	opLabel     = "op"
	create      = "create"
	update      = "update"
	del         = "delete"
	track       = "track"
	oldValue    = "old_value"
	newValue    = "new_value"
	indexLabel  = "index"
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
	index           int
	metricsCount    int
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
	time     int64
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

	return c.populateChangeLogConfig()
}

// populateChangeLogConfig populates the ChangeLog configuration from the plugin parameters
func (c *ChangeLog) populateChangeLogConfig() error {
	var err error
	changeLogYaml, err := yaml.Dump(c.Params)
	if err != nil {
		return err
	}

	c.changeLogConfig, err = getChangeLogConfig(c.ParentParams, changeLogYaml, c.Logger)
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
	// remove all metrics
	err := c.initMatrix()
	if err != nil {
		c.Logger.Warn().Err(err).Msg("error while init matrix")
		return nil, err
	}

	// reset metric count
	c.metricsCount = 0

	// if this is the first poll
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

	currentTime := time.Now().Unix()

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
				time:   currentTime,
			}
			c.updateChangeLogLabels(object, instance, change)
			c.createChangeLogInstance(changeMat, change)
		} else {
			// check for any modification
			cur, old := instance.CompareDiffs(prevInstance, c.changeLogConfig.Track)
			if len(cur) > 0 {
				for currentLabel, nVal := range cur {
					change := &Change{
						key:      uuid + "_" + object + "_" + currentLabel,
						object:   object,
						op:       update,
						labels:   make(map[string]string),
						track:    currentLabel,
						oldValue: old[currentLabel],
						newValue: nVal,
						time:     currentTime,
					}
					c.updateChangeLogLabels(object, instance, change)
					// add changed track and its old, new value
					change.labels[track] = currentLabel
					change.labels[oldValue] = change.oldValue
					change.labels[newValue] = nVal
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
				time:   currentTime,
			}
			c.updateChangeLogLabels(object, prevInstance, change)
			c.createChangeLogInstance(changeMat, change)
		} else {
			c.Logger.Warn().Str("object", object).Str("key", key).Msg("missing instance")
		}
	}

	var matricesArray []*matrix.Matrix
	matricesArray = append(matricesArray, changeMat)

	c.copyPreviousData(data)
	if len(changeMat.GetInstances()) > 0 {
		// The `index` variable is used to differentiate between changes to the same label in a Grafana dashboard.
		// It has a value between 0 and 100 and is used in the `change_log` query as `last_over_time`.
		c.index = (c.index + 1) % 100
		c.Logger.Info().Int("instances", len(changeMat.GetInstances())).
			Int("metrics", c.metricsCount).
			Int("index", c.index).
			Msg("Collected")
	}

	return matricesArray, nil
}

// copyPreviousData creates a copy of the previous data for comparison
func (c *ChangeLog) copyPreviousData(cur *matrix.Matrix) {
	labels := c.changeLogConfig.PublishLabels
	labels = append(labels, c.changeLogConfig.Track...)
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
	cInstance.SetLabel(indexLabel, strconv.Itoa(c.index))
	for k, v := range change.labels {
		cInstance.SetLabel(k, v)
	}
	c.metricsCount += len(cInstance.GetLabels())
	m := mat.GetMetric("log")
	if m == nil {
		if m, err = mat.NewMetricFloat64("log"); err != nil {
			c.Logger.Warn().Err(err).Str("key", "log").Msg("error while creating metric")
			return
		}
	}
	if err = m.SetValueInt64(cInstance, change.time); err != nil {
		c.Logger.Error().Err(err).Int64("val", change.time).Msg("Unable to set value on metric")
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
				change.labels[l] = labelValue
			}
		}
	} else if cl.includeAll {
		maps.Copy(change.labels, instance.GetLabels())
	} else {
		c.Logger.Warn().Str("object", object).Msg("missing publish labels")
	}
}
