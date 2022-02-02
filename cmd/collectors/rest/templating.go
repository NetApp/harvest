package rest

import (
	"fmt"
	"github.com/tidwall/gjson"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (r *Rest) LoadTemplate() error {

	var (
		template *node.Node
		err      error
	)

	// import template
	if template, err = r.ImportSubTemplate("", r.getTemplateFn(), r.client.Cluster().Version); err != nil {
		return err
	}

	r.Params.Union(template)
	return nil
}

func (r *Rest) initCache() error {

	var (
		counters *node.Node
	)

	if x := r.Params.GetChildContentS("object"); x != "" {
		r.prop.object = x
	} else {
		r.prop.object = strings.ToLower(r.Object)
	}

	if e := r.Params.GetChildS("export_options"); e != nil {
		r.Matrix.SetExportOptions(e)
	}

	if r.prop.query = r.Params.GetChildContentS("query"); r.prop.query == "" {
		return errors.New(errors.MISSING_PARAM, "query")
	}

	// create metric cache
	if counters = r.Params.GetChildS("counters"); counters == nil {
		return errors.New(errors.MISSING_PARAM, "counters")
	}

	// default value for ONTAP is 15 sec
	if returnTimeout := r.Params.GetChildContentS("return_timeout"); returnTimeout != "" {
		r.prop.returnTimeOut = returnTimeout
	}

	r.prop.instanceKeys = make([]string, 0)
	r.prop.instanceLabels = make(map[string]string)
	r.prop.counters = make(map[string]string)

	// private end point do not support * as fields. We need to pass fields in endpoint
	query := r.Params.GetChildS("query")
	r.prop.apiType = "public"
	if query != nil {
		r.prop.apiType = checkQueryType(query.GetContentS())
	}

	r.ParseRestCounters(counters, r.prop)

	r.Logger.Info().Strs("extracted Instance Keys", r.prop.instanceKeys).Msg("")
	r.Logger.Info().Int("count metrics", len(r.prop.metrics)).Int("count labels", len(r.prop.instanceLabels)).Msg("initialized metric cache")

	return nil
}

func HandleDuration(value string) float64 {
	// Example: duration: PT8H35M42S
	timeDurationRegex := `^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:.\d+)?)S)?$`

	regexTimeDuration := regexp.MustCompile(timeDurationRegex)
	if match := regexTimeDuration.MatchString(value); match {
		// example: PT8H35M42S   ==>  30942
		matches := regexTimeDuration.FindStringSubmatch(value)
		if matches == nil {
			return 0
		}

		seconds := 0.0

		//years
		//months

		//days
		if matches[3] != "" {
			f, err := strconv.ParseFloat(matches[3], 64)
			if err != nil {
				fmt.Printf("%v", err)
				return 0
			}
			seconds += f * 24 * 60 * 60
		}

		//hours
		if matches[4] != "" {
			f, err := strconv.ParseFloat(matches[4], 64)
			if err != nil {
				fmt.Printf("%v", err)
				return 0
			}
			seconds += f * 60 * 60
		}

		//minutes
		if matches[5] != "" {
			f, err := strconv.ParseFloat(matches[5], 64)
			if err != nil {
				fmt.Printf("%v", err)
				return 0
			}
			seconds += f * 60
		}

		//seconds & milliseconds
		if matches[6] != "" {
			f, err := strconv.ParseFloat(matches[6], 64)
			if err != nil {
				fmt.Printf("%v", err)
				return 0
			}
			seconds += f
		}
		return seconds
	}

	return 0
}

func HandleTimestamp(value string) float64 {
	var timestamp time.Time
	var err error

	// Example: timestamp: 2020-12-02T18:36:19-08:00
	timestampRegex := `[+-]?\d{4}(-[01]\d(-[0-3]\d(T[0-2]\d:[0-5]\d:?([0-5]\d(\.\d+)?)?[+-][0-2]\d:[0-5]\d?)?)?)?`

	regexTimeStamp := regexp.MustCompile(timestampRegex)
	if match := regexTimeStamp.MatchString(value); match {
		// example: 2020-12-02T18:36:19-08:00   ==>  1606962979
		if timestamp, err = time.Parse(time.RFC3339, value); err != nil {
			fmt.Printf("%v", err)
			return 0
		}
		return float64(timestamp.Unix())
	}
	return 0
}

func (r *Rest) ParseRestCounters(counter *node.Node, prop *prop) {
	var (
		display, name, kind, metricType string
	)

	for _, c := range counter.GetAllChildContentS() {
		if c != "" {
			name, display, kind, metricType = util.ParseMetric(c)
			r.Logger.Debug().
				Str("kind", kind).
				Str("name", name).
				Str("display", display).
				Msg("Collected")

			prop.counters[name] = display
			switch kind {
			case "key":
				prop.instanceLabels[name] = display
				prop.instanceKeys = append(prop.instanceKeys, name)
			case "label":
				prop.instanceLabels[name] = display
			case "float":
				m := metric{label: display, name: name, metricType: metricType}
				prop.metrics = append(prop.metrics, m)
			}
		}
	}

	if prop.apiType == "private" {
		counterKey := make([]string, len(prop.counters))
		i := 0
		for k := range prop.counters {
			counterKey[i] = k
			i++
		}
		prop.fields = counterKey
	}

	if prop.apiType == "public" {
		prop.fields = []string{"*"}
		if counter != nil {
			if x := counter.GetChildS("hidden_fields"); x != nil {
				prop.fields = append(prop.fields, x.GetAllChildContentS()...)
			}
		}
	}

}

func (r *Rest) HandleLabelsAndMetrics(instance *matrix.Instance, prop *prop, instanceData gjson.Result, data *matrix.Matrix, instanceKey string) uint64 {
	var (
		err   error
		count uint64
	)

	for label, display := range prop.instanceLabels {
		value := instanceData.Get(label)
		if value.Exists() {
			if value.IsArray() {
				var labelArray []string
				for _, r := range value.Array() {
					labelString := r.String()
					labelArray = append(labelArray, labelString)
				}
				instance.SetLabel(display, strings.Join(labelArray, ","))
			} else {
				instance.SetLabel(display, value.String())
			}
			count++
		} else {
			// spams a lot currently due to missing label mappings. Moved to debug for now till rest gaps are filled
			r.Logger.Debug().Str("Instance key", instanceKey).Str("label", label).Msg("Missing label value")
		}
	}

	for _, metric := range prop.metrics {
		metr, ok := data.GetMetrics()[metric.name]
		if !ok {
			if metr, err = data.NewMetricFloat64(metric.name); err != nil {
				r.Logger.Error().Err(err).
					Str("name", metric.name).
					Msg("NewMetricFloat64")
			}
		}
		f := instanceData.Get(metric.name)
		if f.Exists() {
			metr.SetName(metric.label)

			var floatValue float64
			switch metric.metricType {
			case "duration":
				floatValue = HandleDuration(f.String())
			case "timestamp":
				floatValue = HandleTimestamp(f.String())
			case "":
				floatValue = f.Float()
			default:
				r.Logger.Warn().Str("type", metric.metricType).Str("metric", metric.name).Msg("unknown metric type")
			}

			if err = metr.SetValueFloat64(instance, floatValue); err != nil {
				r.Logger.Error().Err(err).Str("key", metric.name).Str("metric", metric.label).
					Msg("Unable to set float key on metric")
			}
			count++
		}
	}
	return count
}

func (r *Rest) GetRestData(prop *prop, client *rest.Client) ([]interface{}, error) {
	var (
		err     error
		records []interface{}
	)

	href := rest.BuildHref(prop.query, strings.Join(prop.fields, ","), nil, "", "", "", prop.returnTimeOut, prop.query)

	r.Logger.Debug().Str("href", href).Msg("")
	if href == "" {
		return nil, errors.New(errors.ERR_CONFIG, "empty url")
	}

	err = rest.FetchData(client, href, &records)
	if err != nil {
		r.Logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
		return nil, err
	}

	return records, nil
}
