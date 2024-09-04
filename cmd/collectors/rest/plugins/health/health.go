package health

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	constant "github.com/netapp/harvest/v2/pkg/const"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	goversion "github.com/netapp/harvest/v2/third_party/go-version"
	"github.com/tidwall/gjson"
	"strconv"
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
	haHealthMatrix                                = "health_ha"
	networkEthernetPortHealthMatrix               = "health_network_ethernet_port"
	networkFCPortHealthMatrix                     = "health_network_fc_port"
	lifHealthMatrix                               = "health_lif"
	volumeRansomwareHealthMatrix                  = "health_volume_ransomware"
	volumeMoveHealthMatrix                        = "health_volume_move"
	licenseHealthMatrix                           = "health_license"
	severityLabel                                 = "severity"
	defaultDataPollDuration                       = 3 * time.Minute
	alerts                                        = "alerts"
)

type Health struct {
	*plugin.AbstractPlugin
	client         *rest.Client
	data           map[string]*matrix.Matrix
	lastFilterTime int64
	previousData   map[string]*matrix.Matrix
	resolutionData map[string]*matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Health{AbstractPlugin: p}
}

var metrics = []string{
	alerts,
}

func (h *Health) Init() error {

	var err error

	if err := h.InitAbc(); err != nil {
		return err
	}

	if err := h.InitAllMatrix(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if h.client, err = rest.New(conf.ZapiPoller(h.ParentParams), timeout, h.Auth); err != nil {
		return err
	}

	return h.client.Init(5)
}

func (h *Health) InitAllMatrix() error {
	h.data = make(map[string]*matrix.Matrix)
	h.resolutionData = make(map[string]*matrix.Matrix)
	mats := []string{diskHealthMatrix, shelfHealthMatrix, supportHealthMatrix, nodeHealthMatrix,
		networkEthernetPortHealthMatrix, networkFCPortHealthMatrix, lifHealthMatrix,
		volumeRansomwareHealthMatrix, volumeMoveHealthMatrix, licenseHealthMatrix, haHealthMatrix}
	for _, m := range mats {
		if err := h.initMatrix(m, "", h.data); err != nil {
			return err
		}
		if err := h.initMatrix(m, "Resolution", h.resolutionData); err != nil {
			return err
		}
	}
	return nil
}

func (h *Health) initMatrix(name string, prefix string, inputMat map[string]*matrix.Matrix) error {
	inputMat[name] = matrix.New(h.Parent+name+prefix, name, name)
	for _, v1 := range h.data {
		v1.SetExportOptions(matrix.DefaultExportOptions())
	}
	for _, k := range metrics {
		err := matrix.CreateMetric(k, inputMat[name])
		if err != nil {
			h.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
			return err
		}
	}
	return nil
}

func (h *Health) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[h.Object]
	h.client.Metadata.Reset()
	clusterVersion := h.client.Cluster().GetVersion()
	ontapVersion, err := goversion.NewVersion(clusterVersion)
	if err != nil {
		h.Logger.Error().Err(err).
			Str("version", clusterVersion).
			Msg("Failed to parse version")
		return nil, nil, nil
	}
	version96 := "9.6"
	version96After, err := goversion.NewVersion(version96)
	if err != nil {
		h.Logger.Error().Err(err).
			Str("version", version96).
			Msg("Failed to parse version")
		return nil, nil, nil
	}

	if ontapVersion.LessThan(version96After) {
		return nil, nil, nil
	}

	// Purge and reset data
	// remove all metrics as analytics label may change over time
	err = h.InitAllMatrix()
	if err != nil {
		h.Logger.Warn().Err(err).Msg("error while init matrix")
		return nil, nil, err
	}
	for k := range h.data {
		// Set all global labels if already not exist
		h.data[k].SetGlobalLabels(data.GetGlobalLabels())
		h.resolutionData[k].SetGlobalLabels(data.GetGlobalLabels())
	}

	diskAlertCount := h.collectDiskAlerts()
	shelfAlertCount := h.collectShelfAlerts()
	supportAlertCount := h.collectSupportAlerts()
	nodeAlertCount := h.collectNodeAlerts()
	HAAlertCount := h.collectHAAlerts()
	networkEthernetPortAlertCount := h.collectNetworkEthernetPortAlerts()
	networkFcpPortAlertCount := h.collectNetworkFCPortAlerts()
	networkInterfaceAlertCount := h.collectNetworkInterfacesAlerts()
	volumeRansomwareAlertCount := h.collectVolumeRansomwareAlerts()
	volumeMoveAlertCount := h.collectVolumeMoveAlerts()
	licenseAlertCount := h.collectLicenseAlerts()

	resolutionInstancesCount := h.generateResolutionMetrics()

	result := make([]*matrix.Matrix, 0, len(h.data))

	for _, value := range h.data {
		result = append(result, value)
	}

	for _, value := range h.resolutionData {
		result = append(result, value)
	}
	h.Logger.Info().
		Int("numLicenseAlerts", licenseAlertCount).
		Int("numVolumeMoveAlerts", volumeMoveAlertCount).
		Int("numVolumeRansomwareAlerts", volumeRansomwareAlertCount).
		Int("numNetworkInterfaceAlerts", networkInterfaceAlertCount).
		Int("numNetworkFcpPortAlerts", networkFcpPortAlertCount).
		Int("numNetworkEthernetPortAlerts", networkEthernetPortAlertCount).
		Int("numHAAlerts", HAAlertCount).
		Int("numNodeAlerts", nodeAlertCount).
		Int("numSupportAlerts", supportAlertCount).
		Int("numShelfAlerts", shelfAlertCount).
		Int("numDiskAlerts", diskAlertCount).
		Int("numResolutionInstanceCount", resolutionInstancesCount).
		Msg("Collected")

	h.client.Metadata.PluginInstances = uint64(diskAlertCount + shelfAlertCount + supportAlertCount + nodeAlertCount + HAAlertCount + networkEthernetPortAlertCount + networkFcpPortAlertCount +
		networkInterfaceAlertCount + volumeRansomwareAlertCount + volumeMoveAlertCount + licenseAlertCount + resolutionInstancesCount)

	return result, h.client.Metadata, nil
}

func (h *Health) collectLicenseAlerts() int {
	var (
		instance *matrix.Instance
	)
	licenseAlertCount := 0
	records, err := h.getNonCompliantLicense()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			h.Logger.Debug().Err(err).Msg("API not found")
		} else {
			h.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return 0
	}
	mat := h.data[licenseHealthMatrix]
	for _, record := range records {
		name := record.Get("name").String()
		scope := record.Get("scope").String()
		state := record.Get("state").String()
		instance, err = mat.NewInstance(name)
		if err != nil {
			h.Logger.Warn().Str("key", name).Msg("error while creating instance")
			continue
		}
		licenseAlertCount++
		instance.SetLabel("name", name)
		instance.SetLabel("scope", scope)
		instance.SetLabel("state", state)
		instance.SetLabel(severityLabel, string(errr))

		h.setAlertMetric(mat, instance, 1)
	}

	return licenseAlertCount
}

func (h *Health) collectVolumeMoveAlerts() int {
	var (
		instance *matrix.Instance
	)
	volumeMoveAlertCount := 0
	records, err := h.getMoveFailedVolumes()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			h.Logger.Debug().Err(err).Msg("API not found")
		} else {
			h.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return 0
	}
	mat := h.data[volumeMoveHealthMatrix]
	for _, record := range records {
		uuid := record.Get("uuid").String()
		volume := record.Get("name").String()
		svm := record.Get("svm.name").String()
		movementState := record.Get("movement.state").String()
		instance, err = mat.NewInstance(uuid)
		if err != nil {
			h.Logger.Warn().Str("key", uuid).Msg("error while creating instance")
			continue
		}
		volumeMoveAlertCount++
		instance.SetLabel("movement_state", movementState)
		instance.SetLabel("svm", svm)
		instance.SetLabel("volume", volume)
		instance.SetLabel(severityLabel, string(warning))

		h.setAlertMetric(mat, instance, 1)
	}
	return volumeMoveAlertCount
}

func (h *Health) collectVolumeRansomwareAlerts() int {
	var (
		instance *matrix.Instance
	)
	volumeRansomwareAlertCount := 0
	clusterVersion := h.client.Cluster().GetVersion()
	ontapVersion, err := goversion.NewVersion(clusterVersion)
	if err != nil {
		h.Logger.Error().Err(err).
			Str("version", clusterVersion).
			Msg("Failed to parse version")
		return 0
	}
	version910 := "9.10"
	version910After, err := goversion.NewVersion(version910)
	if err != nil {
		h.Logger.Error().Err(err).
			Str("version", version910).
			Msg("Failed to parse version")
		return 0
	}

	if ontapVersion.LessThan(version910After) {
		return 0
	}
	records, err := h.getRansomwareVolumes()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			h.Logger.Debug().Err(err).Msg("API not found")
		} else {
			h.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return 0
	}
	mat := h.data[volumeRansomwareHealthMatrix]
	for _, record := range records {
		uuid := record.Get("uuid").String()
		volume := record.Get("name").String()
		antiRansomwareAttackProbability := record.Get("anti_ransomware.attack_probability").String()
		instance, err = mat.NewInstance(uuid)
		if err != nil {
			h.Logger.Warn().Str("key", uuid).Msg("error while creating instance")
			continue
		}
		volumeRansomwareAlertCount++
		instance.SetLabel("anti_ransomware_attack_probability", antiRansomwareAttackProbability)

		instance.SetLabel("volume", volume)
		instance.SetLabel(severityLabel, string(errr))

		h.setAlertMetric(mat, instance, 1)
	}
	return volumeRansomwareAlertCount
}

func (h *Health) collectNetworkInterfacesAlerts() int {
	var (
		instance *matrix.Instance
	)
	networkInterfaceAlertCount := 0
	records, err := h.getNonHomeLIFs()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			h.Logger.Debug().Err(err).Msg("API not found")
		} else {
			h.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return 0
	}
	mat := h.data[lifHealthMatrix]
	for _, record := range records {
		uuid := record.Get("uuid").String()
		lif := record.Get("name").String()
		svm := record.Get("svm.name").String()
		isHome := record.Get("location.is_home").String()
		instance, err = mat.NewInstance(uuid)
		if err != nil {
			h.Logger.Warn().Str("key", uuid).Msg("error while creating instance")
			continue
		}
		networkInterfaceAlertCount++
		instance.SetLabel("svm", svm)
		instance.SetLabel("isHome", isHome)
		instance.SetLabel("lif", lif)
		instance.SetLabel(severityLabel, string(warning))

		h.setAlertMetric(mat, instance, 1)
	}
	return networkInterfaceAlertCount
}

func (h *Health) collectNetworkFCPortAlerts() int {
	var (
		instance *matrix.Instance
	)
	networkFcpPortAlertCount := 0
	records, err := h.getFCPorts()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			h.Logger.Debug().Err(err).Msg("API not found")
		} else {
			h.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return 0
	}
	mat := h.data[networkFCPortHealthMatrix]
	for _, record := range records {
		uuid := record.Get("uuid").String()
		nodeName := record.Get("node.name").String()
		port := record.Get("name").String()
		state := record.Get("state").String()
		instance, err = mat.NewInstance(uuid)
		if err != nil {
			h.Logger.Warn().Str("key", uuid).Msg("error while creating instance")
			continue
		}
		networkFcpPortAlertCount++
		instance.SetLabel("node", nodeName)
		instance.SetLabel("state", state)
		instance.SetLabel("port", port)
		instance.SetLabel(severityLabel, string(errr))

		h.setAlertMetric(mat, instance, 1)
	}
	return networkFcpPortAlertCount
}

func (h *Health) collectNetworkEthernetPortAlerts() int {
	var (
		instance *matrix.Instance
	)
	networkEthernetPortAlertCount := 0
	records, err := h.getEthernetPorts()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			h.Logger.Debug().Err(err).Msg("API not found")
		} else {
			h.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return 0
	}
	mat := h.data[networkEthernetPortHealthMatrix]
	for _, record := range records {
		uuid := record.Get("uuid").String()
		port := record.Get("name").String()
		nodeName := record.Get("node.name").String()
		portType := record.Get("type").String()
		state := record.Get("state").String()
		instance, err = mat.NewInstance(uuid)
		if err != nil {
			h.Logger.Warn().Str("key", uuid).Msg("error while creating instance")
			continue
		}
		networkEthernetPortAlertCount++
		instance.SetLabel("node", nodeName)
		instance.SetLabel("state", state)
		instance.SetLabel("port", port)
		instance.SetLabel("type", portType)
		instance.SetLabel(severityLabel, string(errr))

		h.setAlertMetric(mat, instance, 1)
	}
	return networkEthernetPortAlertCount
}

func (h *Health) collectNodeAlerts() int {
	var (
		instance *matrix.Instance
	)
	nodeAlertCount := 0
	records, err := h.getNodes()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			h.Logger.Debug().Err(err).Msg("API not found")
		} else {
			h.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return 0
	}
	mat := h.data[nodeHealthMatrix]
	for _, record := range records {
		nodeName := record.Get("node").String()

		instance, err = mat.NewInstance(nodeName)
		if err != nil {
			h.Logger.Warn().Str("key", nodeName).Msg("error while creating instance")
			continue
		}
		nodeAlertCount++
		instance.SetLabel("node", nodeName)
		instance.SetLabel("healthy", "false")
		instance.SetLabel(severityLabel, string(errr))

		h.setAlertMetric(mat, instance, 1)
	}
	return nodeAlertCount
}

func (h *Health) collectHAAlerts() int {
	var (
		instance *matrix.Instance
	)
	HAAlertCount := 0
	records, err := h.getHADown()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			h.Logger.Debug().Err(err).Msg("API not found")
		} else {
			h.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return 0
	}
	mat := h.data[haHealthMatrix]
	for _, record := range records {
		nodeName := record.Get("node").String()
		takeoverPossible := record.Get("possible").String()
		partnerName := record.Get("partner_name").String()
		stateDescription := record.Get("state_description").String()
		partnerState := record.Get("partner_state").String()
		if takeoverPossible == "" {
			takeoverPossible = "false"
		}

		instance, err = mat.NewInstance(nodeName)
		if err != nil {
			h.Logger.Warn().Str("key", nodeName).Msg("error while creating instance")
			continue
		}
		HAAlertCount++
		instance.SetLabel("node", nodeName)
		instance.SetLabel("takeover_possible", takeoverPossible)
		instance.SetLabel("partner", partnerName)
		instance.SetLabel("state_description", stateDescription)
		instance.SetLabel("partner_state", partnerState)
		instance.SetLabel(severityLabel, string(errr))

		h.setAlertMetric(mat, instance, 1)
	}
	return HAAlertCount
}

func (h *Health) collectShelfAlerts() int {
	var (
		instance *matrix.Instance
	)
	shelfAlertCount := 0
	records, err := h.getShelves()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			h.Logger.Debug().Err(err).Msg("API not found")
		} else {
			h.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return 0
	}
	mat := h.data[shelfHealthMatrix]
	for _, record := range records {
		shelf := record.Get("shelf").String()
		errorType := record.Get("error_type").String()
		errorSeverity := record.Get("error_severity").String()
		errorText := record.Get("error_text").String()

		// errorSeverity possible values are unknown|notice|warning|error|critical
		if errorSeverity == "error" || errorSeverity == "critical" || errorSeverity == "warning" {
			instance, err = mat.NewInstance(shelf)
			if err != nil {
				h.Logger.Warn().Str("key", shelf).Msg("error while creating instance")
				continue
			}
			shelfAlertCount++
			instance.SetLabel("shelf", shelf)
			instance.SetLabel("error_type", errorType)
			instance.SetLabel("error_text", errorText)
			if errorSeverity == "error" || errorSeverity == "critical" {
				instance.SetLabel(severityLabel, string(errr))
			} else {
				instance.SetLabel(severityLabel, string(warning))
			}

			h.setAlertMetric(mat, instance, 1)
		}
	}
	return shelfAlertCount
}

func (h *Health) collectSupportAlerts() int {
	var (
		instance *matrix.Instance
	)
	supportAlertCount := 0
	clusterTime, err := collectors.GetClusterTime(h.client, nil, h.Logger)
	if err != nil {
		h.Logger.Error().Err(err).Msg("Failed to collect cluster time")
		return 0
	}
	toTime := clusterTime.Unix()
	timeFilter := h.getTimeStampFilter(clusterTime)
	filter := append([]string{"suppress=false"}, timeFilter)

	records, err := h.getSupportAlerts(filter)
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			h.Logger.Debug().Err(err).Msg("API not found")
		} else {
			h.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return 0
	}
	mat := h.data[supportHealthMatrix]
	for index, record := range records {
		nodeName := record.Get("node.name").String()
		monitor := record.Get("monitor").String()
		name := record.Get("name").String()
		resource := record.Get("resource").String()
		reason := record.Get("cause.message").String()
		correctiveAction := record.Get("corrective_action.message").String()
		instance, err = mat.NewInstance(strconv.Itoa(index))
		if err != nil {
			h.Logger.Warn().Int("key", index).Msg("error while creating instance")
			continue
		}
		supportAlertCount++
		instance.SetLabel("node", nodeName)
		instance.SetLabel("monitor", monitor)
		instance.SetLabel("name", name)
		instance.SetLabel("resource", resource)
		instance.SetLabel("reason", reason)
		instance.SetLabel("correctiveAction", correctiveAction)
		instance.SetLabel(severityLabel, string(warning))

		h.setAlertMetric(mat, instance, 1)
	}
	// update lastFilterTime to current cluster time
	h.lastFilterTime = toTime
	return supportAlertCount
}

func (h *Health) collectDiskAlerts() int {
	var (
		instance *matrix.Instance
	)
	diskAlertCount := 0
	records, err := h.getDisks()
	if err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			h.Logger.Debug().Err(err).Msg("API not found")
		} else {
			h.Logger.Error().Err(err).Msg("Failed to collect analytic data")
		}
		return 0
	}
	mat := h.data[diskHealthMatrix]
	for _, record := range records {
		name := record.Get("name").String()
		containerType := record.Get("container_type").String()
		instance, err = mat.NewInstance(name)
		if err != nil {
			h.Logger.Warn().Str("key", name).Msg("error while creating instance")
			continue
		}
		diskAlertCount++
		instance.SetLabel("disk", name)
		instance.SetLabel("container_type", containerType)
		if containerType == "broken" {
			instance.SetLabel(severityLabel, string(errr))
		} else if containerType == "unassigned" {
			instance.SetLabel(severityLabel, string(warning))
		}

		h.setAlertMetric(mat, instance, 1)
	}
	return diskAlertCount
}

func (h *Health) getDisks() ([]gjson.Result, error) {
	fields := []string{"name", "container_type"}
	query := "api/storage/disks"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Filter([]string{"container_type=broken|unassigned"}).
		Build()

	return collectors.InvokeRestCall(h.client, href, h.Logger)
}

func (h *Health) getShelves() ([]gjson.Result, error) {
	fields := []string{"error_type", "error_severity", "error_text"}
	query := "api/private/cli/storage/shelf"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Build()

	return collectors.InvokeRestCall(h.client, href, h.Logger)
}

func (h *Health) getNodes() ([]gjson.Result, error) {
	fields := []string{"health"}
	query := "api/private/cli/node"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Filter([]string{"health=false"}).
		Build()

	return collectors.InvokeRestCall(h.client, href, h.Logger)
}

func (h *Health) getHADown() ([]gjson.Result, error) {
	fields := []string{"possible,partner_name,state_description,partner_state"}
	query := "api/private/cli/storage/failover"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Filter([]string{"possible=!true"}).
		Build()

	return collectors.InvokeRestCall(h.client, href, h.Logger)
}

func (h *Health) getRansomwareVolumes() ([]gjson.Result, error) {
	query := "api/storage/volumes"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Filter([]string{"anti_ransomware.state=enabled", "anti_ransomware.attack_probability=low|moderate|high"}).
		Build()

	return collectors.InvokeRestCall(h.client, href, h.Logger)
}

func (h *Health) getNonCompliantLicense() ([]gjson.Result, error) {
	query := "api/cluster/licensing/licenses"
	fields := []string{"name,scope,state"}
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Filter([]string{"state=noncompliant"}).
		Build()

	return collectors.InvokeRestCall(h.client, href, h.Logger)
}

func (h *Health) getMoveFailedVolumes() ([]gjson.Result, error) {
	query := "api/storage/volumes"
	fields := []string{"uuid,name,movement.state,svm"}
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Filter([]string{"movement.state=cutover_wait|failed|cutover_pending"}).
		Build()

	return collectors.InvokeRestCall(h.client, href, h.Logger)
}

func (h *Health) getNonHomeLIFs() ([]gjson.Result, error) {
	query := "api/network/ip/interfaces"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields([]string{"svm", "location"}).
		Filter([]string{"location.is_home=false"}).
		Build()

	return collectors.InvokeRestCall(h.client, href, h.Logger)
}

func (h *Health) getFCPorts() ([]gjson.Result, error) {
	fields := []string{"name,node"}
	query := "api/network/fc/ports"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Filter([]string{"enabled=true", "state=offlined_by_system"}).
		Build()

	return collectors.InvokeRestCall(h.client, href, h.Logger)
}

func (h *Health) getEthernetPorts() ([]gjson.Result, error) {
	fields := []string{"name,node"}
	query := "api/network/ethernet/ports"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Filter([]string{"enabled=true", "state=down"}).
		Build()

	return collectors.InvokeRestCall(h.client, href, h.Logger)
}

func (h *Health) getSupportAlerts(filter []string) ([]gjson.Result, error) {
	query := "api/private/support/alerts"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Filter(filter).
		Build()

	return collectors.InvokeRestCall(h.client, href, h.Logger)
}

// returns time filter (clustertime - polldata duration)
func (h *Health) getTimeStampFilter(clusterTime time.Time) string {
	fromTime := h.lastFilterTime
	// check if this is the first request
	if h.lastFilterTime == 0 {
		// if first request fetch cluster time
		dataDuration, err := collectors.GetDataInterval(h.ParentParams, defaultDataPollDuration)
		if err != nil {
			h.Logger.Warn().Err(err).
				Str("defaultDataPollDuration", defaultDataPollDuration.String()).
				Msg("Failed to parse duration. using default")
		}
		fromTime = clusterTime.Add(-dataDuration).Unix()
	}
	return fmt.Sprintf("time=>=%d", fromTime)
}

func (h *Health) setAlertMetric(mat *matrix.Matrix, instance *matrix.Instance, value float64) {
	var err error
	m := mat.GetMetric(alerts)
	if m == nil {
		if m, err = mat.NewMetricFloat64(alerts); err != nil {
			h.Logger.Warn().Err(err).Str("key", alerts).Msg("error while creating metric")
			return
		}
	}
	if err = m.SetValueFloat64(instance, value); err != nil {
		h.Logger.Error().Err(err).Str("metric", alerts).Msg("Unable to set value on metric")
	}
}

func (h *Health) generateResolutionMetrics() int {
	resolutionInstancesCount := 0
	for prevKey, prevMat := range h.previousData {
		curMat, exists := h.data[prevKey]
		if !exists {
			continue
		}

		prevInstances := prevMat.GetInstanceKeys()
		curInstances := make(map[string]struct{})
		for _, instanceKey := range curMat.GetInstanceKeys() {
			curInstances[instanceKey] = struct{}{}
		}

		for _, pInstanceKey := range prevInstances {
			if _, found := curInstances[pInstanceKey]; found {
				continue
			}

			rMat := h.resolutionData[prevKey]
			if rMat == nil {
				h.Logger.Warn().Str("key", prevKey).Msg("empty resolution Matrix")
				continue
			}

			rInstance, err := rMat.NewInstance(pInstanceKey)
			if err != nil {
				h.Logger.Warn().Str("key", pInstanceKey).Msg("error while creating instance")
				continue
			}
			resolutionInstancesCount++

			rInstance.SetLabels(prevMat.GetInstance(pInstanceKey).GetLabels())
			h.setAlertMetric(rMat, rInstance, 0)
		}
	}
	h.previousData = h.data
	return resolutionInstancesCount
}

func (h *Health) GetGeneratedMetrics() []plugin.CustomMetric {
	return []plugin.CustomMetric{
		{
			Name:         alerts,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Provides any issues related to Disks health check if disks are broken or unassigned. Value of 1 means issue is happening and 0 means that issue is resolved.",
			Prefix:       diskHealthMatrix,
		},
		{
			Name:         alerts,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Provides any issues related to Shelf health check. Value of 1 means issue is happening and 0 means that issue is resolved.",
			Prefix:       shelfHealthMatrix,
		},
		{
			Name:         alerts,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Provides any issues related to Support health check. Value of 1 means issue is happening and 0 means that issue is resolved.",
			Prefix:       supportHealthMatrix,
		},
		{
			Name:         alerts,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Provides any issues related to Node health check. Value of 1 means issue is happening and 0 means that issue is resolved.",
			Prefix:       nodeHealthMatrix,
		},
		{
			Name:         alerts,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Provides any issues related to HA health check. Value of 1 means issue is happening and 0 means that issue is resolved.",
			Prefix:       haHealthMatrix,
		},
		{
			Name:         alerts,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Provides any issues related to Network Ethernet Port health check. Value of 1 means issue is happening and 0 means that issue is resolved.",
			Prefix:       networkEthernetPortHealthMatrix,
		},
		{
			Name:         alerts,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Provides any issues related to Network FC Port health check. Value of 1 means issue is happening and 0 means that issue is resolved.",
			Prefix:       networkFCPortHealthMatrix,
		},
		{
			Name:         alerts,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Provides any issues related to LIF health check. Value of 1 means issue is happening and 0 means that issue is resolved.",
			Prefix:       lifHealthMatrix,
		},
		{
			Name:         alerts,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Provides any issues related to Volume Ransomware health check. Value of 1 means issue is happening and 0 means that issue is resolved.",
			Prefix:       volumeRansomwareHealthMatrix,
		},
		{
			Name:         alerts,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Provides any issues related to Volume Move health check. Value of 1 means issue is happening and 0 means that issue is resolved.",
			Prefix:       volumeMoveHealthMatrix,
		},
		{
			Name:         alerts,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Provides any issues related to License health check. Value of 1 means issue is happening and 0 means that issue is resolved.",
			Prefix:       licenseHealthMatrix,
		},
	}
}
