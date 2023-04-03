package health

import (
	"fmt"
	goversion "github.com/hashicorp/go-version"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

type AlertSeverity string

const (
	errr                    AlertSeverity = "error"
	warning                 AlertSeverity = "warning"
	diskHealthMatrix                      = "health_disk"
	shelfHealthMatrix                     = "health_shelf"
	supportHealthMatrix                   = "health_support"
	severityLabel                         = "severity"
	defaultDataPollDuration               = 3 * time.Minute
)

type Health struct {
	*plugin.AbstractPlugin
	client         *rest.Client
	data           map[string]*matrix.Matrix
	lastFilterTime int64
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
	mats := []string{diskHealthMatrix, shelfHealthMatrix, supportHealthMatrix}
	for _, m := range mats {
		if err := v.initMatrix(m); err != nil {
			return err
		}
	}
	return nil
}

func (v *Health) initMatrix(name string) error {
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
	v.collectShelfAlerts()
	v.collectSupportAlerts()

	result := make([]*matrix.Matrix, 0, len(v.data))

	for _, value := range v.data {
		result = append(result, value)
	}
	return result, nil
}

func (v *Health) collectShelfAlerts() {
	var (
		instance *matrix.Instance
	)
	records, err := v.getShelves()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			v.Logger.Debug().Err(err).Msg("API not found")
		} else {
			v.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return
	}
	mat := v.data[shelfHealthMatrix]
	for _, record := range records {
		shelf := record.Get("shelf").String()
		errorType := record.Get("error_type").String()
		errorSeverity := record.Get("error_severity").String()
		errorText := record.Get("error_text").String()

		//errorSeverity possible values are unknown|notice|warning|error|critical
		if errorSeverity == "error" || errorSeverity == "critical" || errorSeverity == "warning" {
			instance, err = mat.NewInstance(shelf)
			if err != nil {
				v.Logger.Warn().Str("key", shelf).Msg("error while creating instance")
				continue
			}
			instance.SetLabel("shelf", shelf)
			instance.SetLabel("error_type", errorType)
			instance.SetLabel("error_text", errorText)
			if errorSeverity == "error" || errorSeverity == "critical" {
				instance.SetLabel(severityLabel, string(errr))
			} else if errorSeverity == "warning" {
				instance.SetLabel(severityLabel, string(warning))
			}

			m := mat.GetMetric("alerts")
			if m == nil {
				if m, err = mat.NewMetricFloat64("alerts"); err != nil {
					v.Logger.Warn().Err(err).Str("key", "alerts").Msg("error while creating metric")
					continue
				}
			}
			if err = m.SetValueFloat64(instance, 1); err != nil {
				v.Logger.Error().Err(err).Str("metric", "alerts").Msg("Unable to set value on metric")
			}
		}
	}
}

func (v *Health) collectSupportAlerts() {
	var (
		instance *matrix.Instance
	)
	clusterTime, err := collectors.GetClusterTime(v.client, "", v.Logger)
	if err != nil {
		v.Logger.Error().Err(err).Msg("Failed to collect cluster time")
		return
	}
	toTime := clusterTime.Unix()
	timeFilter := v.getTimeStampFilter(clusterTime)
	addFilter := []string{"suppress=false"}
	filter := append(addFilter, timeFilter)

	records, err := v.getSupportAlerts(filter)
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			v.Logger.Debug().Err(err).Msg("API not found")
		} else {
			v.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return
	}
	mat := v.data[supportHealthMatrix]
	for index, record := range records {
		nodeName := record.Get("node.name").String()
		monitor := record.Get("monitor").String()
		name := record.Get("name").String()
		resource := record.Get("resource").String()
		reason := record.Get("cause.message").String()
		correctiveAction := record.Get("corrective_action.message").String()
		instance, err = mat.NewInstance(strconv.Itoa(index))
		if err != nil {
			v.Logger.Warn().Int("key", index).Msg("error while creating instance")
			continue
		}
		instance.SetLabel("node", nodeName)
		instance.SetLabel("monitor", monitor)
		instance.SetLabel("name", name)
		instance.SetLabel("resource", resource)
		instance.SetLabel("reason", reason)
		instance.SetLabel("correctiveAction", correctiveAction)
		instance.SetLabel(severityLabel, string(warning))

		m := mat.GetMetric("alerts")
		if m == nil {
			if m, err = mat.NewMetricFloat64("alerts"); err != nil {
				v.Logger.Warn().Err(err).Str("key", "alerts").Msg("error while creating metric")
				continue
			}
		}
		if err = m.SetValueFloat64(instance, 1); err != nil {
			v.Logger.Error().Err(err).Str("metric", "alerts").Msg("Unable to set value on metric")
		}

	}
	// update lastFilterTime to current cluster time
	v.lastFilterTime = toTime
}

func (v *Health) collectDiskAlerts() {
	var (
		instance *matrix.Instance
	)
	records, err := v.getDisks()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			v.Logger.Debug().Err(err).Msg("API not found")
		} else {
			v.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return
	}
	mat := v.data[diskHealthMatrix]
	for _, record := range records {
		name := record.Get("name").String()
		containerType := record.Get("container_type").String()
		instance, err = mat.NewInstance(name)
		if err != nil {
			v.Logger.Warn().Str("key", name).Msg("error while creating instance")
			continue
		}
		instance.SetLabel("disk", name)
		instance.SetLabel("container_type", containerType)
		if containerType == "broken" {
			instance.SetLabel(severityLabel, string(errr))
		} else if containerType == "unassigned" {
			instance.SetLabel(severityLabel, string(warning))
		}
		m := mat.GetMetric("alerts")
		if m == nil {
			if m, err = mat.NewMetricFloat64("alerts"); err != nil {
				v.Logger.Warn().Err(err).Str("key", "alerts").Msg("error while creating metric")
				continue
			}
		}
		if err = m.SetValueFloat64(instance, 1); err != nil {
			v.Logger.Error().Err(err).Str("metric", "alerts").Msg("Unable to set value on metric")
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

func (v *Health) getShelves() ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)

	fields := []string{"error_type", "error_severity", "error_text"}
	query := "api/private/cli/storage/shelf"
	href := rest.BuildHref(query, strings.Join(fields, ","), nil, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(v.client, href, v.Logger); err != nil {
		return nil, err
	}
	return result, nil
}

func (v *Health) getSupportAlerts(filter []string) ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)
	query := "api/private/support/alerts"
	href := rest.BuildHref(query, "", filter, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(v.client, href, v.Logger); err != nil {
		return nil, err
	}

	return result, nil
}

// returns time filter (clustertime - polldata duration)
func (v *Health) getTimeStampFilter(clusterTime time.Time) string {
	fromTime := v.lastFilterTime
	// check if this is the first request
	if v.lastFilterTime == 0 {
		// if first request fetch cluster time
		dataDuration, err := GetDataInterval(v.ParentParams, defaultDataPollDuration)
		if err != nil {
			v.Logger.Warn().Err(err).
				Str("defaultDataPollDuration", defaultDataPollDuration.String()).
				Msg("Failed to parse duration. using default")
		}
		fromTime = clusterTime.Add(-dataDuration).Unix()
	}
	return fmt.Sprintf("time=>=%d", fromTime)
}

// GetDataInterval fetch pollData interval
func GetDataInterval(param *node.Node, defaultInterval time.Duration) (time.Duration, error) {
	var dataIntervalStr string
	var durationVal time.Duration
	var err error
	schedule := param.GetChildS("schedule")
	if schedule != nil {
		dataInterval := schedule.GetChildS("data")
		if dataInterval != nil {
			dataIntervalStr = dataInterval.GetContentS()
			if durationVal, err = time.ParseDuration(dataIntervalStr); err == nil {
				return durationVal, nil
			}
			return defaultInterval, err
		}
	}
	return defaultInterval, nil
}
