package collectors

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"regexp"
	"strconv"
	"strings"
)

var iopsPerUnitRe = regexp.MustCompile(`(\d+)iops/(tb|gb)`)
var iopsRe = regexp.MustCompile(`(\d+)iops`)
var bpsRe = regexp.MustCompile(`(\d+(\.\d+)?)(\w+)/s`)

var unitToMb = map[string]float32{
	"b":  1 / float32(1000*1000),
	"kb": 1 / float32(1000),
	"mb": 1,
	"gb": 1000,
	"tb": 1000 * 1000,
}

type MaxXput struct {
	IOPS string
	Mbps string
}

type QosCommon struct {
}

func (q *QosCommon) SetThroughput(data *matrix.Matrix, instance *matrix.Instance, labelName string, iopLabel string, mbpsLabel string, logger *logging.Logger) {
	val := instance.GetLabel(labelName)
	if val == "" {
		return
	}
	xput, err := ZapiXputToRest(val)
	if err != nil {
		logger.Warn().Str(labelName, val).Msg("Unable to convert label, skipping")
		return
	}
	q.SetLabel(iopLabel, data, instance, xput.IOPS, logger)
	q.SetLabel(mbpsLabel, data, instance, xput.Mbps, logger)
}

func (q *QosCommon) SetLabel(labelName string, data *matrix.Matrix, instance *matrix.Instance, value string, logger *logging.Logger) {
	if value == "" {
		return
	}
	instance.SetLabel(labelName, value)
	m := data.GetMetric(labelName)
	if m != nil {
		err := m.SetValueString(instance, value)
		if err != nil {
			logger.Error().Str(labelName, value).Err(err).Msg("Unable to set metric")
		}
	}
}

func ZapiXputToRest(zapi string) (MaxXput, error) {
	lower := strings.ToLower(zapi)
	empty := MaxXput{IOPS: "", Mbps: ""}
	if lower == "inf" || lower == "0" || lower == "" {
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

	// check for iops per unit (TB or GB)
	matches := iopsPerUnitRe.FindStringSubmatch(lower)
	if len(matches) == 3 {
		iops, err := strconv.Atoi(matches[1])
		if err != nil {
			return empty, fmt.Errorf("failed to convert iops value [%s] of [%s]", matches[1], zapi)
		}
		unit := matches[2]
		if unit == "gb" {
			// Convert from IOPS/GB to IOPS/TB. ONTAP default is IOPS/TB
			iops *= 1000
		}
		return MaxXput{IOPS: strconv.Itoa(iops), Mbps: ""}, nil
	}

	// check for iops
	matches = iopsRe.FindStringSubmatch(lower)
	if len(matches) == 2 {
		return MaxXput{IOPS: matches[1], Mbps: ""}, nil
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
	return MaxXput{Mbps: strconv.Itoa(int(mbps)), IOPS: ""}, nil
}
