package collectors

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
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

func SetThroughput(data *matrix.Matrix, instance *matrix.Instance, labelName string, iopLabel string, mbpsLabel string, logger *slog.Logger) {
	val := instance.GetLabel(labelName)
	if val == "" {
		return
	}
	xput, err := ZapiXputToRest(val)
	if err != nil {
		logger.Warn("Unable to convert label, skipping", slog.String(labelName, val))
		return
	}
	QosSetLabel(iopLabel, data, instance, xput.IOPS, logger)
	QosSetLabel(mbpsLabel, data, instance, xput.Mbps, logger)
}

func QosSetLabel(labelName string, data *matrix.Matrix, instance *matrix.Instance, value string, logger *slog.Logger) {
	if value == "" {
		return
	}
	instance.SetLabel(labelName, value)
	m := data.GetMetric(labelName)
	if m != nil {
		err := m.SetValueString(instance, value)
		if err != nil {
			logger.Error("Unable to set metric", slog.String(labelName, value), slogx.Err(err))
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
	mbpsStr := strconv.FormatFloat(float64(mbps), 'f', 2, 32)

	// Trim unnecessary trailing zeros and decimal points
	mbpsStr = strings.TrimRight(mbpsStr, "0")
	mbpsStr = strings.TrimRight(mbpsStr, ".")
	return MaxXput{Mbps: mbpsStr, IOPS: ""}, nil
}
