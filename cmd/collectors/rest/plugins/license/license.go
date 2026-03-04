package license

import (
	"log/slog"

	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
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

	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	for _, k := range []string{"license", "scope", "owner"} {
		instanceKeys.NewChildS("", k)
	}
	instanceLabels := exportOptions.NewChildS("instance_labels", "")
	for _, il := range []string{
		"description", "entitlement_action", "entitlement_risk",
		"serial_number", "installed_license", "host_id",
		"active", "evaluation", "compliance_state",
	} {
		instanceLabels.NewChildS("", il)
	}
	l.data.SetExportOptions(exportOptions)

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
				if m := l.data.GetMetric("expiry_time"); m != nil {
					m.SetValueFloat64(newInstance, collectors.HandleTimestamp(expiryStr)*1000)
				}
			}
			if lic.Get("capacity").Exists() {
				maxSize := lic.Get("capacity.maximum_size").Float()
				usedSize := lic.Get("capacity.used_size").Float()
				if m := l.data.GetMetric("capacity_maximum_size"); m != nil {
					m.SetValueFloat64(newInstance, maxSize)
				}
				if m := l.data.GetMetric("capacity_used_size"); m != nil {
					m.SetValueFloat64(newInstance, usedSize)
				}
				if m := l.data.GetMetric("capacity_used_percent"); m != nil && maxSize > 0 {
					m.SetValueFloat64(newInstance, usedSize/maxSize*100)
				}
			}
		}
	}
	return []*matrix.Matrix{l.data}, nil, nil
}
