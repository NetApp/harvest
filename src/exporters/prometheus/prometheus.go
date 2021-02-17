package main

import (
    "fmt"
    "strings"
    "bytes"
    "goharvest2/share/logger"
    "goharvest2/poller/exporter"
    "goharvest2/poller/struct/matrix"
)

type Prometheus struct {
    *exporter.AbstractExporter
    cache map[string]*matrix.Matrix
}

func New(abc *exporter.AbstractExporter) exporter.Exporter {
    return &Prometheus{AbstractExporter: abc}
}

func (e *Prometheus) Init() error {

    if err := e.InitAbc(); err != nil {
        return err
    }

    if e.Options.Debug {
        logger.Info(e.Prefix, "Initialized exporter. No HTTP server started since in debug mode")
        return nil
    }
    
    e.cache = make(map[string]*matrix.Matrix)

    url := e.Params.GetChildContentS("url")
    port := e.Params.GetChildContentS("port")
    e.StartHttpd(url, port)
 
    logger.Info(e.Prefix, "Initialized Exporter. HTTP daemon serving at [http://%s:%s]", url, port)

    return nil
}

func (e *Prometheus) Export(data *matrix.Matrix) error {
    e.Lock()
    defer e.Unlock()

	if e.Options.Debug {
		logger.Debug(e.Prefix, "no export since in debug mode")
		if metrics, err := e.Render(data); err == nil {
			for _, m := range metrics {
				logger.Debug(e.Prefix, "M= %s", bytes.TrimRight(m, "\n"))
			}
		} else {
			return err
		}
	}
    key := data.Collector + "." + data.Plugin + "." + data.Object
    if data.IsMetadata {
        key = data.MetadataType + "." + data.MetadataObject + "." + key
    }
    delete(e.cache, key)
    e.cache[key] = data
    logger.Debug(e.Prefix, "added to cache with key [%s]", key)

    return nil
}

func (e *Prometheus) Render(data *matrix.Matrix) ([][]byte, error) {
    var rendered [][]byte
    var metric_labels, key_labels []string
    var object, prefix, instance_tag string
    var include_all_labels, include_instance_names bool

    options := data.ExportOptions

    //options.Print(0)

    rendered = make([][]byte, 0)
    // @TODO check for nil
    
    if options.GetChildS("instance_labels") != nil {
        metric_labels = options.GetChildS("instance_labels").GetAllChildContentS()
        logger.Debug(e.Prefix, "requested instance_labels : %v", metric_labels)
    }

    if options.GetChildS("instance_keys") != nil {
        key_labels = options.GetChildS("instance_keys").GetAllChildContentS()
        logger.Debug(e.Prefix, "requested keys_labels : %v", key_labels)
    }
    
    if options.GetChildContentS("include_all_labels") == "True" {
        include_all_labels = true
    } else {
        include_all_labels = false
    }

    if options.GetChildContentS("include_instance_names") == "True" {
        include_instance_names = true
    } else {
        include_instance_names = false
    }

    object = data.Object
    if data.IsMetadata {
        prefix = "metadata_" + data.MetadataType
        instance_tag = data.MetadataObject
    } else {
        prefix = object
        instance_tag = object
    }

    global_labels := make([]string, 0)
    for key, value := range data.GlobalLabels.Iter() {
        global_labels = append(global_labels, fmt.Sprintf("%s=\"%s\"", key, value))
    }

    for raw_key, instance := range data.Instances {

        logger.Debug(e.Prefix, "Rendering instance [%d] %v", instance.Index, instance.Labels.Iter())

        instance_labels := make([]string, 0)
        instance_keys := make([]string, len(global_labels))
        copy(instance_keys, global_labels)

        // @TODO deprecate
        if include_instance_names {
            instance_keys = append(instance_keys, fmt.Sprintf("%s=\"%s\"", instance_tag, raw_key))
        }
        
        for _, key := range key_labels {
            //value, found := data.GetInstanceLabel(instance, key)
            value, found := instance.Labels.GetHas(key)
            if include_all_labels || (found && value != "") {
                instance_keys = append(instance_keys, fmt.Sprintf("%s=\"%s\"", key, value))
            } else {
                logger.Debug(e.Prefix, "Key [%s] (%s) found=%v", key, value, found)
            }
        }

        for _, label := range metric_labels {
            //value, found := data.GetInstanceLabel(instance, label)
            value, found := instance.Labels.GetHas(label)
            if found {
                instance_labels = append(instance_labels, fmt.Sprintf("%s=\"%s\"", label, value))
            } else {
                logger.Debug(e.Prefix, "Label [%s] (%s) found=%v", label, value, found)
            }
        }

        //logger.Debug(e.Prefix, "Parsed Keys: [%s]", strings.Join(instance_keys, ","))
        //logger.Debug(e.Prefix, "Parsed Labels: [%s]", strings.Join(instance_labels, ","))

        if len(instance_keys) == 0 {
            logger.Debug(e.Prefix, "Skipping instance, no keys parsed (%v) (%v)", instance_keys, instance_labels)
            continue
        }

        if len(instance_labels) > 0 {
            label_data := fmt.Sprintf("%s_labels{%s,%s} 1.0", prefix, strings.Join(instance_keys, ","), strings.Join(instance_labels, ","))
            rendered = append(rendered, []byte(label_data))
        } else {
            logger.Debug(e.Prefix, "Skipping instance metric labels parsed (%v) (%v)", instance_keys, instance_labels)
        }

        for _, metric := range data.Metrics {

            if !metric.Enabled {
                continue
            }

            if metric.Scalar {
                if value, set := data.GetValue(metric, instance); set {
                    metric_data := fmt.Sprintf("%s_%s{%s} %f", prefix, metric.Display, strings.Join(instance_keys, ","), value)
                    rendered = append(rendered, []byte(metric_data))
                }
            } else {
                values := data.GetArrayValues(metric, instance)
                if metric.Dimensions == 1 {
                    for i:=0; i<len(metric.Labels); i+=1 {
                        if values[i] == values[i] {
                            metric_data := fmt.Sprintf("%s_%s{%s,metric=\"%s\"} %f", prefix, metric.Display, strings.Join(instance_keys, ","), metric.Labels[i], values[i])
                            rendered = append(rendered, []byte(metric_data))
                        }
                    }
                } else if metric.Dimensions == 2 {
                    for i:=0; i<len(metric.Labels); i+=1 {
                        for j:=0; j<len(metric.SubLabels); j+=1 {
                            k := i * len(metric.SubLabels) + j
                            if values[k] == values[k] {
                                metric_data := fmt.Sprintf("%s_%s{%s,metric=\"%s\",submetric=\"%s\"} %f", prefix, metric.Display, strings.Join(instance_keys, ","), metric.Labels[i], metric.SubLabels[j], values[k])
                                rendered = append(rendered, []byte(metric_data))
                            }
                        }
                    }
                }
            }
        }
    }
    logger.Debug(e.Prefix, "Renderd %d data points for [%s] %d instances", len(rendered), object, len(data.Instances))
    return rendered, nil
}










