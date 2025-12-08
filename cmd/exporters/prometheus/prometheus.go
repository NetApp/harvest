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
	"github.com/netapp/harvest/v2/cmd/exporters"
	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Default parameters
const (
	// the maximum amount of time to keep metrics in the cache
	cacheMaxKeep = "5m"
	// apply a prefix to metrics globally (default none)
	globalPrefix = ""
)

type Prometheus struct {
	*exporter.AbstractExporter
	aCache          cacher
	allowAddrs      []string
	allowAddrsRegex []*regexp.Regexp
	cacheAddrs      map[string]bool
	checkAddrs      bool
	addMetaTags     bool
	globalPrefix    string
	replacer        *strings.Replacer
}

func New(abc *exporter.AbstractExporter) exporter.Exporter {
	return &Prometheus{AbstractExporter: abc}
}

func (p *Prometheus) createCacher(dur time.Duration) cacher {
	if p.Params.DiskCache != nil && p.Params.DiskCache.Path != "" {
		p.Logger.Debug("disk cache enabled - will use disk-based caching for RSS optimization",
			slog.String("path", p.Params.DiskCache.Path))

		cacheDir := p.Params.DiskCache.Path
		// Include poller name in cache directory to avoid collisions between multiple pollers
		if p.Options.Poller != "" {
			cacheDir = filepath.Join(cacheDir, p.Options.Poller)
		}
		return newDiskCache(dur, cacheDir, p.Logger)
	}

	return newMemCache(p.Logger, dur)
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

	p.replacer = exporters.NewReplacer()

	if instance, err := p.Metadata.NewInstance("info"); err == nil {
		instance.SetLabel("task", "info")
	} else {
		return err
	}

	if x := p.Params.GlobalPrefix; x != nil {
		p.Logger.Debug("use global prefix", slog.String("prefix", *x))
		p.globalPrefix = *x
		if !strings.HasSuffix(p.globalPrefix, "_") {
			p.globalPrefix += "_"
		}
	} else {
		p.globalPrefix = globalPrefix
	}

	// add HELP and TYPE tags to exported metrics if requested
	if p.Params.ShouldAddMetaTags != nil && *p.Params.ShouldAddMetaTags {
		p.addMetaTags = true
	}

	maxKeep := cacheMaxKeep
	var maxKeepDur time.Duration
	if x := p.Params.CacheMaxKeep; x != nil {
		maxKeep = *x
		p.Logger.Debug("using custom cache_max_keep", slog.String("cacheMaxKeep", maxKeep))
	}
	d, err := time.ParseDuration(maxKeep)
	if err != nil {
		p.Logger.Error("failed to use cache_max_keep duration. Using default", slogx.Err(err),
			slog.String("maxKeep", maxKeep),
			slog.String("default", cacheMaxKeep),
		)
		maxKeepDur, _ = time.ParseDuration(cacheMaxKeep)
	} else {
		maxKeepDur = d
	}

	p.aCache = p.createCacher(maxKeepDur)
	if !p.aCache.isValid() {
		return errs.New(errs.ErrInvalidParam, "cache initialization failed")
	}

	// allow access to metrics only from the given plain addresses
	if x := p.Params.AllowedAddrs; x != nil {
		p.allowAddrs = *x
		if len(p.allowAddrs) == 0 {
			p.Logger.Error("allow_addrs without any")
			return errs.New(errs.ErrInvalidParam, "allow_addrs")
		}
		p.checkAddrs = true
		p.Logger.Debug("added plain allow rules", slog.Int("count", len(p.allowAddrs)))
	}

	// allow access only from addresses matching one of defined regular expressions
	if x := p.Params.AllowedAddrsRegex; x != nil {
		p.allowAddrsRegex = make([]*regexp.Regexp, 0)
		for _, r := range *x {
			r = strings.TrimPrefix(strings.TrimSuffix(r, "`"), "`")
			if reg, err := regexp.Compile(r); err == nil {
				p.allowAddrsRegex = append(p.allowAddrsRegex, reg)
			} else {
				p.Logger.Error("parse regex", slogx.Err(err))
				return errs.New(errs.ErrInvalidParam, "allow_addrs_regex")
			}
		}
		if len(p.allowAddrsRegex) == 0 {
			p.Logger.Error("allow_addrs_regex without any")
			return errs.New(errs.ErrInvalidParam, "allow_addrs")
		}
		p.checkAddrs = true
		p.Logger.Debug("added regex allow rules", slog.Int("count", len(p.allowAddrsRegex)))
	}

	// cache addresses that have been allowed or denied already
	if p.checkAddrs {
		p.cacheAddrs = make(map[string]bool)
	}

	// Finally, the most important and only required parameter: port
	// can be passed to us either as an option or as a parameter
	port := p.Options.PromPort
	if port == 0 {
		if promPort := p.Params.Port; promPort == nil {
			p.Logger.Error("missing Prometheus port")
		} else {
			port = *promPort
		}
	}

	// Make sure port is valid
	if port == 0 {
		return errs.New(errs.ErrMissingParam, "port")
	} else if port < 0 {
		return errs.New(errs.ErrInvalidParam, "port")
	}

	// The optional parameter LocalHTTPAddr is the address of the HTTP service, valid values are:
	// - "localhost" or "127.0.0.1", this limits access to local machine
	// - "" (default) or "0.0.0.0", allows access from network
	addr := p.Params.LocalHTTPAddr
	if addr != "" {
		p.Logger.Debug("using custom local addr", slog.String("addr", addr))
	}

	if !p.Params.IsTest {
		go p.startHTTPD(addr, port)
	}

	// @TODO: implement error checking to enter failed state if HTTPd failed
	// (like we did in Alpha)

	p.Logger.Debug("initialized HTTP daemon started", slog.String("addr", addr), slog.Int("port", port))

	return nil
}

// Export - Unlike other Harvest exporters, we don't export data
// but put it in cache. The HTTP daemon serves that cache on request.
//
// An important aspect of the whole mechanism is that all incoming
// data should have a unique UUID and object pair, otherwise they'll
// overwrite other data in the cache.
// This key is also used by the HTTP daemon to trace back the name
// of the collectors and plugins where the metrics come from (for the info page)
func (p *Prometheus) Export(data *matrix.Matrix) (exporter.Stats, error) {

	var (
		metrics     [][]byte
		stats       exporter.Stats
		err         error
		metricNames *set.Set
	)

	// lock the exporter, to prevent other collectors from writing to us
	p.Lock()
	defer p.Unlock()

	// render metrics into Prometheus format
	start := time.Now()
	metrics, stats, metricNames = exporters.Render(data, p.addMetaTags, p.Params.SortLabels, p.globalPrefix, p.Logger, "")

	// fix render time for metadata
	d := time.Since(start)

	// store metrics in cache
	key := data.UUID + "." + data.Object + "." + data.Identifier

	p.aCache.exportMetrics(key, metrics, metricNames)

	// update metadata
	p.AddExportCount(uint64(len(metrics)))
	err = p.Metadata.LazyAddValueInt64("time", "render", d.Microseconds())
	if err != nil {
		p.Logger.Error("error", slogx.Err(err))
	}
	err = p.Metadata.LazyAddValueInt64("time", "export", time.Since(start).Microseconds())
	if err != nil {
		p.Logger.Error("error", slogx.Err(err))
	}

	return stats, nil
}
