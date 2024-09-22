/*
 * Copyright NetApp Inc, 2023 All rights reserved
 */

package netroute

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"log/slog"
	"strconv"
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
	"uuid",
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &NetRoute{AbstractPlugin: p}
}

func (n *NetRoute) Init() error {

	if err := n.InitAbc(); err != nil {
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
		iKeys.NewChildS("", "route_uuid")
	}
	n.data.SetExportOptions(exportOptions)
	return nil
}

func (n *NetRoute) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

	data := dataMap[n.Object]
	// Purge and reset data
	n.data.PurgeInstances()
	n.data.Reset()

	n.data.SetGlobalLabels(data.GetGlobalLabels())

	count := 0
	for key, instance := range data.GetInstances() {
		cluster := data.GetGlobalLabels()["cluster"]
		routeID := instance.GetLabel("uuid")
		interfaces := instance.GetLabel("interfaces")

		interfacesList := gjson.Result{Type: gjson.JSON, Raw: interfaces}
		names := interfacesList.Get("name").Array()
		address := interfacesList.Get("address").Array()

		if len(names) == len(address) {
			for i, name := range names {
				index := cluster + "_" + strconv.Itoa(count)
				interfaceInstance, err := n.data.NewInstance(index)
				if err != nil {
					n.SLogger.Error("add instance failed", slog.Any("err", err), slog.String("key", key))
					return nil, nil, err
				}

				for _, l := range instanceLabels {
					interfaceInstance.SetLabel(l, instance.GetLabel(l))
				}
				interfaceInstance.SetLabel("index", index)
				interfaceInstance.SetLabel("address", address[i].String())
				interfaceInstance.SetLabel("name", name.String())
				interfaceInstance.SetLabel("route_uuid", routeID)
				count++
			}
		}
	}

	return []*matrix.Matrix{n.data}, nil, nil
}
