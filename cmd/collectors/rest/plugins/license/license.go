package license

import (
	"log/slog"

	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

const licenseMatrix = "license"

type License struct {
	*plugin.AbstractPlugin
	data *matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &License{AbstractPlugin: p}
}

func (l *License) Init(_ conf.Remote) error {
	if err := l.InitAbc(); err != nil {
		return err
	}

	l.data = matrix.New(l.Parent+".License", licenseMatrix, licenseMatrix)

	l.data.SetExportOptions(matrix.NewExportOptionsWithLabels(
		[]string{"license", "scope", "owner", "serial_number"},
		[]string{
			"description", "entitlement_action", "entitlement_risk",
			"installed_license", "host_id",
			"active", "evaluation", "compliance_state",
		},
	))

	for _, m := range []string{"capacity_maximum_size", "capacity_used_size", "capacity_used_percent", "expiry_time"} {
		if _, err := l.data.NewMetricFloat64(m, m); err != nil {
			l.SLogger.Error("add metric", slogx.Err(err), slog.String("metric", m))
			return err
		}
	}

	return nil
}

func (l *License) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[l.Object]

	l.data.PurgeInstances()
	l.data.Reset()
	l.data.SetGlobalLabels(data.GetGlobalLabels())

	expiryTimeMetric := l.data.MustGetMetric("expiry_time")
	capacityMaxSizeMetric := l.data.MustGetMetric("capacity_maximum_size")
	capacityUsedSizeMetric := l.data.MustGetMetric("capacity_used_size")
	capacityUsedPercentMetric := l.data.MustGetMetric("capacity_used_percent")

	for _, instance := range data.GetInstances() {
		licenseName := instance.GetLabel("license")
		scope := instance.GetLabel("scope")
		description := instance.GetLabel("description")
		entitlementAction := instance.GetLabel("entitlement_action")
		entitlementRisk := instance.GetLabel("entitlement_risk")

		rawLicenses := instance.GetLabel("licenses")
		if rawLicenses == "" {
			continue
		}

		licensesData := gjson.Result{Type: gjson.JSON, Raw: "[" + rawLicenses + "]"}
		for _, lic := range licensesData.Array() {
			owner := lic.Get("owner").ClonedString()
			serialNumber := lic.Get("serial_number").ClonedString()

			instanceKey := licenseName + "_" + scope + "_" + owner + "_" + serialNumber

			newInstance, err := l.data.NewInstance(instanceKey)
			if err != nil {
				l.SLogger.Error("Failed to create instance", slogx.Err(err), slog.String("key", instanceKey))
				continue
			}

			newInstance.SetLabel("license", licenseName)
			newInstance.SetLabel("scope", scope)
			newInstance.SetLabel("description", description)
			newInstance.SetLabel("entitlement_action", entitlementAction)
			newInstance.SetLabel("entitlement_risk", entitlementRisk)

			newInstance.SetLabel("owner", owner)
			newInstance.SetLabel("serial_number", serialNumber)
			newInstance.SetLabel("installed_license", lic.Get("installed_license").ClonedString())
			newInstance.SetLabel("host_id", lic.Get("host_id").ClonedString())
			newInstance.SetLabel("active", lic.Get("active").ClonedString())
			newInstance.SetLabel("evaluation", lic.Get("evaluation").ClonedString())
			newInstance.SetLabel("compliance_state", lic.Get("compliance.state").ClonedString())

			if expiryStr := lic.Get("expiry_time").ClonedString(); expiryStr != "" {
				expiryTimeMetric.SetValueFloat64(newInstance, collectors.HandleTimestamp(expiryStr)*1000)
			}
			if lic.Get("capacity").Exists() {
				maxSize := lic.Get("capacity.maximum_size").Float()
				usedSize := lic.Get("capacity.used_size").Float()
				capacityMaxSizeMetric.SetValueFloat64(newInstance, maxSize)
				capacityUsedSizeMetric.SetValueFloat64(newInstance, usedSize)
				if maxSize > 0 {
					capacityUsedPercentMetric.SetValueFloat64(newInstance, usedSize/maxSize*100)
				}
			}
		}
	}
	return []*matrix.Matrix{l.data}, nil, nil
}
