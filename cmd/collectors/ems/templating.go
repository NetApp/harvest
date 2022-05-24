package ems

import (
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"strings"
)

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
		display, name, kind string
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
			name, display, kind = ParseEmsMetric(c)
			e.Logger.Debug().
				Str("kind", kind).
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

func ParseEmsMetric(rawName string) (string, string, string) {
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
		display = strings.ReplaceAll(display, "-", "_")
	}

	return name, display, "label"
}

//LoadEmsPlugins loads built-in plugins or dynamically loads custom plugins
//and adds them to the collector
func (e *Ems) LoadEmsPlugins(params *node.Node) ([]plugin.Plugin, error) {

	var p plugin.Plugin
	var abc *plugin.AbstractPlugin
	var plugins []plugin.Plugin

	for _, x := range params.GetChildren() {

		name := x.GetNameS()
		if name == "" {
			name = x.GetContentS() // some plugins are defined as list elements others as dicts
			x.SetNameS(name)
		}

		abc = plugin.New(e.Name, e.Options, x, e.Params, e.Object)

		// case 1: available as built-in plugin
		if p = collector.GetBuiltinPlugin(name, abc); p != nil {
			e.Logger.Debug().Msgf("loaded built-in plugin [%s]", name)
			// case 2: available as dynamic plugin
		} else {
			p = e.LoadPlugin(name, abc)
			e.Logger.Debug().Msgf("loaded plugin [%s]", name)
		}
		if p == nil {
			continue
		}

		if err := p.Init(); err != nil {
			e.Logger.Error().Stack().Err(err).Msgf("init plugin [%s]:", name)
			return plugins, err
		}
		plugins = append(plugins, p)
	}
	e.Logger.Debug().Msgf("initialized %d plugins", len(e.Plugins))
	return plugins, nil
}
