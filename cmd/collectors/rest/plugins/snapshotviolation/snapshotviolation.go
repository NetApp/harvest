package snapshotviolation

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/collectors/plugins/snapshotviolation"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"strconv"
	"time"
)

type SnapshotViolation struct {
	*plugin.AbstractPlugin
	client   *rest.Client
	data     *matrix.Matrix
	schedule int
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SnapshotViolation{AbstractPlugin: p}
}

func (s *SnapshotViolation) Init(remote conf.Remote) error {

	var err error

	if err := s.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if s.client, err = rest.New(conf.ZapiPoller(s.ParentParams), timeout, s.Auth); err != nil {
		s.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := s.client.Init(5, remote); err != nil {
		return err
	}
	s.data, err = snapshotviolation.InitMatrix(s.Parent)
	if err != nil {
		return fmt.Errorf("error while initializing matrix: %w", err)
	}
	s.schedule = s.SetPluginInterval()
	return nil
}

func (s *SnapshotViolation) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	s.client.Metadata.Reset()

	if s.schedule >= s.PluginInvocationRate {
		s.schedule = 0
		s.populateSnapshotViolation(dataMap)
	}
	s.schedule++
	s.client.Metadata.PluginInstances = uint64(len(s.data.GetInstances()))
	return []*matrix.Matrix{s.data}, s.client.Metadata, nil
}

func (s *SnapshotViolation) populateSnapshotViolation(dataMap map[string]*matrix.Matrix) {
	// Purge and reset data
	data := dataMap[s.Object]
	s.data.PurgeInstances()
	s.data.Reset()

	// Set all global labels from Rest.go if already not exist
	s.data.SetGlobalLabels(data.GetGlobalLabels())

	// Map to store prefixes by svm key
	prefixMap := make(map[string]*set.Set)

	for _, instance := range data.GetInstances() {
		copies := instance.GetLabel("copies")
		copiesJSON := gjson.Result{Type: gjson.JSON, Raw: "[" + copies + "]"}

		svm := instance.GetLabel("svm")
		for _, copiesData := range copiesJSON.Array() {
			// Extract prefix from copiesData and add to map
			prefix := copiesData.Get("prefix").ClonedString()
			if prefix != "" {
				if _, ok := prefixMap[svm]; !ok {
					prefixMap[svm] = set.New()
				}
				prefixMap[svm].Add(prefix)
			}
		}
	}

	filteredSnapshotStats := s.getFilteredVolumeSnapshotStats(prefixMap)
	for _, stats := range filteredSnapshotStats {
		svm := stats.Svm
		volume := stats.Volume
		key := svm + snapshotviolation.KeyToken + volume

		instance := s.data.GetInstance(key)
		if instance == nil {
			var err error
			instance, err = s.data.NewInstance(key)
			if err != nil {
				s.SLogger.Error("Failed to create instance", slog.String("key", key), slogx.Err(err))
				continue
			}

			instance.SetLabel("svm", svm)
			instance.SetLabel("volume", volume)
		}

		if metric := s.data.GetMetric(snapshotviolation.ViolationCount); metric != nil {
			metric.SetValueInt64(instance, int64(stats.Count))
		}

		if metric := s.data.GetMetric(snapshotviolation.ViolationTotalSize); metric != nil {
			metric.SetValueInt64(instance, stats.TotalSize)
		}

	}
}

func (s *SnapshotViolation) getFilteredVolumeSnapshotStats(prefixMap map[string]*set.Set) map[string]snapshotviolation.Stats {
	filteredSnapshotStats := make(map[string]snapshotviolation.Stats)

	fields := []string{"vserver", "volume", "snapshot", "size"}
	query := "api/private/cli/volume/snapshot"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(collectors.DefaultBatchSize).
		Build()

	// Define a callback function to process each batch of records
	processBatch := func(records []gjson.Result, _ int64) error {
		for _, sData := range records {
			svm := sData.Get("vserver").ClonedString()
			volume := sData.Get("volume").ClonedString()
			snapshot := sData.Get("snapshot").ClonedString()
			sizeStr := sData.Get("size").ClonedString()

			if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
				snapshotviolation.ProcessSnapshotData(svm, volume, snapshot, size, prefixMap, filteredSnapshotStats)
			} else {
				// If conversion fails, log warning and use original value
				s.SLogger.Warn("Failed to convert snapshot size from KB to bytes",
					slog.String("svm", svm),
					slog.String("volume", volume),
					slog.String("snapshot", snapshot),
					slog.String("size", sizeStr),
					slogx.Err(err))
				continue
			}
		}
		return nil
	}

	err := rest.FetchAllStream(s.client, href, processBatch)
	if err != nil {
		s.SLogger.Error("Failed to fetch data", slogx.Err(err), slog.String("href", href))
		return filteredSnapshotStats
	}

	return filteredSnapshotStats
}
