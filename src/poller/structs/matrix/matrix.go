package matrix

import (
	"fmt"
	"errors"
    "math"
    "strconv"
    "strings"
    "goharvest2/poller/share"
    "goharvest2/poller/yaml"
)

type Metric struct {
	Index int
	Name string
    Display string
    Labels []string
	Scalar bool
	Enabled bool
}

func (m *Metric) Copy() *Metric {
    n := Metric{Index: m.Index, Name: m.Name, Display: m.Display, Scalar: m.Scalar, Enabled: m.Enabled}
    n.Labels = make([]string, len(m.Labels))
    copy(n.Labels, m.Labels)
    return &n
}

type Instance struct {
	Name string
	Index int
	Display string
	Labels map[string]string
}

func (i *Instance) Copy() *Instance {
    n := Instance{ Index: i.Index, Name: i.Name, Display: i.Display}
    n.Labels = map[string]string{}
    n.Labels = copy_ss_map(i.Labels)
    return &n
}

func NewInstance (index int) *Instance {
    var I Instance
    I = Instance{Index: index}
    I.Labels = map[string]string{}
    return &I
}

type Matrix struct {
    Collector string
    Object string
    Plugin string
	GlobalLabels map[string]string
	LabelNames map[string]string
    InstanceKeys [][]string
	ExportOptions *yaml.Node
	Instances map[string]*Instance
	Metrics map[string]*Metric
	MetricsIndex int /* since some metrics are arrays and we can't relay on len(Metrics) */
    Data [][]float64
    IsMetadata bool
    MetadataType string
    MetadataObject string
}

func New(collector, object, plugin string) *Matrix {
    m := Matrix{Collector: collector, Object: object, Plugin: plugin, MetricsIndex: 0 }
    m.GlobalLabels = map[string]string{}
    m.LabelNames = map[string]string{}
    m.InstanceKeys = [][]string{}
    m.Instances = map[string]*Instance{}
    m.Metrics = map[string]*Metric{}
    return &m
}

func (m *Matrix) InitData() error {
	var x, y, i, j int
	x = m.MetricsIndex
	y = len(m.Instances)
	if x == 0 || y == 0 {
		return errors.New("Counter or Instance cache empty")
	}
	m.Data = make([][]float64, x)
	for i=0; i<x; i+=1 {
		m.Data[i] = make([]float64, y)
		for j=0; j<y; j+=1 {
			m.Data[i][j] = math.NaN()
		}
	}
	return nil
}

func (m *Matrix) AddMetric(key, display string, enabled bool) (*Metric, error) {
	var metric Metric
	var exists bool
	var err error
	if _, exists = m.Metrics[key]; exists {
		err = errors.New(fmt.Sprintf("Metric [%s] already in cache", key))
	} else {
        metric = Metric{Index: m.MetricsIndex, Display: display, Scalar: true, Enabled: enabled}
		m.Metrics[key] = &metric
		m.MetricsIndex += 1
	}
	return &metric, err
}


func (m *Matrix) AddMetricArray(key, display string, labels []string, enabled bool) (*Metric, error) {
    metric, err := m.AddMetric(key, display, enabled)
    if err == nil {
        metric.Labels = labels
        metric.Scalar = false
        m.MetricsIndex += len(labels) - 1
    }
    return metric, err
}

func (m *Matrix) GetMetric(key string) (*Metric, bool) {
    var metric *Metric
    var found bool
    metric, found = m.Metrics[key]
	return metric, found
}

/*
func (m *Matrix) GetMetrics() []Metric {
	var M Metric
	var metrics []Metric
	metrics = make([]Metric, m.MetricsIndex+1)
	for _, M = range m.Metrics {
		metrics = append(metrics, M)
	}
	return metrics
}*/

func (m *Matrix) AddInstance(key string) (*Instance, error) {
	var I *Instance
	var exists bool
	var err error
	if _, exists = m.Instances[key]; exists {
		err = errors.New(fmt.Sprintf("Instance [%s] already in cache", key))
	} else {
		I = NewInstance(len(m.Instances))
		m.Instances[key] = I
	}
	return I, err
}

func (m *Matrix) GetInstance(key string) (*Instance, bool) {
    var instance *Instance
    var found bool
    instance, found = m.Instances[key]
    return instance, found
}

/*func (m *Matrix) GetInstances() []Instance {
	var I Instance
	var instances []Instance
	instances = make([]Instance, len(m.Instances))
	for _, I = range m.Instances {
		instances = append(instances, I)
	}
	return instances
}*/

func (m *Matrix) PrintInstances() {
    count := 0
    for key, instance := range m.Instances {
        fmt.Printf("%3d %s\n", instance.Index, key)
        count += 1
    }
    fmt.Printf("   --- Printed %d instances\n\n", count)
}

func (m *Matrix) ResetInstances() {
    m.Instances = make(map[string]*Instance, 0)
}

func (m *Matrix) SetValueString(metric *Metric, instance *Instance, value []byte) error {
	var numeric float64
	var err error

    numeric, err = strconv.ParseFloat(string(value), 64)

    if err == nil {
		m.SetValue(metric, instance, numeric)
    }
	return err
}

func (m *Matrix) SetValue(metric *Metric, instance *Instance, value float64) {
	m.Data[metric.Index][instance.Index] = value
}

func (m *Matrix) SetValueForMetric(key string, instance *Instance, value float64) {
    metric, found := m.GetMetric(key)
    if found {
        m.SetValue(metric, instance, value)
    }
}

func (m *Matrix) SetValueForMetricAndInstance(metric_key, instance_key string, value float64) {
    if instance, found := m.GetInstance(instance_key); found {
        m.SetValueForMetric(metric_key, instance, value)
    }
}

func (m *Matrix) SetArrayValues(metric *Metric, instance *Instance, values []float64) {
    for i:=0; i<len(metric.Labels); i+=1 {
        m.Data[metric.Index+i][instance.Index] = values[i]        
    }
}

func (m *Matrix) GetValue(metric *Metric, instance *Instance) (float64, bool) {
	var value float64
	value = m.Data[metric.Index][instance.Index]
	return value, value==value
}

func (m *Matrix) GetArrayValues(metric *Metric, instance *Instance) []float64 {
    values := make([]float64, len(metric.Labels))
    for i:=0; i<len(metric.Labels); i+=1 {
        values[i] = m.Data[metric.Index+i][instance.Index]        
    }
    return values
}

func (m *Matrix) AddLabelName(name string) {
    m.AddLabelKeyName(name, name)
}

func (m *Matrix) AddLabelKeyName(key, name string) {
	m.LabelNames[key] = name
}

func (m *Matrix) GetLabel(key string) (string, bool) {
    display, found := m.LabelNames[key]
    return display, found
}

func (m *Matrix) AddInstanceKey(key []string) {
    m.InstanceKeys = append(m.InstanceKeys, key)
}

func (m *Matrix) GetInstanceKeys() [][]string {
    return m.InstanceKeys
}

func (m *Matrix) SetInstanceLabel(instance *Instance, display, value string) {
    instance.Labels[display] = value
}

func (m *Matrix) GetInstanceLabel(instance *Instance, display string) (string, bool) {
    var label string
    var found bool
    label, found = instance.Labels[display]
    return label, found
}

func (m *Matrix) GetInstanceLabels(instance *Instance) map[string]string {
	return instance.Labels
}

func (m *Matrix) SetGlobalLabel(label, value string) {
	m.GlobalLabels[label] = value
}

func (m *Matrix) GetGlobalLabels() map[string]string {
	return m.GlobalLabels
}

func (m *Matrix) SetExportOptions(options *yaml.Node) {
    m.ExportOptions = options
}

func DefaultExportOptions() *yaml.Node {
    n := yaml.New("export_options", "")
    n.AddNewChild("include_all_labels", "True")
    return n
}


func (m *Matrix) Copy() *Matrix {
    /* Only export options is passed as pointer */
    /* Everything else undergoes deep copy */


    n := New(m.Object, m.Collector, m.Plugin)

    /* Copy string2string maps */
    n.GlobalLabels = copy_ss_map(m.GlobalLabels)
    n.LabelNames = copy_ss_map(m.LabelNames)

    /* Copy Instance keys */
    n.InstanceKeys = make([][]string, len(m.InstanceKeys))
    for _, key := range m.InstanceKeys {
        n.InstanceKeys = append(n.InstanceKeys, key)
    }

    /* Copy numeric slices */
    n.MetricsIndex = m.MetricsIndex
    n.Data = make([][]float64, m.MetricsIndex)
    for i:=0; i<m.MetricsIndex; i+=1 {
        n.Data[i] = make([]float64, len(m.Instances))
        copy(n.Data[i], m.Data[i])
    }

    n.Metrics = map[string]*Metric{}
    for key, metric := range m.Metrics {
        n.Metrics[key] = metric.Copy()
    }

    n.Instances = map[string]*Instance{}
    for key, instance := range m.Instances {
        n.Instances[key] = instance.Copy()
    }
    return n
}

func copy_ss_map(src map[string]string) map[string]string {
    dest := map[string]string{}
    for key, value := range src {
        dest[key] = value
    }
    return dest
}

func (m *Matrix) Print() {

    fmt.Printf("\n\n")

    /* create local caches */
    lineLen := 8 + 50 + 60 + 15 + 15 + 7

    mSorted := make(map[int]*Metric, 0)
    mKeys := make(map[int]string, 0)
    mCount := 0
    mMaxIndex := 0

    for key, metric := range m.Metrics {
        if _, found := mSorted[metric.Index]; found {
            fmt.Printf("Error: metric index [%d] duplicate\n", metric.Index)
        } else {
            mSorted[metric.Index] = metric
            mKeys[metric.Index] = key
            mCount += 1

            if metric.Index > mMaxIndex {
                mMaxIndex = metric.Index
            }
        }
    }
    fmt.Printf("Sorted metric cache with %d elements (out of %d)\n", mCount, len(m.Metrics))

    lSorted := make([]string, 0)
    lKeys := make([]string, 0)
    lCount := 0

    for key, display := range m.LabelNames {
        lSorted = append(lSorted, display)
        lKeys = append(lKeys, key)
        lCount += 1
    }
    fmt.Printf("Sorted label cache with %d elements (out of %d)\n", lCount, len(m.LabelNames))

    iSorted := make(map[int]*Instance, 0)
    iKeys := make([]string, 0)
    iCount := 0
    for key, instance := range m.Instances {
        if _, found := iSorted[instance.Index]; found {
            fmt.Printf("Error: instance index [%d] is duplicate\n", instance.Index)
        } else {
            iSorted[instance.Index] = instance
            iKeys = append(iKeys, key)
            iCount += 1
        }
    }
    fmt.Printf("Sorted instance cache with %d elements (out of %d)\n", iCount, len(m.Instances))

    /* Print metric cache */
    fmt.Printf("\n\nMetric cache:\n\n")
    fmt.Println(strings.Repeat("+", lineLen))
    fmt.Printf("%-8s %s %s %-50s %s %60s %15s %15s\n", "index", share.Bold, share.Blue, "display", share.End, "key", "enabled", "scalar")
    fmt.Println(strings.Repeat("+", lineLen))

    for i:=0; i<mMaxIndex; i+=1 {
        metric := mSorted[i]
        if metric == nil {
            continue
        }
        if metric.Scalar {
            fmt.Printf("%-8d %s %s %-50s %s %60s %15v %15v\n", metric.Index, share.Bold, share.Blue, metric.Display, share.End, mKeys[i], metric.Enabled, metric.Scalar)
        } else {
            for k:=0; k<len(metric.Labels); k+=1 {
                fmt.Printf("%s %-8d %s %s %s %-50s %s\n", share.Grey, metric.Index+k, share.End, share.Bold, share.Cyan, metric.Labels[k], share.End)
            }
        }
        
    }

    /* Print labels */
    fmt.Printf("\n\nLabel cache:\n\n")
    fmt.Println(strings.Repeat("+", lineLen))
    fmt.Printf("%-8s %s %s %-50s %s %60s\n", "index", share.Bold, share.Yellow, "display", share.End, "key")
    fmt.Println(strings.Repeat("+", lineLen))
    for i:=0; i<lCount; i+=1 {
        fmt.Printf("%-8d %s %s %-50s %s %60s\n", i, share.Bold, share.Yellow, lSorted[i], share.End, lKeys[i])
    }

    /* Print instances with data and labels */
    fmt.Printf("\n\nInstance & Data cache:\n\n")
    for i:=0; i<iCount; i+=1 {
        fmt.Printf("\n")
        fmt.Println(strings.Repeat("-", 100))
        fmt.Printf("%-8d Instance:\n", i)
        fmt.Printf("%s%s%s\n", share.Grey, iKeys[i], share.End)
        fmt.Println(strings.Repeat("-", 100))

        instance := iSorted[i]

        fmt.Println(share.Bold, "\nlabels:\n", share.End)

        for j:=0; j<lCount; j+=1 {
            if lSorted[j] == "uid" {
                continue
            }
            value, found := m.GetInstanceLabel(instance, lSorted[j])
            if !found {
                value = "--"
            }
            fmt.Printf("%-46s %s %s %50s %s\n", lSorted[j], share.Bold, share.Yellow, value, share.End)
        }

        fmt.Println(share.Bold, "\ndata:\n", share.End)

        for k:=0; k<mMaxIndex; k+=1 {
            metric := mSorted[k]

            if metric == nil {
                continue
            }

            if metric.Scalar {
                value, has := m.GetValue(metric, instance)
                if !has {
                    fmt.Printf("%-46s %s %s %50s %s\n", metric.Display, share.Bold, share.Pink, "--", share.End)
                } else {
                    fmt.Printf("%-46s %s %s %50f %s\n", metric.Display, share.Bold, share.Pink, value, share.End)
                }
            } else {
                fmt.Printf("%-46s\n", metric.Display)
                values := m.GetArrayValues(metric, instance)
                for l:=0; l<len(values); l+=1 {
                    fmt.Printf("  %-44s %s %s %50f %s\n", metric.Labels[l], share.Bold, share.Pink, values[l], share.End)
                }
            }
        }
    }
    fmt.Println()
}
