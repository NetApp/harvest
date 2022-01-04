package rest

import (
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
	"strings"
)

func (r *Rest) initCache() error {

	var (
		counters            *node.Node
		display, name, kind string
		metr                matrix.Metric
		err                 error
	)

	if x := r.Params.GetChildContentS("object"); x != "" {
		r.Object = x
	} else {
		r.Object = strings.ToLower(r.Object)
	}
	r.Matrix.Object = r.Object

	if e := r.Params.GetChildS("export_options"); e != nil {
		r.Matrix.SetExportOptions(e)
	}

	if r.apiPath = r.Params.GetChildContentS("query"); r.apiPath == "" {
		return errors.New(errors.MISSING_PARAM, "query")
	}

	// create metric cache
	if counters = r.Params.GetChildS("counters"); counters == nil {
		return errors.New(errors.MISSING_PARAM, "counters")
	}

	// default value for ONTAP is 15 sec
	if returnTimeout := r.Params.GetChildContentS("return_timeout"); returnTimeout != "" {
		r.returnTimeOut = returnTimeout
	}

	r.instanceKeys = make([]string, 0)
	r.instanceLabels = make(map[string]string)
	r.counters = make(map[string]string)

	for _, c := range counters.GetAllChildContentS() {
		name, display, kind = util.ParseMetric(c)
		r.Logger.Debug().
			Str("kind", kind).
			Str("name", name).
			Str("display", display).
			Msg("Collected")

		r.counters[name] = display
		switch kind {
		case "key":
			r.instanceLabels[name] = display
			r.instanceKeys = append(r.instanceKeys, name)
		case "label":
			r.instanceLabels[name] = display
		case "bool":
			if metr, err = r.Matrix.NewMetricUint8(name); err != nil {
				r.Logger.Error().Err(err).
					Str("name", name).
					Msg("NewMetricUint8")
				return err
			}
			metr.SetName(display)
			metr.SetProperty("etl.bool") // to distinct from internally generated metrics, e.g. from plugins
		case "float":
			if metr, err = r.Matrix.NewMetricFloat64(name); err != nil {
				r.Logger.Error().Err(err).
					Str("name", name).
					Msg("NewMetricFloat64")
				return err
			}
			metr.SetName(display)
			metr.SetProperty("etl.float")
		}
	}

	r.Logger.Info().Strs("extracted Instance Keys", r.instanceKeys).Msg("")
	r.Logger.Info().Int("count metrics", len(r.Matrix.GetMetrics())).Int("count labels", len(r.instanceLabels)).Msg("initialized metric cache")

	if len(r.Matrix.GetMetrics()) == 0 && r.Params.GetChildContentS("collect_only_labels") != "true" {
		return errors.New(errors.ERR_NO_METRIC, "failed to parse numeric metrics")
	}
	return nil

}
