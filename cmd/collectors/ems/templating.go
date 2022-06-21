package ems

import (
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"time"
)

var defaultInstanceKey = []string{"index", "message.name"}

func (e *Ems) ParseMatches(matches *node.Node, prop *emsProp) {
	for _, v := range matches.GetChildren() {
		name := v.GetChildContentS("name")
		value := v.GetChildContentS("value")
		if name != "" && value != "" {
			prop.Matches = append(prop.Matches, &Matches{
				Name:  name,
				value: value,
			})
		} else {
			e.Logger.Warn().Str("name", name).Str("value", value).Msg("Match name and value cannot be empty")
			continue
		}
	}
}

func (e *Ems) ParseLabels(labels *node.Node, prop *emsProp) {
	prop.Labels = make(map[string]string)
	for _, v := range labels.GetChildren() {
		labelName := v.GetNameS()
		labelValue := v.GetContentS()
		if labelName != "" && labelValue != "" {
			prop.Labels[labelName] = labelValue
		} else {
			e.Logger.Warn().
				Str("ems", prop.Name).
				Str("labelName", labelName).
				Str("labelValue", labelValue).
				Msg("Label name and value cannot be empty for ems")
			continue
		}
	}
}

func (e *Ems) ParseDefaults(prop *emsProp) {
	var (
		display, name string
	)

	//load default instance keys
	prop.InstanceKeys = append(prop.InstanceKeys, defaultInstanceKey...)

	//process default labels
	for _, c := range e.DefaultLabels {
		if c != "" {
			name, display, _, _ = util.ParseMetric(c)
			e.Logger.Debug().
				Str("name", name).
				Str("display", display).
				Msg("Collected default labels")

			// EMS only supports labels
			prop.InstanceLabels[name] = display
		}
	}
	// add a default placeholder metric for ems
	m := &Metric{Label: "events", Name: "events", MetricType: "", Exportable: true}
	prop.Metrics["events"] = m
	// add an artificial timestamp metric for ems instances
	// used for remove the instances from cache after pre-defined time duration
	timeStampMetric := &Metric{Label: "timestamp", Name: "timestamp", MetricType: "", Exportable: false}
	prop.Metrics["timestamp"] = timeStampMetric
}

func (e *Ems) ParseExports(counter *node.Node, prop *emsProp) {
	var (
		display, name, key string
		bookendKeys        []string
	)

	for _, c := range counter.GetAllChildContentS() {
		if c != "" {
			name, display, key, _ = util.ParseMetric(c)
			e.Logger.Debug().
				Str("name", name).
				Str("display", display).
				Msg("Collected exports")

			// EMS only supports labels
			prop.InstanceLabels[name] = display

			if key == "key" {
				// only for bookend EMS
				bookendKeys = append(bookendKeys, name)
				e.Logger.Debug().
					Str("name", name).
					Str("display", display).
					Msg("Collected bookend keys")
			}
		}
	}

	// For bookend case, instanceKeys are replaced with bookendKeys
	if len(bookendKeys) > 0 {
		prop.InstanceKeys = bookendKeys
	}
}

func (e *Ems) ParseResolveEms(resolveEvent *node.Node, emsName string) {
	var resolveEmsName, resolveAfter string
	var resolveKey *node.Node

	prop := emsProp{}
	prop.InstanceLabels = make(map[string]string)
	prop.InstanceKeys = make([]string, 0)

	// check if resolvedby is present in template
	if resolveEmsName = resolveEvent.GetChildContentS("name"); resolveEmsName == "" {
		e.Logger.Warn().Msg("Missing resolving event name")
		return
	}
	prop.Name = resolveEmsName

	// populate prop counter for asup
	e.Prop.Counters[resolveEmsName] = resolveEmsName

	// check if resolvedby is present in template
	if resolveKey = resolveEvent.GetChildS("resolve_key"); resolveKey == nil {
		e.Logger.Warn().Msg("Missing resolving event exports")
		return
	}

	// check if resolveAfter is present in template
	e.resolveAfter = DefaultBookendResolutionDuration
	if resolveAfter = resolveEvent.GetChildContentS("resolve_bookend_after"); resolveAfter != "" {
		if durationVal, err := time.ParseDuration(resolveAfter); err == nil {
			e.resolveAfter = durationVal
		}
	}
	e.Logger.Debug().Str("bookend ems resolve After", e.resolveAfter.String()).Msg("")

	e.ParseExports(resolveKey, &prop)
	e.bookendEmsMap[resolveEmsName] = append(e.bookendEmsMap[resolveEmsName], emsName)
	e.emsProp[resolveEmsName] = append(e.emsProp[resolveEmsName], &prop)
}

func (e *Ems) InstanceOlderThan(metr matrix.Metric, instance *matrix.Instance) bool {
	if metricTimestamp, ok := metr.GetValueFloat64(instance); ok {
		if time.Since(time.UnixMicro(int64(metricTimestamp))).Hours() > e.resolveAfter.Hours() {
			return true
		}
	}
	return false
}
