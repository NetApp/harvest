/*
 * Copyright NetApp Inc, 2021 All rights reserved

Package Description:

   The Prometheus exporter exposes metrics to the Prometheus DB
   over an HTTP server. It consists of two concurrent components:

      - the "actual" exporter (this file): receives metrics from collectors,
        renders into the Prometheus format and stores in cache

      - the HTTP daemon (httpd.go): will listen for incoming requests and
        will serve metrics from that cache.

   Strictly speaking this is an HTTP-exporter, simply using the exposition
   format accepted by Prometheus.

   Special thanks Yann Bizeul who helped to identify that having no lock
   on the cache creates a race-condition (not caught on all Linux systems).
*/

package prometheus

import (
	"fmt"
	"goharvest2/cmd/poller/exporter"
	"goharvest2/pkg/color"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/set"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Default parameters
const (
	// maximum amount of time we will keep metrics in cache
	cacheMaxKeep = "300s"
	// apply a prefix to metrics globally (default none)
	globalPrefix = ""
)

type Prometheus struct {
	*exporter.AbstractExporter
	cache           *cache
	allowAddrs      []string
	allowAddrsRegex []*regexp.Regexp
	cacheAddrs      map[string]bool
	checkAddrs      bool
	addMetaTags     bool
	globalPrefix    string
}

func New(abc *exporter.AbstractExporter) exporter.Exporter {
	return &Prometheus{AbstractExporter: abc}
}

func (me *Prometheus) Init() error {

	if err := me.InitAbc(); err != nil {
		return err
	}

	// from abstract class, we get "export" and "render" time
	// some additional metadata instances
	if instance, err := me.Metadata.NewInstance("http"); err == nil {
		instance.SetLabel("task", "http")
	} else {
		return err
	}

	if instance, err := me.Metadata.NewInstance("info"); err == nil {
		instance.SetLabel("task", "info")
	} else {
		return err
	}

	if x := me.Params.GlobalPrefix; x != nil {
		me.Logger.Debug().Msgf("will use global prefix [%s]", *x)
		me.globalPrefix = *x
		if !strings.HasSuffix(me.globalPrefix, "_") {
			me.globalPrefix += "_"
		}
	} else {
		me.globalPrefix = globalPrefix
	}

	if me.Options.Debug {
		me.Logger.Debug().Msg("initialized without HTTP server since in debug mode")
		return nil
	}

	// add HELP and TYPE tags to exported metrics if requested
	if me.Params.ShouldAddMetaTags != nil && *me.Params.ShouldAddMetaTags {
		me.addMetaTags = true
	}

	// all other parameters are only relevant to the HTTP daemon
	if x := me.Params.CacheMaxKeep; x != nil {
		if d, err := time.ParseDuration(*x); err == nil {
			me.Logger.Debug().Msgf("using cache_max_keep [%s]", *x)
			me.cache = newCache(d)
		} else {
			me.Logger.Error().Stack().Err(err).Msgf("cache_max_keep [%s]", *x)
		}
	}

	if me.cache == nil {
		me.Logger.Debug().Msgf("using default cache_max_keep [%s]", cacheMaxKeep)
		if d, err := time.ParseDuration(cacheMaxKeep); err == nil {
			me.cache = newCache(d)
		} else {
			return err
		}
	}

	// allow access to metrics only from the given plain addresses
	if x := me.Params.AllowedAddrs; x != nil {
		me.allowAddrs = *x
		if len(me.allowAddrs) == 0 {
			me.Logger.Error().Stack().Err(nil).Msg("allow_addrs without any")
			return errors.New(errors.INVALID_PARAM, "allow_addrs")
		}
		me.checkAddrs = true
		me.Logger.Debug().Msgf("added %d plain allow rules", len(me.allowAddrs))
	}

	// allow access only from addresses matching one of defined regular expressions
	if x := me.Params.AllowedAddrsRegex; x != nil {
		me.allowAddrsRegex = make([]*regexp.Regexp, 0)
		for _, r := range *x {
			r = strings.TrimPrefix(strings.TrimSuffix(r, "`"), "`")
			if reg, err := regexp.Compile(r); err == nil {
				me.allowAddrsRegex = append(me.allowAddrsRegex, reg)
			} else {
				me.Logger.Error().Stack().Err(err).Msg("parse regex")
				return errors.New(errors.INVALID_PARAM, "allow_addrs_regex")
			}
		}
		if len(me.allowAddrsRegex) == 0 {
			me.Logger.Error().Stack().Err(nil).Msg("allow_addrs_regex without any")
			return errors.New(errors.INVALID_PARAM, "allow_addrs")
		}
		me.checkAddrs = true
		me.Logger.Debug().Msgf("added %d regex allow rules", len(me.allowAddrsRegex))
	}

	// cache addresses that have been allowed or denied already
	if me.checkAddrs {
		me.cacheAddrs = make(map[string]bool)
	}

	// finally the most important and only required parameter: port
	// can be passed to us either as an option or as a parameter
	port := me.Options.PromPort
	if port == 0 {
		if p := me.Params.Port; p == nil {
			me.Logger.Error().Stack().Err(nil).Msg("Issue while reading prometheus port")
		} else {
			port = *p
		}
	}

	// sanity check on port
	if port == 0 {
		return errors.New(errors.MISSING_PARAM, "port")
	} else if port < 0 {
		return errors.New(errors.INVALID_PARAM, "port")
	}

	// The optional parameter LocalHttpAddr is the address of the HTTP service, valid values are:
	//- "localhost" or "127.0.0.1", this limits access to local machine
	//- "" (default) or "0.0.0.0", allows access from network
	addr := me.Params.LocalHttpAddr
	if addr != "" {
		me.Logger.Debug().Str("addr", addr).Msg("Using custom local addr")
	}
	go me.startHttpD(addr, port)

	// @TODO: implement error checking to enter failed state if HTTPd failed
	// (like we did in Alpha)

	me.Logger.Debug().Msgf("initialized, HTTP daemon started at [http://%s:%d]", addr, port)

	return nil
}

// Unlike other Harvest exporters, we don't actually export data
// but put it in cache, for the HTTP daemon to serve on request
//
// An important aspect of the whole mechanism is that all incoming
// data should have a unique UUID and object pair, otherwise they'll
// will overwrite other data in the cache.
// This key is also used by the HTTP daemon to trace back the name
// of the collectors and plugins where the metrics come from (for the info page)
func (me *Prometheus) Export(data *matrix.Matrix) error {

	var (
		metrics [][]byte
		err     error
	)

	// lock the exporter, to prevent other collectors from writing to us
	me.Lock()
	defer me.Unlock()

	me.Logger.Trace().Msgf("incoming %s%s(%s) (%s)%s", color.Bold, color.Cyan, data.UUID, data.Object, color.End)

	// render metrics into Prometheus format
	start := time.Now()
	if metrics, err = me.render(data); err != nil {
		return err
	}
	// fix render time for metadata
	d := time.Since(start)

	// simulate export in debug mode
	if me.Options.Debug {
		me.Logger.Debug().Msg("no export since in debug mode")
		for _, m := range metrics {
			me.Logger.Debug().Msgf("M= %s", string(m))
		}
		return nil
	}

	// store metrics in cache
	key := data.UUID + "." + data.Object + "." + data.Identifier

	// lock cache, to prevent HTTPd reading while we are mutating it
	me.cache.Lock()
	me.cache.Put(key, metrics)
	me.cache.Unlock()
	me.Logger.Debug().Msgf("added to cache with key [%s%s%s%s]", color.Bold, color.Red, key, color.End)

	// update metadata
	me.AddExportCount(uint64(len(metrics)))
	err = me.Metadata.LazyAddValueInt64("time", "render", d.Microseconds())
	if err != nil {
		me.Logger.Error().Stack().Err(err).Msg("error")
	}
	err = me.Metadata.LazyAddValueInt64("time", "export", time.Since(start).Microseconds())
	if err != nil {
		me.Logger.Error().Stack().Err(err).Msg("error")
	}

	return nil
}

// Render metrics and labels into the exposition format, as described in
// https://prometheus.io/docs/instrumenting/exposition_formats/
//
// All metrics are implicitly "Gauge" counters. If requested we also submit
// HELP and TYPE metadata (see add_meta_tags in config).
//
// Metric name is concatenation of the collector object (e.g. "volume",
// "fcp_lif") + the metric name (e.g. "read_ops" => "volume_read_ops").
// We do this since same metrics for different object can have
// different set of labels and Prometheus does not allow this.
//
// Example outputs:
//
// volume_read_ops{node="my-node",vol="some_vol"} 2523
// fcp_lif_read_ops{vserver="nas_svm",port_id="e02"} 771

func (me *Prometheus) render(data *matrix.Matrix) ([][]byte, error) {
	var (
		rendered                                          [][]byte
		tagged                                            *set.Set
		labels_to_include, keys_to_include, global_labels []string
		prefix                                            string
		err                                               error
	)

	rendered = make([][]byte, 0)
	global_labels = make([]string, 0)

	if me.addMetaTags {
		tagged = set.New()
	}

	options := data.GetExportOptions()

	if x := options.GetChildS("instance_labels"); x != nil {
		labels_to_include = x.GetAllChildContentS()
		me.Logger.Debug().Msgf("requested instance_labels : %v", labels_to_include)
	}

	if x := options.GetChildS("instance_keys"); x != nil {
		keys_to_include = x.GetAllChildContentS()
		me.Logger.Debug().Msgf("requested keys_labels : %v", keys_to_include)
	}

	include_all_labels := false
	require_instance_keys := true

	if x := options.GetChildContentS("include_all_labels"); x != "" {
		if include_all_labels, err = strconv.ParseBool(x); err != nil {
			me.Logger.Error().Stack().Err(err).Msg("parameter: include_all_labels")
		}
	}

	if x := options.GetChildContentS("require_instance_keys"); x != "" {
		if require_instance_keys, err = strconv.ParseBool(x); err != nil {
			me.Logger.Error().Stack().Err(err).Msg("parameter: require_instance_keys")
		}
	}

	prefix = me.globalPrefix + data.Object

	for key, value := range data.GetGlobalLabels().Map() {
		global_labels = append(global_labels, fmt.Sprintf("%s=\"%s\"", key, value))
	}

	for key, instance := range data.GetInstances() {

		if !instance.IsExportable() {
			me.Logger.Trace().Msgf("skip instance [%s]: disabled for export", key)
			continue
		}

		me.Logger.Trace().Msgf("rendering instance [%s] (%v)", key, instance.GetLabels())

		instance_keys := make([]string, len(global_labels))
		copy(instance_keys, global_labels)
		instance_keys_ok := false
		instance_labels := make([]string, 0)

		if include_all_labels {
			for label, value := range instance.GetLabels().Map() {
				// temporary fix for the rarely happening duplicate labels
				// known case is: ZapiPerf -> 7mode -> disk.yaml
				// actual cause is the Aggregator plugin, which is adding node as
				// instance label (even though it's already a global label for 7modes)
				if !data.GetGlobalLabels().Has(label) {
					instance_keys = append(instance_keys, fmt.Sprintf("%s=\"%s\"", label, value))
				}
			}
		} else {
			for _, key := range keys_to_include {
				value := instance.GetLabel(key)
				instance_keys = append(instance_keys, fmt.Sprintf("%s=\"%s\"", key, value))
				if !instance_keys_ok && value != "" {
					instance_keys_ok = true
				}
				me.Logger.Trace().Msgf("++ key [%s] (%s) found=%v", key, value, value != "")
			}

			for _, label := range labels_to_include {
				value := instance.GetLabel(label)
				instance_labels = append(instance_labels, fmt.Sprintf("%s=\"%s\"", label, value))
				me.Logger.Trace().Msgf("++ label [%s] (%s) %t", label, value, value != "")
			}

			// @TODO, probably be strict, and require all keys to be present
			if !instance_keys_ok && require_instance_keys {
				me.Logger.Trace().Msgf("skip instance, no keys parsed (%v) (%v)", instance_keys, instance_labels)
				continue
			}

			// @TODO, check at least one label is found?
			if len(instance_labels) != 0 {
				var labelData string
				if me.Params.SortLabels {
					allLabels := make([]string, len(instance_labels))
					copy(allLabels, instance_labels)
					for _, instanceKey := range instance_keys {
						allLabels = append(allLabels, instanceKey)
					}
					sort.Strings(allLabels)
					labelData = fmt.Sprintf("%s_labels{%s} 1.0", prefix, strings.Join(allLabels, ","))
				} else {
					labelData = fmt.Sprintf("%s_labels{%s,%s} 1.0", prefix, strings.Join(instance_keys, ","), strings.Join(instance_labels, ","))
				}
				if me.addMetaTags && !tagged.Has(prefix+"_labels") {
					tagged.Add(prefix + "_labels")
					rendered = append(rendered, []byte("# HELP "+prefix+"_labels Pseudo-metric for "+data.Object+" labels"))
					rendered = append(rendered, []byte("# TYPE "+prefix+"_labels gauge"))
				}
				rendered = append(rendered, []byte(labelData))
			} else {
				me.Logger.Trace().Msgf("skip instance labels, no labels parsed (%v) (%v)", instance_keys, instance_labels)
			}
		}

		if me.Params.SortLabels {
			sort.Strings(instance_keys)
		}
		for mkey, metric := range data.GetMetrics() {

			if !metric.IsExportable() {
				me.Logger.Debug().Msgf("skip metric [%s]: disabled for export", mkey)
				continue
			}

			me.Logger.Trace().Msgf("rendering metric [%s]", mkey)

			if value, ok := metric.GetValueString(instance); ok {

				// metric is histogram
				if metric.HasLabels() {
					metric_labels := make([]string, 0)
					for k, v := range metric.GetLabels().Map() {
						metric_labels = append(metric_labels, fmt.Sprintf("%s=\"%s\"", k, v))
					}
					x := fmt.Sprintf("%s_%s{%s,%s} %s", prefix, metric.GetName(), strings.Join(instance_keys, ","), strings.Join(metric_labels, ","), value)

					if me.addMetaTags && !tagged.Has(prefix+"_"+metric.GetName()) {
						tagged.Add(prefix + "_" + metric.GetName())
						rendered = append(rendered, []byte("# HELP "+prefix+"_"+metric.GetName()+" Metric for "+data.Object))
						rendered = append(rendered, []byte("# TYPE "+prefix+"_"+metric.GetName()+" histogram"))
					}

					rendered = append(rendered, []byte(x))
					// scalar metric
				} else {
					x := fmt.Sprintf("%s_%s{%s} %s", prefix, metric.GetName(), strings.Join(instance_keys, ","), value)

					if me.addMetaTags && !tagged.Has(prefix+"_"+metric.GetName()) {
						tagged.Add(prefix + "_" + metric.GetName())
						rendered = append(rendered, []byte("# HELP "+prefix+"_"+metric.GetName()+" Metric for "+data.Object))
						rendered = append(rendered, []byte("# TYPE "+prefix+"_"+metric.GetName()+" gauge"))
					}

					rendered = append(rendered, []byte(x))
				}
			} else {
				me.Logger.Trace().Msg("skipped: no data value")
			}
		}
	}
	me.Logger.Debug().
		Str("object", data.Object).
		Int("rendered", len(rendered)).
		Int("instances", len(data.GetInstances())).
		Msg("Rendered data points for instances")
	return rendered, nil
}
