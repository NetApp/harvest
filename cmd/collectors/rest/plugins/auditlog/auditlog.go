package auditlog

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
	"strings"
	"time"
)

const (
	auditMatrix             = "audit"
	defaultDataPollDuration = 3 * time.Minute
)

type CacheRefresher interface {
	RefreshCache() error
}

var defaultFields = []string{"application", "location", "user", "timestamp", "state"}
var volumeCache = make(map[string]VolumeInfo)
var volumeCacheCopy = make(map[string]VolumeInfo)
var cacheForceRefresh bool

type VolumeInfo struct {
	Name string
	SVM  string
}

type AuditLog struct {
	*plugin.AbstractPlugin
	schedule        int
	data            *matrix.Matrix
	client          *rest.Client
	rootConfig      RootConfig
	lastFilterTimes map[string]int64
}

var metrics = []string{
	"log",
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &AuditLog{AbstractPlugin: p}
}

func (a *AuditLog) Init(remote conf.Remote) error {
	if err := a.AbstractPlugin.Init(remote); err != nil {
		return err
	}

	err := a.initMatrix()
	if err != nil {
		return err
	}

	err = a.populateAuditLogConfig()
	if err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if a.client, err = rest.New(conf.ZapiPoller(a.ParentParams), timeout, a.Auth); err != nil {
		return err
	}
	a.schedule = a.SetPluginInterval()
	a.lastFilterTimes = make(map[string]int64)

	return a.client.Init(5, remote)
}

func (a *AuditLog) initMatrix() error {
	a.data = matrix.New(a.Parent+auditMatrix, auditMatrix, auditMatrix)
	a.data.SetExportOptions(matrix.DefaultExportOptions())
	for _, k := range metrics {
		err := matrix.CreateMetric(k, a.data)
		if err != nil {
			a.SLogger.Warn("error while creating metric", slogx.Err(err), slog.String("key", k))
			return err
		}
	}
	return nil
}

func (a *AuditLog) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	a.client.Metadata.Reset()

	if a.schedule >= a.PluginInvocationRate {
		a.schedule = 0
		err := a.populateVolumeCache()
		if err != nil {
			return nil, nil, err
		}
	}
	a.schedule++
	cacheForceRefresh = false

	err := a.initMatrix()
	if err != nil {
		return nil, nil, err
	}
	data := dataMap[a.Object]
	a.data.SetGlobalLabels(data.GetGlobalLabels())

	clusterTime, err := collectors.GetClusterTime(a.client, nil, a.SLogger)
	if err != nil {
		return nil, nil, err
	}

	if a.HasVolumeConfig() {
		// process volume rootConfig
		volume := a.rootConfig.AuditLog.Volume
		var actions = make([]string, len(volume.Action))
		for _, action := range volume.Action {
			actions = append(actions, fmt.Sprintf("*%s*", action))
		}
		state := volume.Filter.State
		timestampFilter := a.getTimeStampFilter(clusterTime, a.lastFilterTimes["volume"])
		href := a.constructAuditLogURL(actions, state, timestampFilter)
		records, err := collectors.InvokeRestCall(a.client, href)
		if err != nil {
			return nil, nil, err
		}
		a.parseVolumeRecords(records)
		// update lastFilterTime to current cluster time
		a.lastFilterTimes["volume"] = clusterTime.Unix()
	}

	return []*matrix.Matrix{a.data}, a.client.Metadata, nil
}

func (a *AuditLog) HasVolumeConfig() bool {
	return len(a.rootConfig.AuditLog.Volume.Action) > 0
}

func (a *AuditLog) populateAuditLogConfig() error {
	var err error

	a.rootConfig, err = InitAuditLogConfig()
	if err != nil {
		return err
	}
	return nil
}

func (a *AuditLog) constructAuditLogURL(actions []string, state string, timestampFilter string) string {
	actionFilter := "input=" + strings.Join(actions, "|")
	stateFilter := "state=" + state

	// Construct the Href
	href := rest.NewHrefBuilder().
		APIPath("api/security/audit/messages").
		Fields(defaultFields).
		Filter([]string{timestampFilter, actionFilter, stateFilter}).
		Build()

	return href
}

func (a *AuditLog) getTimeStampFilter(clusterTime time.Time, lastFilterTime int64) string {
	fromTime := lastFilterTime
	// check if this is the first request
	if lastFilterTime == 0 {
		// if first request fetch cluster time
		dataDuration, err := collectors.GetDataInterval(a.ParentParams, defaultDataPollDuration)
		if err != nil {
			a.SLogger.Warn(
				"Failed to parse duration. using default",
				slogx.Err(err),
				slog.String("defaultDataPollDuration", defaultDataPollDuration.String()),
			)
		}
		fromTime = clusterTime.Add(-dataDuration).Unix()
	}
	return fmt.Sprintf("timestamp=>=%d", fromTime)
}

func (a *AuditLog) populateVolumeCache() error {
	volumeCacheCopy = make(map[string]VolumeInfo)
	for key, value := range volumeCache {
		volumeCacheCopy[key] = value
	}
	volumeCache = make(map[string]VolumeInfo)
	query := "api/storage/volumes"
	href := rest.NewHrefBuilder().
		APIPath(query).
		MaxRecords(collectors.DefaultBatchSize).
		Fields([]string{"svm.name", "uuid", "name"}).
		Build()

	records, err := rest.FetchAll(a.client, href)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	for _, volume := range records {

		if !volume.IsObject() {
			a.SLogger.Warn("volume is not object, skipping", slog.String("type", volume.Type.String()))
			continue
		}
		uuid := volume.Get("uuid").ClonedString()
		name := volume.Get("name").ClonedString()
		svm := volume.Get("svm.name").ClonedString()
		volumeCache[uuid] = VolumeInfo{
			Name: name,
			SVM:  svm,
		}
	}
	return nil
}

func GetVolumeInfo(uuid string) (VolumeInfo, bool) {
	volumeInfo, exists := volumeCache[uuid]
	if exists {
		return volumeInfo, true
	}
	volumeInfo, exists = volumeCacheCopy[uuid]
	return volumeInfo, exists
}

func (a *AuditLog) setLogMetric(mat *matrix.Matrix, instance *matrix.Instance, value float64) error {
	m := mat.GetMetric("log")
	if m != nil {
		if err := m.SetValueFloat64(instance, value); err != nil {
			return err
		}
	}
	return nil
}

func (a *AuditLog) RefreshCache() error {
	if !cacheForceRefresh {
		a.SLogger.Debug("refreshing cache via handler")
		err := a.populateVolumeCache()
		if err != nil {
			return err
		}
		cacheForceRefresh = true
	}
	return nil
}
