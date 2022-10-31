package tenant

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type Tenant struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Tenant{AbstractPlugin: p}
}

func (t *Tenant) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		used, quota, usedPercent matrix.Metric
		err                      error
	)

	if used = data.GetMetric("dataBytes"); used == nil {
		return nil, errs.New(errs.ErrNoMetric, "logical_used")
	}

	if quota = data.GetMetric("policy.quotaObjectBytes"); quota == nil {
		return nil, errs.New(errs.ErrNoMetric, "logical_quota")
	}

	if usedPercent = data.GetMetric("used_percent"); usedPercent == nil {
		if usedPercent, err = data.NewMetricFloat64("used_percent"); err == nil {
			usedPercent.SetProperty("raw")
		} else {
			return nil, err
		}
	}

	for _, instance := range data.GetInstances() {

		var (
			usedBytes, quotaBytes, percentage float64
			usedOK, quotaOK                   bool
		)

		usedBytes, usedOK = used.GetValueFloat64(instance)
		quotaBytes, quotaOK = quota.GetValueFloat64(instance)
		if (usedOK) && (quotaOK) {
			percentage = usedBytes / quotaBytes * 100
			if quotaBytes == 0 {
				percentage = 0
			}
			err := usedPercent.SetValueFloat64(instance, percentage)
			if err != nil {
				t.Logger.Error().Err(err).Float64("percentage", percentage).Msg("failed to set percentage")
			}
		}
	}

	return nil, nil
}
