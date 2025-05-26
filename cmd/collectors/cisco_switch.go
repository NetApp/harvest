package collectors

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

const (
	adminUp = "admin_up"
	labels  = "labels"
)

var versionRegex = regexp.MustCompile(`\d+(\.\d+)+\(([0-9]+?)\)`)

func NewCiscoSwitch(p *plugin.AbstractPlugin) plugin.Plugin {
	return &CiscoSwitch{AbstractPlugin: p}
}

type SwitchData struct {
	isCiscoSwitch bool
	adminState    bool
	osVersion     string
}

type CiscoSwitch struct {
	*plugin.AbstractPlugin
	client *rest.Client
	data   *matrix.Matrix
}

func (c *CiscoSwitch) Init(_ conf.Remote) error {

	var err error
	if err := c.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if c.client, err = rest.New(conf.ZapiPoller(c.ParentParams), timeout, c.Auth); err != nil {
		c.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	return nil
}

func (c *CiscoSwitch) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[c.Object]
	datacenter := data.GetGlobalLabels()["datacenter"]
	// metrics with cisco_interface prefix are available to handle cisco switch dashboard
	ciscoFormatedData := data.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
	ciscoFormatedData.UUID = c.Parent + ".Cisco_interface"
	ciscoFormatedData.Object = "cisco_interface"
	ciscoFormatedData.Identifier = "cisco_interface"
	globalLabels := ciscoFormatedData.GetGlobalLabels()
	clear(globalLabels)
	ciscoFormatedData.SetGlobalLabel("datacenter", datacenter)

	if _, err := ciscoFormatedData.NewMetricFloat64(adminUp); err != nil {
		return nil, nil, err
	}

	c.data = matrix.New(c.Parent+".Cisco_interface", "cisco_switch", "cisco_switch")
	c.data.SetGlobalLabels(ciscoFormatedData.GetGlobalLabels())

	// create cisco_switch_labels metric
	err := matrix.CreateMetric(labels, c.data)
	if err != nil {
		c.SLogger.Warn("error while creating metric", slogx.Err(err), slog.String("key", labels))
	}

	switchMap := c.collectSwitches()
	c.GenerateCiscoMetrics(ciscoFormatedData, data, switchMap)
	if len(ciscoFormatedData.GetInstances()) > 0 {
		return []*matrix.Matrix{ciscoFormatedData, c.data}, nil, nil
	}
	return nil, nil, nil
}

// CollectSwitches is here consumed from both REST and KeyPerf
func (c *CiscoSwitch) collectSwitches() map[string]SwitchData {
	switchMap := make(map[string]SwitchData)
	var adminState bool
	var isCiscoSwitch bool
	var osVersion string
	fields := []string{"name", "version", "monitoring.enabled"}
	query := "api/network/ethernet/switches"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(DefaultBatchSize).
		Build()

	result, err := rest.FetchAll(c.client, href)
	if err != nil {
		c.SLogger.Error("failed to fetch data", slog.String("href", href), slog.Any("err", err))
		return switchMap
	}

	for _, r := range result {
		adminState = false
		isCiscoSwitch = false
		osVersion = ""
		switchName := r.Get("name").String()
		version := r.Get("version")
		if state := r.Get("monitoring.enabled"); state.Exists() {
			adminState = state.Bool()
		}
		if version.Exists() && strings.Contains(version.String(), "NX-OS") {
			isCiscoSwitch = true
			allMatches := versionRegex.FindAllStringSubmatch(version.String(), -1)
			for _, match := range allMatches {
				m := match[0]
				if m == "" {
					continue
				}
				osVersion = m
			}
		}
		switchMap[switchName] = SwitchData{isCiscoSwitch: isCiscoSwitch, adminState: adminState, osVersion: osVersion}
	}
	return switchMap
}

func (c *CiscoSwitch) GenerateCiscoMetrics(ciscoFormatedData *matrix.Matrix, data *matrix.Matrix, switchMap map[string]SwitchData) {
	for key, ciscoData := range ciscoFormatedData.GetInstances() {
		switchName := ciscoData.GetLabel("switch")
		switchData := switchMap[switchName]
		if switchData.isCiscoSwitch {
			c.SLogger.Info("Cisco Switches have been found")
			data.RemoveInstance(key)

			// Only Rest template plugin can add below 2 metrics not KeyPerf
			if data.UUID == "Rest" {
				// Add cisco_interface_admin_up metric
				adminUpMetric := ciscoFormatedData.GetMetric(adminUp)
				if switchData.adminState {
					adminUpMetric.SetValueFloat64(ciscoData, 1)
				} else {
					adminUpMetric.SetValueFloat64(ciscoData, 0)
				}

				// Add cisco_switch_labels metric
				labelsMetric := c.data.GetMetric(labels)
				labelsInstance, err := c.data.NewInstance(switchName)
				if err != nil {
					c.SLogger.Error("", slogx.Err(err), slog.String("instanceKey", switchName))
					continue
				}
				labelsInstance.SetLabel("switch", ciscoData.GetLabel("switch"))
				labelsMetric.SetLabel("osVersion", switchData.osVersion)
				labelsMetric.SetValueFloat64(labelsInstance, 1)
			}
		} else {
			c.SLogger.Info("Cisco Switches have not been found")
			ciscoFormatedData.RemoveInstance(key)
		}
	}
}
