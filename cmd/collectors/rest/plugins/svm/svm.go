/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package svm

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/tidwall/gjson"
	"sort"
	"strings"
	"time"
)

type SVM struct {
	*plugin.AbstractPlugin
	nsswitchInfo   map[string]nsswitch
	kerberosConfig map[string]string
	client         *rest.Client
}

type nsswitch struct {
	nsdb     []string
	nssource []string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SVM{AbstractPlugin: p}
}

func (my *SVM) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if my.client, err = rest.New(conf.ZapiPoller(my.ParentParams), timeout, my.Auth); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}
	my.nsswitchInfo = make(map[string]nsswitch)
	my.kerberosConfig = make(map[string]string)

	return nil
}

func (my *SVM) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		err error
	)

	data := dataMap[my.Object]
	// update nsswitch info
	if my.nsswitchInfo, err = my.GetNSSwitchInfo(data); err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			my.Logger.Debug().Err(err).Msg("Failed to collect nsswitch info")
		} else {
			my.Logger.Warn().Err(err).Msg("Failed to collect nsswitch info")
		}
	}

	// invoke api/protocols/nfs/kerberos/interfaces rest and get nfs_kerberos_protocol_enabled
	if my.kerberosConfig, err = my.GetKerberosConfig(); err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			my.Logger.Debug().Err(err).Msg("Failed to collect kerberos config")
		} else {
			my.Logger.Error().Err(err).Msg("Failed to collect kerberos config")
		}
	}

	// update svm instance based on the above zapi response
	for _, svmInstance := range data.GetInstances() {
		svmName := svmInstance.GetLabel("svm")

		// Update nameservice_switch and nis_domain label in svm
		if nsswitchInfo, ok := my.nsswitchInfo[svmName]; ok {
			sort.Strings(nsswitchInfo.nsdb)
			sort.Strings(nsswitchInfo.nssource)
			nsDB := strings.Join(nsswitchInfo.nsdb, ",")
			nsSource := strings.Join(nsswitchInfo.nssource, ",")
			nisDomain := svmInstance.GetLabel("nis_domain")
			svmInstance.SetLabel("ns_source", nsSource)
			svmInstance.SetLabel("ns_db", nsDB)
			collectors.SetNameservice(nsDB, nsSource, nisDomain, svmInstance)
		}

		// Update nfs_kerberos_protocol_enabled label in svm
		if kerberosEnabled, ok := my.kerberosConfig[svmName]; ok {
			svmInstance.SetLabel("nfs_kerberos_protocol_enabled", kerberosEnabled)
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
		replaceStr := strings.NewReplacer("[", "", "]", "", "\"", "", "\n", "", " ", "")

		for nsdb, nssource := range config.Map() {
			nssourcelist := strings.Split(replaceStr.Replace(nssource.String()), ",")

			if ns, ok = vserverNsswitchMap[svmName]; ok {
				ns.nsdb = append(ns.nsdb, nsdb)
				ns.nssource = append(ns.nssource, nssourcelist...)
			} else {
				ns = nsswitch{nsdb: []string{nsdb}, nssource: nssourcelist}
			}
			vserverNsswitchMap[svmName] = ns
		}
	}
	return vserverNsswitchMap, nil
}

func (my *SVM) GetKerberosConfig() (map[string]string, error) {
	var (
		result         []gjson.Result
		svmKerberosMap map[string]string
		err            error
	)

	svmKerberosMap = make(map[string]string)
	query := "api/protocols/nfs/kerberos/interfaces"
	kerberosFields := []string{"svm.name", "enabled"}
	href := rest.BuildHref("", strings.Join(kerberosFields, ","), nil, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(my.client, href, my.Logger); err != nil {
		return nil, err
	}

	for _, kerberosConfig := range result {
		enable := kerberosConfig.Get("enabled").String()
		svmName := kerberosConfig.Get("svm.name").String()
		if _, ok := svmKerberosMap[svmName]; !ok {
			svmKerberosMap[svmName] = enable
		} else {
			// If any interface on the svm has kerberos on, then only set to true
			if enable == "true" {
				svmKerberosMap[svmName] = enable
			}
		}
	}

	return svmKerberosMap, nil
}
