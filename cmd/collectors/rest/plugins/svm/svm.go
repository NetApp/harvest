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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var weakCiphers = regexp.MustCompile("(.*)_cbc.*")

type SVM struct {
	*plugin.AbstractPlugin
	nsswitchInfo        map[string]nsswitch
	kerberosInfo        map[string]string
	fpolicyInfo         map[string]fpolicy
	iscsiServiceInfo    map[string]string
	iscsiCredentialInfo map[string]string
	client              *rest.Client
}

type nsswitch struct {
	nsdb     []string
	nssource []string
}

type fpolicy struct {
	name   string
	enable string
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
	my.kerberosInfo = make(map[string]string)
	my.fpolicyInfo = make(map[string]fpolicy)
	my.iscsiServiceInfo = make(map[string]string)
	my.iscsiCredentialInfo = make(map[string]string)

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
	if my.kerberosInfo, err = my.GetKerberosConfig(); err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			my.Logger.Debug().Err(err).Msg("Failed to collect kerberos config")
		} else {
			my.Logger.Error().Err(err).Msg("Failed to collect kerberos config")
		}
	}

	// invoke api/protocols/fpolicy rest and get fpolicy_enabled, fpolicy_name
	if my.fpolicyInfo, err = my.GetFpolicy(); err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			my.Logger.Debug().Err(err).Msg("Failed to collect fpolicy info")
		} else {
			my.Logger.Error().Err(err).Msg("Failed to collect fpolicy info")
		}
	}

	// invoke api/protocols/san/iscsi/services rest and get iscsi_service_enabled
	if my.iscsiServiceInfo, err = my.GetIscsiServices(); err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			my.Logger.Debug().Err(err).Msg("Failed to collect fpolicy info")
		} else {
			my.Logger.Error().Err(err).Msg("Failed to collect fpolicy info")
		}
	}

	// invoke api/protocols/san/iscsi/credentials rest and get iscsi_authentication_type
	if my.iscsiCredentialInfo, err = my.GetIscsiCredentials(); err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			my.Logger.Debug().Err(err).Msg("Failed to collect fpolicy info")
		} else {
			my.Logger.Error().Err(err).Msg("Failed to collect fpolicy info")
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
		if kerberosEnabled, ok := my.kerberosInfo[svmName]; ok {
			svmInstance.SetLabel("nfs_kerberos_protocol_enabled", kerberosEnabled)
		}

		// Update fpolicy_enabled, fpolicy_name label in svm
		if fpolicyInfo, ok := my.fpolicyInfo[svmName]; ok {
			svmInstance.SetLabel("fpolicy_enabled", fpolicyInfo.enable)
			svmInstance.SetLabel("fpolicy_name", fpolicyInfo.name)
		}

		// Update iscsi_service_enabled label in svm
		if iscsiServiceEnabled, ok := my.iscsiServiceInfo[svmName]; ok {
			svmInstance.SetLabel("iscsi_service_enabled", iscsiServiceEnabled)
		}

		// Update iscsi_authentication_type label in svm
		if iscsiAuthenticationType, ok := my.iscsiCredentialInfo[svmName]; ok {
			svmInstance.SetLabel("iscsi_authentication_type", iscsiAuthenticationType)
		}

		ciphersVal := svmInstance.GetLabel("ciphers")
		insecured := weakCiphers.MatchString(ciphersVal)
		svmInstance.SetLabel("insecured", strconv.FormatBool(insecured))
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

func (my *SVM) GetFpolicy() (map[string]fpolicy, error) {
	var (
		result        []gjson.Result
		svmFpolicyMap map[string]fpolicy
		err           error
	)

	svmFpolicyMap = make(map[string]fpolicy)
	query := "api/protocols/fpolicy"
	fpolicyFields := []string{"svm.name", "policies.enabled", "policies.name"}
	href := rest.BuildHref("", strings.Join(fpolicyFields, ","), nil, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(my.client, href, my.Logger); err != nil {
		return nil, err
	}

	for _, fpolicyData := range result {
		fpolicyEnable := fpolicyData.Get("policies.enabled").String()
		fpolicyName := fpolicyData.Get("policies.name").String()
		svmName := fpolicyData.Get("svm.name").String()
		if _, ok := svmFpolicyMap[svmName]; !ok {
			svmFpolicyMap[svmName] = fpolicy{name: fpolicyName, enable: fpolicyEnable}
		} else {
			// If svm is already present, update the status value only if it is false
			if svmFpolicyMap[svmName].enable == "false" {
				svmFpolicyMap[svmName] = fpolicy{name: fpolicyName, enable: fpolicyEnable}
			}
		}
	}

	return svmFpolicyMap, nil
}

func (my *SVM) GetIscsiServices() (map[string]string, error) {
	var (
		result             []gjson.Result
		svmIscsiServiceMap map[string]string
		err                error
	)

	svmIscsiServiceMap = make(map[string]string)
	query := "api/protocols/san/iscsi/services"
	iscsiServiceFields := []string{"svm.name", "enabled"}
	href := rest.BuildHref("", strings.Join(iscsiServiceFields, ","), nil, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(my.client, href, my.Logger); err != nil {
		return nil, err
	}

	for _, iscsiData := range result {
		iscsiServiceEnable := iscsiData.Get("enabled").String()
		svmName := iscsiData.Get("svm.name").String()
		if _, ok := svmIscsiServiceMap[svmName]; !ok {
			svmIscsiServiceMap[svmName] = iscsiServiceEnable
		} else {
			// If svm is already present, update the map value only if previous value is false
			if svmIscsiServiceMap[svmName] == "false" {
				svmIscsiServiceMap[svmName] = iscsiServiceEnable
			}
		}
	}

	return svmIscsiServiceMap, nil
}

func (my *SVM) GetIscsiCredentials() (map[string]string, error) {
	var (
		result                []gjson.Result
		svmIscsiCredentialMap map[string]string
		err                   error
	)

	svmIscsiCredentialMap = make(map[string]string)
	query := "api/protocols/san/iscsi/credentials"
	iscsiCredentialFields := []string{"svm.name", "authentication_type"}
	href := rest.BuildHref("", strings.Join(iscsiCredentialFields, ","), nil, "", "", "", "", query)

	if result, err = collectors.InvokeRestCall(my.client, href, my.Logger); err != nil {
		return nil, err
	}

	for _, iscsiData := range result {
		authenticationType := iscsiData.Get("authentication_type").String()
		svmName := iscsiData.Get("svm.name").String()
		if _, ok := svmIscsiCredentialMap[svmName]; !ok {
			svmIscsiCredentialMap[svmName] = authenticationType
		} else {
			// If svm is already present, update the map value with append this authenticationType to previous value
			svmIscsiCredentialMap[svmName] = svmIscsiCredentialMap[svmName] + "," + authenticationType
		}
	}

	return svmIscsiCredentialMap, nil
}
