package rest

import (
	"fmt"
	"goharvest2/pkg/errors"
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
		counters            *node.Node
		display, name, kind string
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

	for _, c := range counters.GetAllChildContentS() {
		if c != "" {
			mType := ""
			name, display, kind = util.ParseMetric(c)
			r.Logger.Debug().
				Str("kind", kind).
				Str("name", name).
				Str("display", display).
				Msg("Collected")

			if strings.Contains(name, "(") {
				metricName := strings.Split(name, "(")
				name = metricName[0]
				mType = strings.TrimRight(metricName[1], ")")
			}

			r.prop.counters[name] = display
			switch kind {
			case "key":
				r.prop.instanceLabels[name] = display
				r.prop.instanceKeys = append(r.prop.instanceKeys, name)
			case "label":
				r.prop.instanceLabels[name] = display
			case "float":
				m := metric{label: display, name: name, metricType: mType}
				r.prop.metrics = append(r.prop.metrics, m)
			}
		}
	}

	// private end point do not support * as fields. We need to pass fields in endpoint
	query := r.Params.GetChildS("query")
	r.prop.apiType = "public"
	if query != nil {
		r.prop.apiType = checkQueryType(query.GetContentS())
	}

	if r.prop.apiType == "private" {
		counterKey := make([]string, len(r.prop.counters))
		i := 0
		for k := range r.prop.counters {
			counterKey[i] = k
			i++
		}
		r.prop.fields = counterKey
	}

	if r.prop.apiType == "public" {
		r.prop.fields = []string{"*"}
		if c := r.Params.GetChildS("counters"); c != nil {
			if x := c.GetChildS("hidden_fields"); x != nil {
				r.prop.fields = append(r.prop.fields, x.GetAllChildContentS()...)
			}
		}
	}

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
