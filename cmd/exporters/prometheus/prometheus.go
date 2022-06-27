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
	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
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

func (p *Prometheus) Init() error {

	if err := p.InitAbc(); err != nil {
		return err
	}

	// from abstract class, we get "export" and "render" time
	// some additional metadata instances
	if instance, err := p.Metadata.NewInstance("http"); err == nil {
		instance.SetLabel("task", "http")
	} else {
		return err
	}

	if instance, err := p.Metadata.NewInstance("info"); err == nil {
		instance.SetLabel("task", "info")
	} else {
		return err
	}

	if x := p.Params.GlobalPrefix; x != nil {
		p.Logger.Debug().Msgf("will use global prefix [%s]", *x)
		p.globalPrefix = *x
		if !strings.HasSuffix(p.globalPrefix, "_") {
			p.globalPrefix += "_"
		}
	} else {
		p.globalPrefix = globalPrefix
	}

	if p.Options.Debug {
		p.Logger.Debug().Msg("initialized without HTTP server since in debug mode")
		return nil
	}

	// add HELP and TYPE tags to exported metrics if requested
	if p.Params.ShouldAddMetaTags != nil && *p.Params.ShouldAddMetaTags {
		p.addMetaTags = true
	}

	// all other parameters are only relevant to the HTTP daemon
	if x := p.Params.CacheMaxKeep; x != nil {
		if d, err := time.ParseDuration(*x); err == nil {
			p.Logger.Debug().Msgf("using cache_max_keep [%s]", *x)
			p.cache = newCache(d)
		} else {
			p.Logger.Error().Stack().Err(err).Msgf("cache_max_keep [%s]", *x)
		}
	}

	if p.cache == nil {
		p.Logger.Debug().Msgf("using default cache_max_keep [%s]", cacheMaxKeep)
		if d, err := time.ParseDuration(cacheMaxKeep); err == nil {
			p.cache = newCache(d)
		} else {
			return err
		}
	}

	// allow access to metrics only from the given plain addresses
	if x := p.Params.AllowedAddrs; x != nil {
		p.allowAddrs = *x
		if len(p.allowAddrs) == 0 {
			p.Logger.Error().Stack().Err(nil).Msg("allow_addrs without any")
			return errs.New(errs.ErrInvalidParam, "allow_addrs")
		}
		p.checkAddrs = true
		p.Logger.Debug().Msgf("added %d plain allow rules", len(p.allowAddrs))
	}

	// allow access only from addresses matching one of defined regular expressions
	if x := p.Params.AllowedAddrsRegex; x != nil {
		p.allowAddrsRegex = make([]*regexp.Regexp, 0)
		for _, r := range *x {
			r = strings.TrimPrefix(strings.TrimSuffix(r, "`"), "`")
			if reg, err := regexp.Compile(r); err == nil {
				p.allowAddrsRegex = append(p.allowAddrsRegex, reg)
			} else {
				p.Logger.Error().Stack().Err(err).Msg("parse regex")
				return errs.New(errs.ErrInvalidParam, "allow_addrs_regex")
			}
		}
		if len(p.allowAddrsRegex) == 0 {
			p.Logger.Error().Stack().Err(nil).Msg("allow_addrs_regex without any")
			return errs.New(errs.ErrInvalidParam, "allow_addrs")
		}
		p.checkAddrs = true
		p.Logger.Debug().Msgf("added %d regex allow rules", len(p.allowAddrsRegex))
	}

	// cache addresses that have been allowed or denied already
	if p.checkAddrs {
		p.cacheAddrs = make(map[string]bool)
	}

	// finally the most important and only required parameter: port
	// can be passed to us either as an option or as a parameter
	port := p.Options.PromPort
	if port == 0 {
		if promPort := p.Params.Port; promPort == nil {
			p.Logger.Error().Stack().Err(nil).Msg("Issue while reading prometheus port")
		} else {
			port = *promPort
		}
	}

	// sanity check on port
	if port == 0 {
		return errs.New(errs.ErrMissingParam, "port")
	} else if port < 0 {
		return errs.New(errs.ErrInvalidParam, "port")
	}

	// The optional parameter LocalHTTPAddr is the address of the HTTP service, valid values are:
	//- "localhost" or "127.0.0.1", this limits access to local machine
	//- "" (default) or "0.0.0.0", allows access from network
	addr := p.Params.LocalHTTPAddr
	if addr != "" {
		p.Logger.Debug().Str("addr", addr).Msg("Using custom local addr")
	}
	go p.startHTTPD(addr, port)

	// @TODO: implement error checking to enter failed state if HTTPd failed
	// (like we did in Alpha)

	p.Logger.Debug().Msgf("initialized, HTTP daemon started at [http://%s:%d]", addr, port)

	return nil
}

// Export - Unlike other Harvest exporters, we don't actually export data
// but put it in cache, for the HTTP daemon to serve on request
//
// An important aspect of the whole mechanism is that all incoming
// data should have a unique UUID and object pair, otherwise they'll
// overwrite other data in the cache.
// This key is also used by the HTTP daemon to trace back the name
// of the collectors and plugins where the metrics come from (for the info page)
func (p *Prometheus) Export(data *matrix.Matrix) error {

	var (
		metrics [][]byte
		err     error
	)

	// lock the exporter, to prevent other collectors from writing to us
	p.Lock()
	defer p.Unlock()

	p.Logger.Trace().Msgf("incoming %s%s(%s) (%s)%s", color.Bold, color.Cyan, data.UUID, data.Object, color.End)

	// render metrics into Prometheus format
	start := time.Now()
	if metrics, err = p.render(data); err != nil {
		return err
	}
	// fix render time for metadata
	d := time.Since(start)

	// simulate export in debug mode
	if p.Options.Debug {
		p.Logger.Debug().Msg("no export since in debug mode")
		for _, m := range metrics {
			p.Logger.Debug().Msgf("M= %s", string(m))
		}
		return nil
	}

	// store metrics in cache
	key := data.UUID + "." + data.Object + "." + data.Identifier

	// lock cache, to prevent HTTPd reading while we are mutating it
	p.cache.Lock()
	p.cache.Put(key, metrics)
	p.cache.Unlock()
	p.Logger.Trace().Msgf("added to cache with key [%s%s%s%s]", color.Bold, color.Red, key, color.End)

	// update metadata
	p.AddExportCount(uint64(len(metrics)))
	err = p.Metadata.LazyAddValueInt64("time", "render", d.Microseconds())
	if err != nil {
		p.Logger.Error().Stack().Err(err).Msg("error")
	}
	err = p.Metadata.LazyAddValueInt64("time", "export", time.Since(start).Microseconds())
	if err != nil {
		p.Logger.Error().Stack().Err(err).Msg("error")
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

func (p *Prometheus) render(data *matrix.Matrix) ([][]byte, error) {
	var (
		rendered                                     [][]byte
		tagged                                       *set.Set
		labelsToInclude, keysToInclude, globalLabels []string
		prefix                                       string
		err                                          error
	)

	rendered = make([][]byte, 0)
	globalLabels = make([]string, 0)

	if p.addMetaTags {
		tagged = set.New()
	}

	options := data.GetExportOptions()

	if x := options.GetChildS("instance_labels"); x != nil {
		labelsToInclude = x.GetAllChildContentS()
		p.Logger.Trace().Strs("requested instance_labels", labelsToInclude).Msg("")
	}

	if x := options.GetChildS("instance_keys"); x != nil {
		keysToInclude = x.GetAllChildContentS()
		p.Logger.Trace().Strs("requested keys_labels", keysToInclude).Msg("")
	}

	includeAllLabels := false
	requireInstanceKeys := true

	if x := options.GetChildContentS("include_all_labels"); x != "" {
		if includeAllLabels, err = strconv.ParseBool(x); err != nil {
			p.Logger.Error().Stack().Err(err).Msg("parameter: include_all_labels")
		}
	}

	if x := options.GetChildContentS("require_instance_keys"); x != "" {
		if requireInstanceKeys, err = strconv.ParseBool(x); err != nil {
			p.Logger.Error().Stack().Err(err).Msg("parameter: require_instance_keys")
		}
	}

	prefix = p.globalPrefix + data.Object

	for key, value := range data.GetGlobalLabels().Map() {
		globalLabels = append(globalLabels, fmt.Sprintf("%s=\"%s\"", key, value))
	}

	for key, instance := range data.GetInstances() {

		if !instance.IsExportable() {
			p.Logger.Trace().Msgf("skip instance [%s]: disabled for export", key)
			continue
		}

		p.Logger.Trace().Msgf("rendering instance [%s] (%v)", key, instance.GetLabels())

		instanceKeys := make([]string, len(globalLabels))
		copy(instanceKeys, globalLabels)
		instanceKeysOk := false
		instanceLabels := make([]string, 0)

		if includeAllLabels {
			for label, value := range instance.GetLabels().Map() {
				// temporary fix for the rarely happening duplicate labels
				// known case is: ZapiPerf -> 7mode -> disk.yaml
				// actual cause is the Aggregator plugin, which is adding node as
				// instance label (even though it's already a global label for 7modes)
				if !data.GetGlobalLabels().Has(label) {
					instanceKeys = append(instanceKeys, fmt.Sprintf("%s=\"%s\"", label, value)) //nolint:makezero
				}
			}
		} else {
			for _, key := range keysToInclude {
				value := instance.GetLabel(key)
				instanceKeys = append(instanceKeys, fmt.Sprintf("%s=\"%s\"", key, value)) //nolint:makezero
				if !instanceKeysOk && value != "" {
					instanceKeysOk = true
				}
				p.Logger.Trace().Msgf("++ key [%s] (%s) found=%v", key, value, value != "")
			}

			for _, label := range labelsToInclude {
				value := instance.GetLabel(label)
				instanceLabels = append(instanceLabels, fmt.Sprintf("%s=\"%s\"", label, value))
				p.Logger.Trace().Msgf("++ label [%s] (%s) %t", label, value, value != "")
			}

			// @TODO, probably be strict, and require all keys to be present
			if !instanceKeysOk && requireInstanceKeys {
				p.Logger.Trace().Msgf("skip instance, no keys parsed (%v) (%v)", instanceKeys, instanceLabels)
				continue
			}

			// @TODO, check at least one label is found?
			if len(instanceLabels) != 0 {
				var labelData string
				if p.Params.SortLabels {
					allLabels := make([]string, len(instanceLabels))
					copy(allLabels, instanceLabels)
					allLabels = append(allLabels, instanceKeys...) //nolint:makezero
					sort.Strings(allLabels)
					labelData = fmt.Sprintf("%s_labels{%s} 1.0", prefix, strings.Join(allLabels, ","))
				} else {
					labelData = fmt.Sprintf("%s_labels{%s,%s} 1.0", prefix, strings.Join(instanceKeys, ","), strings.Join(instanceLabels, ","))
				}
				if p.addMetaTags && !tagged.Has(prefix+"_labels") {
					tagged.Add(prefix + "_labels")
					rendered = append(rendered, []byte("# HELP "+prefix+"_labels Pseudo-metric for "+data.Object+" labels"))
					rendered = append(rendered, []byte("# TYPE "+prefix+"_labels gauge"))
				}
				rendered = append(rendered, []byte(labelData))
			} else {
				p.Logger.Trace().Msgf("skip instance labels, no labels parsed (%v) (%v)", instanceKeys, instanceLabels)
			}
		}

		if p.Params.SortLabels {
			sort.Strings(instanceKeys)
		}
		for mkey, metric := range data.GetMetrics() {

			if !metric.IsExportable() {
				p.Logger.Trace().Msgf("skip metric [%s]: disabled for export", mkey)
				continue
			}

			p.Logger.Trace().Msgf("rendering metric [%s]", mkey)

			if value, ok := metric.GetValueString(instance); ok {

				// metric is histogram
				if metric.HasLabels() {
					metricLabels := make([]string, 0)
					for k, v := range metric.GetLabels().Map() {
						metricLabels = append(metricLabels, fmt.Sprintf("%s=\"%s\"", k, v))
					}
					x := fmt.Sprintf("%s_%s{%s,%s} %s", prefix, metric.GetName(), strings.Join(instanceKeys, ","), strings.Join(metricLabels, ","), value)

					if p.addMetaTags && !tagged.Has(prefix+"_"+metric.GetName()) {
						tagged.Add(prefix + "_" + metric.GetName())
						rendered = append(rendered, []byte("# HELP "+prefix+"_"+metric.GetName()+" Metric for "+data.Object))
						rendered = append(rendered, []byte("# TYPE "+prefix+"_"+metric.GetName()+" histogram"))
					}

					rendered = append(rendered, []byte(x))
					// scalar metric
				} else {
					x := fmt.Sprintf("%s_%s{%s} %s", prefix, metric.GetName(), strings.Join(instanceKeys, ","), value)

					if p.addMetaTags && !tagged.Has(prefix+"_"+metric.GetName()) {
						tagged.Add(prefix + "_" + metric.GetName())
						rendered = append(rendered, []byte("# HELP "+prefix+"_"+metric.GetName()+" Metric for "+data.Object))
						rendered = append(rendered, []byte("# TYPE "+prefix+"_"+metric.GetName()+" gauge"))
					}

					rendered = append(rendered, []byte(x))
				}
			} else {
				p.Logger.Trace().Msg("skipped: no data value")
			}
		}
	}
	p.Logger.Debug().
		Str("object", data.Object).
		Int("rendered", len(rendered)).
		Int("instances", len(data.GetInstances())).
		Msg("Rendered data points for instances")
	return rendered, nil
}
