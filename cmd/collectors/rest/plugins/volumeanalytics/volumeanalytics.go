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

const explorer = "volume_analytics"

var MaxDirCollectCount = "100"

// const activityFiles = "volume_analytics_files"
// const activityDir = "volume_analytics_dir"
// const activityClient = "volume_analytics_client"
// const activityUser = "volume_analytics_user"

type VolumeAnalytics struct {
	*plugin.AbstractPlugin
	currentVal int
	client     *rest.Client
	data       map[string]*matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &VolumeAnalytics{AbstractPlugin: p}
}

var metrics = []string{
	"dir_bytes_used",
	"dir_file_count",
	"dir_subdir_count",
}

//var topMetricMap = []TopMetrics{
//	{objType: "files", topMetricCounter: "iops.read", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "iops_read", matrixName: activityFiles},
//	{objType: "files", topMetricCounter: "iops.write", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "iops_write", matrixName: activityFiles},
//	{objType: "files", topMetricCounter: "throughput.read", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "throughput_read", matrixName: activityFiles},
//	{objType: "files", topMetricCounter: "throughput.write", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "throughput_write", matrixName: activityFiles},
//	{objType: "directories", topMetricCounter: "iops.read", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "iops_read", matrixName: activityDir},
//	{objType: "directories", topMetricCounter: "iops.write", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "iops_write", matrixName: activityDir},
//	{objType: "directories", topMetricCounter: "throughput.read", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "throughput_read", matrixName: activityDir},
//	{objType: "directories", topMetricCounter: "throughput.write", fields: []string{"path"}, instanceKey: []string{"path"}, displayName: "throughput_write", matrixName: activityDir},
//	{objType: "clients", topMetricCounter: "iops.read", fields: []string{"client_ip"}, instanceKey: []string{"client_ip"}, displayName: "iops_read", matrixName: activityClient},
//	{objType: "clients", topMetricCounter: "iops.write", fields: []string{"client_ip"}, instanceKey: []string{"client_ip"}, displayName: "iops_write", matrixName: activityClient},
//	{objType: "clients", topMetricCounter: "throughput.read", fields: []string{"client_ip"}, instanceKey: []string{"client_ip"}, displayName: "throughput_read", matrixName: activityClient},
//	{objType: "clients", topMetricCounter: "throughput.write", fields: []string{"client_ip"}, instanceKey: []string{"client_ip"}, displayName: "throughput_write", matrixName: activityClient},
//	{objType: "users", topMetricCounter: "iops.read", fields: []string{"user_name"}, instanceKey: []string{"user_id"}, displayName: "iops_read", matrixName: activityUser},
//	{objType: "users", topMetricCounter: "iops.write", fields: []string{"user_name"}, instanceKey: []string{"user_id"}, displayName: "iops_write", matrixName: activityUser},
//	{objType: "users", topMetricCounter: "throughput.read", fields: []string{"user_name"}, instanceKey: []string{"user_id"}, displayName: "throughput_read", matrixName: activityUser},
//	{objType: "users", topMetricCounter: "throughput.write", fields: []string{"user_name"}, instanceKey: []string{"user_id"}, displayName: "throughput_write", matrixName: activityUser},
//}

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

	if err = v.initMatrix(); err != nil {
		return err
	}

	m := v.Params.GetChildS("MaxDirectoryCount")
	if m != nil {
		MaxDirCollectCount = m.GetContentS()
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if v.client, err = rest.New(conf.ZapiPoller(v.ParentParams), timeout); err != nil {
		v.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = v.client.Init(5); err != nil {
		return err
	}

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	v.currentVal = v.SetPluginInterval()
	return nil
}

func (v *VolumeAnalytics) initMatrix() error {
	v.data = make(map[string]*matrix.Matrix)

	v.data[explorer] = matrix.New(v.Parent+explorer, explorer, explorer)
	//for _, v1 := range topMetricMap {
	//	v.data[v1.matrixName+v1.topMetricCounter] = matrix.New(v.Parent+v1.matrixName, v1.matrixName, v1.topMetricCounter)
	//}
	for _, v1 := range v.data {
		v1.SetExportOptions(matrix.DefaultExportOptions())
	}
	for _, k := range metrics {
		err := matrix.CreateMetric(k, v.data[explorer])
		if err != nil {
			v.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
			return err
		}
	}
	return nil
}

func (v *VolumeAnalytics) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	// Purge and reset data
	// remove all metrics as analytics label may change over time
	err := v.initMatrix()
	if err != nil {
		v.Logger.Warn().Err(err).Msg("error while init matrix")
		return nil, err
	}
	for k := range v.data {
		// Set all global labels if already not exist
		v.data[k].SetGlobalLabels(data.GetGlobalLabels())
	}
	//for _, v1 := range topMetricMap {
	//	err = matrix.CreateMetric(v1.displayName, v.data[v1.matrixName+v1.topMetricCounter])
	//	if err != nil {
	//		v.Logger.Warn().Err(err).Str("key", v1.displayName).Msg("error while creating metric")
	//		return err
	//	}
	//}

	for _, k := range metrics {
		err := matrix.CreateMetric(k, v.data[explorer])
		if err != nil {
			v.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
			return nil, err
		}
	}
	cluster, _ := data.GetGlobalLabels().GetHas("cluster")
	v.currentVal = 0
	ver := v.client.Cluster().Version
	clusterVersion := fmt.Sprintf("%d.%d.%d", ver[0], ver[1], ver[2])
	ontapVersion, err := goversion.NewVersion(clusterVersion)
	version98 := "9.8"
	//version910 := "9.10"
	version98After, _ := goversion.NewVersion(version98)
	//version910After, _ := goversion.NewVersion(version910)
	if err != nil {
		v.Logger.Error().Err(err).
			Str("version", clusterVersion).
			Msg("Failed to parse version")
		return nil, nil
	}
	if ontapVersion.GreaterThanOrEqual(version98After) {
		for instanceID, dataInstance := range data.GetInstances() {
			if records, analytics, err := v.getAnalyticsData(instanceID); err != nil {
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
					mtBytesUsedValues := strings.Split(util.ArrayMetricToString(record.Get("analytics.by_modified_time.bytes_used.values").String()), ",")
					mtBytesUsedPercentages := strings.Split(util.ArrayMetricToString(record.Get("analytics.by_modified_time.bytes_used.percentages").String()), ",")
					mtBytesUsedNewestLabel := record.Get("analytics.by_modified_time.bytes_used.newest_label").String()
					mtBytesUsedOldestLabel := record.Get("analytics.by_modified_time.bytes_used.oldest_label").String()
					mtBytesUsedLabels := strings.Split(util.ArrayMetricToString(analytics.Get("by_modified_time.bytes_used.labels").String()), ",")

					atBytesUsedValues := strings.Split(util.ArrayMetricToString(record.Get("analytics.by_accessed_time.bytes_used.values").String()), ",")
					atBytesUsedPercentages := strings.Split(util.ArrayMetricToString(record.Get("analytics.by_accessed_time.bytes_used.percentages").String()), ",")
					atBytesUsedNewestLabel := record.Get("analytics.by_accessed_time.bytes_used.newest_label").String()
					atBytesUsedOldestLabel := record.Get("analytics.by_accessed_time.bytes_used.oldest_label").String()
					atBytesUsedLabels := strings.Split(util.ArrayMetricToString(analytics.Get("by_accessed_time.bytes_used.labels").String()), ",")

					instance, err := explorerMatrix.NewInstance(instanceID + name)
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
					if bytesUsed != "" {
						if err = explorerMatrix.GetMetric("dir_bytes_used").SetValueString(instance, bytesUsed); err != nil {
							v.Logger.Error().Err(err).Str("value", bytesUsed).Msg("set metric")
						}
					}
					if fileCount != "" {
						if err = explorerMatrix.GetMetric("dir_file_count").SetValueString(instance, fileCount); err != nil {
							v.Logger.Error().Err(err).Str("value", fileCount).Msg("set metric")
						}
					}
					if subDirCount != "" {
						if name == "." {
							if err = explorerMatrix.GetMetric("dir_subdir_count").SetValueString(instance, util.AddIntString(subDirCount, 1)); err != nil {
								v.Logger.Error().Err(err).Str("value", subDirCount).Msg("set metric")
							}
						} else {
							if err = explorerMatrix.GetMetric("dir_subdir_count").SetValueString(instance, subDirCount); err != nil {
								v.Logger.Error().Err(err).Str("value", subDirCount).Msg("set metric")
							}
						}
					}
					if len(mtBytesUsedValues) == len(mtBytesUsedPercentages) && len(mtBytesUsedValues) == len(mtBytesUsedLabels) {
						instance.SetLabel("newest_label_by_Modified_time", mtBytesUsedNewestLabel)
						instance.SetLabel("oldest_label_by_Modified_time", mtBytesUsedOldestLabel)

						for i, mv := range mtBytesUsedValues {
							key := "modified_value_" + mtBytesUsedLabels[i]
							if m := explorerMatrix.GetMetric(key); m != nil {
								continue
							}
							if m, err := explorerMatrix.NewMetricFloat64(key, "bytes_used_by_modified_time"); err != nil {
								return nil, err
							} else {
								m.SetLabel("modified_time", mtBytesUsedLabels[i])
								if err = m.SetValueString(instance, mv); err != nil {
									v.Logger.Error().Err(err).Str("value", mv).Msg("set metric")
								}
							}
						}
						for i, mp := range mtBytesUsedPercentages {
							key := "modified_percent_" + mtBytesUsedLabels[i]
							if m := explorerMatrix.GetMetric(key); m != nil {
								continue
							}
							if m, err := explorerMatrix.NewMetricFloat64(key, "bytes_used_percent_by_modified_time"); err != nil {
								return nil, err
							} else {
								m.SetLabel("modified_time", mtBytesUsedLabels[i])
								if err = m.SetValueString(instance, mp); err != nil {
									v.Logger.Error().Err(err).Str("value", mp).Msg("set metric")
								}
							}
						}
					}

					if len(atBytesUsedValues) == len(atBytesUsedPercentages) && len(atBytesUsedValues) == len(atBytesUsedLabels) {
						instance.SetLabel("newest_label_by_accessed_time", atBytesUsedNewestLabel)
						instance.SetLabel("oldest_label_by_accessed_time", atBytesUsedOldestLabel)

						for i, av := range atBytesUsedValues {
							key := "access_value_" + atBytesUsedLabels[i]
							if m := explorerMatrix.GetMetric(key); m != nil {
								continue
							}
							if m, err := explorerMatrix.NewMetricFloat64(key, "bytes_used_by_accessed_time"); err != nil {
								return nil, err
							} else {
								m.SetLabel("access_time", atBytesUsedLabels[i])
								if err = m.SetValueString(instance, av); err != nil {
									v.Logger.Error().Err(err).Str("value", av).Msg("set metric")
								}
							}
						}
						for i, ap := range atBytesUsedPercentages {
							key := "access_percent_" + atBytesUsedLabels[i]
							if m := explorerMatrix.GetMetric(key); m != nil {
								continue
							}
							if m, err := explorerMatrix.NewMetricFloat64(key, "bytes_used_percent_by_accessed_time"); err != nil {
								return nil, err
							} else {
								m.SetLabel("access_time", atBytesUsedLabels[i])
								if err = m.SetValueString(instance, ap); err != nil {
									v.Logger.Error().Err(err).Str("value", ap).Msg("set metric")
								}
							}
						}
					}

				}
			}
			//if ontapVersion.GreaterThanOrEqual(version910After) {
			//	for _, v1 := range topMetricMap {
			//		if records, err := v.getTopMetrics(instanceID, v1.objType, v1.topMetricCounter, v1.fields); err != nil {
			//			if errs.IsRestErr(err, errs.APINotFound) {
			//				v.Logger.Debug().Err(err).Msg("API not found")
			//			} else {
			//				v.Logger.Error().Err(err).Msg("Failed to collect top metrics data")
			//			}
			//			continue
			//		} else {
			//			mat := v.data[v1.matrixName+v1.topMetricCounter]
			//			for _, record := range records {
			//				var instanceKey string
			//				if len(v1.instanceKey) != 0 {
			//					// extract instance key(s)
			//					for _, k := range v1.instanceKey {
			//						value := record.Get(k)
			//						if value.Exists() {
			//							instanceKey += value.String()
			//						} else {
			//							v.Logger.Trace().Str("key", k).Msg("missing key")
			//						}
			//					}
			//
			//					if instanceKey == "" {
			//						v.Logger.Trace().Msg("Instance key is empty, skipping")
			//						continue
			//					}
			//				}
			//				instance := mat.GetInstance(instanceKey)
			//				if instance == nil {
			//					if instance, err = mat.NewInstance(instanceKey); err != nil {
			//						v.Logger.Error().Err(err).Str("instKey", instanceKey).Msg("Failed to create new instance")
			//						continue
			//					}
			//				}
			//				// copy all labels
			//				for k2, v2 := range dataInstance.GetLabels().Map() {
			//					instance.SetLabel(k2, v2)
			//				}
			//				topMetric := record.Get(v1.topMetricCounter).String()
			//				for _, f := range v1.fields {
			//					fv := record.Get(f).String()
			//					instance.SetLabel(f, fv)
			//				}
			//				err = mat.GetMetric(v1.displayName).SetValueString(instance, topMetric)
			//			}
			//		}
			//	}
			//}
		}
	}

	result := make([]*matrix.Matrix, 0, len(v.data))

	for _, value := range v.data {
		result = append(result, value)
	}
	return result, nil
}

func (v *VolumeAnalytics) getAnalyticsData(instanceId string) ([]gjson.Result, gjson.Result, error) {
	var (
		result    []gjson.Result
		analytics gjson.Result
		err       error
	)

	fields := []string{"analytics.file_count", "analytics.bytes_used", "analytics.subdir_count", "analytics.by_modified_time.bytes_used", "analytics.by_accessed_time.bytes_used"}
	query := path.Join("api/storage/volumes", instanceId, "files/")
	href := rest.BuildHref(query, strings.Join(fields, ","), []string{"order_by=analytics.bytes_used+desc", "type=directory"}, "", "", MaxDirCollectCount, "", query)

	if result, analytics, err = collectors.InvokeRestCallAnalyticsLimited(v.client, href); err != nil {
		return nil, gjson.Result{}, err
	}
	return result, analytics, nil
}

//func (v *VolumeAnalytics) getTopMetrics(instanceId string, objType string, topMetric string, fields []string) ([]gjson.Result, error) {
//	var (
//		result []gjson.Result
//		err    error
//	)
//
//	query := path.Join("api/storage/volumes", instanceId, "top-metrics", objType)
//	href := rest.BuildHref("", strings.Join(fields, ","), []string{"top_metric=" + topMetric, "return_timeout=120"}, "", "", "", "", query)
//
//	if result, err = collectors.InvokeRestCallLimited(v.client, href, v.Logger); err != nil {
//		return nil, err
//	}
//	return result, nil
//}
