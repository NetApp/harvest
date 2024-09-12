/*
 *
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
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"math"
	"strconv"
	"strings"
	"time"
)

type Nic struct {
	*plugin.AbstractPlugin
	data         *matrix.Matrix
	client       *rest.Client
	testFilePath string // Used only from unit test
}

var ifgrpMetrics = []string{
	"rx_bytes",
	"tx_bytes",
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Nic{AbstractPlugin: p}
}

func (n *Nic) Init() error {
	err := n.InitAbc()
	if err != nil {
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
		metricName, display, _, _ := util.ParseMetric(obj)
		_, err := n.data.NewMetricFloat64(metricName, display)
		if err != nil {
			n.Logger.Error().Err(err).Msg("add metric")
			return err
		}
	}

	if n.Options.IsTest {
		return nil
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if n.client, err = rest.New(conf.ZapiPoller(n.ParentParams), timeout, n.Auth); err != nil {
		n.Logger.Error().Err(err).Msg("connecting")
		return err
	}

	if err := n.client.Init(5); err != nil {
		return err
	}

	return nil
}

// Run speed label is reported in bits-per-second and rx/tx is reported as bytes-per-second
func (n *Nic) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

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

	if read = data.GetMetric("receive_bytes"); read == nil {
		return nil, nil, errs.New(errs.ErrNoMetric, "receive_bytes")
	}

	if write = data.GetMetric("transmit_bytes"); write == nil {
		return nil, nil, errs.New(errs.ErrNoMetric, "transmit_bytes")
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

		var speed, base int
		var s, port, nodeName string
		var err error
		s = instance.GetLabel("speed")
		nodeName = instance.GetLabel("node")

		// example name = cluster_name:e0a
		// nic          = e0a
		if i := instance.GetLabel("id"); i != "" {
			if split := strings.Split(instance.GetLabel("id"), ":"); len(split) >= 2 {
				instance.SetLabel("nic", split[1])
				port = instance.GetLabel("nic")
			}
		}

		if s != "" {
			if strings.HasSuffix(s, "M") {
				base, err = strconv.Atoi(strings.TrimSuffix(s, "M"))
				if err != nil {
					n.Logger.Warn().Msgf("convert speed [%s]", s)
				} else {
					// NIC speed value converted from Mbps to Bps(bytes per second)
					speed = base * 125000
				}
			} else if speed, err = strconv.Atoi(s); err != nil {
				n.Logger.Warn().Msgf("convert speed [%s]", s)
			}

			if speed != 0 {

				var rxBytes, txBytes, rxPercent, txPercent float64
				var rxOk, txOk bool

				if rxBytes, rxOk = read.GetValueFloat64(instance); rxOk {
					rxPercent = rxBytes / float64(speed)
					err := rx.SetValueFloat64(instance, rxPercent)
					if err != nil {
						n.Logger.Error().Err(err).Msg("error")
					}
				}

				if txBytes, txOk = write.GetValueFloat64(instance); txOk {
					txPercent = txBytes / float64(speed)
					err := tx.SetValueFloat64(instance, txPercent)
					if err != nil {
						n.Logger.Error().Err(err).Msg("error")
					}
				}

				portDataMap[nodeName+port] = collectors.PortData{Node: nodeName, Port: port, Read: rxBytes, Write: txBytes}

				if rxOk || txOk {
					err := utilPercent.SetValueFloat64(instance, math.Max(rxPercent, txPercent))
					if err != nil {
						n.Logger.Error().Err(err).Msg("error")
					}
				}
			}
		}

		if s = instance.GetLabel("speed"); strings.HasSuffix(s, "M") {
			base, err = strconv.Atoi(strings.TrimSuffix(s, "M"))
			if err != nil {
				n.Logger.Warn().Msgf("convert speed [%s]", s)
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
	if err := collectors.PopulateIfgroupMetrics(portIfgroupMap, portDataMap, n.data, n.Logger); err != nil {
		return nil, nil, err
	}

	return []*matrix.Matrix{n.data}, n.client.Metadata, nil
}

func (n *Nic) getIfgroupInfo() map[string]string {
	var (
		err            error
		ifgroupsData   []gjson.Result
		portIfgroupMap map[string]string // Node+port to ifgroup mapping map
	)

	portIfgroupMap = make(map[string]string)
	query := "api/private/cli/network/port/ifgrp"
	fields := []string{"ifgrp", "node", "ports"}
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Build()

	if ifgroupsData, err = collectors.InvokeRestCallWithTestFile(n.client, href, n.Logger, n.testFilePath); err != nil {
		return portIfgroupMap
	}

	for _, ifgroup := range ifgroupsData {
		nodeName := ifgroup.Get("node").String()
		ifgrp := ifgroup.Get("ifgrp").String()
		ports := ifgroup.Get("ports").Array()
		for _, portName := range ports {
			portIfgroupMap[nodeName+portName.String()] = ifgrp
		}
	}
	return portIfgroupMap
}
