package ems

import (
	"github.com/netapp/harvest/v2/pkg/set"
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
	)
	bookendKeys := make(map[string]string)

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
				bookendKeys[display] = name
				e.Logger.Debug().
					Str("name", name).
					Str("display", display).
					Msg("Collected bookend keys")
			}
		}
	}

	// For bookend case, instanceKeys are replaced with bookendKeys
	if len(bookendKeys) > 0 {
		sortedBookendKeys := util.GetSortedKeys(bookendKeys)
		// Append instance keys to ems prop
		for _, k := range sortedBookendKeys {
			prop.InstanceKeys = append(prop.InstanceKeys, bookendKeys[k])
		}
	}
}

func (e *Ems) ParseResolveEms(resolveEvent *node.Node, issueEmsProp emsProp) {
	var resolveEmsName, resolveAfter string
	var resolveKey *node.Node

	prop := emsProp{}
	prop.InstanceLabels = make(map[string]string)
	prop.InstanceKeys = make([]string, 0)

	// check if resolvedby is present in template
	if resolveEmsName = resolveEvent.GetChildContentS("name"); resolveEmsName == "" {
		e.Logger.Error().Msg("Missing resolving event name")
		return
	}
	prop.Name = resolveEmsName

	// populate prop counter for asup
	e.Prop.Counters[resolveEmsName] = resolveEmsName

	// check if resolved_key is present in template, if not then use the issue ems resolve key
	if resolveKey = resolveEvent.GetChildS("resolve_key"); resolveKey == nil {
		// IssuingEmsKey: index-messageName-bookendKey, ResolvingEmsKey would be bookendKey
		if len(issueEmsProp.InstanceKeys) > 2 {
			prop.InstanceKeys = issueEmsProp.InstanceKeys[2:]
		} else {
			// If bookendKey is missing in IssueEms, the  default bookendKey is index of IssueEMs
			prop.InstanceKeys = issueEmsProp.InstanceKeys[0:1]
			e.Logger.Error().Str("Ems name", issueEmsProp.Name).Msg("Missing bookend keys")
		}
	} else {
		e.ParseExports(resolveKey, &prop)
	}

	// check if resolveAfter is present in template
	e.resolveAfter[issueEmsProp.Name] = DefaultBookendResolutionDuration
	if resolveAfter = resolveEvent.GetChildContentS("resolve_after"); resolveAfter != "" {
		if durationVal, err := time.ParseDuration(resolveAfter); err == nil {
			e.resolveAfter[issueEmsProp.Name] = durationVal
		}
	}
	e.Logger.Debug().Str("bookend ems resolve After", e.resolveAfter[issueEmsProp.Name].String()).Msg("")

	// Using Set to ensure it has slice of unique issuing ems
	if _, ok := e.bookendEmsMap[resolveEmsName]; !ok {
		e.bookendEmsMap[resolveEmsName] = set.New()
	}
	e.bookendEmsMap[resolveEmsName].Add(issueEmsProp.Name)
	e.emsProp[resolveEmsName] = append(e.emsProp[resolveEmsName], &prop)
}
