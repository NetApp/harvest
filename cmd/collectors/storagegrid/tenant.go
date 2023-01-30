package storagegrid

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type Tenant struct {
	*plugin.AbstractPlugin
	sg *StorageGrid
}

func NewTenant(p *plugin.AbstractPlugin, s *StorageGrid) plugin.Plugin {
	return &Tenant{AbstractPlugin: p, sg: s}
}

func (t *Tenant) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		used, quota, usedPercent *matrix.Metric
		err                      error
		tenantNamesByID          map[string]string
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

	tenantNamesByID = make(map[string]string)
	for _, instance := range data.GetInstances() {

		var (
			usedBytes, quotaBytes, percentage float64
			usedOK, quotaOK                   bool
		)

		usedBytes, usedOK = used.GetValueFloat64(instance)
		quotaBytes, quotaOK = quota.GetValueFloat64(instance)
		if usedOK && quotaOK {
			percentage = usedBytes / quotaBytes * 100
			if quotaBytes == 0 {
				percentage = 0
			}
			err := usedPercent.SetValueFloat64(instance, percentage)
			if err != nil {
				t.Logger.Error().Err(err).Float64("percentage", percentage).Msg("failed to set percentage")
			}
		}

		id := instance.GetLabel("id")
		name := instance.GetLabel("tenant")
		if id != "" && name != "" {
			tenantNamesByID[id] = name
		}
	}

	promMetrics := t.collectPromMetrics(tenantNamesByID)
	return promMetrics, nil
}

func (t *Tenant) collectPromMetrics(tenantNamesByID map[string]string) []*matrix.Matrix {
	metrics := make(map[string]*matrix.Matrix)
	promMetrics := []string{
		"storagegrid_tenant_usage_data_bytes",
		"storagegrid_tenant_usage_quota_bytes",
	}

	for _, metric := range promMetrics {
		mat, err := t.sg.GetMetric(metric, metric[12:], tenantNamesByID)
		if err != nil {
			t.Logger.Error().Err(err).Str("metric", metric).Msg("Unable to get metric")
			continue
		}
		mat.Object = "storagegrid"
		metrics[metric] = mat
	}

	all := make([]*matrix.Matrix, 0, len(promMetrics))
	for _, m := range metrics {
		all = append(all, m)
	}
	return all
}
