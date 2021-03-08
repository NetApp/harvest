package main

import (
    "fmt"
    "strings"
    "strconv"
    "bytes"
    "goharvest2/share/logger"
    "goharvest2/share/matrix"
    "goharvest2/share/errors"
    "goharvest2/share/util"
    "goharvest2/poller/exporter"
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

    e.cache = make(map[string]*matrix.Matrix)

    if e.Options.Debug {
        logger.Debug(e.Prefix, "Initialized exporter. No HTTP server started since in debug mode")
        return nil
    }

    addr := e.Params.GetChildContentS("addr")
    if addr == "" {
        addr = "0.0.0.0"
    }

    port := e.Options.PrometheusPort
    if port == "" {
        port = e.Params.GetChildContentS("port")
    }

    // sanity check on port
    if port == "" {
        return errors.New(errors.MISSING_PARAM, "port") 
    } else if _, err := strconv.Atoi(port); err != nil {
        return errors.New(errors.INVALID_PARAM, "port (" + port + ")")
    }
    
    e.StartHttpd(addr, port)

    logger.Debug(e.Prefix, "Initialized Exporter. HTTP daemon serving at [http://%s:%s]", addr, port)

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
        key += "." + data.MetadataType + "." + data.MetadataObject
    }
    delete(e.cache, key)
    e.cache[key] = data
    logger.Debug(e.Prefix, "added to cache with key [%s%s%s%s]", util.Bold, util.Red, key, util.End)

    return nil
}

func (e *Prometheus) Render(data *matrix.Matrix) ([][]byte, error) {
    var count int
    var rendered [][]byte
    var labels_to_include, keys_to_include, global_labels []string
    var object, prefix string
    var include_all_labels bool

    options := data.ExportOptions
    // @TODO check for nil

    rendered = make([][]byte, 0)
    
    if options.GetChildS("instance_labels") != nil {
        labels_to_include = options.GetChildS("instance_labels").GetAllChildContentS()
        logger.Debug(e.Prefix, "requested instance_labels : %v", labels_to_include)
    }

    if options.GetChildS("instance_keys") != nil {
        keys_to_include = options.GetChildS("instance_keys").GetAllChildContentS()
        logger.Debug(e.Prefix, "requested keys_labels : %v", keys_to_include)
    }

    if options.GetChildContentS("include_all_labels") == "True" {
        include_all_labels = true
    } else {
        include_all_labels = false
    }

    object = data.Object
    if data.IsMetadata {
        prefix = "metadata_" + data.MetadataType
        //instance_tag = data.MetadataObject
    } else {
        prefix = object
        //instance_tag = object
    }

    global_labels = make([]string, 0)
    for key, value := range data.GlobalLabels.Iter() {
        global_labels = append(global_labels, fmt.Sprintf("%s=\"%s\"", key, value))
    }

    for _, instance := range data.Instances {

        logger.Debug(e.Prefix, "render instance [%d] %v", instance.Index, instance.Labels.Iter())

        instance_keys := make([]string, len(global_labels))
        instance_labels := make([]string, 0)
        copy(instance_keys, global_labels)

        if include_all_labels {
            for label, value := range instance.Labels.Iter() {
                instance_keys = append(instance_keys, fmt.Sprintf("%s=\"%s\"", label, value))
            }
        } else {
            for _, key := range keys_to_include {
                value, found := instance.Labels.GetHas(key)
                if found {
                    instance_keys = append(instance_keys, fmt.Sprintf("%s=\"%s\"", key, value))
                }
                logger.Debug(e.Prefix, "++ key [%s] (%s) found=%v", key, value, found)
            }

            for _, label := range labels_to_include {
                value, found := instance.Labels.GetHas(label)
                instance_labels = append(instance_labels, fmt.Sprintf("%s=\"%s\"", label, value))
                logger.Debug(e.Prefix, "++ label [%s] (%s) found=%v", label, value, found)
            }

            // @TODO, probably be strict, and require all keys to be present
            if len(instance_keys) == 0 {
                logger.Debug(e.Prefix, "skip instance, no keys parsed (%v) (%v)", instance_keys, instance_labels)
                continue
            }

            // @TODO, check at least one label is found?
            if len(instance_labels) != 0 {
                label_data := fmt.Sprintf("%s_labels{%s,%s} 1.0", prefix, strings.Join(instance_keys, ","), strings.Join(instance_labels, ","))
                rendered = append(rendered, []byte(label_data))
            } else {
                logger.Debug(e.Prefix, "skip instance labels, no labels parsed (%v) (%v)", instance_keys, instance_labels)
            }
        }


        for key, metric := range data.Metrics {

            if !metric.Enabled {

            logger.Debug(e.Prefix, "skip metric [%d] %s: disabled", metric.Index, key)
                continue
            }

            logger.Debug(e.Prefix, "render metric [%d] %s", metric.Index, key)

            if value, ok := data.GetValue(metric, instance); ok {

                if metric.Labels != nil && metric.Labels.Size() != 0 {
                    metric_labels := make([]string, 0, metric.Labels.Size())
                    for key, value := range metric.Labels.Iter() {
                        metric_labels = append(metric_labels, fmt.Sprintf("%s=\"%s\"", key, value))
                    }
                    x := fmt.Sprintf("%s_%s{%s,%s} %f", prefix, metric.Name, strings.Join(instance_keys, ","), strings.Join(metric_labels, ","), value)
                    rendered = append(rendered, []byte(x))
                    count += 1
                } else {
                    x := fmt.Sprintf("%s_%s{%s} %f", prefix, metric.Name, strings.Join(instance_keys, ","), value) 
                    rendered = append(rendered, []byte(x))
                    count += 1
                }
            } else {
                logger.Debug(e.Prefix, "skipped: no data value")
            }
        }
    }
    e.AddCount(count)
    logger.Debug(e.Prefix, "rendered %d data points for [%s] %d instances", len(rendered), object, len(data.Instances))
    return rendered, nil
}










