package volumeanalytics

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	goversion "github.com/netapp/harvest/v2/third_party/go-version"
	"github.com/tidwall/gjson"
	"log/slog"
	"path"
	"strconv"
	"strings"
	"time"
)

const explorer = "volume_analytics"

var maxDirCollectCount = "100"

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

	if err := v.InitAbc(); err != nil {
		return err
	}

	if err := v.initMatrix(); err != nil {
		return err
	}

	m := v.Params.GetChildS("MaxDirectoryCount")
	if m != nil {
		count := m.GetContentS()
		_, err := strconv.Atoi(count)
		if err != nil {
			v.SLogger.Warn("using default", slog.String("MaxDirectoryCount", count))
		} else {
			maxDirCollectCount = count
		}
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if v.client, err = rest.New(conf.ZapiPoller(v.ParentParams), timeout, v.Auth); err != nil {
		v.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := v.client.Init(5); err != nil {
		return err
	}

	// Assigned the value to currentVal so that plugin would be invoked the first time to populate cache.
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
			v.SLogger.Warn("error while creating metric", slog.String("key", k))
			return err
		}
	}
	return nil
}

func (v *VolumeAnalytics) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[v.Object]
	v.client.Metadata.Reset()

	cluster := data.GetGlobalLabels()["cluster"]
	clusterVersion := v.client.Cluster().GetVersion()
	ontapVersion, err := goversion.NewVersion(clusterVersion)
	if err != nil {
		v.SLogger.Error("Failed to parse version",
			slogx.Err(err),
			slog.String("version", clusterVersion),
		)
		return nil, nil, nil
	}
	version98 := "9.8"
	version98After, err := goversion.NewVersion(version98)
	if err != nil {
		v.SLogger.Error("Failed to parse version",
			slogx.Err(err),
			slog.String("version", version98),
		)
		return nil, nil, nil
	}

	if ontapVersion.LessThan(version98After) {
		return nil, nil, nil
	}

	// Purge and reset data
	// remove all metrics as analytics label may change over time
	err = v.initMatrix()
	if err != nil {
		v.SLogger.Warn("error while init matrix", slogx.Err(err))
		return nil, nil, err
	}
	for k := range v.data {
		// Set all global labels if already not exist
		v.data[k].SetGlobalLabels(data.GetGlobalLabels())
	}

	for instanceID, dataInstance := range data.GetInstances() {
		if records, analytics, err := v.getAnalyticsData(instanceID); err != nil {
			if errs.IsRestErr(err, errs.APINotFound) {
				v.SLogger.Debug("API not found", slogx.Err(err))
			} else {
				v.SLogger.Error("Failed to collect analytic data", slogx.Err(err))
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
					v.SLogger.Warn("error while creating instance", slog.String("key", name))
					continue
				}
				instance.SetLabel("dir_name", name)
				instance.SetLabel("index", cluster+"_"+strconv.Itoa(index))
				// copy all labels
				for k1, v1 := range dataInstance.GetLabels() {
					instance.SetLabel(k1, v1)
				}
				if bytesUsed != "" {
					if err = explorerMatrix.GetMetric("dir_bytes_used").SetValueString(instance, bytesUsed); err != nil {
						v.SLogger.Error("set metric", slogx.Err(err), slog.String("value", bytesUsed))
					}
				}
				if fileCount != "" {
					if err = explorerMatrix.GetMetric("dir_file_count").SetValueString(instance, fileCount); err != nil {
						v.SLogger.Error("set metric", slogx.Err(err), slog.String("value", fileCount))
					}
				}
				if subDirCount != "" {
					if name == "." {
						if err = explorerMatrix.GetMetric("dir_subdir_count").SetValueString(instance, util.AddIntString(subDirCount, 1)); err != nil {
							v.SLogger.Error("set metric", slogx.Err(err), slog.String("value", subDirCount))
						}
					} else {
						if err = explorerMatrix.GetMetric("dir_subdir_count").SetValueString(instance, subDirCount); err != nil {
							v.SLogger.Error("set metric", slogx.Err(err), slog.String("value", subDirCount))
						}
					}
				}
				if len(mtBytesUsedValues) > 0 && len(mtBytesUsedValues) == len(mtBytesUsedPercentages) && len(mtBytesUsedValues) == len(mtBytesUsedLabels) {

					for i, mv := range mtBytesUsedValues {
						if mv == "" {
							continue
						}
						key := "modified_value_" + mtBytesUsedLabels[i]
						m := explorerMatrix.GetMetric(key)
						if m == nil {
							if m, err = explorerMatrix.NewMetricFloat64(key, "bytes_used_by_modified_time"); err != nil {
								return nil, nil, err
							}
						}
						m.SetLabel("time", mtBytesUsedLabels[i])
						m.SetLabel("order", strconv.Itoa(i))
						m.SetLabel("activity", v.getLabelBucket(atBytesUsedLabels[i]))
						if err = m.SetValueString(instance, mv); err != nil {
							v.SLogger.Error("set metric", slogx.Err(err), slog.String("value", mv))
						}

					}
					for i, mp := range mtBytesUsedPercentages {
						if mp == "" {
							continue
						}
						key := "modified_percent_" + mtBytesUsedLabels[i]
						m := explorerMatrix.GetMetric(key)
						if m == nil {
							if m, err = explorerMatrix.NewMetricFloat64(key, "bytes_used_percent_by_modified_time"); err != nil {
								return nil, nil, err
							}
						}
						m.SetLabel("time", mtBytesUsedLabels[i])
						m.SetLabel("order", strconv.Itoa(i))
						m.SetLabel("activity", v.getLabelBucket(atBytesUsedLabels[i]))
						if err = m.SetValueString(instance, mp); err != nil {
							v.SLogger.Error("set metric", slogx.Err(err), slog.String("value", mp))
						}

					}
				}

				if len(atBytesUsedValues) == len(atBytesUsedPercentages) && len(atBytesUsedValues) == len(atBytesUsedLabels) { //nolint:gocritic

					for i, av := range atBytesUsedValues {
						if av == "" {
							continue
						}
						key := "access_value_" + atBytesUsedLabels[i]
						m := explorerMatrix.GetMetric(key)
						if m == nil {
							if m, err = explorerMatrix.NewMetricFloat64(key, "bytes_used_by_accessed_time"); err != nil {
								return nil, nil, err
							}
						}
						m.SetLabel("time", atBytesUsedLabels[i])
						m.SetLabel("order", strconv.Itoa(i))
						m.SetLabel("activity", v.getLabelBucket(atBytesUsedLabels[i]))
						if err = m.SetValueString(instance, av); err != nil {
							v.SLogger.Error("set metric", slogx.Err(err), slog.String("value", av))
						}

					}
					for i, ap := range atBytesUsedPercentages {
						if ap == "" {
							continue
						}
						key := "access_percent_" + atBytesUsedLabels[i]
						m := explorerMatrix.GetMetric(key)
						if m == nil {
							if m, err = explorerMatrix.NewMetricFloat64(key, "bytes_used_percent_by_accessed_time"); err != nil {
								return nil, nil, err
							}
						}
						m.SetLabel("time", atBytesUsedLabels[i])
						m.SetLabel("order", strconv.Itoa(i))
						m.SetLabel("activity", v.getLabelBucket(atBytesUsedLabels[i]))
						if err = m.SetValueString(instance, ap); err != nil {
							v.SLogger.Error("set metric", slogx.Err(err), slog.String("value", ap))
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
	return result, v.client.Metadata, nil
}

func (v *VolumeAnalytics) getLabelBucket(label string) string {
	switch {
	case strings.Contains(label, "-W"):
		return "Weekly"
	case strings.Contains(label, "-Q"):
		return "Quarterly"
	case strings.Contains(label, "-") && !strings.Contains(label, "--"):
		return "Monthly"
	case strings.Contains(label, "unknown"):
		return "Unknown"
	}
	return "Yearly"
}

func (v *VolumeAnalytics) getAnalyticsData(instanceID string) ([]gjson.Result, gjson.Result, error) {
	var (
		result    []gjson.Result
		analytics gjson.Result
		err       error
	)

	fields := []string{"analytics.file_count", "analytics.bytes_used", "analytics.subdir_count", "analytics.by_modified_time.bytes_used", "analytics.by_accessed_time.bytes_used"}
	query := path.Join("api/storage/volumes", instanceID, "files/")

	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Filter([]string{"order_by=analytics.bytes_used+desc", "type=directory"}).
		MaxRecords(maxDirCollectCount).
		Build()
	if result, analytics, err = rest.FetchAnalytics(v.client, href); err != nil {
		return nil, gjson.Result{}, err
	}
	return result, analytics, nil
}
