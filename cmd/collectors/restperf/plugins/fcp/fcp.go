package fcp

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	constant "github.com/netapp/harvest/v2/pkg/const"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"math"
	"strconv"
	"strings"
)

const (
	readPercent  = "read_percent"
	writePercent = "write_percent"
	utilPercent  = "util_percent"
)

type Fcp struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Fcp{AbstractPlugin: p}
}

func (f *Fcp) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

	var rx, tx, up, read, write *matrix.Metric
	var err error
	data := dataMap[f.Object]

	if read = data.GetMetric("read_data"); read == nil {
		return nil, nil, errs.New(errs.ErrNoMetric, "read_data")
	}

	if write = data.GetMetric("write_data"); write == nil {
		return nil, nil, errs.New(errs.ErrNoMetric, "write_data")
	}

	if rx = data.GetMetric(readPercent); rx == nil {
		if rx, err = data.NewMetricFloat64(readPercent); err == nil {
			rx.SetProperty("raw")
		} else {
			return nil, nil, err
		}

	}
	if tx = data.GetMetric(writePercent); tx == nil {
		if tx, err = data.NewMetricFloat64(writePercent); err == nil {
			tx.SetProperty("raw")
		} else {
			return nil, nil, err
		}
	}

	if up = data.GetMetric(utilPercent); up == nil {
		if up, err = data.NewMetricFloat64(utilPercent); err == nil {
			up.SetProperty("raw")
		} else {
			return nil, nil, err
		}
	}

	for _, instance := range data.GetInstances() {

		instance.SetLabel("port", strings.TrimPrefix(instance.GetLabel("port"), "port."))

		var speed int
		var s string
		var err error

		if speed, err = strconv.Atoi(instance.GetLabel("speed")); err != nil {
			f.Logger.Debug().Msgf("skip, can't convert speed (%s) to numeric", s)
		}

		if speed != 0 {

			var rxBytes, txBytes, rxPercent, txPercent float64
			var rxOk, txOk bool

			if rxBytes, rxOk = write.GetValueFloat64(instance); rxOk {
				rxPercent = rxBytes / float64(speed)
				err := rx.SetValueFloat64(instance, rxPercent)
				if err != nil {
					f.Logger.Error().Err(err).Msg("error")
				}
			}

			if txBytes, txOk = read.GetValueFloat64(instance); txOk {
				txPercent = txBytes / float64(speed)
				err := tx.SetValueFloat64(instance, txPercent)
				if err != nil {
					f.Logger.Error().Err(err).Msg("error")
				}
			}

			if rxOk || txOk {
				err := up.SetValueFloat64(instance, math.Max(rxPercent, txPercent))
				if err != nil {
					f.Logger.Error().Err(err).Msg("error")
				}
			}
		}
	}
	return nil, nil, nil
}

func (f *Fcp) GetGeneratedMetrics() []plugin.CustomMetric {
	return []plugin.CustomMetric{
		{
			Name:         readPercent,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Bytes received percentage.",
		},
		{
			Name:         writePercent,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Bytes sent percentage.",
		},
		{
			Name:         utilPercent,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Max of Bytes received percentage and Bytes sent percentage.",
		},
	}
}
