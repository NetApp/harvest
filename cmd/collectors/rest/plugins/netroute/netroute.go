/*
 * Copyright NetApp Inc, 2023 All rights reserved
 */

package netroute

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"strconv"
	"strings"
)

type NetRoute struct {
	*plugin.AbstractPlugin
	data *matrix.Matrix
}

var instanceLabels = []string{
	"name",
	"address",
	"svm",
	"scope",
	"gateway",
	"destination",
	"family",
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &NetRoute{AbstractPlugin: p}
}

func (n *NetRoute) Init() error {
	var err error

	if err = n.InitAbc(); err != nil {
		return err
	}

	n.data = matrix.New(n.Parent+".NetRouteInterface", "net_route_interface", "net_route_interface")

	exportOptions := node.NewS("export_options")
	iLabels := exportOptions.NewChildS("instance_labels", "")
	iKeys := exportOptions.NewChildS("instance_keys", "")

	if exportOption := n.ParentParams.GetChildS("export_options"); exportOption != nil {
		for _, label := range instanceLabels {
			iLabels.NewChildS("", label)
		}
		iKeys.NewChildS("", "index")
	}
	n.data.SetExportOptions(exportOptions)
	return nil
}

func (n *NetRoute) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {

	data := dataMap[n.Object]
	// Purge and reset data
	n.data.PurgeInstances()
	n.data.Reset()

	n.data.SetGlobalLabels(data.GetGlobalLabels())

	count := 0
	for key, instance := range data.GetInstances() {
		cluster := data.GetGlobalLabels().Get("cluster")
		interfaceName := instance.GetLabel("interface_name")
		interfaceAddress := instance.GetLabel("interface_address")
		if interfaceName != "" && interfaceAddress != "" {

			names := strings.Split(interfaceName, ",")
			address := strings.Split(interfaceAddress, ",")
			if len(names) == len(address) {
				for i, name := range names {
					index := strings.Join([]string{cluster, strconv.Itoa(count)}, "_")
					interfaceInstance, err := n.data.NewInstance(index)
					if err != nil {
						n.Logger.Error().Err(err).Str("add instance failed for instance key", key).Send()
						return nil, err
					}

					for _, l := range instanceLabels {
						interfaceInstance.SetLabel(l, instance.GetLabel(l))
					}
					interfaceInstance.SetLabel("index", index)
					interfaceInstance.SetLabel("address", address[i])
					interfaceInstance.SetLabel("name", name)
					count++
				}
			}
		}
	}

	return []*matrix.Matrix{n.data}, nil
}
