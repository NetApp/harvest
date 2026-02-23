package volumetopmetrics

import (
	"cmp"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

type VolumeTracker interface {
	fetchTopClients(volumes *set.Set, svms *set.Set, metric string) ([]gjson.Result, error)
	fetchTopFiles(volumes *set.Set, svms *set.Set, metric string) ([]gjson.Result, error)
	fetchTopUsers(volumes *set.Set, svms *set.Set, metric string) ([]gjson.Result, error)
	fetchVolumesWithActivityTrackingEnabled() (*set.Set, error)
	processTopClients(data *TopMetricsData) error
	processTopFiles(data *TopMetricsData) error
	processTopUsers(data *TopMetricsData) error
}

const (
	keyToken                 = "?#"
	topClientReadOPSMatrix   = "volume_top_clients_read_ops"
	topClientWriteOPSMatrix  = "volume_top_clients_write_ops"
	topClientReadDataMatrix  = "volume_top_clients_read_data"
	topClientWriteDataMatrix = "volume_top_clients_write_data"
	topFileReadOPSMatrix     = "volume_top_files_read_ops"
	topFileWriteOPSMatrix    = "volume_top_files_write_ops"
	topFileReadDataMatrix    = "volume_top_files_read_data"
	topFileWriteDataMatrix   = "volume_top_files_write_data"
	topUserReadOPSMatrix     = "volume_top_users_read_ops"
	topUserWriteOPSMatrix    = "volume_top_users_write_ops"
	topUserReadDataMatrix    = "volume_top_users_read_data"
	topUserWriteDataMatrix   = "volume_top_users_write_data"
	defaultTopN              = 5
	maxTopN                  = 50
)

var opMetric = "ops"

var dataMetric = "data"

type TopMetrics struct {
	*plugin.AbstractPlugin
	schedule             int
	client               *rest.Client
	data                 map[string]*matrix.Matrix
	cache                *VolumeCache
	maxVolumeCount       int
	tracker              VolumeTracker
	clientMetricsEnabled bool
	fileMetricsEnabled   bool
	userMetricsEnabled   bool
}

type TopMetricsData struct {
	readOpsVolumes   *set.Set
	readOpsSvms      *set.Set
	writeOpsVolumes  *set.Set
	writeOpsSvms     *set.Set
	readDataVolumes  *set.Set
	readDataSvms     *set.Set
	writeDataVolumes *set.Set
	writeDataSvms    *set.Set
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
	return &TopMetrics{AbstractPlugin: p}
}

func (t *TopMetrics) InitAllMatrix() error {
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
		{topFileReadOPSMatrix, "volume_top_files_read", opMetric},
		{topFileWriteOPSMatrix, "volume_top_files_write", opMetric},
		{topFileReadDataMatrix, "volume_top_files_read", dataMetric},
		{topFileWriteDataMatrix, "volume_top_files_write", dataMetric},
		{topUserReadOPSMatrix, "volume_top_users_read", opMetric},
		{topUserWriteOPSMatrix, "volume_top_users_write", opMetric},
		{topUserReadDataMatrix, "volume_top_users_read", dataMetric},
		{topUserWriteDataMatrix, "volume_top_users_write", dataMetric},
	}

	for _, m := range mats {
		if err := t.initMatrix(m.name, m.object, t.data, m.metric); err != nil {
			return err
		}
	}
	return nil
}

func (t *TopMetrics) initMatrix(name string, object string, inputMat map[string]*matrix.Matrix, metric string) error {
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

func (t *TopMetrics) Init(remote conf.Remote) error {
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

	if err := t.client.Init(5, remote); err != nil {
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

	// enable client, file and user metrics collection by default
	t.clientMetricsEnabled = true
	t.fileMetricsEnabled = true
	t.userMetricsEnabled = true

	if objects := t.Params.GetChildS("objects"); objects != nil {
		o := objects.GetAllChildContentS()
		if !slices.Contains(o, "client") {
			t.clientMetricsEnabled = false
		}
		if !slices.Contains(o, "file") {
			t.fileMetricsEnabled = false
		}
		if !slices.Contains(o, "user") {
			t.userMetricsEnabled = false
		}
	}
	t.schedule = t.SetPluginInterval()
	t.SLogger.Info("Using", slog.Int("maxVolumeCount", t.maxVolumeCount))
	return nil
}

func (t *TopMetrics) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	// if client, file and user metrics are all disabled then return
	if !t.clientMetricsEnabled && !t.fileMetricsEnabled && !t.userMetricsEnabled {
		return nil, nil, nil
	}

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

	metricsData, err := t.processTopMetrics(data)
	if err != nil {
		return nil, t.client.Metadata, err
	}

	if metricsData == nil {
		return nil, t.client.Metadata, nil
	}

	if t.clientMetricsEnabled {
		err = t.processTopClients(metricsData)
		if err != nil {
			return nil, t.client.Metadata, err
		}
	}
	if t.fileMetricsEnabled {
		err = t.processTopFiles(metricsData)
		if err != nil {
			return nil, t.client.Metadata, err
		}
	}
	if t.userMetricsEnabled {
		err = t.processTopUsers(metricsData)
		if err != nil {
			return nil, t.client.Metadata, err
		}
	}

	result := make([]*matrix.Matrix, 0, len(t.data))

	var pluginInstances uint64
	for _, value := range t.data {
		instCount := len(value.GetInstances())
		if instCount == 0 {
			continue
		}
		result = append(result, value)
		pluginInstances += uint64(len(value.GetInstances()))
	}

	t.client.Metadata.PluginInstances = pluginInstances
	return result, t.client.Metadata, err
}

func (t *TopMetrics) getCachedVolumesWithActivityTracking() (*set.Set, error) {

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

func (t *TopMetrics) processTopClients(metricsData *TopMetricsData) error {
	if err := t.processTopClientsByMetric(metricsData.readOpsVolumes, metricsData.readOpsSvms, topClientReadOPSMatrix, "iops.read", opMetric); err != nil {
		return err
	}

	if err := t.processTopClientsByMetric(metricsData.writeOpsVolumes, metricsData.writeOpsSvms, topClientWriteOPSMatrix, "iops.write", opMetric); err != nil {
		return err
	}

	if err := t.processTopClientsByMetric(metricsData.readDataVolumes, metricsData.readDataSvms, topClientReadDataMatrix, "throughput.read", dataMetric); err != nil {
		return err
	}

	if err := t.processTopClientsByMetric(metricsData.writeDataVolumes, metricsData.writeDataSvms, topClientWriteDataMatrix, "throughput.write", dataMetric); err != nil {
		return err
	}

	return nil
}

func (t *TopMetrics) processTopFiles(metricsData *TopMetricsData) error {
	if err := t.processTopFilesByMetric(metricsData.readOpsVolumes, metricsData.readOpsSvms, topFileReadOPSMatrix, "iops.read", opMetric); err != nil {
		return err
	}

	if err := t.processTopFilesByMetric(metricsData.writeOpsVolumes, metricsData.writeOpsSvms, topFileWriteOPSMatrix, "iops.write", opMetric); err != nil {
		return err
	}

	if err := t.processTopFilesByMetric(metricsData.readDataVolumes, metricsData.readDataSvms, topFileReadDataMatrix, "throughput.read", dataMetric); err != nil {
		return err
	}

	if err := t.processTopFilesByMetric(metricsData.writeDataVolumes, metricsData.writeDataSvms, topFileWriteDataMatrix, "throughput.write", dataMetric); err != nil {
		return err
	}

	return nil
}

func (t *TopMetrics) processTopUsers(metricsData *TopMetricsData) error {
	if err := t.processTopUsersByMetric(metricsData.readOpsVolumes, metricsData.readOpsSvms, topUserReadOPSMatrix, "iops.read", opMetric); err != nil {
		return err
	}

	if err := t.processTopUsersByMetric(metricsData.writeOpsVolumes, metricsData.writeOpsSvms, topUserWriteOPSMatrix, "iops.write", opMetric); err != nil {
		return err
	}

	if err := t.processTopUsersByMetric(metricsData.readDataVolumes, metricsData.readDataSvms, topUserReadDataMatrix, "throughput.read", dataMetric); err != nil {
		return err
	}

	if err := t.processTopUsersByMetric(metricsData.writeDataVolumes, metricsData.writeDataSvms, topUserWriteDataMatrix, "throughput.write", dataMetric); err != nil {
		return err
	}

	return nil
}

func (t *TopMetrics) processTopMetrics(data *matrix.Matrix) (*TopMetricsData, error) {
	va, err := t.getCachedVolumesWithActivityTracking()
	if err != nil {
		return nil, err
	}

	if va.Size() == 0 {
		return nil, nil
	}

	filteredDataInstances := set.New()
	for key, value := range data.GetInstances() {
		if !value.IsExportable() {
			continue
		}
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

	return &TopMetricsData{
		readOpsVolumes:   readOpsVolumes,
		readOpsSvms:      readOpsSvms,
		writeOpsVolumes:  writeOpsVolumes,
		writeOpsSvms:     writeOpsSvms,
		readDataVolumes:  readDataVolumes,
		readDataSvms:     readDataSvms,
		writeDataVolumes: writeDataVolumes,
		writeDataSvms:    writeDataSvms,
	}, nil
}

func (t *TopMetrics) collectMetricValues(data *matrix.Matrix, filteredDataInstances *set.Set) ([]MetricValue, []MetricValue, []MetricValue, []MetricValue) {
	readOpsMetric := data.DisplayMetric("read_ops")
	writeOpsMetric := data.DisplayMetric("write_ops")
	readDataMetric := data.DisplayMetric("read_data")
	writeDataMetric := data.DisplayMetric("write_data")

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

func (t *TopMetrics) getTopN(metricList []MetricValue) []MetricValue {
	slices.SortFunc(metricList, func(a, b MetricValue) int {
		return cmp.Compare(b.value, a.value) // Sort in descending order
	})

	if len(metricList) > t.maxVolumeCount {
		return metricList[:t.maxVolumeCount]
	}
	return metricList
}

func (t *TopMetrics) extractVolumesAndSvms(data *matrix.Matrix, topMetrics []MetricValue) (*set.Set, *set.Set) {
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

func (t *TopMetrics) processTopFilesByMetric(volumes, svms *set.Set, matrixName, metric, metricType string) error {
	if svms.Size() == 0 || volumes.Size() == 0 {
		return nil
	}

	topFiles, err := t.fetchTopFiles(volumes, svms, metric)
	if err != nil {
		return err
	}

	mat := t.data[matrixName]
	if mat == nil {
		return nil
	}
	for _, client := range topFiles {
		path := client.Get("path").ClonedString()
		vol := client.Get("volume.name").ClonedString()
		svm := client.Get("svm.name").ClonedString()
		value := client.Get(metric).Float()
		instanceKey := path + keyToken + vol + keyToken + svm
		instance, err := mat.NewInstance(instanceKey)
		if err != nil {
			t.SLogger.Warn("error while creating instance", slogx.Err(err), slog.String("volume", vol))
			continue
		}
		instance.SetLabel("volume", vol)
		instance.SetLabel("svm", svm)
		instance.SetLabel("path", path)
		t.setMetric(mat, instance, value, metricType)
	}
	return nil
}

func (t *TopMetrics) processTopClientsByMetric(volumes, svms *set.Set, matrixName, metric, metricType string) error {
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
		clientIP := client.Get("client_ip").ClonedString()
		vol := client.Get("volume.name").ClonedString()
		svm := client.Get("svm.name").ClonedString()
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

func (t *TopMetrics) processTopUsersByMetric(volumes, svms *set.Set, matrixName, metric, metricType string) error {
	if svms.Size() == 0 || volumes.Size() == 0 {
		return nil
	}

	topUsers, err := t.fetchTopUsers(volumes, svms, metric)
	if err != nil {
		return err
	}

	mat := t.data[matrixName]
	if mat == nil {
		return nil
	}
	for _, user := range topUsers {
		userName := user.Get("user_name").ClonedString()
		userID := user.Get("user_id").ClonedString()
		vol := user.Get("volume.name").ClonedString()
		svm := user.Get("svm.name").ClonedString()
		value := user.Get(metric).Float()
		instanceKey := userID + keyToken + vol + keyToken + svm
		instance, err := mat.NewInstance(instanceKey)
		if err != nil {
			t.SLogger.Warn("error while creating instance", slogx.Err(err), slog.String("volume", vol))
			continue
		}
		instance.SetLabel("volume", vol)
		instance.SetLabel("svm", svm)
		instance.SetLabel("user_name", userName)
		instance.SetLabel("user_id", userID)
		t.setMetric(mat, instance, value, metricType)
	}
	return nil
}

func (t *TopMetrics) setMetric(mat *matrix.Matrix, instance *matrix.Instance, value float64, metricType string) {
	var err error
	m := mat.GetMetric(metricType)
	if m == nil {
		if m, err = mat.NewMetricFloat64(metricType); err != nil {
			t.SLogger.Warn("error while creating metric", slogx.Err(err), slog.String("key", metricType))
			return
		}
	}
	m.SetValueFloat64(instance, value)
}

func (t *TopMetrics) fetchVolumesWithActivityTrackingEnabled() (*set.Set, error) {
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

	if result, err = collectors.InvokeRestCall(t.client, href); err != nil {
		return va, err
	}

	for _, volume := range result {
		volName := volume.Get("name").ClonedString()
		svmName := volume.Get("svm.name").ClonedString()
		va.Add(svmName + keyToken + volName)
	}
	return va, nil
}

func (t *TopMetrics) fetchTopClients(volumes *set.Set, svms *set.Set, metric string) ([]gjson.Result, error) {
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
		Filter([]string{"top_metric=" + metric, "volume.name=" + strings.Join(volumes.Values(), "|"), "svm.name=" + strings.Join(svms.Values(), "|")}).
		Build()

	if result, err = collectors.InvokeRestCall(t.client, href); err != nil {
		return result, err
	}

	return result, nil
}

func (t *TopMetrics) fetchTopFiles(volumes *set.Set, svms *set.Set, metric string) ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)
	if t.tracker != nil {
		return t.tracker.fetchTopFiles(volumes, svms, metric)
	}
	query := "api/storage/volumes/*/top-metrics/files"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields([]string{"path", "svm", "volume.name", metric}).
		MaxRecords(collectors.DefaultBatchSize).
		Filter([]string{"top_metric=" + metric, "volume.name=" + strings.Join(volumes.Values(), "|"), "svm.name=" + strings.Join(svms.Values(), "|")}).
		Build()

	if result, err = collectors.InvokeRestCall(t.client, href); err != nil {
		return result, err
	}

	return result, nil
}

func (t *TopMetrics) fetchTopUsers(volumes *set.Set, svms *set.Set, metric string) ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)
	if t.tracker != nil {
		return t.tracker.fetchTopUsers(volumes, svms, metric)
	}
	query := "api/storage/volumes/*/top-metrics/users"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields([]string{"user_name", "user_id", "svm", "volume.name", metric}).
		MaxRecords(collectors.DefaultBatchSize).
		Filter([]string{"top_metric=" + metric, "volume.name=" + strings.Join(volumes.Values(), "|"), "svm.name=" + strings.Join(svms.Values(), "|")}).
		Build()

	if result, err = collectors.InvokeRestCall(t.client, href); err != nil {
		return result, err
	}

	return result, nil
}
