package label_agent

import (
    "strings"
    "strconv"
    "regexp"
    "goharvest2/share/logger"
    "goharvest2/share/matrix"
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

    for _, c := range me.Params.GetChildren() {
        name := c.GetNameS()
        rule := strings.TrimSpace(c.GetContentS())

        switch name {
        case "split":
            me.parseSplitSimpleRule(rule)
        case "split_regex":
            me.parseSplitRegexRule(rule)
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
        default:
            logger.Warn(me.Prefix, "unknown rule (%s)", name)
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

    return count
}

type splitSimpleRule struct {
	sep string
	source string
	targets []string
}

// example rule:
// node `/` ,aggr,plex,disk
// if node="jamaica1/ag1/p1/d1", then:
// aggr="ag1", plex="p1", disk="d1"

func (me *LabelAgent) parseSplitSimpleRule(rule string) {
    if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
        r := splitSimpleRule{source: strings.TrimSpace(fields[0])}
        logger.Debug(me.Prefix, "fields := %v", fields)
        if fields = strings.SplitN(fields[1], "` ", 2); len(fields) == 2 {
            logger.Debug(me.Prefix, "fields = %v", fields)
            r.sep = fields[0]
            if r.targets = strings.Split(fields[1], ","); len(r.targets) != 0 {
                me.splitSimpleRules = append(me.splitSimpleRules, r)
                logger.Debug(me.Prefix, "(split) parsed rule [%v]", r)
                return
            }
        }
    }
    logger.Warn(me.Prefix, "(split) rule has invalid format [%s]", rule)
}

type splitRegexRule struct {
	reg *regexp.Regexp
	source string
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
                logger.Error(me.Prefix, "(split_regex) invalid regex: %v", err)
                return
            }
            logger.Trace(me.Prefix, "(split_regex) compule regex [%s]", r.reg.String())
            if r.targets = strings.Split(fields[1], ","); len(r.targets) != 0 {
                me.splitRegexRules = append(me.splitRegexRules, r)
                logger.Debug(me.Prefix, "(split_regex) parsed rule [%v]", r)
                return
            }
        }
    }
    logger.Warn(me.Prefix, "(split_regex) rule has invalid format [%s]", rule)
}

type joinSimpleRule struct {
	sep string
	target string
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
                logger.Debug(me.Prefix, "(join) parsed rule [%v]", r)
                return
            }
        }
    }
    logger.Warn(me.Prefix, "(join) rule has invalid format [%s]", rule)
}

type replaceSimpleRule struct {
	source string
    target string
	old string
	new string
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
            logger.Debug(me.Prefix, "(replace) parsed rule [%v]", r)
            return
        }
    }
    logger.Warn(me.Prefix, "(replace) rule has invalid format [%s]", rule)
}

type replaceRegexRule struct {
	reg *regexp.Regexp
    source string
    target string
    indices []int
    format string
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
                logger.Error(me.Prefix, "(replace_regex) invalid regex: %v", err)
                return
            }
            logger.Trace(me.Prefix, "(replace_regex) compiled regular expression [%s]", r.reg.String())

            r.indices = make([]int, 0)
            err_pos := -1

            if fields[2] = strings.TrimSuffix(fields[2], "`"); len(fields[2]) != 0 {
                logger.Trace(me.Prefix, "(replace_regex) parsing substitution string [%s] (%d)", fields[2], len(fields[2]))
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
                        r.form
}

// example rule
// vol_type `flexgroup_constituent`
// all instances with matching label type, will not be exported

func (me *LabelAgent) parseExcludeEqualsRule(rule string) {
    if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
        r := excludeEqualsRule{label: fields[0]}
        r.value = strings.TrimSuffix(fields[1], "`")
        me.excludeEqualsRules = append(me.excludeEqualsRules, r)
        logger.Debug(me.Prefix, "(exclude_equals) parsed rule [%v]", r)
    } else {
        logger.Warn(me.Prefix, "(exclude_equals) rule definition [%s] should have two fields", rule)
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
        logger.Debug(me.Prefix, "(exclude_contains) parsed rule [%v]", r)
    } else {
        logger.Error(me.Prefix, "(exclude_contains) rule definition [%s] should have two fields", rule)
    }
}

type excludeRegexRule struct {
    label string
    reg *regexp.Regexp
}

func (me *LabelAgent) parseExcludeRegexRule(rule string) {
    if fields := strings.SplitN(rule, " `", 2); len(fields) == 2 {
        r := excludeRegexRule{label: fields[0]}
        var err error
        if r.reg, err = regexp.Compile(strings.TrimSuffix(fields[1], "`")); err == nil {
            me.excludeRegexRules = append(me.excludeRegexRules, r)
            logger.Trace(me.Prefix, "(exclude_regex) compiled regex: [%s]", r.reg.String())
            logger.Debug(me.Prefix, "(exclude_regex) parsed rule [%v]", r)
        } else {
            logger.Error(me.Prefix, "(exclude_regex) compule regex: %v", err)
        }
    } else {
        logger.Error(me.Prefix, "(exclude_regex) rule definition [%s] should have two fields", rule)
    }
}
