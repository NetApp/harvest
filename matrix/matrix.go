package matrix

import (
	"fmt"
	"errors"
    "math"
    "strconv"
    "strings"
    "local.host/share"
)

type Metric struct {
	Index int
	Name string
	Display string
	Scalar bool
	Enabled bool
}

type Instance struct {
	Name string
	Index int
	Display string
	Labels map[string]string
}

func NewInstance (index int) *Instance {
    var I Instance
    I = Instance{Index: index}
    I.Labels = map[string]string{}
    return &I
}

type Matrix struct {
	Object string
	GlobalLabels map[string]string
	LabelNames map[string]string
    InstanceKeys [][]string
	ExportOptions map[string]string
	Instances map[string]Instance
	Metrics map[string]Metric
	MetricsIndex int /* since some metrics are arrays and we can't relay on len(Metrics) */
	Data [][]float64
}

func New(object string) *Matrix {
    m := Matrix{Object: object, MetricsIndex: 0 }
    m.GlobalLabels = map[string]string{}
    m.LabelNames = map[string]string{}
    m.InstanceKeys = [][]string{}
    m.Instances = map[string]Instance{}
    m.Metrics = map[string]Metric{}
    return &m
}

func (m *Matrix) InitData() error {
	var x, y, i, j int
	x = m.MetricsIndex+1
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
		m.Metrics[key] = metric
		m.MetricsIndex += 1
	}
	return &metric, err
}

func (m *Matrix) GetMetric(key string) (*Metric, bool) {
    var M Metric
    var found bool
    M, found = m.Metrics[key]
	return &M, found
}

func (m *Matrix) GetMetrics() []Metric {
	var M Metric
	var metrics []Metric
	metrics = make([]Metric, m.MetricsIndex+1)
	for _, M = range m.Metrics {
		metrics = append(metrics, M)
	}
	return metrics
}

func (m *Matrix) AddInstance(key string) (*Instance, error) {
	var I *Instance
	var exists bool
	var err error
	if _, exists = m.Instances[key]; exists {
		err = errors.New(fmt.Sprintf("Instance [%s] already in cache", key))
	} else {
		I = NewInstance(len(m.Instances))
		m.Instances[key] = *I
	}
	return I, err
}

func (m *Matrix) GetInstance(key string) (*Instance, bool) {
    var I Instance
    var found bool
    I, found = m.Instances[key]
    return &I, found
}

func (m *Matrix) GetInstances() []Instance {
	var I Instance
	var instances []Instance
	instances = make([]Instance, len(m.Instances))
	for _, I = range m.Instances {
		instances = append(instances, I)
	}
	return instances
}

func (m *Matrix) PrintInstances() {
    count := 0
    for key, instance := range m.Instances {
        fmt.Printf("%3d %s\n", instance.Index, key)
        count += 1
    }
    fmt.Printf("   --- Printed %d instances\n\n", count)
}

func (m *Matrix) ResetInstances() {
    m.Instances = make(map[string]Instance, 0)
}

func (m *Matrix) SetValueString(M *Metric, I *Instance, value []byte) error {
	var numeric float64
	var err error

    numeric, err = strconv.ParseFloat(string(value), 64)

    if err == nil {
		m.SetValue(M, I, numeric)
    }
	return err
}

func (m *Matrix) SetValue(M *Metric, I *Instance, value float64) {
	m.Data[M.Index][I.Index] = value
}

func (m *Matrix) GetValue(M Metric, I Instance) (float64, bool) {
	var value float64
	value = m.Data[M.Index][I.Index]
	return value, value==value
}

func (m *Matrix) AddLabel(key, name string) {
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

func (m *Matrix) SetInstanceLabel(I *Instance, display, value string) {
    I.Labels[display] = value
}

func (m *Matrix) GetInstanceLabel(I *Instance, display string) (string, bool) {
    var label string
    var found bool
    label, found = I.Labels[display]
    return label, found
}

func (m *Matrix) GetInstanceLabels(I *Instance) map[string]string {
	return I.Labels
}

func (m *Matrix) SetGlobalLabel(label, value string) {
	m.GlobalLabels[label] = value
}

func (m *Matrix) GetGlobalLabels() map[string]string {
	return m.GlobalLabels
}


func (m *Matrix) Print() {

    fmt.Printf("\n\n")

    /* create local caches */
    lineLen := 8 + 50 + 60 + 15 + 15 + 7

    mSorted := make(map[int]Metric, 0)
    mKeys := make([]string, 0)
    mCount := 0

    for key, metric := range m.Metrics {
        if _, found := mSorted[metric.Index]; found {
            fmt.Printf("Error: metric index [%d] duplicate\n", metric.Index)
        } else {
            mSorted[metric.Index] = metric
            mKeys = append(mKeys, key)
            mCount += 1
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

    iSorted := make(map[int]Instance, 0)
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

    for i:=0; i<mCount; i+=1 {
        metric := mSorted[i]
        fmt.Printf("%-8d %s %s %-50s %s %60s %15v %15v\n", metric.Index, share.Bold, share.Blue, metric.Display, share.End, mKeys[i], metric.Enabled, metric.Scalar)
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
            value, found := m.GetInstanceLabel(&instance, lSorted[j])
            if !found {
                value = "--"
            }
            fmt.Printf("%-46s %s %s %50s %s\n", lSorted[j], share.Bold, share.Yellow, value, share.End)
        }

        fmt.Println(share.Bold, "\ndata:\n", share.End)

        for k:=0; k<mCount; k+=1 {
            metric := mSorted[k]
            value, has := m.GetValue(metric, instance)
            if !has {
                fmt.Printf("%-46s %s %s %50s %s\n", metric.Display, share.Bold, share.Pink, "--", share.End)
            } else {
                fmt.Printf("%-46s %s %s %50f %s\n", metric.Display, share.Bold, share.Pink, value, share.End)
            }
        }
    }
    fmt.Println()
}
