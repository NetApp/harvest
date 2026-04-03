package ssdcachecapacity

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

const (
	volumeMatrixName = "eseries_ssd_cache_volume"
	driveMatrixName  = "eseries_ssd_cache_drive"
)

type driveData struct {
	location    string
	rawCapacity float64
}

type SsdCacheCapacity struct {
	*plugin.AbstractPlugin
	client            *rest.Client
	schedule          int
	volumeNames       map[string]string
	driveInfo         map[string]driveData
	volumeMat         *matrix.Matrix
	driveMat          *matrix.Matrix
	maxFlashCacheSize float64
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SsdCacheCapacity{AbstractPlugin: p}
}

func (s *SsdCacheCapacity) Init(remote conf.Remote) error {
	if err := s.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	poller, err := conf.PollerNamed(s.Options.Poller)
	if err != nil {
		return err
	}

	credentials := auth.NewCredentials(poller, s.SLogger)
	if s.client, err = rest.New(poller, timeout, credentials, ""); err != nil {
		return err
	}

	if err := s.client.Init(1, remote); err != nil {
		return err
	}

	s.volumeNames = make(map[string]string)
	s.driveInfo = make(map[string]driveData)
	s.schedule = s.SetPluginInterval()

	s.initVolumeMatrix()
	s.initDriveMatrix()

	return nil
}

func (s *SsdCacheCapacity) initVolumeMatrix() {
	mat := matrix.New(s.Parent+"."+volumeMatrixName, volumeMatrixName, volumeMatrixName)
	exportOptions := node.NewS("export_options")

	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "ssd_cache")
	instanceKeys.NewChildS("", "ssd_cache_id")
	instanceKeys.NewChildS("", "volume")

	instanceLabels := exportOptions.NewChildS("instance_labels", "")
	instanceLabels.NewChildS("", "volume")

	mat.SetExportOptions(exportOptions)

	s.volumeMat = mat
}

func (s *SsdCacheCapacity) initDriveMatrix() {
	mat := matrix.New(s.Parent+"."+driveMatrixName, driveMatrixName, driveMatrixName)
	exportOptions := node.NewS("export_options")

	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "ssd_cache")
	instanceKeys.NewChildS("", "ssd_cache_id")
	instanceKeys.NewChildS("", "drive")

	instanceLabels := exportOptions.NewChildS("instance_labels", "")
	instanceLabels.NewChildS("", "drive")

	mat.SetExportOptions(exportOptions)
	_, _ = mat.NewMetricFloat64("raw_capacity")

	s.driveMat = mat
}

func (s *SsdCacheCapacity) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[s.Object]

	arrayID := s.ParentParams.GetChildContentS("array_id")
	if arrayID == "" {
		s.SLogger.Warn("arrayID not found in ParentParams, skipping SSD cache capacity")
		return nil, nil, nil
	}

	if s.schedule >= s.PluginInvocationRate {
		s.schedule = 0
		s.rebuildCaches(arrayID)
	}
	s.schedule++

	globalLabels := data.GetGlobalLabels()
	s.volumeMat.PurgeInstances()
	s.volumeMat.Reset()
	for k, v := range globalLabels {
		s.volumeMat.SetGlobalLabel(k, v)
	}

	s.driveMat.PurgeInstances()
	s.driveMat.Reset()
	for k, v := range globalLabels {
		s.driveMat.SetGlobalLabel(k, v)
	}

	s.populateMappings(data)
	s.populateCapacityMetrics(data)

	totalInstances := len(s.volumeMat.GetInstances()) + len(s.driveMat.GetInstances())
	s.SLogger.Debug("SsdCacheCapacity plugin",
		slog.Int("volumes", len(s.volumeMat.GetInstances())),
		slog.Int("drives", len(s.driveMat.GetInstances())),
	)

	metadata := &collector.Metadata{}
	//nolint:gosec
	metadata.PluginInstances = uint64(totalInstances)

	return []*matrix.Matrix{s.volumeMat, s.driveMat}, metadata, nil
}

func (s *SsdCacheCapacity) rebuildCaches(arrayID string) {
	s.volumeNames = make(map[string]string)
	s.driveInfo = make(map[string]driveData)

	if err := s.buildVolumeCache(arrayID); err != nil {
		s.SLogger.Warn("Failed to build volume cache", slogx.Err(err))
	}

	if err := s.buildDriveCache(arrayID); err != nil {
		s.SLogger.Warn("Failed to build drive cache", slogx.Err(err))
	}

	if err := s.buildCapabilitiesCache(arrayID); err != nil {
		s.SLogger.Warn("Failed to build capabilities cache", slogx.Err(err))
	}
}

func (s *SsdCacheCapacity) buildCapabilitiesCache(arrayID string) error {
	apiPath := s.client.APIPath + "/storage-systems/" + arrayID + "/capabilities"
	results, err := s.client.Fetch(apiPath, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch capabilities: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("empty capabilities response for array %s", arrayID)
	}

	maxFlashCacheSizeStr := results[0].Get("featureParameters.maxFlashCacheSize").ClonedString()
	if maxFlashCacheSizeStr == "" {
		s.SLogger.Warn("featureParameters.maxFlashCacheSize not found in capabilities response",
			slog.String("array_id", arrayID))
		return nil
	}

	v, parseErr := strconv.ParseFloat(maxFlashCacheSizeStr, 64)
	if parseErr != nil {
		s.SLogger.Warn("Failed to parse maxFlashCacheSize",
			slog.String("value", maxFlashCacheSizeStr), slogx.Err(parseErr))
		return nil
	}

	s.maxFlashCacheSize = v
	s.SLogger.Debug("Built capabilities cache",
		slog.Float64("maxFlashCacheSize", v))
	return nil
}

func (s *SsdCacheCapacity) buildVolumeCache(arrayID string) error {
	apiPath := s.client.APIPath + "/storage-systems/" + arrayID + "/volumes"
	volumes, err := s.client.Fetch(apiPath, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch volumes: %w", err)
	}

	for _, vol := range volumes {
		id := vol.Get("id").ClonedString()
		if id == "" {
			id = vol.Get("volumeRef").ClonedString()
		}
		name := vol.Get("name").ClonedString()
		if id != "" && name != "" {
			s.volumeNames[id] = name
		}
	}

	s.SLogger.Debug("Built volume cache", slog.Int("count", len(s.volumeNames)))
	return nil
}

func (s *SsdCacheCapacity) buildDriveCache(arrayID string) error {
	apiPath := s.client.APIPath + "/storage-systems/" + arrayID + "/drives"
	drives, err := s.client.Fetch(apiPath, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch drives: %w", err)
	}

	for _, drive := range drives {
		id := drive.Get("id").ClonedString()
		if id == "" {
			id = drive.Get("driveRef").ClonedString()
		}
		if id == "" {
			continue
		}

		location := drive.Get("physicalLocation.label").ClonedString()
		if location == "" {
			slot := drive.Get("physicalLocation.slot").ClonedString()
			if slot != "" {
				location = "Slot " + slot
			} else {
				location = id
			}
		}

		var rawCap float64
		if capStr := drive.Get("rawCapacity").ClonedString(); capStr != "" {
			if v, parseErr := strconv.ParseFloat(capStr, 64); parseErr == nil {
				rawCap = v
			}
		}

		s.driveInfo[id] = driveData{location: location, rawCapacity: rawCap}
	}

	s.SLogger.Debug("Built drive cache", slog.Int("count", len(s.driveInfo)))
	return nil
}

func (s *SsdCacheCapacity) populateMappings(data *matrix.Matrix) {
	driveCapMetric := s.driveMat.GetMetric("raw_capacity")

	for cacheID, cacheInstance := range data.GetInstances() {
		cacheName := cacheInstance.GetLabel("ssd_cache")
		if cacheName == "" {
			cacheName = cacheID
		}

		volumeIDsJSON := cacheInstance.GetLabel("cached_volume_ids")
		if volumeIDsJSON != "" && volumeIDsJSON != "[]" {
			parsed := gjson.Parse(volumeIDsJSON)
			parsed.ForEach(func(_, vol gjson.Result) bool {
				volID := vol.ClonedString()
				if volID == "" {
					return true
				}
				volName := volID
				if n, ok := s.volumeNames[volID]; ok {
					volName = n
				}
				instKey := cacheID + "_" + volID
				inst, err := s.volumeMat.NewInstance(instKey)
				if err != nil {
					s.SLogger.Error("Failed to create volume instance", slogx.Err(err), slog.String("key", instKey))
					return true
				}
				inst.SetLabelTrimmed("ssd_cache", cacheName)
				inst.SetLabelTrimmed("ssd_cache_id", cacheID)
				inst.SetLabelTrimmed("volume", volName)
				return true
			})
		}

		driveIDsJSON := cacheInstance.GetLabel("drive_ids")
		if driveIDsJSON != "" && driveIDsJSON != "[]" {
			parsed := gjson.Parse(driveIDsJSON)
			parsed.ForEach(func(_, drv gjson.Result) bool {
				driveID := drv.ClonedString()
				if driveID == "" {
					return true
				}
				location := driveID
				var rawCap float64
				if info, ok := s.driveInfo[driveID]; ok {
					location = info.location
					rawCap = info.rawCapacity
				}
				instKey := cacheID + "_" + driveID
				inst, err := s.driveMat.NewInstance(instKey)
				if err != nil {
					s.SLogger.Error("Failed to create drive instance", slogx.Err(err), slog.String("key", instKey))
					return true
				}
				inst.SetLabelTrimmed("ssd_cache", cacheName)
				inst.SetLabelTrimmed("ssd_cache_id", cacheID)
				inst.SetLabelTrimmed("drive", location)
				if driveCapMetric != nil {
					driveCapMetric.SetValueFloat64(inst, rawCap)
				}
				return true
			})
		}
	}
}

func (s *SsdCacheCapacity) populateCapacityMetrics(data *matrix.Matrix) {
	if s.maxFlashCacheSize == 0 {
		s.SLogger.Debug("maxFlashCacheSize is 0, skipping capacity metrics")
		return
	}

	maxCapMetric := data.GetMetric("max_capacity")
	if maxCapMetric == nil {
		var err error
		maxCapMetric, err = data.NewMetricFloat64("max_capacity")
		if err != nil {
			s.SLogger.Error("Failed to create max_capacity metric", slogx.Err(err))
			return
		}
	}

	additionalCapMetric := data.GetMetric("additional_capacity")
	if additionalCapMetric == nil {
		var err error
		additionalCapMetric, err = data.NewMetricFloat64("additional_capacity")
		if err != nil {
			s.SLogger.Error("Failed to create additional_capacity metric", slogx.Err(err))
			return
		}
	}

	usedCapMetric := data.GetMetric("usedCapacity")

	for _, instance := range data.GetInstances() {
		maxCapMetric.SetValueFloat64(instance, s.maxFlashCacheSize)

		var usedCap float64
		if usedCapMetric != nil {
			if v, ok := usedCapMetric.GetValueFloat64(instance); ok {
				usedCap = v
			}
		}
		additionalCapMetric.SetValueFloat64(instance, s.maxFlashCacheSize-usedCap)
	}
}
