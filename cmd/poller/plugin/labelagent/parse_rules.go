/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package labelagent

import (
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
)

// parse rules from plugin parameters and return number of rules parsed
func (a *LabelAgent) parseRules() int {

	a.splitSimpleRules = make([]splitSimpleRule, 0)
	a.splitRegexRules = make([]splitRegexRule, 0)
	a.joinSimpleRules = make([]joinSimpleRule, 0)
	a.replaceSimpleRules = make([]replaceSimpleRule, 0)
	a.replaceRegexRules = make([]replaceRegexRule, 0)
	a.excludeEqualsRules = make([]excludeEqualsRule, 0)
	a.excludeContainsRules = make([]excludeContainsRule, 0)
	a.excludeRegexRules = make([]excludeRegexRule, 0)
	a.includeEqualsRules = make([]includeEqualsRule, 0)
	a.includeContainsRules = make([]includeContainsRule, 0)
	a.includeRegexRules = make([]includeRegexRule, 0)
	a.splitPairsRules = make([]splitPairsRule, 0)
	a.valueToNumRules = make([]valueToNumRule, 0)
	a.valueToNumRegexRules = make([]valueToNumRegexRule, 0)

	for _, c := range a.Params.GetChildren() {
		name := c.GetNameS()

		rules := c.GetChildren()
		// loop over all rules
		for _, rc := range rules {
			rule := strings.TrimSpace(rc.GetContentS())

			switch name {
			case "split":
				a.parseSplitSimpleRule(rule)
			case "split_regex":
				a.parseSplitRegexRule(rule)
			case "split_pairs":
				a.parseSplitPairsRule(rule)
			case "join":
				a.parseJoinSimpleRule(rule)
			case "replace":
				a.parseReplaceSimpleRule(rule)
			case "replace_regex":
				a.parseReplaceRegexRule(rule)
			case "exclude_equals":
				a.parseExcludeEqualsRule(rule)
			case "exclude_contains":
				a.parseExcludeContainsRule(rule)
			case "exclude_regex":
				a.parseExcludeRegexRule(rule)
			case "include_equals":
				a.parseIncludeEqualsRule(rule)
			case "include_contains":
				a.parseIncludeContainsRule(rule)
			case "include_regex":
				a.parseIncludeRegexRule(rule)
			case "value_to_num":
				a.parseValueToNumRule(rule)
			case "value_to_num_regex":
				a.parseValueToNumRegexRule(rule)
			default:
				a.SLogger.Warn(
					"Unknown rule name",
					slog.String("object", a.ParentParams.GetChildContentS("object")),
					slog.String("name", name),
				)
			}
		}
	}

	a.actions = make([]func(matrix *matrix.Matrix) error, 0)
	count := 0

	for _, c := range a.Params.GetChildren() {
		name := c.GetNameS()
		switch name {
		case "split":
			if len(a.splitSimpleRules) != 0 {
				a.actions = append(a.actions, a.splitSimple)
				count += len(a.splitSimpleRules)
			}
		case "split_regex":
			if len(a.splitRegexRules) != 0 {
				a.actions = append(a.actions, a.splitRegex)
				count += len(a.splitRegexRules)
			}
		case "split_pairs":
			if len(a.splitPairsRules) != 0 {
				a.actions = append(a.actions, a.splitPairs)
				count += len(a.splitPairsRules)
			}
		case "join":
			if len(a.joinSimpleRules) != 0 {
				a.actions = append(a.actions, a.joinSimple)
				count += len(a.joinSimpleRules)
			}
		case "replace":
			if len(a.replaceSimpleRules) != 0 {
				a.actions = append(a.actions, a.replaceSimple)
				count += len(a.replaceSimpleRules)
			}
		case "replace_regex":
			if len(a.replaceRegexRules) != 0 {
				a.actions = append(a.actions, a.replaceRegex)
				count += len(a.replaceRegexRules)
			}
		case "exclude_equals":
			if len(a.excludeEqualsRules) != 0 {
				a.actions = append(a.actions, a.excludeEquals)
				count += len(a.excludeEqualsRules)
			}
		case "exclude_contains":
			if len(a.excludeContainsRules) != 0 {
				a.actions = append(a.actions, a.excludeContains)
				count += len(a.excludeContainsRules)
			}
		case "exclude_regex":
			if len(a.excludeRegexRules) != 0 {
				a.actions = append(a.actions, a.excludeRegex)
				count += len(a.excludeRegexRules)
			}
		case "include_equals":
			if len(a.includeEqualsRules) != 0 {
				a.actions = append(a.actions, a.includeEquals)
				count += len(a.includeEqualsRules)
			}
		case "include_contains":
			if len(a.includeContainsRules) != 0 {
				a.actions = append(a.actions, a.includeContains)
				count += len(a.includeContainsRules)
			}
		case "include_regex":
			if len(a.includeRegexRules) != 0 {
				a.actions = append(a.actions, a.includeRegex)
				count += len(a.includeRegexRules)
			}
		case "value_to_num":
			if len(a.valueToNumRules) != 0 {
				a.actions = append(a.actions, a.mapValueToNum)
				count += len(a.valueToNumRules)
			}
		case "value_to_num_regex":
			if len(a.valueToNumRegexRules) != 0 {
				a.actions = append(a.actions, a.mapValueToNumRegex)
				count += len(a.valueToNumRegexRules)
			}
		default:
			a.SLogger.Warn(
				"Unknown rule name",
				slog.String("object", a.ParentParams.GetChildContentS("object")),
				slog.String("name", name),
			)
		}
	}

	return count
}

type splitSimpleRule struct {
	sep     string
	source  string
	targets []string
}

// example rule:
// node `/` ,aggr,plex,disk
// if node="jamaica1/ag1/p1/d1", then:
// aggr="ag1", plex="p1", disk="d1"

func (a *LabelAgent) parseSplitSimpleRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := splitSimpleRule{source: strings.TrimSpace(fields[0])}
		a.SLogger.Debug("fields", slog.Any("fields", fields))
		if fields = strings.SplitN(fields[1], "` ", 2); len(fields) == 2 {
			a.SLogger.Debug("fields", slog.Any("fields", fields))
			r.sep = fields[0]
			if r.targets = strings.Split(fields[1], ","); len(r.targets) != 0 {
				a.splitSimpleRules = append(a.splitSimpleRules, r)
				a.addNewLabels(r.targets)
				a.SLogger.Debug("(split) parsed rule", slog.Any("rule", r))
				return
			}
		}
	}
	a.SLogger.Warn("(split) rule has invalid format", slog.String("rule", rule))
}

func (a *LabelAgent) addNewLabels(targets []string) {
	for _, target := range targets {
		if target != "" {
			a.newLabelNames = append(a.newLabelNames, target)
		}
	}
}

type splitPairsRule struct {
	source string
	sep1   string
	sep2   string
}

// example rule:
// node ` ` `:`
// will use single space to extract pairs
// will use colon to extract key-value
func (a *LabelAgent) parseSplitPairsRule(rule string) {
	if fields := strings.Split(rule, "`"); len(fields) == 5 {
		r := splitPairsRule{source: strings.TrimSpace(fields[0])}
		r.sep1 = fields[1]
		r.sep2 = fields[3]
		a.SLogger.Debug("(split_pairs) parsed rule", slog.Any("rule", r))
		a.splitPairsRules = append(a.splitPairsRules, r)
		return
	}
	a.SLogger.Warn("(split_pairs) rule has invalid format", slog.String("rule", rule))
}

type splitRegexRule struct {
	reg     *regexp.Regexp
	source  string
	targets []string
}

// example rule:
// node `.*_(ag\d+)_(p\d+)_(d\d+)` aggr,plex,disk
// if node="jamaica1_ag1_p1_d1", then:
// aggr="ag1", plex="p1", disk="d1"

func (a *LabelAgent) parseSplitRegexRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := splitRegexRule{source: strings.TrimSpace(fields[0])}
		if fields = strings.SplitN(fields[1], "` ", 2); len(fields) == 2 {
			var err error
			if r.reg, err = regexp.Compile(fields[0]); err != nil {
				a.SLogger.Error("(split_regex) invalid regex", slogx.Err(err))
				return
			}
			if r.targets = strings.Split(fields[1], ","); len(r.targets) != 0 {
				a.splitRegexRules = append(a.splitRegexRules, r)
				a.addNewLabels(r.targets)
				a.SLogger.Debug("(split_regex) parsed rule", slog.String("rule", r.source))
				return
			}
		}
	}
	a.SLogger.Warn("(split_regex) rule has invalid format", slog.String("rule", rule))
}

type joinSimpleRule struct {
	sep     string
	target  string
	sources []string
}

// example rule:
// plex_long `_` aggr,plex
// if aggr="aggr1" and plex="plex1"; then
// plex_long="aggr1_plex1"

func (a *LabelAgent) parseJoinSimpleRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := joinSimpleRule{target: strings.TrimSpace(fields[0])}
		a.newLabelNames = append(a.newLabelNames, r.target)
		if fields = strings.SplitN(fields[1], "` ", 2); len(fields) == 2 {
			r.sep = fields[0]
			if r.sources = strings.Split(fields[1], ","); len(r.sources) != 0 {
				a.joinSimpleRules = append(a.joinSimpleRules, r)
				a.SLogger.Debug(
					"(join) parsed rule",
					slog.String("sep", r.sep),
					slog.String("target", r.target),
				)
				return
			}
		}
	}
	a.SLogger.Warn("(join) rule has invalid format", slog.String("rule", rule))
}

type replaceSimpleRule struct {
	source string
	target string
	old    string
	new    string
}

// example rule:
// node node_short `node_` ``
// if node="node_jamaica1"; then:
// node_short="jamaica1"

func (a *LabelAgent) parseReplaceSimpleRule(rule string) {
	if fields := strings.SplitN(rule, " `", 3); len(fields) == 3 {
		if labels := strings.Fields(fields[0]); len(labels) == 2 {
			r := replaceSimpleRule{source: labels[0], target: labels[1]}
			r.old = strings.TrimSuffix(fields[1], "`")
			r.new = strings.TrimSuffix(fields[2], "`")
			a.replaceSimpleRules = append(a.replaceSimpleRules, r)
			a.newLabelNames = append(a.newLabelNames, r.target)
			a.SLogger.Debug("(replace) parsed rule", slog.Any("rule", r))
			return
		}
	}
	a.SLogger.Warn("(replace) rule has invalid format", slog.String("rule", rule))
}

type replaceRegexRule struct {
	reg     *regexp.Regexp
	source  string
	target  string
	indices []int
	format  string
}

// example rule:
//nolint:dupword
// node node `^(node)_(\d+)_.*$` `Node-$2`
// if node="node_10_dc2"; then:
// node="Node-10"

func (a *LabelAgent) parseReplaceRegexRule(rule string) {
	if fields := strings.SplitN(rule, " `", 3); len(fields) == 3 {
		if labels := strings.Fields(fields[0]); len(labels) == 2 {
			r := replaceRegexRule{source: labels[0], target: labels[1]}
			a.newLabelNames = append(a.newLabelNames, r.target)
			var err error
			if r.reg, err = regexp.Compile(strings.TrimSuffix(fields[1], "`")); err != nil {
				a.SLogger.Error("(replace_regex) invalid regex", slogx.Err(err))
				return
			}

			r.indices = make([]int, 0)
			errPos := -1

			if fields[2] = strings.TrimSuffix(fields[2], "`"); fields[2] != "" {
				insideNum := false
				num := ""
				for i, b := range fields[2] {
					ch := string(b)
					if insideNum {
						if _, err := strconv.Atoi(ch); err == nil {
							num += ch
							continue
						} else if index, err := strconv.Atoi(num); err == nil && index > 0 {
							r.indices = append(r.indices, index-1)
							r.format += "%s"
							insideNum = false
							num = ""
						} else {
							errPos = i
							break
						}
					}
					if ch == "$" {
						if strings.HasSuffix(r.format, `\`) {
							r.format = strings.TrimSuffix(r.format, `\`) + "$"
						} else {
							insideNum = true
						}
					} else {
						r.format += ch
					}
				}
			}
			if errPos != -1 {
				a.SLogger.Error(
					"(replace_regex) invalid char in substitution string",
					slog.Int("errPos", errPos),
					slog.String("sub", string(fields[2][errPos])),
				)
				return
			}
			a.replaceRegexRules = append(a.replaceRegexRules, r)
			a.SLogger.Debug("(replace_regex) parsed rule", slog.Any("rule", r))
			return
		}
	}
	a.SLogger.Warn("(replace_regex) rule has invalid format", slog.String("rule", rule))
}

type excludeEqualsRule struct {
	label string
	value string
}

// example rule
// vol_type `flexgroup_constituent`
// all instances with matching label type, will not be exported

func (a *LabelAgent) parseExcludeEqualsRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := excludeEqualsRule{label: fields[0]}
		r.value = strings.TrimSuffix(fields[1], "`")
		a.excludeEqualsRules = append(a.excludeEqualsRules, r)
		a.SLogger.Debug(
			"(exclude_equals) parsed rule",
			slog.String("label", r.label),
			slog.String("value", r.value),
		)
	} else {
		a.SLogger.Debug("(exclude_equals) rule definition should have two fields", slog.String("rule", rule))
	}
}

type excludeContainsRule struct {
	label string
	value string
}

func (a *LabelAgent) parseExcludeContainsRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := excludeContainsRule{label: fields[0]}
		r.value = strings.TrimSuffix(fields[1], "`")
		a.excludeContainsRules = append(a.excludeContainsRules, r)
		a.SLogger.Debug("exclude_contains parsed rule", slog.String("rule", r.label), slog.String("value", r.value))
	} else {
		a.SLogger.Error("exclude_contains rule definition should have two fields", slog.String("rule", rule))
	}
}

type excludeRegexRule struct {
	label string
	reg   *regexp.Regexp
}

func (a *LabelAgent) parseExcludeRegexRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := excludeRegexRule{label: fields[0]}
		var err error
		pattern := strings.TrimSuffix(fields[1], "`")
		if r.reg, err = regexp.Compile(pattern); err == nil {
			a.excludeRegexRules = append(a.excludeRegexRules, r)
			a.SLogger.Debug("parsed exclude_regex", slog.String("pattern", pattern), slog.String("rule", r.label))
		} else {
			a.SLogger.Error("ignore exclude_regex", slogx.Err(err))
		}
	} else {
		a.SLogger.Error("exclude_regex rule definition should have two fields", slog.String("rule", rule))
	}
}

type includeEqualsRule struct {
	label string
	value string
}

// example rule
// vol_type `flexgroup_constituent`
// all instances with matching label type, will be exported

func (a *LabelAgent) parseIncludeEqualsRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := includeEqualsRule{label: fields[0]}
		r.value = strings.TrimSuffix(fields[1], "`")
		a.includeEqualsRules = append(a.includeEqualsRules, r)
		a.SLogger.Debug("(include_equals) parsed rule", slog.String("label", r.label), slog.String("value", r.value))
	} else {
		a.SLogger.Warn("(include_equals) rule definition should have two fields", slog.String("rule", rule))
	}
}

type includeContainsRule struct {
	label string
	value string
}

func (a *LabelAgent) parseIncludeContainsRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := includeContainsRule{label: fields[0]}
		r.value = strings.TrimSuffix(fields[1], "`")
		a.includeContainsRules = append(a.includeContainsRules, r)
		a.SLogger.Debug("(include_contains) parsed rule", slog.String("label", r.label), slog.String("value", r.value))
	} else {
		a.SLogger.Error("(include_contains) rule definition should have two fields", slog.String("rule", rule))
	}
}

type includeRegexRule struct {
	label string
	reg   *regexp.Regexp
}

func (a *LabelAgent) parseIncludeRegexRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := includeRegexRule{label: fields[0]}
		var err error
		if r.reg, err = regexp.Compile(strings.TrimSuffix(fields[1], "`")); err == nil {
			a.includeRegexRules = append(a.includeRegexRules, r)
			a.SLogger.Debug(
				"(include_regex) parsed rule",
				slog.String("regex", r.reg.String()),
				slog.String("label", r.label),
			)
		} else {
			a.SLogger.Error("(include_regex) compile regex", slogx.Err(err))
		}
	} else {
		a.SLogger.Error("(include_regex) rule definition should have two fields", slog.String("rule", rule))
	}
}

type valueToNumRule struct {
	metric       string
	label        string
	defaultValue uint8
	hasDefault   bool
	mapping      map[string]uint8
}

// example rule:
// metric label zapi_value rest_value `default_value`
// status state normal ok `0`
// will create a new metric "status" of type uint8
// if value of label "state" is normal or ok
// the metric value will be 1, otherwise it will be 0.

func (a *LabelAgent) parseValueToNumRule(rule string) {
	if fields := strings.Fields(rule); len(fields) == 4 || len(fields) == 5 {
		r := valueToNumRule{metric: fields[0], label: fields[1]}
		r.mapping = make(map[string]uint8)

		// This '-' is used for handling special case in disk.yaml, rest all are handled normally with assigning to 1.
		for v := range strings.SplitSeq(fields[2], "-") {
			r.mapping[v] = uint8(1)
		}
		for v := range strings.SplitSeq(fields[3], "-") {
			r.mapping[v] = uint8(1)
		}

		if len(fields) == 5 {

			fields[4] = strings.TrimPrefix(strings.TrimSuffix(fields[4], "`"), "`")

			v, err := strconv.ParseUint(fields[4], 10, 8)
			if err != nil {
				a.SLogger.Error("(value_to_num) parse", slogx.Err(err), slog.String("default", fields[4]))
				return
			}
			r.hasDefault = true
			r.defaultValue = uint8(v) // #nosec G115
		}

		a.valueToNumRules = append(a.valueToNumRules, r)
		a.SLogger.Debug("(value_to_num) parsed rule", slog.Any("rule", r))
		return
	}
	a.SLogger.Warn("(value_to_num) rule has invalid format", slog.String("rule", rule))
}

type valueToNumRegexRule struct {
	metric       string
	label        string
	defaultValue uint8
	hasDefault   bool
	reg          []*regexp.Regexp
}

// example rule:
// metric label zapi_value rest_value `default_value`
// status state ^normal$ ^ok$ `0`
// will create a new metric "status" of type uint8
// if value of label "state" contains normal or ok
// the metric value will be 1, otherwise it will be 0.

func (a *LabelAgent) parseValueToNumRegexRule(rule string) {
	var err error

	if fields := strings.Fields(rule); len(fields) == 4 || len(fields) == 5 {
		r := valueToNumRegexRule{metric: fields[0], label: fields[1], reg: make([]*regexp.Regexp, 2)}
		if r.reg[0], err = regexp.Compile(fields[2]); err != nil {
			a.SLogger.Error(
				"(value_to_num_regex) compile regex",
				slogx.Err(err),
				slog.String("regex", r.reg[0].String()),
				slog.String("value", fields[2]),
			)
		}

		if r.reg[1], err = regexp.Compile(fields[3]); err != nil {
			a.SLogger.Error(
				"(value_to_num_regex) compile regex",
				slogx.Err(err),
				slog.String("regex", r.reg[1].String()),
				slog.String("value", fields[3]),
			)
		}

		if len(fields) == 5 {
			fields[4] = strings.TrimPrefix(strings.TrimSuffix(fields[4], "`"), "`")
			v, err := strconv.ParseUint(fields[4], 10, 8)
			if err != nil {
				a.SLogger.Error("(value_to_num_regex) parse default value", slogx.Err(err), slog.String("regex", fields[4]))
				return
			}
			r.hasDefault = true
			r.defaultValue = uint8(v) // #nosec G115
		}

		a.valueToNumRegexRules = append(a.valueToNumRegexRules, r)
		a.SLogger.Debug("(value_to_num_regex) parsed rule", slog.Any("rule", r))
		return
	}
	a.SLogger.Warn("(value_to_num_regex) rule has invalid format", slog.String("rule", rule))
}
