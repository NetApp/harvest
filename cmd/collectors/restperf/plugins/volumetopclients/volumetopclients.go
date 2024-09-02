package volumetopclients

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

type VolumeInterface interface {
	fetchTopClients(volumes []string, svms []string, metric string) ([]gjson.Result, error)
	fetchVolumesWithActivityTrackingEnabled() (*set.Set, error)
	processTopClients(data *matrix.Matrix) error
}

const (
	keyToken                 = "?#"
	topClientReadOPSMatrix   = "volume_top_clients_read_ops"
	topClientWriteOPSMatrix  = "volume_top_clients_write_ops"
	topClientReadDataMatrix  = "volume_top_clients_read_data"
	topClientWriteDataMatrix = "volume_top_clients_write_data"
	defaultTopN              = 5
	maxTopN                  = 50
)

var opMetrics = []string{
	"ops",
}
var dataMetrics = []string{
	"data",
}

type Volume struct {
	*plugin.AbstractPlugin
	client          *rest.Client
	data            map[string]*matrix.Matrix
	cache           *VolumeCache
	maxVolumeCount  int
	volumeInterface VolumeInterface
}

type VolumeCache struct {
	volumesWithActivityTrackingEnabled *set.Set
	lastFetched                        time.Time
}

type MetricValue struct {
	key   string
	value float64
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (v *Volume) InitAllMatrix() error {
	v.data = make(map[string]*matrix.Matrix)
	mats := []struct {
		name    string
		object  string
		metrics []string
	}{
		{topClientReadOPSMatrix, "volume_top_clients_read", opMetrics},
		{topClientWriteOPSMatrix, "volume_top_clients_write", opMetrics},
		{topClientReadDataMatrix, "volume_top_clients_read", dataMetrics},
		{topClientWriteDataMatrix, "volume_top_clients_write", dataMetrics},
	}

	for _, m := range mats {
		if err := v.initMatrix(m.name, m.object, v.data, m.metrics); err != nil {
			return err
		}
	}
	return nil
}

func (v *Volume) initMatrix(name string, object string, inputMat map[string]*matrix.Matrix, metrics []string) error {
	matrixName := v.Parent + name
	inputMat[name] = matrix.New(matrixName, object, name)
	for _, v1 := range v.data {
		v1.SetExportOptions(matrix.DefaultExportOptions())
	}
	for _, k := range metrics {
		err := matrix.CreateMetric(k, inputMat[name])
		if err != nil {
			v.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
			return err
		}
	}
	return nil
}

func (v *Volume) Init() error {
	var err error
	if err := v.InitAbc(); err != nil {
		return err
	}

	if err := v.InitAllMatrix(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if v.client, err = rest.New(conf.ZapiPoller(v.ParentParams), timeout, v.Auth); err != nil {
		return err
	}

	if err := v.client.Init(5); err != nil {
		return err
	}

	if maxVol := v.Params.GetChildContentS("MaxVolumeCount"); maxVol != "" {
		if maxVolCount, err := strconv.Atoi(maxVol); err != nil {
			v.maxVolumeCount = defaultTopN
		} else {
			v.maxVolumeCount = int(math.Min(float64(maxVolCount), float64(maxTopN)))
		}
	}
	v.Logger.Info().Int("maxVolumeCount", v.maxVolumeCount).Msg("Using maxVolumeCount")
	return nil
}

func (v *Volume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[v.Object]
	v.client.Metadata.Reset()
	err := v.InitAllMatrix()
	if err != nil {
		return nil, nil, err
	}
	for k := range v.data {
		// Set all global labels if already not exist
		v.data[k].SetGlobalLabels(data.GetGlobalLabels())
	}
	err = v.processTopClients(data)
	if err != nil {
		return nil, nil, err
	}
	result := make([]*matrix.Matrix, 0, len(v.data))

	var pluginInstances uint64
	for _, value := range v.data {
		instCount := len(value.GetInstances())
		if instCount == 0 {
			continue
		}
		result = append(result, value)
		pluginInstances = +uint64(len(value.GetInstances()))
	}

	v.client.Metadata.PluginInstances = pluginInstances
	return result, v.client.Metadata, err
}

func (v *Volume) getCachedVolumesWithActivityTracking() (*set.Set, error) {
	const cacheDuration = time.Hour

	// Check if the cache is still valid
	if v.cache != nil && time.Since(v.cache.lastFetched) < cacheDuration {
		return v.cache.volumesWithActivityTrackingEnabled, nil
	}

	va, err := v.fetchVolumesWithActivityTrackingEnabled()
	if err != nil {
		return nil, err
	}

	// Update the cache
	v.cache = &VolumeCache{
		volumesWithActivityTrackingEnabled: va,
		lastFetched:                        time.Now(),
	}

	return va, nil
}

func (v *Volume) processTopClients(data *matrix.Matrix) error {
	va, err := v.getCachedVolumesWithActivityTracking()
	if err != nil {
		return err
	}

	if va.Size() == 0 {
		return nil
	}

	filteredDataInstances := set.New()
	for key, value := range data.GetInstances() {
		svmName := value.GetLabel("svm")
		volName := value.GetLabel("volume")
		if va.Has(svmName + keyToken + volName) {
			filteredDataInstances.Add(key)
		}
	}

	readOpsList, writeOpsList, readDataList, writeDataList := v.collectMetricValues(data, filteredDataInstances)

	topReadOps := v.getTopN(readOpsList)
	topWriteOps := v.getTopN(writeOpsList)
	topReadData := v.getTopN(readDataList)
	topWriteData := v.getTopN(writeDataList)

	readOpsVolumes, readOpsSvms := v.extractVolumesAndSvms(data, topReadOps)
	writeOpsVolumes, writeOpsSvms := v.extractVolumesAndSvms(data, topWriteOps)
	readDataVolumes, readDataSvms := v.extractVolumesAndSvms(data, topReadData)
	writeDataVolumes, writeDataSvms := v.extractVolumesAndSvms(data, topWriteData)

	if err := v.processTopClientsByMetric(readOpsVolumes, readOpsSvms, topClientReadOPSMatrix, "iops.read", v.setOpsMetric); err != nil {
		return err
	}

	if err := v.processTopClientsByMetric(writeOpsVolumes, writeOpsSvms, topClientWriteOPSMatrix, "iops.write", v.setOpsMetric); err != nil {
		return err
	}

	if err := v.processTopClientsByMetric(readDataVolumes, readDataSvms, topClientReadDataMatrix, "throughput.read", v.setDataMetric); err != nil {
		return err
	}

	if err := v.processTopClientsByMetric(writeDataVolumes, writeDataSvms, topClientWriteDataMatrix, "throughput.write", v.setDataMetric); err != nil {
		return err
	}

	return nil
}

func (v *Volume) collectMetricValues(data *matrix.Matrix, filteredDataInstances *set.Set) ([]MetricValue, []MetricValue, []MetricValue, []MetricValue) {
	readOpsMetric := data.GetMetric("total_read_ops")
	writeOpsMetric := data.GetMetric("total_write_ops")
	readDataMetric := data.GetMetric("bytes_read")
	writeDataMetric := data.GetMetric("bytes_written")

	var readOpsList, writeOpsList, readDataList, writeDataList []MetricValue
	for _, key := range filteredDataInstances.Slice() {
		var readOps, writeOps, readData, writeData float64
		var ro, wo, rb, wb bool
		if readOpsMetric != nil {
			readOps, ro = readOpsMetric.GetValueFloat64(data.GetInstance(key))
		}
		if writeOpsMetric != nil {
			writeOps, wo = writeOpsMetric.GetValueFloat64(data.GetInstance(key))
		}
		if ro && readOps > 0 {
			readOpsList = append(readOpsList, MetricValue{key: key, value: readOps})
		}
		if wo && writeOps > 0 {
			writeOpsList = append(writeOpsList, MetricValue{key: key, value: writeOps})
		}

		if readDataMetric != nil {
			readData, rb = readDataMetric.GetValueFloat64(data.GetInstance(key))
		}
		if writeDataMetric != nil {
			writeData, wb = writeDataMetric.GetValueFloat64(data.GetInstance(key))
		}

		if rb && readData > 0 {
			readDataList = append(readDataList, MetricValue{key: key, value: readData})
		}
		if wb && writeData > 0 {
			writeDataList = append(writeDataList, MetricValue{key: key, value: writeData})
		}
	}
	return readOpsList, writeOpsList, readDataList, writeDataList
}

func (v *Volume) getTopN(metricList []MetricValue) []MetricValue {
	sort.Slice(metricList, func(i, j int) bool {
		return metricList[i].value > metricList[j].value
	})

	if len(metricList) > v.maxVolumeCount {
		return metricList[:v.maxVolumeCount]
	}
	return metricList
}

func (v *Volume) extractVolumesAndSvms(data *matrix.Matrix, topMetrics []MetricValue) (*set.Set, *set.Set) {
	volumes := set.New()
	svms := set.New()
	for _, item := range topMetrics {
		instance := data.GetInstance(item.key)
		if instance != nil {
			volName := instance.GetLabel("volume")
			svmName := instance.GetLabel("svm")
			volumes.Add(volName)
			svms.Add(svmName)
		}
	}
	return volumes, svms
}

func (v *Volume) processTopClientsByMetric(volumes, svms *set.Set, matrixName, metric string, setMetricFunc func(*matrix.Matrix, *matrix.Instance, float64)) error {
	if svms.Size() == 0 || volumes.Size() == 0 {
		return nil
	}

	topClients, err := v.fetchTopClients(volumes.Values(), svms.Values(), metric)
	if err != nil {
		return err
	}

	mat := v.data[matrixName]
	if mat == nil {
		return nil
	}
	for i, client := range topClients {
		clientIP := client.Get("client_ip").String()
		vol := client.Get("volume.name").String()
		svm := client.Get("svm.name").String()
		value := client.Get(metric).Float()
		instance, err := mat.NewInstance(strconv.Itoa(i))
		if err != nil {
			v.Logger.Warn().Str("volume", vol).Msg("error while creating instance")
			continue
		}
		instance.SetLabel("volume", vol)
		instance.SetLabel("svm", svm)
		instance.SetLabel("client_ip", clientIP)
		setMetricFunc(mat, instance, value)
	}
	return nil
}

func (v *Volume) setOpsMetric(mat *matrix.Matrix, instance *matrix.Instance, value float64) {
	var err error
	m := mat.GetMetric("ops")
	if m == nil {
		if m, err = mat.NewMetricFloat64("ops"); err != nil {
			v.Logger.Warn().Err(err).Str("key", "ops").Msg("error while creating metric")
			return
		}
	}
	if err = m.SetValueFloat64(instance, value); err != nil {
		v.Logger.Error().Err(err).Str("metric", "ops").Msg("Unable to set value on metric")
	}
}

func (v *Volume) setDataMetric(mat *matrix.Matrix, instance *matrix.Instance, value float64) {
	var err error
	m := mat.GetMetric("data")
	if m == nil {
		if m, err = mat.NewMetricFloat64("data"); err != nil {
			v.Logger.Warn().Err(err).Str("key", "data").Msg("error while creating metric")
			return
		}
	}
	if err = m.SetValueFloat64(instance, value); err != nil {
		v.Logger.Error().Err(err).Str("metric", "data").Msg("Unable to set value on metric")
	}
}

func (v *Volume) fetchVolumesWithActivityTrackingEnabled() (*set.Set, error) {
	var (
		result []gjson.Result
		err    error
	)
	if v.volumeInterface != nil {
		return v.volumeInterface.fetchVolumesWithActivityTrackingEnabled()
	}
	va := set.New()
	query := "api/storage/volumes"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields([]string{"svm.name", "name"}).
		Filter([]string{"activity_tracking.state=on"}).
		Build()

	if result, err = collectors.InvokeRestCall(v.client, href, v.Logger); err != nil {
		return va, err
	}

	for _, volume := range result {
		volName := volume.Get("name").String()
		svmName := volume.Get("svm.name").String()
		va.Add(svmName + keyToken + volName)
	}
	return va, nil
}

func (v *Volume) fetchTopClients(volumes []string, svms []string, metric string) ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)
	if v.volumeInterface != nil {
		return v.volumeInterface.fetchTopClients(volumes, svms, metric)
	}
	query := "api/storage/volumes/*/top-metrics/clients"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields([]string{"client_ip", "svm", "volume.name", metric}).
		Filter([]string{"top_metric=" + metric, "volume=" + strings.Join(volumes, "|"), "svm=" + strings.Join(svms, "|")}).
		Build()

	if result, err = collectors.InvokeRestCall(v.client, href, v.Logger); err != nil {
		return result, err
	}

	return result, nil
}
