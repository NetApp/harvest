/*
 * Copyright NetApp Inc, 2023 All rights reserved
 */
package externalserviceop

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type ExternalServiceOp struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &ExternalServiceOp{AbstractPlugin: p}
}

func (e *ExternalServiceOp) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	data := dataMap[e.Object]
	keyIndex := 0
	datacenterClusterKey := data.GetGlobalLabels().Get("datacenter") + "-" + data.GetGlobalLabels().Get("cluster") + "-"
	for _, instance := range data.GetInstances() {
		// generate unique key by appending datacenter, cluster, svm, service_name and operation to support topk in grafana dashboard
		key := datacenterClusterKey + instance.GetLabel("svm") + "-" + instance.GetLabel("service_name") + "-" + instance.GetLabel("operation")
		instance.SetLabel("key", key)
		keyIndex++
	}
	return nil, nil
}
