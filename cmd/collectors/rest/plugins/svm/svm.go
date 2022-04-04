/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package svm

import (
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/matrix"
	"strings"
)

type SVM struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SVM{AbstractPlugin: p}
}

func (my *SVM) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	my.setNameservice(data)
	return nil, nil
}

func (my *SVM) setNameservice(data *matrix.Matrix) {
	for _, instance := range data.GetInstances() {
		requiredNSDb := false
		requiredNSSource := false
		ns := instance.GetLabel("nameservice_switch")
		nisDomain := instance.GetLabel("nis_domain")

		if strings.Contains(ns, "passwd") || strings.Contains(ns, "group") || strings.Contains(ns, "netgroup") {
			requiredNSDb = true
			if strings.Contains(ns, "nis") {
				requiredNSSource = true
			}
		}

		if nisDomain != "" && requiredNSDb && requiredNSSource {
			instance.SetLabel("nis_authentication_enabled", "true")
		} else {
			instance.SetLabel("nis_authentication_enabled", "false")
		}

	}
}
