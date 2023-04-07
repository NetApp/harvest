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
	errr                            AlertSeverity = "error"
	warning                         AlertSeverity = "warning"
	diskHealthMatrix                              = "health_disk"
	shelfHealthMatrix                             = "health_shelf"
	supportHealthMatrix                           = "health_support"
	nodeHealthMatrix                              = "health_node"
	networkEthernetPortHealthMatrix               = "health_network_ethernet_port"
	networkFCPortHealthMatrix                     = "health_network_fc_port"
	networkInterfaceHealthMatrix                  = "health_network_interface"
	volumeRansomwareHealthMatrix                  = "health_volume_ransomware"
	volumeMoveHealthMatrix                        = "health_volume_move"
	licenseHealthMatrix                           = "health_license"
	severityLabel                                 = "severity"
	defaultDataPollDuration                       = 3 * time.Minute
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
	mats := []string{diskHealthMatrix, shelfHealthMatrix, supportHealthMatrix, nodeHealthMatrix,
		networkEthernetPortHealthMatrix, networkFCPortHealthMatrix, networkInterfaceHealthMatrix,
		volumeRansomwareHealthMatrix, volumeMoveHealthMatrix, licenseHealthMatrix}
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
	v.collectNodeAlerts()
	v.collectNetworkEthernetPortAlerts()
	v.collectNetworkFCPortAlerts()
	v.collectNetworkInterfacesAlerts()
	v.collectVolumeRansomwareAlerts()
	v.collectVolumeMoveAlerts()
	v.collectLicenseAlerts()

	result := make([]*matrix.Matrix, 0, len(v.data))

	for _, value := range v.data {
		result = append(result, value)
	}
	return result, nil
}

func (v *Health) collectLicenseAlerts() {
	var (
		instance *matrix.Instance
	)

	records, err := v.getNonCompliantLicense()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			v.Logger.Debug().Err(err).Msg("API not found")
		} else {
			v.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return
	}
	mat := v.data[licenseHealthMatrix]
	for _, record := range records {
		name := record.Get("name").String()
		scope := record.Get("scope").String()
		state := record.Get("state").String()
		instance, err = mat.NewInstance(name)
		if err != nil {
			v.Logger.Warn().Str("key", name).Msg("error while creating instance")
			continue
		}
		instance.SetLabel("name", name)
		instance.SetLabel("scope", scope)
		instance.SetLabel("state", state)
		instance.SetLabel(severityLabel, string(errr))

		v.setAlertMetric(mat, instance)
	}
}

func (v *Health) collectVolumeMoveAlerts() {
	var (
		instance *matrix.Instance
	)

	records, err := v.getMoveFailedVolumes()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			v.Logger.Debug().Err(err).Msg("API not found")
		} else {
			v.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return
	}
	mat := v.data[volumeMoveHealthMatrix]
	for _, record := range records {
		uuid := record.Get("uuid").String()
		volume := record.Get("name").String()
		svm := record.Get("svm.name").String()
		movementState := record.Get("movement.state").String()
		instance, err = mat.NewInstance(uuid)
		if err != nil {
			v.Logger.Warn().Str("key", uuid).Msg("error while creating instance")
			continue
		}
		instance.SetLabel("movement_state", movementState)
		instance.SetLabel("svm", svm)
		instance.SetLabel("volume", volume)
		instance.SetLabel(severityLabel, string(warning))

		v.setAlertMetric(mat, instance)
	}
}

func (v *Health) collectVolumeRansomwareAlerts() {
	var (
		instance *matrix.Instance
	)
	clusterVersion := v.client.Cluster().GetVersion()
	ontapVersion, err := goversion.NewVersion(clusterVersion)
	if err != nil {
		v.Logger.Error().Err(err).
			Str("version", clusterVersion).
			Msg("Failed to parse version")
		return
	}
	version910 := "9.10"
	version910After, err := goversion.NewVersion(version910)
	if err != nil {
		v.Logger.Error().Err(err).
			Str("version", version910).
			Msg("Failed to parse version")
		return
	}

	if ontapVersion.LessThan(version910After) {
		return
	}
	records, err := v.getRansomwareVolumes()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			v.Logger.Debug().Err(err).Msg("API not found")
		} else {
			v.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return
	}
	mat := v.data[volumeRansomwareHealthMatrix]
	for _, record := range records {
		uuid := record.Get("uuid").String()
		volume := record.Get("name").String()
		antiRansomwareAttackProbability := record.Get("anti_ransomware.attack_probability").String()
		instance, err = mat.NewInstance(uuid)
		if err != nil {
			v.Logger.Warn().Str("key", uuid).Msg("error while creating instance")
			continue
		}
		instance.SetLabel("anti_ransomware_attack_probability", antiRansomwareAttackProbability)

		instance.SetLabel("volume", volume)
		instance.SetLabel(severityLabel, string(errr))

		v.setAlertMetric(mat, instance)
	}
}

func (v *Health) collectNetworkInterfacesAlerts() {
	var (
		instance *matrix.Instance
	)
	records, err := v.getLIFs()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			v.Logger.Debug().Err(err).Msg("API not found")
		} else {
			v.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return
	}
	mat := v.data[networkInterfaceHealthMatrix]
	for _, record := range records {
		uuid := record.Get("uuid").String()
		lif := record.Get("name").String()
		isHome := record.Get("location.is_home").String()
		instance, err = mat.NewInstance(uuid)
		if err != nil {
			v.Logger.Warn().Str("key", uuid).Msg("error while creating instance")
			continue
		}
		instance.SetLabel("isHome", isHome)
		instance.SetLabel("lif", lif)
		instance.SetLabel(severityLabel, string(warning))

		v.setAlertMetric(mat, instance)
	}
}

func (v *Health) collectNetworkFCPortAlerts() {
	var (
		instance *matrix.Instance
	)
	records, err := v.getFCPorts()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			v.Logger.Debug().Err(err).Msg("API not found")
		} else {
			v.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return
	}
	mat := v.data[networkFCPortHealthMatrix]
	for _, record := range records {
		uuid := record.Get("uuid").String()
		nodeName := record.Get("node.name").String()
		port := record.Get("name").String()
		state := record.Get("state").String()
		instance, err = mat.NewInstance(uuid)
		if err != nil {
			v.Logger.Warn().Str("key", uuid).Msg("error while creating instance")
			continue
		}
		instance.SetLabel("node", nodeName)
		instance.SetLabel("state", state)
		instance.SetLabel("port", port)
		instance.SetLabel(severityLabel, string(errr))

		v.setAlertMetric(mat, instance)
	}
}

func (v *Health) collectNetworkEthernetPortAlerts() {
	var (
		instance *matrix.Instance
	)
	records, err := v.getEthernetPorts()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			v.Logger.Debug().Err(err).Msg("API not found")
		} else {
			v.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return
	}
	mat := v.data[networkEthernetPortHealthMatrix]
	for _, record := range records {
		uuid := record.Get("uuid").String()
		port := record.Get("name").String()
		nodeName := record.Get("node.name").String()
		portType := record.Get("type").String()
		state := record.Get("state").String()
		instance, err = mat.NewInstance(uuid)
		if err != nil {
			v.Logger.Warn().Str("key", uuid).Msg("error while creating instance")
			continue
		}
		instance.SetLabel("node", nodeName)
		instance.SetLabel("state", state)
		instance.SetLabel("port", port)
		instance.SetLabel("type", portType)
		instance.SetLabel(severityLabel, string(errr))

		v.setAlertMetric(mat, instance)
	}
}

func (v *Health) collectNodeAlerts() {
	var (
		instance *matrix.Instance
	)
	records, err := v.getNodes()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			v.Logger.Debug().Err(err).Msg("API not found")
		} else {
			v.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return
	}
	mat := v.data[nodeHealthMatrix]
	for _, record := range records {
		nodeName := record.Get("node").String()

		instance, err = mat.NewInstance(nodeName)
		if err != nil {
			v.Logger.Warn().Str("key", nodeName).Msg("error while creating instance")
			continue
		}
		instance.SetLabel("node", nodeName)
		instance.SetLabel("healthy", "false")
		instance.SetLabel(severityLabel, string(errr))

		v.setAlertMetric(mat, instance)
	}
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

			v.setAlertMetric(mat, instance)
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

		v.setAlertMetric(mat, instance)
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

		v.setAlertMetric(mat, instance)
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

func (v *Health) getNodes() ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)

	fields := []string{"health"}
	query := "api/private/cli/node"
	href := rest.BuildHref(query, strings.Join(fields, ","), []string{"health=false"}, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(v.client, href, v.Logger); err != nil {
		return nil, err
	}
	return result, nil
}

func (v *Health) getRansomwareVolumes() ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)

	query := "api/storage/volumes"
	href := rest.BuildHref(query, "", []string{"anti_ransomware.state=enabled", "anti_ransomware.attack_probability=low|moderate|high"}, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(v.client, href, v.Logger); err != nil {
		return nil, err
	}
	return result, nil
}

func (v *Health) getNonCompliantLicense() ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)

	query := "api/cluster/licensing/licenses"
	fields := []string{"name,scope,state"}
	href := rest.BuildHref(query, strings.Join(fields, ","), []string{"state=noncompliant"}, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(v.client, href, v.Logger); err != nil {
		return nil, err
	}
	return result, nil
}

func (v *Health) getMoveFailedVolumes() ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)

	query := "api/storage/volumes"
	fields := []string{"uuid,name,movement.state,svm"}
	href := rest.BuildHref(query, strings.Join(fields, ","), []string{"movement.state=cutover_wait|failed|cutover_pending"}, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(v.client, href, v.Logger); err != nil {
		return nil, err
	}
	return result, nil
}

func (v *Health) getLIFs() ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)

	query := "api/network/ip/interfaces"
	href := rest.BuildHref(query, "", []string{"location.is_home=false"}, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(v.client, href, v.Logger); err != nil {
		return nil, err
	}
	return result, nil
}

func (v *Health) getFCPorts() ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)

	fields := []string{"name,node"}
	query := "api/network/fc/ports"
	href := rest.BuildHref(query, strings.Join(fields, ","), []string{"enabled=true", "state=offlined_by_system"}, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(v.client, href, v.Logger); err != nil {
		return nil, err
	}
	return result, nil
}

func (v *Health) getEthernetPorts() ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)

	fields := []string{"name,node"}
	query := "api/network/ethernet/ports"
	href := rest.BuildHref(query, strings.Join(fields, ","), []string{"enabled=true", "state=down"}, "", "", "", "", query)

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

func (v *Health) setAlertMetric(mat *matrix.Matrix, instance *matrix.Instance) {
	var err error
	m := mat.GetMetric("alerts")
	if m == nil {
		if m, err = mat.NewMetricFloat64("alerts"); err != nil {
			v.Logger.Warn().Err(err).Str("key", "alerts").Msg("error while creating metric")
			return
		}
	}
	if err = m.SetValueFloat64(instance, 1); err != nil {
		v.Logger.Error().Err(err).Str("metric", "alerts").Msg("Unable to set value on metric")
	}
}
