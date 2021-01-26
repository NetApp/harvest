package prometheus

import (
    "fmt"
    "strings"
    "poller/share"
    "poller/yaml"
    "poller/share/logger"
    "poller/structs/matrix"
    "poller/structs/opts"
)

var Log *logger.Logger = logger.New(1, "")

type Prometheus struct {
    Class string
    Name string
    Params *yaml.Node
    Options *opts.Opts
    Cache []*matrix.Matrix
}

func New(class string, params *yaml.Node, options *opts.Opts) *Prometheus {
    name := params.Name
    e := Prometheus{Class: class, Name: name, Params: params, Options: options}
    return &e
}

func (p* Prometheus) GetName() string {
    return p.Name
}

func (e *Prometheus) Init() error {
    Log = logger.New(0, e.Name)

    if e.Options.Debug {
        Log.Info("Initialized exporter. No HTTP server started since in debug mode")
        return nil
    }
    
    url := e.Params.GetChildValue("url")
    port := e.Params.GetChildValue("port")
    e.StartHttpd(url, port)

    Log.Info("Initialized Exporter. HTTP daemon serving at [http://%s:%s]", url, port)
    return nil
}

func (e *Prometheus) Export(data *matrix.Matrix) error {

    if e.Options.Debug {
        rendered := e.Render(data)
        Log.Debug("Simulating export of %d data points", len(rendered))
        for _, m := range rendered {
            fmt.Printf("M= %s%s%s\n", share.Pink, m, share.End)
        }
    } else {
        e.Cache = append(e.Cache, data)
        Log.Debug("Added data to cache")
    }

    return nil
}

func (e *Prometheus) Render(data *matrix.Matrix) []string {
    var rendered []string
    var metric_labels, key_labels []string
    var object string

    options := data.ExportOptions

    rendered = make([]string, 0)
    metric_labels = options.GetChildValues("include_labels")
    key_labels = options.GetChildValues("include_keys")
    object = data.Object

    for _, instance := range data.Instances {
        Log.Debug("Rendering instance [%d]", instance.Index)

        instance_labels := make([]string, 0)
        instance_keys := make([]string, 0)

        for _, key := range key_labels {
            value, found := data.GetInstanceLabel(instance, key)
            if found && value != "" {
                instance_keys = append(instance_keys, fmt.Sprintf("%s=\"%s\"", key, value))
            } else {
                Log.Debug("Skipped Key [%s] (%s) found=%v", key, value, found)
            }
        }

        for _, label := range metric_labels {
            value, found := data.GetInstanceLabel(instance, label)
            if found {
                instance_labels = append(instance_labels, fmt.Sprintf("%s=\"%s\"", label, value))
            } else {
                Log.Debug("Skipped Label [%s] (%s) found=%v", label, value, found)
            }
        }

        //Log.Debug("Parsed Keys: [%s]", strings.Join(instance_keys, ","))
        //Log.Debug("Parsed Labels: [%s]", strings.Join(instance_labels, ","))

        if len(instance_keys) == 0 {
            Log.Debug("Skipping instance, no keys parsed (%v) (%v)", instance_keys, instance_labels)
            continue
        }

        if len(instance_labels) > 0 {
            label_data := fmt.Sprintf("%s_labels{%s,%s} 1.0", object, strings.Join(instance_keys, ","), strings.Join(instance_labels, ","))
            rendered = append(rendered, label_data)
        } else {
            Log.Debug("Skipping instance labels (%v) (%v)", instance_keys, instance_labels)
        }

        for _, metric := range data.Metrics {

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
    Log.Debug("Renderd %d data points for [%s] %d instances", len(rendered), object, len(data.Instances))
    return rendered
}










