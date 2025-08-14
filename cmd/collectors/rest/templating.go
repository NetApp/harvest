package rest

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/template"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"strings"
)

func (r *Rest) LoadTemplate() (string, error) {

	jitter := r.Params.GetChildContentS("jitter")
	subTemplate, path, err := r.ImportSubTemplate("", TemplateFn(r.Params, r.Object), jitter, r.Remote.Version)
	if err != nil {
		return "", err
	}

	r.Params.Union(subTemplate)
	return path, nil
}

func (r *Rest) InitCache() error {

	var (
		counters *node.Node
	)

	if x := r.Params.GetChildContentS("object"); x != "" {
		r.Prop.Object = x
	} else {
		r.Prop.Object = strings.ToLower(r.Object)
	}

	if shouldIgnore := r.Params.GetChildContentS("ignore"); shouldIgnore == "true" {
		return nil
	}

	if e := r.Params.GetChildS("export_options"); e != nil {
		r.Matrix[r.Object].SetExportOptions(e)
	}

	if r.Prop.Query = r.Params.GetChildContentS("query"); r.Prop.Query == "" {
		return errs.New(errs.ErrMissingParam, "query")
	}

	// create metric cache
	if counters = r.Params.GetChildS("counters"); counters == nil {
		return errs.New(errs.ErrMissingParam, "counters")
	}

	// default value for ONTAP is 15 sec
	if returnTimeout := r.Params.GetChildContentS("return_timeout"); returnTimeout != "" {
		iReturnTimeout, err := strconv.Atoi(returnTimeout)
		if err != nil {
			r.Logger.Warn("Invalid value of returnTimeout", slog.String("returnTimeout", returnTimeout))
		} else {
			r.Prop.ReturnTimeOut = &iReturnTimeout
		}
	}

	if b := r.Params.GetChildContentS("batch_size"); b != "" {
		if _, err := strconv.Atoi(b); err == nil {
			r.BatchSize = b
		}
	}
	if r.BatchSize == "" && r.Params.GetChildContentS("no_max_records") != "true" {
		r.BatchSize = collectors.DefaultBatchSize
	}

	allowPartialAggregation := r.Params.GetChildContentS("allow_partial_aggregation")
	if allowPartialAggregation == "true" {
		r.AllowPartialAggregation = true
	}

	// Private end points do not support * as fields. We need to pass fields in endpoint
	query := r.Params.GetChildS("query")
	r.Prop.IsPublic = true
	if query != nil {
		r.Prop.IsPublic = requests.IsPublicAPI(query.GetContentS())
	}

	r.ParseRestCounters(counters, r.Prop)

	r.Logger.Debug(
		"Initialized metric cache",
		slog.Any("extracted Instance Keys", r.Prop.InstanceKeys),
		slog.Int("numMetrics", len(r.Prop.Metrics)),
		slog.Int("numLabels", len(r.Prop.InstanceLabels)),
	)

	return nil
}

func (r *Rest) ParseRestCounters(counter *node.Node, prop *prop) {
	var (
		display, name, kind, metricType string
	)

	instanceKeys := make(map[string]string)

	for _, c := range counter.GetAllChildContentS() {
		if c != "" {
			name, display, kind, metricType = template.ParseMetric(c)
			prop.Counters[name] = display
			switch kind {
			case "key":
				prop.InstanceLabels[name] = display
				instanceKeys[display] = name
			case "label":
				prop.InstanceLabels[name] = display
			case "float":
				m := &Metric{Label: display, Name: name, MetricType: metricType, Exportable: true}
				prop.Metrics[name] = m
			}
		}
	}

	// populate prop.instanceKeys
	// sort keys by display name. This is needed to match counter and endpoints keys
	keys := slices.Sorted(maps.Keys(instanceKeys))

	// Append instance keys to prop
	for _, k := range keys {
		prop.InstanceKeys = append(prop.InstanceKeys, instanceKeys[k])
	}

	counterKey := make([]string, len(prop.Counters))
	i := 0
	for k := range prop.Counters {
		counterKey[i] = k
		i++
	}
	prop.Fields = counterKey
	if counter != nil {
		if x := counter.GetChildS("filter"); x != nil {
			prop.Filter = append(prop.Filter, x.GetAllChildContentS()...)
		}
	}

	if prop.IsPublic {
		if counter != nil {
			if x := counter.GetChildS("hidden_fields"); x != nil {
				prop.HiddenFields = append(prop.HiddenFields, x.GetAllChildContentS()...)
			}
		}
	}

}
