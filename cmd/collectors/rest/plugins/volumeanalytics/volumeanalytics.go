package volumeanalytics

import (
	goversion "github.com/hashicorp/go-version"
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

	cluster, _ := data.GetGlobalLabels().GetHas("cluster")
	clusterVersion := v.client.Cluster().GetVersion()
	ontapVersion, err := goversion.NewVersion(clusterVersion)
	if err != nil {
		v.Logger.Error().Err(err).
			Str("version", clusterVersion).
			Msg("Failed to parse version")
		return nil, nil
	}
	version98 := "9.8"
	version98After, err := goversion.NewVersion(version98)
	if err != nil {
		v.Logger.Error().Err(err).
			Str("version", version98).
			Msg("Failed to parse version")
		return nil, nil
	}

	if ontapVersion.LessThan(version98After) {
		return nil, nil
	}

	// Purge and reset data
	// remove all metrics as analytics label may change over time
	err = v.initMatrix()
	if err != nil {
		v.Logger.Warn().Err(err).Msg("error while init matrix")
		return nil, err
	}
	for k := range v.data {
		// Set all global labels if already not exist
		v.data[k].SetGlobalLabels(data.GetGlobalLabels())
	}

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
				mtBytesUsedLabels := strings.Split(util.ArrayMetricToString(analytics.Get("by_modified_time.bytes_used.labels").String()), ",")

				atBytesUsedValues := strings.Split(util.ArrayMetricToString(record.Get("analytics.by_accessed_time.bytes_used.values").String()), ",")
				atBytesUsedPercentages := strings.Split(util.ArrayMetricToString(record.Get("analytics.by_accessed_time.bytes_used.percentages").String()), ",")
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

					for i, mv := range mtBytesUsedValues {
						key := "modified_value_" + mtBytesUsedLabels[i]
						m := explorerMatrix.GetMetric(key)
						if m == nil {
							if m, err = explorerMatrix.NewMetricFloat64(key, "bytes_used_by_modified_time"); err != nil {
								return nil, err
							}
						}
						m.SetLabel("time", mtBytesUsedLabels[i])
						m.SetLabel("order", strconv.Itoa(i))
						m.SetLabel("activity", v.getLabelBucket(atBytesUsedLabels[i]))
						if err = m.SetValueString(instance, mv); err != nil {
							v.Logger.Error().Err(err).Str("value", mv).Msg("set metric")
						}

					}
					for i, mp := range mtBytesUsedPercentages {
						key := "modified_percent_" + mtBytesUsedLabels[i]
						m := explorerMatrix.GetMetric(key)
						if m == nil {
							if m, err = explorerMatrix.NewMetricFloat64(key, "bytes_used_percent_by_modified_time"); err != nil {
								return nil, err
							}
						}
						m.SetLabel("time", mtBytesUsedLabels[i])
						m.SetLabel("order", strconv.Itoa(i))
						m.SetLabel("activity", v.getLabelBucket(atBytesUsedLabels[i]))
						if err = m.SetValueString(instance, mp); err != nil {
							v.Logger.Error().Err(err).Str("value", mp).Msg("set metric")
						}

					}
				}

				if len(atBytesUsedValues) == len(atBytesUsedPercentages) && len(atBytesUsedValues) == len(atBytesUsedLabels) {

					for i, av := range atBytesUsedValues {
						key := "access_value_" + atBytesUsedLabels[i]
						m := explorerMatrix.GetMetric(key)
						if m == nil {
							if m, err = explorerMatrix.NewMetricFloat64(key, "bytes_used_by_accessed_time"); err != nil {
								return nil, err
							}
						}
						m.SetLabel("time", atBytesUsedLabels[i])
						m.SetLabel("order", strconv.Itoa(i))
						m.SetLabel("activity", v.getLabelBucket(atBytesUsedLabels[i]))
						if err = m.SetValueString(instance, av); err != nil {
							v.Logger.Error().Err(err).Str("value", av).Msg("set metric")
						}

					}
					for i, ap := range atBytesUsedPercentages {
						key := "access_percent_" + atBytesUsedLabels[i]
						m := explorerMatrix.GetMetric(key)
						if m == nil {
							if m, err = explorerMatrix.NewMetricFloat64(key, "bytes_used_percent_by_accessed_time"); err != nil {
								return nil, err
							}
						}
						m.SetLabel("time", atBytesUsedLabels[i])
						m.SetLabel("order", strconv.Itoa(i))
						m.SetLabel("activity", v.getLabelBucket(atBytesUsedLabels[i]))
						if err = m.SetValueString(instance, ap); err != nil {
							v.Logger.Error().Err(err).Str("value", ap).Msg("set metric")
						}

					}
				}
			}
		}
	}

	result := make([]*matrix.Matrix, 0, len(v.data))

	for _, value := range v.data {
		result = append(result, value)
	}
	return result, nil
}

func (v *VolumeAnalytics) getLabelBucket(label string) string {
	if strings.Contains(label, "-W") {
		return "Weekly"
	} else if strings.Contains(label, "-Q") {
		return "Quarterly"
	} else if strings.Contains(label, "-") && !strings.Contains(label, "--") {
		return "Monthly"
	} else if strings.Contains(label, "unknown") {
		return "Unknown"
	} else {
		return "Yearly"
	}
}

func (v *VolumeAnalytics) getAnalyticsData(instanceID string) ([]gjson.Result, gjson.Result, error) {
	var (
		result    []gjson.Result
		analytics gjson.Result
		err       error
	)

	fields := []string{"analytics.file_count", "analytics.bytes_used", "analytics.subdir_count", "analytics.by_modified_time.bytes_used", "analytics.by_accessed_time.bytes_used"}
	query := path.Join("api/storage/volumes", instanceID, "files/")
	href := rest.BuildHref(query, strings.Join(fields, ","), []string{"order_by=analytics.bytes_used+desc", "type=directory"}, "", "", MaxDirCollectCount, "", query)

	if result, analytics, err = rest.FetchAnalytics(v.client, href); err != nil {
		return nil, gjson.Result{}, err
	}
	return result, analytics, nil
}
