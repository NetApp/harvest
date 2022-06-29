/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package svm

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/tidwall/gjson"
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
		if errs.IsApiNotFound(err) {
			my.Logger.Debug().Err(err).Msg("Failed to collect nsswitch info")
		} else {
			my.Logger.Warn().Err(err).Msg("Failed to collect nsswitch info")
		}
	}

	// update svm instance based on the above zapi response
	for _, svmInstance := range data.GetInstances() {
		svmName := svmInstance.GetLabel("svm")

		// Update nameservice_switch and nis_domain label in svm
		if nsswitchInfo, ok := my.nsswitchInfo[svmName]; ok {
			nsDb := strings.Join(nsswitchInfo.nsdb, ",")
			nsSource := strings.Join(nsswitchInfo.nssource, ",")
			nisDomain := svmInstance.GetLabel("nis_domain")
			svmInstance.SetLabel("ns_source", nsSource)
			svmInstance.SetLabel("ns_db", nsDb)
			collectors.SetNameservice(nsDb, nsSource, nisDomain, svmInstance)
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
