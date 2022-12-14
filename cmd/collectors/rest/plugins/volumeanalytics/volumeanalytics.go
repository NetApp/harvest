package volumeanalytics

import (
	"fmt"
	goversion "github.com/hashicorp/go-version"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"path"
	"strconv"
	"strings"
	"time"
)

const DefaultPluginDuration = 30 * time.Minute
const DefaultDataPollDuration = 3 * time.Minute
const explorer = "volume_analytics"
const activityFiles = "volume_analytics_files"
const activityDir = "volume_analytics_dir"
const activityClient = "volume_analytics_client"
const activityUser = "volume_analytics_user"
const MaxDirCollectCount = "101"

type VolumeAnalytics struct {
	*plugin.AbstractPlugin
	pluginInvocationRate int
	currentVal           int
	client               *rest.Client
	data                 map[string]*matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &VolumeAnalytics{AbstractPlugin: p}
}

var metrics = []string{
	"dir_bytes_used",
	"dir_file_count",
	"dir_subdir_count",
}

var topMetricMap = []TopMetrics{
	{objType: "files", topMetricCounter: "iops.read", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "iops_read", matrixName: activityFiles},
	{objType: "files", topMetricCounter: "iops.write", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "iops_write", matrixName: activityFiles},
	{objType: "files", topMetricCounter: "throughput.read", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "throughput_read", matrixName: activityFiles},
	{objType: "files", topMetricCounter: "throughput.write", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "throughput_write", matrixName: activityFiles},
	{objType: "directories", topMetricCounter: "iops.read", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "iops_read", matrixName: activityDir},
	{objType: "directories", topMetricCounter: "iops.write", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "iops_write", matrixName: activityDir},
	{objType: "directories", topMetricCounter: "throughput.read", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "throughput_read", matrixName: activityDir},
	{objType: "directories", topMetricCounter: "throughput.write", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "throughput_write", matrixName: activityDir},
	{objType: "clients", topMetricCounter: "iops.read", fields: []string{"client_ip"}, instanceKey: []string{"client_ip"}, displayName: "iops_read", matrixName: activityClient},
	{objType: "clients", topMetricCounter: "iops.write", fields: []string{"client_ip"}, instanceKey: []string{"client_ip"}, displayName: "iops_write", matrixName: activityClient},
	{objType: "clients", topMetricCounter: "throughput.read", fields: []string{"client_ip"}, instanceKey: []string{"client_ip"}, displayName: "throughput_read", matrixName: activityClient},
	{objType: "clients", topMetricCounter: "throughput.write", fields: []string{"client_ip"}, instanceKey: []string{"client_ip"}, displayName: "throughput_write", matrixName: activityClient},
	{objType: "users", topMetricCounter: "iops.read", fields: []string{"user_name"}, instanceKey: []string{"user_id"}, displayName: "iops_read", matrixName: activityUser},
	{objType: "users", topMetricCounter: "iops.write", fields: []string{"user_name"}, instanceKey: []string{"user_id"}, displayName: "iops_write", matrixName: activityUser},
	{objType: "users", topMetricCounter: "throughput.read", fields: []string{"user_name"}, instanceKey: []string{"user_id"}, displayName: "throughput_read", matrixName: activityUser},
	{objType: "users", topMetricCounter: "throughput.write", fields: []string{"user_name"}, instanceKey: []string{"user_id"}, displayName: "throughput_write", matrixName: activityUser},
}

type TopMetrics struct {
	objType          string
	instanceKey      []string
	topMetricCounter string
	displayName      string
	fields           []string
	matrixName       string
}

func (v *VolumeAnalytics) Init() error {

	var err error

	if err = v.InitAbc(); err != nil {
		return err
	}

	v.data = make(map[string]*matrix.Matrix)

	v.data[explorer] = matrix.New(v.Parent+explorer, explorer, explorer)
	for _, v1 := range topMetricMap {
		v.data[v1.matrixName+v1.topMetricCounter] = matrix.New(v.Parent+v1.matrixName, v1.matrixName, v1.topMetricCounter)
	}

	for _, v1 := range v.data {
		v1.SetExportOptions(matrix.DefaultExportOptions())
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if v.client, err = rest.New(conf.ZapiPoller(v.ParentParams), timeout); err != nil {
		v.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	for _, k := range metrics {
		err = matrix.CreateMetric(k, v.data[explorer])
		if err != nil {
			v.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
			return err
		}
	}
	for _, v1 := range topMetricMap {
		err = matrix.CreateMetric(v1.displayName, v.data[v1.matrixName+v1.topMetricCounter])
		if err != nil {
			v.Logger.Warn().Err(err).Str("key", v1.displayName).Msg("error while creating metric")
			return err
		}
	}
	if err = v.client.Init(5); err != nil {
		return err
	}

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	if v.currentVal, err = collectors.SetPluginInterval(v.ParentParams, v.Params, v.Logger, DefaultDataPollDuration, DefaultPluginDuration); err != nil {
		v.Logger.Error().Err(err).Stack().Msg("Failed while setting the plugin interval")
		return err
	}

	return nil
}

func (v *VolumeAnalytics) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	if v.currentVal >= v.pluginInvocationRate {
		v.Logger.Info().Msg("Invoking FSA")
		// Purge and reset data
		for k := range v.data {
			v.data[k].PurgeInstances()
			v.data[k].Reset()
			// Set all global labels if already not exist
			v.data[k].SetGlobalLabels(data.GetGlobalLabels())
		}
		cluster, _ := data.GetGlobalLabels().GetHas("cluster")
		v.currentVal = 0
		ver := v.client.Cluster().Version
		clusterVersion := fmt.Sprintf("%d.%d.%d", ver[0], ver[1], ver[2])
		ontapVersion, err := goversion.NewVersion(clusterVersion)
		version98 := "9.8"
		version910 := "9.10"
		version98After, _ := goversion.NewVersion(version98)
		version910After, _ := goversion.NewVersion(version910)
		if err != nil {
			v.Logger.Error().Err(err).
				Str("version", clusterVersion).
				Msg("Failed to parse version")
			return nil, nil
		}
		if ontapVersion.GreaterThanOrEqual(version98After) {
			for instanceId, dataInstance := range data.GetInstances() {
				if records, err := v.getAnalyticsData(instanceId); err != nil {
					if errs.IsRestErr(err, errs.APINotFound) {
						v.Logger.Debug().Err(err).Msg("API not found")
					} else {
						v.Logger.Error().Err(err).Msg("Failed to collect analytic data")
					}
				} else {
					explorerMatrix := v.data[explorer]
					for index, record := range records {
						name := record.Get("name").String()
						fileCount := record.Get("analytics.file_count").String()
						bytesUsed := record.Get("analytics.bytes_used").String()
						subDirCount := record.Get("analytics.subdir_count").String()
						instance, err := explorerMatrix.NewInstance(instanceId + name)
						if err != nil {
							v.Logger.Warn().Str("key", name).Msg("error while creating instance")
							continue
						}
						instance.SetLabel("dir_name", name)
						instance.SetLabel("index", cluster+"_"+strconv.Itoa(index))
						// copy all labels
						for k1, v1 := range dataInstance.GetLabels().Map() {
							instance.SetLabel(k1, v1)
						}
						explorerMatrix.GetMetric("dir_bytes_used").SetValueString(instance, bytesUsed)
						explorerMatrix.GetMetric("dir_file_count").SetValueString(instance, fileCount)
						if name == "." {
							explorerMatrix.GetMetric("dir_subdir_count").SetValueString(instance, util.AddIntString(subDirCount, 1))
						} else {
							explorerMatrix.GetMetric("dir_subdir_count").SetValueString(instance, subDirCount)
						}
					}
				}
				if ontapVersion.GreaterThanOrEqual(version910After) {
					for _, v1 := range topMetricMap {
						if records, err := v.getTopMetrics(instanceId, v1.objType, v1.topMetricCounter, v1.fields); err != nil {
							if errs.IsRestErr(err, errs.APINotFound) {
								v.Logger.Debug().Err(err).Msg("API not found")
							} else {
								v.Logger.Error().Err(err).Msg("Failed to collect top metrics data")
							}
							continue
						} else {
							mat := v.data[v1.matrixName+v1.topMetricCounter]
							for _, record := range records {
								var instanceKey string
								if len(v1.instanceKey) != 0 {
									// extract instance key(s)
									for _, k := range v1.instanceKey {
										value := record.Get(k)
										if value.Exists() {
											instanceKey += value.String()
										} else {
											v.Logger.Trace().Str("key", k).Msg("missing key")
										}
									}

									if instanceKey == "" {
										v.Logger.Trace().Msg("Instance key is empty, skipping")
										continue
									}
								}
								instance := mat.GetInstance(instanceKey)
								if instance == nil {
									if instance, err = mat.NewInstance(instanceKey); err != nil {
										v.Logger.Error().Err(err).Str("instKey", instanceKey).Msg("Failed to create new instance")
										continue
									}
								}
								// copy all labels
								for k2, v2 := range dataInstance.GetLabels().Map() {
									instance.SetLabel(k2, v2)
								}
								topMetric := record.Get(v1.topMetricCounter).String()
								for _, f := range v1.fields {
									fv := record.Get(f).String()
									instance.SetLabel(f, fv)
								}
								mat.GetMetric(v1.displayName).SetValueString(instance, topMetric)
							}
						}
					}
				}
			}
		}
	}
	v.currentVal++

	result := make([]*matrix.Matrix, 0, len(v.data))

	for _, value := range v.data {
		result = append(result, value)
	}
	return result, nil
}

func (v *VolumeAnalytics) getAnalyticsData(instanceId string) ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)

	fields := []string{"analytics.file_count", "analytics.bytes_used", "analytics.subdir_count"}
	query := path.Join("api/storage/volumes", instanceId, "files/")
	href := rest.BuildHref("", strings.Join(fields, ","), []string{"order_by=analytics.bytes_used+desc", "type=directory"}, "", "", MaxDirCollectCount, "", query)

	if result, err = collectors.InvokeRestCallLimited(v.client, href, v.Logger); err != nil {
		return nil, err
	}
	return result, nil
}

func (v *VolumeAnalytics) getTopMetrics(instanceId string, objType string, topMetric string, fields []string) ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)

	query := path.Join("api/storage/volumes", instanceId, "top-metrics", objType)
	href := rest.BuildHref("", strings.Join(fields, ","), []string{"top_metric=" + topMetric, "return_timeout=120"}, "", "", "", "", query)

	if result, err = collectors.InvokeRestCallLimited(v.client, href, v.Logger); err != nil {
		return nil, err
	}
	return result, nil
}
