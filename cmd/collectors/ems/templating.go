package ems

import (
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
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
	for _, v := range defaultInstanceKey {
		prop.InstanceKeys = append(prop.InstanceKeys, v)
	}

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
}

func (e *Ems) ParseExports(counter *node.Node, prop *emsProp) {
	var (
		display, name string
	)

	for _, c := range counter.GetAllChildContentS() {
		if c != "" {
			name, display, _, _ = util.ParseMetric(c)
			e.Logger.Debug().
				Str("name", name).
				Str("display", display).
				Msg("Collected exports")

			// EMS only supports labels
			prop.InstanceLabels[name] = display
		}
	}
}
