package matrix

// Print countents of matrix: metrics, instances and data
// For debugging / testing
// @BUG: prints all but last metric

import (
    "fmt"
    "strings"
    "goharvest2/share/util"
)

func (m *Matrix) PrintDimensions(msg string) {
    fmt.Printf("\n%s%s#########################  %s  ###########################%s\n", util.Bold, util.Pink, msg, util.End)

    if m.Data != nil && len(m.Data) != 0 {
        fmt.Printf("%scloning matrix %dx%d (arrays: %dx%d)%s\n\n", util.Red, m.SizeMetrics(), m.SizeInstances(), len(m.Data), len(m.Data[0]), util.End)
    } else {
        fmt.Printf("%scloning matrix %dx%d (arrays: --)%s\n\n", util.Red, m.SizeMetrics(), m.SizeInstances(), util.End)
    }
}

func (m *Matrix) Print() {

    fmt.Printf("\n%s%s%s - %s - %s%s\n", util.Bold, util.Red, m.Collector, m.Plugin, m.Object, util.End)
    fmt.Printf("%s%sArray dimensions: ", util.Bold, util.Red)
    if m.Data == nil {
        fmt.Printf("nil")
    } else if len(m.Data) == 0 {
        fmt.Printf("0")
    } else {
        fmt.Printf("%d x %d", len(m.Data), len(m.Data[0]))
    }
    fmt.Printf("%s\n", util.End)

    /* create sorted caches for metrics, labels and instances */
    lineLen := 8 + 50 + 60 + 15 + 15 + 7

    // sorted metrics cache by index
    mSorted := make(map[int]*Metric, 0)
    mKeys := make(map[int]string, 0)
    mCount := 0
    mMaxIndex := 0

    for key, metric := range m.GetMetrics() {
        if _, found := mSorted[metric.Index]; found {
            fmt.Printf("Error: metric (%s): duplicate index [%d]\n", key, metric.Index)
        } else {
            mSorted[metric.Index] = metric
            mKeys[metric.Index] = key
            mCount += 1

            if metric.Index > mMaxIndex {
                mMaxIndex = metric.Index
            }
        }
    }
    fmt.Printf("sorted metric cache, count: %d, max index: %d\n", mCount, mMaxIndex)
    fmt.Printf("     compare Matrix, count: %d, metrics index: %d\n", len(m.Metrics), m.SizeMetrics())

    // sorted label cache
    lSorted := make([]string, 0)
    lKeys := make([]string, 0)
    lCount := 0

    for key, display := range m.Labels.Iter() {
        lSorted = append(lSorted, display)
        lKeys = append(lKeys, key)
        lCount += 1
    }
    fmt.Printf("sorted label cache, count: %d\n", lCount)

    // sorted instance cache by index
    iSorted := make(map[int]*Instance, 0)
    iKeys := make([]string, m.SizeInstances())
    iCount := 0

    for key, instance := range m.GetInstances() {
        if _, found := iSorted[instance.Index]; found {
            fmt.Printf("Error: instance (%s): duplicate index [%d]\n", key, instance.Index)
        } else {
            iSorted[instance.Index] = instance
            iKeys[instance.Index] = key
            iCount += 1
        }
    }
    fmt.Printf("sorted instance cache, count: %d (out of %d)\n", iCount, m.SizeInstances())

    /* Print metric cache */
    fmt.Printf("\n\nMetric cache:\n\n")
    fmt.Println(strings.Repeat("+", lineLen))
    fmt.Printf("%-8s %s %s %-50s %s %60s %15s %10s\n", "index", util.Bold, util.Blue, "name", util.End, "key", "enabled", "size")
    fmt.Println(strings.Repeat("+", lineLen))

    for i:=0; i<=mMaxIndex; i+=1 {
        metric := mSorted[i]
        if metric == nil {
            continue
        }
        fmt.Printf("%-8d %s %s %-50s %s %60s %15v %10d\n",
            metric.Index,
            util.Bold,
            util.Blue,
            metric.Name,
            util.End,
            mKeys[i],
            metric.Enabled,
            metric.Size,
        )
    }

    /* Print labels */
    fmt.Printf("\n\nLabel cache:\n\n")
    fmt.Println(strings.Repeat("+", lineLen))
    fmt.Printf("%-8s %s %s %-50s %s %60s\n", "index", util.Bold, util.Yellow, "display", util.End, "key")
    fmt.Println(strings.Repeat("+", lineLen))
    for i:=0; i<lCount; i+=1 {
        fmt.Printf("%-8d %s %s %-50s %s %60s\n", i, util.Bold, util.Yellow, lSorted[i], util.End, lKeys[i])
    }

    /* Print instances with data and labels */
    fmt.Printf("\n\nInstance & Data cache:\n\n")
    for i:=0; i<iCount; i+=1 {
        fmt.Printf("\n")
        fmt.Println(strings.Repeat("-", 100))
        fmt.Printf("%-8d Instance:\n", i)
        fmt.Printf("%s%s%s\n", util.Grey, iKeys[i], util.End)
        fmt.Println(strings.Repeat("-", 100))

        instance := iSorted[i]

        fmt.Println(util.Bold, "\nlabels:\n", util.End)
        //fmt.Printf("\n%s%s%s\n", util.Grey, instance.Labels.String(), util.End)

        for j:=0; j<lCount; j+=1 {

            value, found := m.GetInstanceLabel(instance, lSorted[j])
            if !found {
                value = "--"
            }
            fmt.Printf("%-46s %s %s %50s %s\n", lSorted[j], util.Bold, util.Yellow, value, util.End)
        }

        fmt.Println(util.Bold, "\ndata:\n", util.End)

        for k:=0; k<=mMaxIndex; k+=1 {
            metric := mSorted[k]

            if metric == nil {
                continue
            }

            value, has := m.GetValue(metric, instance)
            if !has {
                fmt.Printf("%-46s %s %s %50s %s\n", metric.Name, util.Bold, util.Pink, "--", util.End)
            } else {
                fmt.Printf("%-46s %s %s %50f %s\n", metric.Name, util.Bold, util.Pink, value, util.End)
            }
        }
    }
    fmt.Println()
}
