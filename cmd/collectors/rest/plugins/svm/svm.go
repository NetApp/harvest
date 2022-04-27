/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package svm

import (
	"github.com/tidwall/gjson"
	"goharvest2/cmd/collectors"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/matrix"
	"strings"
)

type SVM struct {
	*plugin.AbstractPlugin
	nsswitchInfo map[string]nsswitch
}

type nsswitch struct {
	nsdb     []string
	nssource []string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SVM{AbstractPlugin: p}
}

func (my *SVM) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		err error
	)

	// invoke nameservice-nsswitch-get-iter zapi and get nsswitch info
	if my.nsswitchInfo, err = my.GetNSSwitchInfo(data); err != nil {
		my.Logger.Warn().Stack().Err(err).Msg("Failed to collect nsswitch info")
		//return nil, nil
	}

	// update svm instance based on the above zapi response
	for _, svmInstance := range data.GetInstances() {
		svmName := svmInstance.GetLabel("svm")

		// Update nameservice_switch and nis_domain label in svm
		if nsswitchInfo, ok := my.nsswitchInfo[svmName]; ok {
			ns_db := strings.Join(nsswitchInfo.nsdb, ",")
			ns_source := strings.Join(nsswitchInfo.nssource, ",")
			nis_domain := svmInstance.GetLabel("nis_domain")
			svmInstance.SetLabel("ns_source", ns_source)
			svmInstance.SetLabel("ns_db", ns_db)
			collectors.SetNameservice(ns_db, ns_source, nis_domain, svmInstance)
		}
	}
	return nil, nil
}

func (my *SVM) GetNSSwitchInfo(data *matrix.Matrix) (map[string]nsswitch, error) {

	var (
		vserverNsswitchMap map[string]nsswitch
		ns                 nsswitch
		ok                 bool
	)

	vserverNsswitchMap = make(map[string]nsswitch)

	for _, svmInstance := range data.GetInstances() {
		svmName := svmInstance.GetLabel("svm")
		nsswitchConfig := svmInstance.GetLabel("nameservice_switch")

		config := gjson.Result{Type: gjson.JSON, Raw: nsswitchConfig}
		replaceStr := strings.NewReplacer("[", "", "]", "", "\"", "")

		for nsdb, nssource := range config.Map() {
			nssourcelist := replaceStr.Replace(nssource.String())

			if ns, ok = vserverNsswitchMap[svmName]; ok {
				ns.nsdb = append(ns.nsdb, nsdb)
				ns.nssource = append(ns.nssource, nssourcelist)
			} else {
				ns = nsswitch{nsdb: []string{nsdb}, nssource: []string{nssourcelist}}
			}
			vserverNsswitchMap[svmName] = ns
		}
	}
	return vserverNsswitchMap, nil
}
