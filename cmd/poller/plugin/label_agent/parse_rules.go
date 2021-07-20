/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package label_agent

import (
	"goharvest2/pkg/matrix"
	"regexp"
	"strconv"
	"strings"
)

// parse rules from plugin parameters and return number of rules parsed
func (me *LabelAgent) parseRules() int {

	me.splitSimpleRules = make([]splitSimpleRule, 0)
	me.splitRegexRules = make([]splitRegexRule, 0)
	me.joinSimpleRules = make([]joinSimpleRule, 0)
	me.replaceSimpleRules = make([]replaceSimpleRule, 0)
	me.replaceRegexRules = make([]replaceRegexRule, 0)
	me.excludeEqualsRules = make([]excludeEqualsRule, 0)
	me.excludeContainsRules = make([]excludeContainsRule, 0)
	me.excludeRegexRules = make([]excludeRegexRule, 0)
	me.splitPairsRules = make([]splitPairsRule, 0)
	me.valueMappingRules = make([]valueMappingRule, 0)
	me.valueToNumRules = make([]valueToNumRule, 0)

	for _, c := range me.Params.GetChildren() {
		name := c.GetNameS()
		rule := strings.TrimSpace(c.GetContentS())

		switch name {
		case "split":
			me.parseSplitSimpleRule(rule)
		case "split_regex":
			me.parseSplitRegexRule(rule)
		case "split_pairs":
			me.parseSplitPairsRule(rule)
		case "join":
			me.parseJoinSimpleRule(rule)
		case "replace":
			me.parseReplaceSimpleRule(rule)
		case "replace_regex":
			me.parseReplaceRegexRule(rule)
		case "exclude_equals":
			me.parseExcludeEqualsRule(rule)
		case "exclude_contains":
			me.parseExcludeContainsRule(rule)
		case "exclude_regex":
			me.parseExcludeRegexRule(rule)
		case "value_mapping":
			me.parseValueMappingRule(rule)
		case "value_to_num":
			me.parseValueToNumRule(rule)
		default:
			me.Logger.Warn().Msgf("unknown rule (%s)", name)
		}
	}

	me.actions = make([]func(*matrix.Instance), 0)
	count := 0

	if len(me.splitSimpleRules) != 0 {
		me.actions = append(me.actions, me.splitSimple)
		count += len(me.splitSimpleRules)
	}

	if len(me.splitRegexRules) != 0 {
		me.actions = append(me.actions, me.splitRegex)
		count += len(me.splitRegexRules)
	}

	if len(me.splitPairsRules) != 0 {
		me.actions = append(me.actions, me.splitPairs)
		count += len(me.splitPairsRules)
	}

	if len(me.joinSimpleRules) != 0 {
		me.actions = append(me.actions, me.joinSimple)
		count += len(me.joinSimpleRules)
	}

	if len(me.replaceSimpleRules) != 0 {
		me.actions = append(me.actions, me.replaceSimple)
		count += len(me.replaceSimpleRules)
	}

	if len(me.replaceRegexRules) != 0 {
		me.actions = append(me.actions, me.replaceRegex)
		count += len(me.replaceRegexRules)
	}

	if len(me.excludeEqualsRules) != 0 {
		me.actions = append(me.actions, me.excludeEquals)
		count += len(me.excludeEqualsRules)
	}

	if len(me.excludeContainsRules) != 0 {
		me.actions = append(me.actions, me.excludeContains)
		count += len(me.excludeContainsRules)
	}

	if len(me.excludeRegexRules) != 0 {
		me.actions = append(me.actions, me.excludeRegex)
		count += len(me.excludeRegexRules)
	}

	count += len(me.valueMappingRules)
	count += len(me.valueToNumRules)

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

func (me *LabelAgent) parseSplitSimpleRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := splitSimpleRule{source: strings.TrimSpace(fields[0])}
		me.Logger.Debug().Msgf("fields := %v", fields)
		if fields = strings.SplitN(fields[1], "` ", 2); len(fields) == 2 {
			me.Logger.Debug().Msgf("fields = %v", fields)
			r.sep = fields[0]
			if r.targets = strings.Split(fields[1], ","); len(r.targets) != 0 {
				me.splitSimpleRules = append(me.splitSimpleRules, r)
				me.Logger.Debug().Msgf("(split) parsed rule [%v]", r)
				return
			}
		}
	}
	me.Logger.Warn().Msgf("(split) rule has invalid format [%s]", rule)
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
func (me *LabelAgent) parseSplitPairsRule(rule string) {
	if fields := strings.Split(rule, "`"); len(fields) == 5 {
		r := splitPairsRule{source: strings.TrimSpace(fields[0])}
		r.sep1 = fields[1]
		r.sep2 = fields[3]
		me.Logger.Debug().Msgf("(split_pairs) parsed rule [%v]", r)
		me.splitPairsRules = append(me.splitPairsRules, r)
		return
	}
	me.Logger.Warn().Msgf("(split_pairs) rule has invalid format [%s]", rule)
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

func (me *LabelAgent) parseSplitRegexRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := splitRegexRule{source: strings.TrimSpace(fields[0])}
		if fields = strings.SplitN(fields[1], "` ", 2); len(fields) == 2 {
			var err error
			if r.reg, err = regexp.Compile(fields[0]); err != nil {
				me.Logger.Error().Stack().Err(err).Msg("(split_regex) invalid regex")
				return
			}
			me.Logger.Trace().Msgf("(split_regex) compule regex [%s]", r.reg.String())
			if r.targets = strings.Split(fields[1], ","); len(r.targets) != 0 {
				me.splitRegexRules = append(me.splitRegexRules, r)
				me.Logger.Debug().Msgf("(split_regex) parsed rule [%v]", r)
				return
			}
		}
	}
	me.Logger.Warn().Msgf("(split_regex) rule has invalid format [%s]", rule)
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

func (me *LabelAgent) parseJoinSimpleRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := joinSimpleRule{target: strings.TrimSpace(fields[0])}
		if fields = strings.SplitN(fields[1], "` ", 2); len(fields) == 2 {
			r.sep = fields[0]
			if r.sources = strings.Split(fields[1], ","); len(r.sources) != 0 {
				me.joinSimpleRules = append(me.joinSimpleRules, r)
				me.Logger.Debug().Msgf("(join) parsed rule [%v]", r)
				return
			}
		}
	}
	me.Logger.Warn().Msgf("(join) rule has invalid format [%s]", rule)
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

func (me *LabelAgent) parseReplaceSimpleRule(rule string) {
	if fields := strings.SplitN(rule, " `", 3); len(fields) == 3 {
		if labels := strings.Fields(fields[0]); len(labels) == 2 {
			r := replaceSimpleRule{source: labels[0], target: labels[1]}
			r.old = strings.TrimSuffix(fields[1], "`")
			r.new = strings.TrimSuffix(fields[2], "`")
			me.replaceSimpleRules = append(me.replaceSimpleRules, r)
			me.Logger.Debug().Msgf("(replace) parsed rule [%v]", r)
			return
		}
	}
	me.Logger.Warn().Msgf("(replace) rule has invalid format [%s]", rule)
}

type replaceRegexRule struct {
	reg     *regexp.Regexp
	source  string
	target  string
	indices []int
	format  string
}

// example rule:
// node node `^(node)_(\d+)_.*$` `Node-$2`
// if node="node_10_dc2"; then:
// node="Node-10"

func (me *LabelAgent) parseReplaceRegexRule(rule string) {
	if fields := strings.SplitN(rule, " `", 3); len(fields) == 3 {
		if labels := strings.Fields(fields[0]); len(labels) == 2 {
			r := replaceRegexRule{source: labels[0], target: labels[1]}
			var err error
			if r.reg, err = regexp.Compile(strings.TrimSuffix(fields[1], "`")); err != nil {
				me.Logger.Error().Stack().Err(err).Msg("(replace_regex) invalid regex")
				return
			}
			me.Logger.Trace().Msgf("(replace_regex) compiled regular expression [%s]", r.reg.String())

			r.indices = make([]int, 0)
			err_pos := -1

			if fields[2] = strings.TrimSuffix(fields[2], "`"); len(fields[2]) != 0 {
				me.Logger.Trace().Msgf("(replace_regex) parsing substitution string [%s] (%d)", fields[2], len(fields[2]))
				inside_num := false
				num := ""
				for i, b := range fields[2] {
					ch := string(b)
					if inside_num {
						if _, err := strconv.Atoi(ch); err == nil {
							num += ch
							continue
						} else if index, err := strconv.Atoi(num); err == nil && index > 0 {
							r.indices = append(r.indices, index-1)
							r.format += "%s"
							inside_num = false
							num = ""
						} else {
							err_pos = i
							break
						}
					}
					if ch == "$" {
						if strings.HasSuffix(r.format, `\`) {
							r.format = strings.TrimSuffix(r.format, `\`) + "$"
						} else {
							inside_num = true
						}
					} else {
						r.format += ch
					}
				}
			}
			if err_pos != -1 {
				me.Logger.Error().Stack().Err(nil).Msgf("(replace_regex) invalid char in substitution string at pos %d (%s)", err_pos, string(fields[2][err_pos]))
				return
			}
			me.replaceRegexRules = append(me.replaceRegexRules, r)
			me.Logger.Debug().Msgf("(replace_regex) parsed rule [%v]", r)
			return
		}
	}
	me.Logger.Warn().Msgf("(replace_regex) rule has invalid format [%s]", rule)
}

type excludeEqualsRule struct {
	label string
	value string
}

// example rule
// vol_type `flexgroup_constituent`
// all instances with matching label type, will not be exported

func (me *LabelAgent) parseExcludeEqualsRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := excludeEqualsRule{label: fields[0]}
		r.value = strings.TrimSuffix(fields[1], "`")
		me.excludeEqualsRules = append(me.excludeEqualsRules, r)
		me.Logger.Debug().Msgf("(exclude_equals) parsed rule [%v]", r)
	} else {
		me.Logger.Warn().Msgf("(exclude_equals) rule definition [%s] should have two fields", rule)
	}
}

type excludeContainsRule struct {
	label string
	value string
}

func (me *LabelAgent) parseExcludeContainsRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := excludeContainsRule{label: fields[0]}
		r.value = strings.TrimSuffix(fields[1], "`")
		me.excludeContainsRules = append(me.excludeContainsRules, r)
		me.Logger.Debug().Msgf("(exclude_contains) parsed rule [%v]", r)
	} else {
		me.Logger.Error().Stack().Err(nil).Msgf("(exclude_contains) rule definition [%s] should have two fields", rule)
	}
}

type excludeRegexRule struct {
	label string
	reg   *regexp.Regexp
}

func (me *LabelAgent) parseExcludeRegexRule(rule string) {
	if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
		r := excludeRegexRule{label: fields[0]}
		var err error
		if r.reg, err = regexp.Compile(strings.TrimSuffix(fields[1], "`")); err == nil {
			me.excludeRegexRules = append(me.excludeRegexRules, r)
			me.Logger.Debug().Msgf("(exclude_regex) compiled regex: [%s]", r.reg.String())
			me.Logger.Debug().Msgf("(exclude_regex) parsed rule [%v]", r)
		} else {
			me.Logger.Error().Stack().Err(err).Msgf("(exclude_regex) compile regex:")
		}
	} else {
		me.Logger.Error().Stack().Err(nil).Msgf("(exclude_regex) rule definition [%s] should have two fields", rule)
	}
}

type valueMappingRule struct {
	metric       string
	label        string
	defaultValue uint8
	hasDefault   bool
	mapping      map[string]uint8
}

// example rule:
// status state ok,pending,failed `8`
// will create a new metric "status" of type uint8
// if value of label "state" is any of ok,pending,failed
// the metric value will be respectively 0, 1 or 2

func (me *LabelAgent) parseValueMappingRule(rule string) {
	if fields := strings.Fields(rule); len(fields) == 3 || len(fields) == 4 {
		r := valueMappingRule{metric: fields[0], label: fields[1]}
		r.mapping = make(map[string]uint8)
		for i, v := range strings.Split(fields[2], ",") {
			r.mapping[v] = uint8(i)
		}

		if len(fields) == 4 {

			fields[3] = strings.TrimPrefix(strings.TrimSuffix(fields[3], "`"), "`")

			if v, err := strconv.ParseUint(fields[3], 10, 8); err != nil {
				me.Logger.Error().Stack().Err(err).Msgf("(value_mapping) parse default value (%s): ", fields[3])
				return
			} else {
				r.hasDefault = true
				r.defaultValue = uint8(v)
			}
		}

		me.valueMappingRules = append(me.valueMappingRules, r)
		me.Logger.Debug().Msgf("(value_mapping) parsed rule [%v]", r)
		return
	}
	me.Logger.Warn().Msgf("(value_mapping) rule has invalid format [%s]", rule)
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

func (me *LabelAgent) parseValueToNumRule(rule string) {
	if fields := strings.Fields(rule); len(fields) == 4 || len(fields) == 5 {
		r := valueToNumRule{metric: fields[0], label: fields[1]}
		r.mapping = make(map[string]uint8)

		// This '-' is used for handling special case in disk.yaml, rest all are handled normally with assigning to 1.
		for _, v := range strings.Split(fields[2], "-") {
			r.mapping[v] = uint8(1)
		}
		for _, v := range strings.Split(fields[3], "-") {
			r.mapping[v] = uint8(1)
		}

		if len(fields) == 5 {

			fields[4] = strings.TrimPrefix(strings.TrimSuffix(fields[4], "`"), "`")

			if v, err := strconv.ParseUint(fields[4], 10, 8); err != nil {
				me.Logger.Error().Stack().Err(err).Msgf("(value_to_num) parse default value (%s): ", fields[4])
				return
			} else {
				r.hasDefault = true
				r.defaultValue = uint8(v)
			}
		}

		me.valueToNumRules = append(me.valueToNumRules, r)
		me.Logger.Debug().Msgf("(value_to_num) parsed rule [%v]", r)
		return
	}
	me.Logger.Warn().Msgf("(value_to_num) rule has invalid format [%s]", rule)
}
