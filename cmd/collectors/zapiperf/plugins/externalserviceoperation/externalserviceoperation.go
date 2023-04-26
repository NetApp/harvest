/*
 * Copyright NetApp Inc, 2023 All rights reserved
 */
package externalserviceoperation

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

const Hyphen = "-"

type ExternalServiceOperation struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &ExternalServiceOperation{AbstractPlugin: p}
}

func (e *ExternalServiceOperation) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	data := dataMap[e.Object]
	datacenterClusterKey := data.GetGlobalLabels().Get("datacenter") + Hyphen + data.GetGlobalLabels().Get("cluster") + Hyphen
	for _, instance := range data.GetInstances() {
		// generate unique key by appending datacenter, cluster, svm, service_name and operation to support topk in grafana dashboard
		key := datacenterClusterKey + instance.GetLabel("svm") + Hyphen + instance.GetLabel("service_name") + Hyphen + instance.GetLabel("operation")
		instance.SetLabel("key", key)
	}
	return nil, nil
}
