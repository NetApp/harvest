package rest

import (
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"strings"
)

func (r *Rest) initCache(templateFn string, version [3]int) error {

	var (
		template, counters  *node.Node
		display, name, kind string
		metr                matrix.Metric
		err                 error
	)

	// import template
	if template, err = r.ImportSubTemplate("", templateFn, version); err != nil {
		return err
	}

	r.Logger.Info().Msg("imported subtemplate")
	r.Params.Union(template)

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
		name, display, kind = parseMetric(c)
		r.Logger.Debug().Msgf("extracted [%s] (%s) (%s)", kind, name, display)
		r.counters[name] = display
		switch kind {
		case "key":
			r.instanceLabels[name] = display
			r.instanceKeys = append(r.instanceKeys, name)
		case "label":
			r.instanceLabels[name] = display
		case "bool":
			if metr, err = r.Matrix.NewMetricUint8(name); err != nil {
				r.Logger.Error().Msgf("NewMetricUint8 [%s]: %v", name, err)
				return err
			}
			metr.SetName(display)
			metr.SetProperty("etl.bool") // to distinct from internally generated metrics, e.g. from plugins
		case "float":
			if metr, err = r.Matrix.NewMetricFloat64(name); err != nil {
				r.Logger.Error().Msgf("NewMetricFloat64 [%s]: %v", name, err)
				return err
			}
			metr.SetName(display)
			metr.SetProperty("etl.float")
		}
	}

	r.Logger.Info().Msgf("extracted instance keys: %v", r.instanceKeys)
	r.Logger.Info().Msgf("initialized metric cache with %d metrics and %d labels", len(r.Matrix.GetMetrics()), len(r.instanceLabels))

	if len(r.Matrix.GetMetrics()) == 0 && r.Params.GetChildContentS("collect_only_labels") != "true" {
		return errors.New(errors.ERR_NO_METRIC, "failed to parse numeric metrics")
	}
	return nil

}

func parseMetric(rawName string) (string, string, string) {
	var (
		name, display string
		values        []string
	)
	if values = strings.SplitN(rawName, "=>", 2); len(values) == 2 {
		name = strings.TrimSpace(values[0])
		display = strings.TrimSpace(values[1])
	} else {
		name = rawName
		display = strings.ReplaceAll(rawName, ".", "_")
	}

	if strings.HasPrefix(name, "^^") {
		return strings.TrimPrefix(name, "^^"), strings.TrimPrefix(display, "^^"), "key"
	}

	if strings.HasPrefix(name, "^") {
		return strings.TrimPrefix(name, "^"), strings.TrimPrefix(display, "^"), "label"
	}

	if strings.HasPrefix(name, "?") {
		return strings.TrimPrefix(name, "?"), strings.TrimPrefix(display, "?"), "bool"
	}

	return name, display, "float"
}
