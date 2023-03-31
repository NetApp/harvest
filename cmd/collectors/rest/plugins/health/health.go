package health

import (
	goversion "github.com/hashicorp/go-version"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

type AlertSeverity string

const (
	errr    AlertSeverity = "error"
	warning AlertSeverity = "warning"
)

const diskHealthMatrix = "health_disk"
const severityLabel = "severity"

type Health struct {
	*plugin.AbstractPlugin
	client *rest.Client
	data   map[string]*matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Health{AbstractPlugin: p}
}

var metrics = []string{
	"alerts",
}

func (v *Health) Init() error {

	var err error

	if err = v.InitAbc(); err != nil {
		return err
	}

	if err = v.initAllMatrix(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if v.client, err = rest.New(conf.ZapiPoller(v.ParentParams), timeout, v.Auth); err != nil {
		v.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = v.client.Init(5); err != nil {
		return err
	}

	return nil
}

func (v *Health) initAllMatrix() error {
	v.data = make(map[string]*matrix.Matrix)
	if err := v.initMatrix(diskHealthMatrix); err != nil {
		return err
	}
	return nil
}

func (v *Health) initMatrix(name string) error {
	v.data = make(map[string]*matrix.Matrix)
	v.data[name] = matrix.New(v.Parent+name, name, name)
	for _, v1 := range v.data {
		v1.SetExportOptions(matrix.DefaultExportOptions())
	}
	for _, k := range metrics {
		err := matrix.CreateMetric(k, v.data[name])
		if err != nil {
			v.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
			return err
		}
	}
	return nil
}

func (v *Health) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	data := dataMap[v.Object]
	clusterVersion := v.client.Cluster().GetVersion()
	ontapVersion, err := goversion.NewVersion(clusterVersion)
	if err != nil {
		v.Logger.Error().Err(err).
			Str("version", clusterVersion).
			Msg("Failed to parse version")
		return nil, nil
	}
	version96 := "9.6"
	version96After, err := goversion.NewVersion(version96)
	if err != nil {
		v.Logger.Error().Err(err).
			Str("version", version96).
			Msg("Failed to parse version")
		return nil, nil
	}

	if ontapVersion.LessThan(version96After) {
		return nil, nil
	}

	// Purge and reset data
	// remove all metrics as analytics label may change over time
	err = v.initAllMatrix()
	if err != nil {
		v.Logger.Warn().Err(err).Msg("error while init matrix")
		return nil, err
	}
	for k := range v.data {
		// Set all global labels if already not exist
		v.data[k].SetGlobalLabels(data.GetGlobalLabels())
	}

	v.collectDiskAlerts()

	result := make([]*matrix.Matrix, 0, len(v.data))

	for _, value := range v.data {
		result = append(result, value)
	}
	return result, nil
}

func (v *Health) collectDiskAlerts() {
	var (
		instance *matrix.Instance
	)
	if records, err := v.getDisks(); err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			v.Logger.Debug().Err(err).Msg("API not found")
		} else {
			v.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
	} else {
		dMatrix := v.data[diskHealthMatrix]
		for _, record := range records {
			name := record.Get("name").String()
			containerType := record.Get("container_type").String()
			instance, err = dMatrix.NewInstance(name)
			if err != nil {
				v.Logger.Warn().Str("key", name).Msg("error while creating instance")
				continue
			}
			instance.SetLabel("disk", name)
			instance.SetLabel("container_type", containerType)
			if containerType == "broken" {
				instance.SetLabel(severityLabel, string(errr))
			} else {
				instance.SetLabel(severityLabel, string(warning))
			}
			m := dMatrix.GetMetric("alerts")
			if m == nil {
				if m, err = dMatrix.NewMetricFloat64("alerts"); err != nil {
					v.Logger.Warn().Err(err).Str("key", "alerts").Msg("error while creating metric")
					continue
				}
			}
			if err := m.SetValueFloat64(instance, 1); err != nil {
				v.Logger.Error().Err(err).Str("metric", "alerts").Msg("Unable to set value on metric")
			}
		}
	}
}

func (v *Health) getDisks() ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)

	fields := []string{"name", "container_type"}
	query := "api/storage/disks"
	href := rest.BuildHref(query, strings.Join(fields, ","), []string{"container_type=broken|unassigned"}, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(v.client, href, v.Logger); err != nil {
		return nil, err
	}
	return result, nil
}
