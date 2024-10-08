package volumetopclients

import (
	"cmp"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"
)

type VolumeTracker interface {
	fetchTopClients(volumes *set.Set, svms *set.Set, metric string) ([]gjson.Result, error)
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

var opMetric = "ops"

var dataMetric = "data"

type TopClients struct {
	*plugin.AbstractPlugin
	schedule       int
	client         *rest.Client
	data           map[string]*matrix.Matrix
	cache          *VolumeCache
	maxVolumeCount int
	tracker        VolumeTracker
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
	return &TopClients{AbstractPlugin: p}
}

func (t *TopClients) InitAllMatrix() error {
	t.data = make(map[string]*matrix.Matrix)
	mats := []struct {
		name   string
		object string
		metric string
	}{
		{topClientReadOPSMatrix, "volume_top_clients_read", opMetric},
		{topClientWriteOPSMatrix, "volume_top_clients_write", opMetric},
		{topClientReadDataMatrix, "volume_top_clients_read", dataMetric},
		{topClientWriteDataMatrix, "volume_top_clients_write", dataMetric},
	}

	for _, m := range mats {
		if err := t.initMatrix(m.name, m.object, t.data, m.metric); err != nil {
			return err
		}
	}
	return nil
}

func (t *TopClients) initMatrix(name string, object string, inputMat map[string]*matrix.Matrix, metric string) error {
	matrixName := t.Parent + name
	inputMat[name] = matrix.New(matrixName, object, name)
	for _, v1 := range t.data {
		v1.SetExportOptions(matrix.DefaultExportOptions())
	}
	err := matrix.CreateMetric(metric, inputMat[name])
	if err != nil {
		t.SLogger.Warn("error while creating metric", slogx.Err(err), slog.String("key", metric))
		return err
	}
	return nil
}

func (t *TopClients) Init() error {
	var err error
	if err := t.InitAbc(); err != nil {
		return err
	}

	if err := t.InitAllMatrix(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if t.client, err = rest.New(conf.ZapiPoller(t.ParentParams), timeout, t.Auth); err != nil {
		return err
	}

	if err := t.client.Init(5); err != nil {
		return err
	}

	t.maxVolumeCount = defaultTopN

	if maxVol := t.Params.GetChildContentS("max_volumes"); maxVol != "" {
		if maxVolCount, err := strconv.Atoi(maxVol); err != nil {
			t.maxVolumeCount = defaultTopN
		} else {
			t.maxVolumeCount = min(maxVolCount, maxTopN)
		}
	}
	t.schedule = t.SetPluginInterval()
	t.SLogger.Info("Using", slog.Int("maxVolumeCount", t.maxVolumeCount))
	return nil
}

func (t *TopClients) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[t.Object]
	t.client.Metadata.Reset()
	err := t.InitAllMatrix()
	if err != nil {
		return nil, nil, err
	}
	for k := range t.data {
		// Set all global labels if already not exist
		t.data[k].SetGlobalLabels(data.GetGlobalLabels())
	}
	err = t.processTopClients(data)
	if err != nil {
		return nil, nil, err
	}
	result := make([]*matrix.Matrix, 0, len(t.data))

	var pluginInstances uint64
	for _, value := range t.data {
		instCount := len(value.GetInstances())
		if instCount == 0 {
			continue
		}
		result = append(result, value)
		pluginInstances = +uint64(len(value.GetInstances()))
	}

	t.client.Metadata.PluginInstances = pluginInstances
	return result, t.client.Metadata, err
}

func (t *TopClients) getCachedVolumesWithActivityTracking() (*set.Set, error) {

	var (
		va  *set.Set
		err error
	)
	if t.schedule >= t.PluginInvocationRate {
		va, err = t.fetchVolumesWithActivityTrackingEnabled()
		if err != nil {
			return nil, err
		}
		t.schedule = 0
		// Update the cache
		t.cache = &VolumeCache{
			volumesWithActivityTrackingEnabled: va,
			lastFetched:                        time.Now(),
		}
	} else {
		va = t.cache.volumesWithActivityTrackingEnabled
	}

	t.schedule++
	return va, nil
}

func (t *TopClients) processTopClients(data *matrix.Matrix) error {
	va, err := t.getCachedVolumesWithActivityTracking()
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

	readOpsList, writeOpsList, readDataList, writeDataList := t.collectMetricValues(data, filteredDataInstances)

	topReadOps := t.getTopN(readOpsList)
	topWriteOps := t.getTopN(writeOpsList)
	topReadData := t.getTopN(readDataList)
	topWriteData := t.getTopN(writeDataList)

	readOpsVolumes, readOpsSvms := t.extractVolumesAndSvms(data, topReadOps)
	writeOpsVolumes, writeOpsSvms := t.extractVolumesAndSvms(data, topWriteOps)
	readDataVolumes, readDataSvms := t.extractVolumesAndSvms(data, topReadData)
	writeDataVolumes, writeDataSvms := t.extractVolumesAndSvms(data, topWriteData)

	if err := t.processTopClientsByMetric(readOpsVolumes, readOpsSvms, topClientReadOPSMatrix, "iops.read", opMetric); err != nil {
		return err
	}

	if err := t.processTopClientsByMetric(writeOpsVolumes, writeOpsSvms, topClientWriteOPSMatrix, "iops.write", opMetric); err != nil {
		return err
	}

	if err := t.processTopClientsByMetric(readDataVolumes, readDataSvms, topClientReadDataMatrix, "throughput.read", dataMetric); err != nil {
		return err
	}

	if err := t.processTopClientsByMetric(writeDataVolumes, writeDataSvms, topClientWriteDataMatrix, "throughput.write", dataMetric); err != nil {
		return err
	}

	return nil
}

func (t *TopClients) collectMetricValues(data *matrix.Matrix, filteredDataInstances *set.Set) ([]MetricValue, []MetricValue, []MetricValue, []MetricValue) {
	readOpsMetric := data.GetMetric("total_read_ops")
	writeOpsMetric := data.GetMetric("total_write_ops")
	readDataMetric := data.GetMetric("bytes_read")
	writeDataMetric := data.GetMetric("bytes_written")

	var readOpsList, writeOpsList, readDataList, writeDataList []MetricValue
	for key := range filteredDataInstances.Iter() {
		instanceKey := data.GetInstance(key)
		var readOps, writeOps, readData, writeData float64
		var ro, wo, rb, wb bool
		if readOpsMetric != nil {
			readOps, ro = readOpsMetric.GetValueFloat64(instanceKey)
		}
		if writeOpsMetric != nil {
			writeOps, wo = writeOpsMetric.GetValueFloat64(instanceKey)
		}
		if ro && readOps > 0 {
			readOpsList = append(readOpsList, MetricValue{key: key, value: readOps})
		}
		if wo && writeOps > 0 {
			writeOpsList = append(writeOpsList, MetricValue{key: key, value: writeOps})
		}

		if readDataMetric != nil {
			readData, rb = readDataMetric.GetValueFloat64(instanceKey)
		}
		if writeDataMetric != nil {
			writeData, wb = writeDataMetric.GetValueFloat64(instanceKey)
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

func (t *TopClients) getTopN(metricList []MetricValue) []MetricValue {
	slices.SortFunc(metricList, func(a, b MetricValue) int {
		return cmp.Compare(b.value, a.value) // Sort in descending order
	})

	if len(metricList) > t.maxVolumeCount {
		return metricList[:t.maxVolumeCount]
	}
	return metricList
}

func (t *TopClients) extractVolumesAndSvms(data *matrix.Matrix, topMetrics []MetricValue) (*set.Set, *set.Set) {
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

func (t *TopClients) processTopClientsByMetric(volumes, svms *set.Set, matrixName, metric, metricType string) error {
	if svms.Size() == 0 || volumes.Size() == 0 {
		return nil
	}

	topClients, err := t.fetchTopClients(volumes, svms, metric)
	if err != nil {
		return err
	}

	mat := t.data[matrixName]
	if mat == nil {
		return nil
	}
	for _, client := range topClients {
		clientIP := client.Get("client_ip").String()
		vol := client.Get("volume.name").String()
		svm := client.Get("svm.name").String()
		value := client.Get(metric).Float()
		instanceKey := clientIP + keyToken + vol + keyToken + svm
		instance, err := mat.NewInstance(instanceKey)
		if err != nil {
			t.SLogger.Warn("error while creating instance", slogx.Err(err), slog.String("volume", vol))
			continue
		}
		instance.SetLabel("volume", vol)
		instance.SetLabel("svm", svm)
		instance.SetLabel("client_ip", clientIP)
		t.setMetric(mat, instance, value, metricType)
	}
	return nil
}

func (t *TopClients) setMetric(mat *matrix.Matrix, instance *matrix.Instance, value float64, metricType string) {
	var err error
	m := mat.GetMetric(metricType)
	if m == nil {
		if m, err = mat.NewMetricFloat64(metricType); err != nil {
			t.SLogger.Warn("error while creating metric", slogx.Err(err), slog.String("key", metricType))
			return
		}
	}
	if err = m.SetValueFloat64(instance, value); err != nil {
		t.SLogger.Error("error while setting value", slogx.Err(err), slog.String("metric", metricType))
	}
}

func (t *TopClients) fetchVolumesWithActivityTrackingEnabled() (*set.Set, error) {
	var (
		result []gjson.Result
		err    error
	)
	if t.tracker != nil {
		return t.tracker.fetchVolumesWithActivityTrackingEnabled()
	}
	va := set.New()
	query := "api/storage/volumes"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields([]string{"svm.name", "name"}).
		MaxRecords(collectors.DefaultBatchSize).
		Filter([]string{"activity_tracking.state=on"}).
		Build()

	if result, err = collectors.InvokeRestCall(t.client, href, t.SLogger); err != nil {
		return va, err
	}

	for _, volume := range result {
		volName := volume.Get("name").String()
		svmName := volume.Get("svm.name").String()
		va.Add(svmName + keyToken + volName)
	}
	return va, nil
}

func (t *TopClients) fetchTopClients(volumes *set.Set, svms *set.Set, metric string) ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)
	if t.tracker != nil {
		return t.tracker.fetchTopClients(volumes, svms, metric)
	}
	query := "api/storage/volumes/*/top-metrics/clients"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields([]string{"client_ip", "svm", "volume.name", metric}).
		MaxRecords(collectors.DefaultBatchSize).
		Filter([]string{"top_metric=" + metric, "volume=" + strings.Join(volumes.Values(), "|"), "svm=" + strings.Join(svms.Values(), "|")}).
		Build()

	if result, err = collectors.InvokeRestCall(t.client, href, t.SLogger); err != nil {
		return result, err
	}

	return result, nil
}
