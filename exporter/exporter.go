package exporter

import (
    "fmt"
    "strings"
    "local.host/share"
    "local.host/matrix"
    "local.host/template"
)

type Exporter struct {
    Class string
    Name string
}

func New(class, name string) *Exporter {
    var e Exporter
    e = Exporter{Class: class, Name: name}
    return &e
}

func (e *Exporter) Init() error {
    e.Log("initialized exporter!")
    return nil
}

func (e *Exporter) Log(format string, vars ...interface{}) {
    fmt.Printf("[%s:%s ", e.Class, e.Name)
    fmt.Printf(format, vars...)
    fmt.Println()
}

func (e *Exporter) Export(data *matrix.Matrix, options *template.Element) error {
    rendered := e.Render(data, options)
    for _, m := range rendered {
        fmt.Printf("M= %s%s%s\n", share.Pink, m, share.End)
    }

    e.Log("Export completed: exported %d data points", len(rendered))
    return nil
}

func (e *Exporter) Render(data *matrix.Matrix, options *template.Element) []string {
    var rendered []string
    var metric_labels, key_labels []string
    var object string

    rendered = make([]string, 0)
    metric_labels = options.GetChildValues("include_labels")
    key_labels = options.GetChildValues("include_keys")
    object = data.Object

    for _, instance := range data.GetInstances() {
        e.Log("Rendering instance [%d]", instance.Index)

        instance_labels := make([]string, 0)
        instance_keys := make([]string, 0)

        for _, key := range key_labels {
            value, found := data.GetInstanceLabel(&instance, key)
            if found && value != "" {
                instance_keys = append(instance_keys, fmt.Sprintf("%s=\"%s\"", key, value))
            } else {
                e.Log("Skipped Key [%s] (%s) found=%v", key, value, found)
            }
        }

        for _, label := range metric_labels {
            value, found := data.GetInstanceLabel(&instance, label)
            if found {
                instance_labels = append(instance_labels, fmt.Sprintf("%s=\"%s\"", label, value))
            } else {
                e.Log("Skipped Label [%s] (%s) found=%v", label, value, found)
            }
        }

        //e.Log("Parsed Keys: [%s]", strings.Join(instance_keys, ","))
        //e.Log("Parsed Labels: [%s]", strings.Join(instance_labels, ","))

        if len(instance_keys) == 0 {
            e.Log("Skipping instance, no keys parsed (%v) (%v)", instance_keys, instance_labels)
            continue
        }

        if len(instance_labels) > 0 {
            label_data := fmt.Sprintf("%s_labels{%s,%s} 1.0", object, strings.Join(instance_keys, ","), strings.Join(instance_labels, ","))
            rendered = append(rendered, label_data)
        } else {
            e.Log("Skipping instance labels (%v) (%v)", instance_keys, instance_labels)
        }

        for _, metric := range data.GetMetrics() {

            if !metric.Enabled {
                continue
            }

            value, set := data.GetValue(metric, instance)

            if set {
                metric_data := fmt.Sprintf("%s{%s} %f", metric.Display, strings.Join(instance_keys, ","), value)
                rendered = append(rendered, metric_data)
            }
        }
    }
    e.Log("Renderd %d data points for [%s] %d instances", len(rendered), object, len(data.Instances))
    return rendered
}










