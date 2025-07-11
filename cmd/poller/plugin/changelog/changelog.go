package changelog

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree/yaml"
	"log/slog"
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
	ObjectChangeLog = "change"
	objectLabel     = "object"
	opLabel         = "op"
	create          = "create"
	update          = "update"
	del             = "delete"
	Track           = "track"
	oldValue        = "old_value"
	newValue        = "new_value"
	indexLabel      = "index"
	Metric          = "metric"
	Label           = "label"
	Category        = "category"
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
	category string
}

// New initializes a new instance of the ChangeLog plugin
func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &ChangeLog{AbstractPlugin: p}
}

// Init initializes the ChangeLog plugin
func (c *ChangeLog) Init(remote conf.Remote) error {

	// Initialize the abstract plugin
	if err := c.AbstractPlugin.Init(remote); err != nil {
		return err
	}

	object := c.ParentParams.GetChildS("object")
	c.matrixName = object.GetContentS() + "_" + ObjectChangeLog

	return c.populateChangeLogConfig()
}

// populateChangeLogConfig populates the ChangeLog configuration from the plugin parameters
func (c *ChangeLog) populateChangeLogConfig() error {
	var err error
	changeLogYaml, err := yaml.Dump(c.Params)
	if err != nil {
		return err
	}

	c.changeLogConfig, err = getChangeLogConfig(c.ParentParams, changeLogYaml, c.SLogger)
	if err != nil {
		return err
	}
	return nil
}

// initMatrix initializes a new matrix with the given name
func (c *ChangeLog) initMatrix() (map[string]*matrix.Matrix, error) {
	changeLogMap := make(map[string]*matrix.Matrix)
	changeLogMap[c.matrixName] = matrix.New(c.Parent+c.matrixName, ObjectChangeLog, c.matrixName)
	for _, changeLogMatrix := range changeLogMap {
		changeLogMatrix.SetExportOptions(matrix.DefaultExportOptions())
	}
	for _, k := range metrics {
		err := matrix.CreateMetric(k, changeLogMap[c.matrixName])
		if err != nil {
			c.SLogger.Warn("error while creating metric", slog.Any("err", err), slog.String("key", k))
			return nil, err
		}
	}
	return changeLogMap, nil
}

// Run processes the data and generates ChangeLog instances
func (c *ChangeLog) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[c.Object]
	changeLogMap, err := c.initMatrix()
	if err != nil {
		c.SLogger.Warn("error while init matrix", slog.Any("err", err))
		return nil, nil, err
	}

	// reset metric count
	c.metricsCount = 0

	// if this is the first poll
	if c.previousData == nil {
		c.copyPreviousData(data)
		return nil, nil, nil
	}

	changeMat := changeLogMap[c.matrixName]
	changeMat.SetGlobalLabels(data.GetGlobalLabels())
	object := data.Object
	if c.changeLogConfig.Object == "" {
		c.SLogger.Warn("ChangeLog is not supported. Missing correct configuration", slog.String("object", object))
		return nil, nil, nil
	}

	prevMat := c.previousData
	oldInstances := set.New()
	prevInstancesUUIDKey := make(map[string]string)
	for key, prevInstance := range prevMat.GetInstances() {
		uuid := prevInstance.GetLabel("uuid")
		if uuid == "" {
			c.SLogger.Warn("missing uuid", slog.String("object", object), slog.String("key", key))
			continue
		}
		prevInstancesUUIDKey[uuid] = key
		oldInstances.Add(key)
	}

	currentTime := time.Now().Unix()

	metricChanges := c.CompareMetrics(data)
	// loop over current instances
	for key, instance := range data.GetInstances() {
		uuid := instance.GetLabel("uuid")
		if uuid == "" {
			c.SLogger.Warn(
				"missing uuid. ChangeLog is not supported",
				slog.String("object", object),
				slog.String("key", key),
			)
			continue
		}

		prevKey := prevInstancesUUIDKey[uuid]
		if prevKey != "" {
			// instance already in cache
			oldInstances.Remove(prevKey)
		}

		prevInstance := c.previousData.GetInstance(prevKey)

		if prevInstance == nil {
			// instance created
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
			// check for any label modification
			cur, old := instance.CompareDiffs(prevInstance, c.changeLogConfig.Track)
			if len(cur) > 0 {
				for currentLabel, nVal := range cur {
					change := &Change{
						key:      uuid + "_" + object + "_" + currentLabel,
						object:   object,
						op:       update,
						labels:   make(map[string]string),
						track:    currentLabel,
						category: Label,
						oldValue: old[currentLabel],
						newValue: nVal,
						time:     currentTime,
					}
					c.updateChangeLogLabels(object, instance, change)
					// add changed Track and its old, new value
					change.labels[Category] = change.category
					change.labels[Track] = currentLabel
					change.labels[oldValue] = change.oldValue
					change.labels[newValue] = nVal
					c.createChangeLogInstance(changeMat, change)
				}
			}

			// check for any metric modification
			if changes, ok := metricChanges[key]; ok {
				for metricName := range changes {
					change := &Change{
						key:      uuid + "_" + object + "_" + metricName,
						object:   object,
						op:       update,
						labels:   make(map[string]string),
						track:    metricName,
						category: Metric,
						// Enabling tracking of both old and new values results in the creation of a new time series each time the pair of values changes. For metrics tracking, it is not suitable.
						time: currentTime,
					}
					c.updateChangeLogLabels(object, instance, change)
					// add changed Track and its old, new value
					change.labels[Category] = change.category
					change.labels[Track] = metricName
					change.labels[oldValue] = change.oldValue
					change.labels[newValue] = change.newValue
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
			c.SLogger.Warn(
				"missing uuid. ChangeLog is not supported",
				slog.String("object", object),
				slog.String("key", key),
			)
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
			c.SLogger.Warn("missing instance", slog.String("object", object), slog.String("key", key))
		}
	}

	var matricesArray []*matrix.Matrix
	matricesArray = append(matricesArray, changeMat)

	metadata := &collector.Metadata{}
	metadata.PluginInstances = uint64(len(changeMat.GetInstances()))

	c.copyPreviousData(data)
	if len(changeMat.GetInstances()) > 0 {
		// The `index` variable is used to differentiate between changes to the same label in a Grafana dashboard.
		// It has a value between 0 and 100 and is used in the `change_log` query as `last_over_time`.
		c.index = (c.index + 1) % 100
		c.SLogger.Info(
			"Collected",
			slog.Int("instances", len(changeMat.GetInstances())),
			slog.Int("metrics", c.metricsCount),
			slog.Int("index", c.index),
		)
	}

	return matricesArray, metadata, nil
}

// CompareMetrics compares the metrics of the current and previous instances
func (c *ChangeLog) CompareMetrics(curMat *matrix.Matrix) map[string]map[string]struct{} {
	metricChanges := make(map[string]map[string]struct{})
	prevMat := c.previousData
	met := maps.Keys(c.previousData.GetMetrics())

	for metricKey := range met {
		prevMetric := prevMat.GetMetric(metricKey)
		curMetric := curMat.GetMetric(metricKey)
		for key, currInstance := range curMat.GetInstances() {
			prevInstance := prevMat.GetInstance(key)
			if prevInstance == nil {
				continue
			}
			prevIndex := prevInstance.GetIndex()
			currIndex := currInstance.GetIndex()
			curVal := curMetric.GetValues()[currIndex]
			prevVal := prevMetric.GetValues()[prevIndex]
			if curVal != prevVal {
				if _, ok := metricChanges[key]; !ok {
					metricChanges[key] = make(map[string]struct{})
				}
				metName := curMat.Object + "_" + curMetric.GetName()
				metricChanges[key][metName] = struct{}{}
			}
		}
	}
	return metricChanges
}

// copyPreviousData creates a copy of the previous data for comparison
func (c *ChangeLog) copyPreviousData(cur *matrix.Matrix) {
	labels := c.changeLogConfig.PublishLabels
	var met []string
	for _, t := range c.changeLogConfig.Track {
		mKey := cur.DisplayMetricKey(t)
		if mKey == "" {
			labels = append(labels, t)
		} else {
			met = append(met, mKey)
		}
	}
	labels = append(labels, "uuid")
	withMetrics := len(met) > 0
	c.previousData = cur.Clone(matrix.With{Data: true, Metrics: withMetrics, Instances: true, ExportInstances: false, Labels: labels, MetricsNames: met})
}

// createChangeLogInstance creates a new ChangeLog instance with the given change data
func (c *ChangeLog) createChangeLogInstance(mat *matrix.Matrix, change *Change) {
	cInstance, err := mat.NewInstance(change.key)
	if err != nil {
		c.SLogger.Warn(
			"error while creating instance",
			slog.Any("err", err),
			slog.String("object", change.object),
			slog.String("key", change.key),
		)
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
			c.SLogger.Warn("error while creating metric", slog.Any("err", err), slog.String("key", "log"))
			return
		}
	}
	m.SetValueInt64(cInstance, change.time)
}

// updateChangeLogLabels populates change log labels
func (c *ChangeLog) updateChangeLogLabels(object string, instance *matrix.Instance, change *Change) {
	cl := c.changeLogConfig
	switch {
	case len(cl.PublishLabels) > 0:
		for _, l := range cl.PublishLabels {
			labelValue := instance.GetLabel(l)
			if labelValue == "" {
				c.SLogger.Warn("missing label", slog.String("object", object), slog.String("label", l))
			} else {
				change.labels[l] = labelValue
			}
		}
	case cl.includeAll:
		maps.Copy(change.labels, instance.GetLabels())
	default:
		c.SLogger.Warn("missing publish labels", slog.String("object", object))
	}
}
