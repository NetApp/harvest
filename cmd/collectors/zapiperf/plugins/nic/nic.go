/*
 * Copyright NetApp Inc, 2021 All rights reserved

Package Description:
    Some postprocessing on counter data "nic_common"
      Converts link_speed to numeric MBs
      Adds custom metrics:
          - "rc_percent":    receive data utilization percent
          - "tx_percent":    sent data utilization percent
          - "util_percent":  max utilization percent
		  - "nic_state":     0 if port is up, 1 otherwise

*/

package nic

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/template"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"math"
	"strconv"
	"strings"
)

const batchSize = "500"

type Nic struct {
	*plugin.AbstractPlugin
	data   *matrix.Matrix
	client *zapi.Client
}

var ifgrpMetrics = []string{
	"rx_bytes",
	"tx_bytes",
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Nic{AbstractPlugin: p}
}

func (n *Nic) Init(remote conf.Remote) error {
	var err error
	if err := n.InitAbc(); err != nil {
		return err
	}

	if n.client, err = zapi.New(conf.ZapiPoller(n.ParentParams), n.Auth); err != nil {
		n.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := n.client.Init(5, remote); err != nil {
		return err
	}

	n.data = matrix.New(n.Parent+".NicCommon", "nic_ifgrp", "nic_ifgrp")

	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "node")
	instanceKeys.NewChildS("", "ifgroup")
	instanceKeys.NewChildS("", "ports")
	n.data.SetExportOptions(exportOptions)

	for _, obj := range ifgrpMetrics {
		metricName, display, _, _ := template.ParseMetric(obj)
		_, err := n.data.NewMetricFloat64(metricName, display)
		if err != nil {
			n.SLogger.Error("add metric", slogx.Err(err))
			return err
		}
	}

	return nil
}

// Run speed label is reported in bits-per-second and rx/tx is reported as bytes-per-second
func (n *Nic) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {

	var read, write, rx, tx, utilPercent *matrix.Metric
	var err error
	portDataMap := make(map[string]collectors.PortData)

	data := dataMap[n.Object]
	n.client.Metadata.Reset()

	// Purge and reset data
	n.data.PurgeInstances()
	n.data.Reset()

	// Set all global labels from zapi.go if already not exist
	n.data.SetGlobalLabels(data.GetGlobalLabels())

	if read = data.GetMetric("rx_bytes"); read == nil {
		return nil, nil, errs.New(errs.ErrNoMetric, "rx_bytes")
	}

	if write = data.GetMetric("tx_bytes"); write == nil {
		return nil, nil, errs.New(errs.ErrNoMetric, "tx_bytes")
	}

	if rx = data.GetMetric("rx_percent"); rx == nil {
		if rx, err = data.NewMetricFloat64("rx_percent"); err == nil {
			rx.SetProperty("raw")
		} else {
			return nil, nil, err
		}

	}
	if tx = data.GetMetric("tx_percent"); tx == nil {
		if tx, err = data.NewMetricFloat64("tx_percent"); err == nil {
			tx.SetProperty("raw")
		} else {
			return nil, nil, err
		}
	}

	if utilPercent = data.GetMetric("util_percent"); utilPercent == nil {
		if utilPercent, err = data.NewMetricFloat64("util_percent"); err == nil {
			utilPercent.SetProperty("raw")
		} else {
			return nil, nil, err
		}
	}

	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}

		var speed, base int
		var s, port, nodeName string
		var err error

		s = instance.GetLabel("speed")
		port = instance.GetLabel("nic")
		nodeName = instance.GetLabel("node")

		if s != "" {
			if before, ok := strings.CutSuffix(s, "M"); ok {
				base, err = strconv.Atoi(before)
				if err != nil {
					n.SLogger.Warn("convert", slog.String("speed", s))
					n.SLogger.Warn("convert", slog.String("speed", s))
				} else {
					// NIC speed value converted from Mbps to Bps(bytes per second)
					speed = base * 125000
				}
			} else if speed, err = strconv.Atoi(s); err != nil {
				n.SLogger.Warn("convert", slog.String("speed", s))
			}

			if speed != 0 {

				var rxBytes, txBytes, rxPercent, txPercent float64
				var rxOk, txOk bool

				if rxBytes, rxOk = read.GetValueFloat64(instance); rxOk {
					rxPercent = rxBytes / float64(speed)
					rx.SetValueFloat64(instance, rxPercent)
				}

				if txBytes, txOk = write.GetValueFloat64(instance); txOk {
					txPercent = txBytes / float64(speed)
					tx.SetValueFloat64(instance, txPercent)
				}

				portDataMap[nodeName+port] = collectors.PortData{Node: nodeName, Port: port, Read: rxBytes, Write: txBytes}

				if rxOk || txOk {
					utilPercent.SetValueFloat64(instance, math.Max(rxPercent, txPercent))
				}
			}
		}

		if s = instance.GetLabel("speed"); strings.HasSuffix(s, "M") {
			base, err = strconv.Atoi(strings.TrimSuffix(s, "M"))
			if err != nil {
				n.SLogger.Warn("convert", slog.String("speed", s))
			} else {
				// NIC speed value converted from Mbps to bps(bits per second)
				speed = base * 1_000_000
				instance.SetLabel("speed", strconv.Itoa(speed))
			}
		}

		// truncate redundant prefix in nic type
		if t := instance.GetLabel("type"); strings.HasPrefix(t, "nic_") {
			instance.SetLabel("type", strings.TrimPrefix(t, "nic_"))
		}

	}

	// populate ifgrp metrics
	portIfgroupMap := n.getIfgroupInfo()
	if err := collectors.PopulateIfgroupMetrics(portIfgroupMap, portDataMap, n.data, n.SLogger); err != nil {
		return nil, nil, err
	}

	return []*matrix.Matrix{n.data}, n.client.Metadata, nil
}

func (n *Nic) getIfgroupInfo() map[string]string {
	var (
		result         *node.Node
		ifgroups       []*node.Node
		ifgroupsData   []*node.Node
		portIfgroupMap map[string]string // Node+port to ifgroup mapping map
	)

	portIfgroupMap = make(map[string]string)
	query := "net-port-get-iter"
	tag := "initial"
	request := node.NewXMLS(query)
	request.NewChildS("max-records", batchSize)
	desired := node.NewXMLS("desired-attributes")
	ifgroupAttributes := node.NewXMLS("desired-attributes")
	ifgroupInfoAttributes := node.NewXMLS("net-port-info")
	ifgroupInfoAttributes.NewChildS("node", "")
	ifgroupInfoAttributes.NewChildS("port", "")
	ifgroupInfoAttributes.NewChildS("ifgrp-port", "")
	ifgroupAttributes.AddChild(ifgroupInfoAttributes)
	desired.AddChild(ifgroupAttributes)
	request.AddChild(desired)

	for {
		responseData, err := n.client.InvokeBatchRequest(request, tag, "")
		if err != nil {
			n.SLogger.Error("Failed to invoke batch zapi call", slogx.Err(err))
			return portIfgroupMap
		}
		result = responseData.Result
		tag = responseData.Tag

		if result == nil {
			break
		}

		if x := result.GetChildS("attributes-list"); x != nil {
			ifgroups = x.GetChildren()
		}
		if len(ifgroups) == 0 {
			break
		}
		ifgroupsData = append(ifgroupsData, ifgroups...)
	}
	for _, ifgroup := range ifgroupsData {
		if ifgrpPort := ifgroup.GetChildContentS("ifgrp-port"); ifgrpPort != "" {
			nodeName := ifgroup.GetChildContentS("node")
			port := ifgroup.GetChildContentS("port")
			portIfgroupMap[nodeName+port] = ifgrpPort
		}
	}
	return portIfgroupMap
}
