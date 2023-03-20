package qospolicyfixed

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"regexp"
	"strconv"
	"strings"
)

type QosPolicyFixed struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &QosPolicyFixed{AbstractPlugin: p}
}

func (p *QosPolicyFixed) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	// Change ZAPI max-throughput/min-throughput to match what REST publishes

	data := dataMap[p.Object]
	for _, instance := range data.GetInstances() {
		policyClass := instance.GetLabel("class")
		if policyClass != "user_defined" {
			// Only export user_defined policy classes - ignore all others
			// REST only returns user_defined while ZAPI returns all
			instance.SetExportable(false)
			continue
		}
		p.setThroughput(instance, "max_xput", "max_throughput_iops", "max_throughput_mbps")
		p.setThroughput(instance, "min_xput", "min_throughput_iops", "min_throughput_mbps")
	}

	return nil, nil
}

func (p *QosPolicyFixed) setThroughput(instance *matrix.Instance, labelName string, iopLabel string, mbpsLabel string) {
	val := instance.GetLabel(labelName)
	xput, err := ZapiXputToRest(val)
	if err != nil {
		p.Logger.Warn().Str(labelName, val).Msg("Unable to convert label, skipping")
		return
	}
	instance.SetLabel(iopLabel, xput.IOPS)
	instance.SetLabel(mbpsLabel, xput.Mbps)
}

var iopsRe = regexp.MustCompile(`(\d+)iops`)
var bpsRe = regexp.MustCompile(`(\d+(\.\d+)?)(\w+)/s`)

var unitToMb = map[string]float32{
	"b":  1 / float32(1024*1024),
	"kb": 1 / float32(1024),
	"mb": 1,
	"gb": 1024,
	"tb": 1024 * 1024,
}

type MaxXput struct {
	IOPS string
	Mbps string
}

func ZapiXputToRest(zapi string) (MaxXput, error) {
	lower := strings.ToLower(zapi)
	empty := MaxXput{IOPS: "0", Mbps: "0"}
	if lower == "inf" || lower == "0" {
		return empty, nil
	}

	// check for a combination
	before, after, found := strings.Cut(lower, ",")
	if found {
		l, err1 := ZapiXputToRest(before)
		r, err2 := ZapiXputToRest(after)
		if err1 != nil || err2 != nil {
			return empty, errors.Join(err1, err2)
		}
		return MaxXput{
			IOPS: l.IOPS,
			Mbps: r.Mbps,
		}, nil
	}

	// check for iops
	matches := iopsRe.FindStringSubmatch(lower)
	if len(matches) == 2 {
		return MaxXput{IOPS: matches[1], Mbps: "0"}, nil
	}

	// check for bps and normalize units to Mbps
	matches = bpsRe.FindStringSubmatch(lower)
	if len(matches) != 4 {
		return empty, fmt.Errorf("unknown qos-policy format [%s]", zapi)
	}
	numStr := matches[1]
	unit := matches[3]
	multiple, ok := unitToMb[unit]
	if !ok {
		return empty, fmt.Errorf("unknown qos-policy unit [%s] of [%s]", unit, zapi)
	}
	num, err := strconv.ParseFloat(numStr, 32)
	if err != nil {
		return empty, fmt.Errorf("failed to convert qos-policy unit [%s] of [%s]", numStr, zapi)
	}
	mbps := float32(num) * multiple
	return MaxXput{Mbps: strconv.Itoa(int(mbps)), IOPS: "0"}, nil
}
