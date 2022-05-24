package ems

import (
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
)

// default label set collected for each ems
var defaultLabels = map[string]string{
	"message.name":     "message",
	"node.name":        "node",
	"message.severity": "severity",
	"time":             "time",
	"index":            "index",
}

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
			e.Logger.Warn().Str("name", prop.Name).Msg("Missing matches for ems")
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
			e.Logger.Warn().Str("name", prop.Name).Msg("Missing labels for ems")
			continue
		}
	}
}

func (e *Ems) ParseRestCounters(counter *node.Node, prop *emsProp) {
	var (
		display, name string
	)

	//load default ems labels
	for k, v := range defaultLabels {
		prop.InstanceLabels[k] = v
	}

	//load default instance keys
	for _, v := range defaultInstanceKey {
		prop.InstanceKeys = append(prop.InstanceKeys, v)
	}

	for _, c := range counter.GetAllChildContentS() {
		if c != "" {
			name, display, _, _ = util.ParseMetric(c)
			e.Logger.Debug().
				Str("name", name).
				Str("display", display).
				Msg("Collected")

			// EMS only supports labels
			prop.InstanceLabels[name] = display
		}
	}
	// add a placeholder metric for ems
	m := &Metric{Label: "events", Name: "events", MetricType: "", Exportable: true}
	prop.Metrics["events"] = m
}
