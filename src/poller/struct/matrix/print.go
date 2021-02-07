package matrix

// Print countents of matrix: metrics, instances and data
// For debugging / testing
// @BUG: prints all but last metric

import (
    "fmt"
    "strings"
    "goharvest2/poller/util"
)

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

    for key, display := range m.LabelNames.Iter() {
        lSorted = append(lSorted, display)
        lKeys = append(lKeys, key)
        lCount += 1
    }
    fmt.Printf("Sorted label cache with %d elements (out of %d)\n", lCount, m.LabelNames.Size())

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
    fmt.Printf("%-8s %s %s %-50s %s %60s %15s %15s\n", "index", util.Bold, util.Blue, "display", util.End, "key", "enabled", "scalar")
    fmt.Println(strings.Repeat("+", lineLen))

    for i:=0; i<mMaxIndex; i+=1 {
        metric := mSorted[i]
        if metric == nil {
            continue
        }
        if metric.Scalar {
            fmt.Printf("%-8d %s %s %-50s %s %60s %15v %15v\n", metric.Index, util.Bold, util.Blue, metric.Display, util.End, mKeys[i], metric.Enabled, metric.Scalar)
        } else {
            for k:=0; k<len(metric.Labels); k+=1 {
                fmt.Printf("%s %-8d %s %s %s %-50s %s\n", util.Grey, metric.Index+k, util.End, util.Bold, util.Cyan, metric.Labels[k], util.End)
            }
        }
        
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

        fmt.Printf("\n%s%s%s\n", util.Grey, instance.Labels.String(), util.End)

        for j:=0; j<lCount; j+=1 {

            value, found := m.GetInstanceLabel(instance, lSorted[j])
            if !found {
                value = "--"
            }
            fmt.Printf("%-46s %s %s %50s %s\n", lSorted[j], util.Bold, util.Yellow, value, util.End)
        }

        fmt.Println(util.Bold, "\ndata:\n", util.End)

        for k:=0; k<mMaxIndex; k+=1 {
            metric := mSorted[k]

            if metric == nil {
                continue
            }

            if metric.Scalar {
                value, has := m.GetValue(metric, instance)
                if !has {
                    fmt.Printf("%-46s %s %s %50s %s\n", metric.Display, util.Bold, util.Pink, "--", util.End)
                } else {
                    fmt.Printf("%-46s %s %s %50f %s\n", metric.Display, util.Bold, util.Pink, value, util.End)
                }
            } else {
                fmt.Printf("%-46s\n", metric.Display)
                values := m.GetArrayValues(metric, instance)
                for l:=0; l<len(values); l+=1 {
                    fmt.Printf("  %-44s %s %s %50f %s\n", metric.Labels[l], util.Bold, util.Pink, values[l], util.End)
                }
            }
        }
    }
    fmt.Println()
}
