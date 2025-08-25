/*
 * Copyright NetApp Inc, 2024 All rights reserved
 */

package snapshotviolation

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/collectors/plugins/snapshotviolation"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"strconv"
	"strings"
)

type SnapshotViolation struct {
	*plugin.AbstractPlugin
	data     *matrix.Matrix
	client   *zapi.Client
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

	if s.client, err = zapi.New(conf.ZapiPoller(s.ParentParams), s.Auth); err != nil {
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
		prefixes := strings.Split(instance.GetLabel("prefix"), ",")
		policyOwner := instance.GetLabel("policy_owner")
		svm := instance.GetLabel("svm")
		for _, prefix := range prefixes {
			if policyOwner == "cluster-admin" {
				svm = ""
			}
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

	request := node.NewXMLS("snapshot-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)
	desired := node.NewXMLS("desired-attributes")
	snapshotInfo := node.NewXMLS("snapshot-info")
	snapshotInfo.NewChildS("name", "")
	snapshotInfo.NewChildS("vserver", "")
	snapshotInfo.NewChildS("volume", "")
	snapshotInfo.NewChildS("total", "")
	desired.AddChild(snapshotInfo)
	request.AddChild(desired)
	query := request.NewChildS("query", "")
	snapshotInfoQuery := query.NewChildS("snapshot-info", "")
	snapshotInfoQuery.NewChildS("is-constituent-snapshot", "false")

	// Define a callback function to process each batch of records
	processBatch := func(batch []*node.Node) error {
		for _, sData := range batch {
			svm := sData.GetChildContentS("vserver")
			volume := sData.GetChildContentS("volume")
			snapshot := sData.GetChildContentS("name")
			sizeKBStr := sData.GetChildContentS("total")

			// Convert size from KB to bytes (multiply by 1024)
			if sizeKB, err := strconv.ParseInt(sizeKBStr, 10, 64); err == nil {
				sizeBytes := sizeKB * 1024
				snapshotviolation.ProcessSnapshotData(svm, volume, snapshot, sizeBytes, prefixMap, filteredSnapshotStats)
			} else {
				// If conversion fails, log warning and use original value
				s.SLogger.Warn("Failed to convert snapshot size from KB to bytes",
					slog.String("svm", svm),
					slog.String("volume", volume),
					slog.String("snapshot", snapshot),
					slog.String("sizeKB", sizeKBStr),
					slogx.Err(err))
				continue
			}
		}
		return nil
	}

	err := s.client.InvokeZapiCallStream(request, processBatch)
	if err != nil {
		s.SLogger.Error("Failed to fetch data", slogx.Err(err))
		return filteredSnapshotStats
	}

	return filteredSnapshotStats
}
