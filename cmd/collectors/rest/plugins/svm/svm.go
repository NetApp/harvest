/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package svm

import (
	"github.com/hashicorp/go-version"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/matrix"
	"strconv"
	"strings"
)

const Version_9_10 = "9.10.0"

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
	isHigher, _ := my.isHigherThan9_10(data.GetGlobalLabels().Get("clusterversion"))

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

		instance.SetLabel("isHigherThan9_10", strconv.FormatBool(isHigher))
	}

}

func (my *SVM) isHigherThan9_10(current string) (bool, error) {
	my.Logger.Info().Str("Version", current).Msg("cluster")

	versionToCompare, err := version.NewVersion(Version_9_10)
	if err != nil {
		return false, err
	}
	currentVersion, err := version.NewVersion(current)
	if err != nil {
		return false, err
	}
	if currentVersion.GreaterThan(versionToCompare) {
		return true, nil
	} else {
		return false, nil
	}
}
