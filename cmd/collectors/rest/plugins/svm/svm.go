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
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var weakCiphers = regexp.MustCompile("(.*)_cbc.*")
var replaceStr = strings.NewReplacer("[", "", "]", "", "\"", "", "\n", "", " ", "")

type SVM struct {
	*plugin.AbstractPlugin
	nsswitchInfo        map[string]Nsswitch
	kerberosInfo        map[string]string
	fpolicyInfo         map[string]Fpolicy
	iscsiServiceInfo    map[string]string
	iscsiCredentialInfo map[string]string
	client              *rest.Client
}

type Nsswitch struct {
	nsdb     []string
	nssource []string
}

type Fpolicy struct {
	name   string
	enable string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SVM{AbstractPlugin: p}
}

func (s *SVM) Init(remote conf.Remote) error {

	var err error

	if err := s.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if s.client, err = rest.New(conf.ZapiPoller(s.ParentParams), timeout, s.Auth); err != nil {
		s.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := s.client.Init(5, remote); err != nil {
		return err
	}
	s.nsswitchInfo = make(map[string]Nsswitch)
	s.kerberosInfo = make(map[string]string)
	s.fpolicyInfo = make(map[string]Fpolicy)
	s.iscsiServiceInfo = make(map[string]string)
	s.iscsiCredentialInfo = make(map[string]string)

	return nil
}

func (s *SVM) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	var (
		err error
	)

	data := dataMap[s.Object]
	s.client.Metadata.Reset()

	// update nsswitch info
	if s.nsswitchInfo, err = s.GetNSSwitchInfo(data); err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			s.SLogger.Debug("Failed to collect nsswitch info", slogx.Err(err))
		} else {
			s.SLogger.Warn("Failed to collect nsswitch info", slogx.Err(err))
		}
	}

	// invoke api/protocols/nfs/kerberos/interfaces rest and get nfs_kerberos_protocol_enabled
	if s.kerberosInfo, err = s.GetKerberosConfig(); err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			s.SLogger.Debug("Failed to collect kerberos config", slogx.Err(err))
		} else {
			s.SLogger.Error("Failed to collect kerberos config", slogx.Err(err))
		}
	}

	// invoke api/protocols/fpolicy rest and get fpolicy_enabled, fpolicy_name
	if s.fpolicyInfo, err = s.GetFpolicy(); err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			s.SLogger.Debug("Failed to collect fpolicy info", slogx.Err(err))
		} else {
			s.SLogger.Error("Failed to collect fpolicy info", slogx.Err(err))
		}
	}

	// invoke api/protocols/san/iscsi/services rest and get iscsi_service_enabled
	if s.iscsiServiceInfo, err = s.GetIscsiServices(); err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			s.SLogger.Debug("Failed to collect iscsi service info", slogx.Err(err))
		} else {
			s.SLogger.Error("Failed to collect iscsi service info", slogx.Err(err))
		}
	}

	// invoke api/protocols/san/iscsi/credentials rest and get iscsi_authentication_type
	if s.iscsiCredentialInfo, err = s.GetIscsiCredentials(); err != nil {
		if errs.IsRestErr(err, errs.APINotFound) {
			s.SLogger.Debug("Failed to collect iscsi credential info", slogx.Err(err))
		} else {
			s.SLogger.Error("Failed to collect iscsi credential info", slogx.Err(err))
		}
	}

	s.updateSVM(data)

	return nil, s.client.Metadata, nil
}

func (s *SVM) updateSVM(data *matrix.Matrix) {
	// update svm instance based on the above zapi response
	for _, svmInstance := range data.GetInstances() {
		svmName := svmInstance.GetLabel("svm")
		svmState := svmInstance.GetLabel("state")

		// SVM names ending with "-mc" are MetroCluster SVMs. We should only export svm metrics from these SVMs if the svm is online.
		if svmState == "offline" && strings.HasSuffix(svmName, "-mc") {
			svmInstance.SetExportable(false)
			continue
		}

		// Update nameservice_switch and nis_domain label in svm
		if nsswitchInfo, ok := s.nsswitchInfo[svmName]; ok {
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
		if kerberosEnabled, ok := s.kerberosInfo[svmName]; ok {
			svmInstance.SetLabel("nfs_kerberos_protocol_enabled", kerberosEnabled)
		}

		// Update fpolicy_enabled, fpolicy_name label in svm
		if fpolicyInfo, ok := s.fpolicyInfo[svmName]; ok {
			svmInstance.SetLabel("fpolicy_enabled", fpolicyInfo.enable)
			svmInstance.SetLabel("fpolicy_name", fpolicyInfo.name)
		}

		// Update iscsi_service_enabled label in svm
		if iscsiServiceEnabled, ok := s.iscsiServiceInfo[svmName]; ok {
			svmInstance.SetLabel("iscsi_service_enabled", iscsiServiceEnabled)
		}

		// Update iscsi_authentication_type label in svm
		if iscsiAuthenticationType, ok := s.iscsiCredentialInfo[svmName]; ok {
			svmInstance.SetLabel("iscsi_authentication_type", iscsiAuthenticationType)
		}

		if ciphersVal := svmInstance.GetLabel("ciphers"); ciphersVal != "" {
			insecured := weakCiphers.MatchString(ciphersVal)
			svmInstance.SetLabel("insecured", strconv.FormatBool(insecured))
		}
	}
}

func (s *SVM) GetNSSwitchInfo(data *matrix.Matrix) (map[string]Nsswitch, error) {

	var (
		vserverNsswitchMap map[string]Nsswitch
		ok                 bool
	)

	vserverNsswitchMap = make(map[string]Nsswitch)

	for _, svmInstance := range data.GetInstances() {
		var ns Nsswitch
		svmName := svmInstance.GetLabel("svm")
		if nsswitchConfig := svmInstance.GetLabel("nameservice_switch"); nsswitchConfig != "" {
			config := gjson.Result{Type: gjson.JSON, Raw: nsswitchConfig}
			for nsdb, nssource := range config.Map() {
				if nssource.Exists() {
					nssourcelist := strings.Split(replaceStr.Replace(nssource.ClonedString()), ",")
					if ns, ok = vserverNsswitchMap[svmName]; ok {
						ns.nsdb = append(ns.nsdb, nsdb)
						ns.nssource = append(ns.nssource, nssourcelist...)
					} else {
						ns = Nsswitch{nsdb: []string{nsdb}, nssource: nssourcelist}
					}
				}
				vserverNsswitchMap[svmName] = ns
			}
		}
	}
	return vserverNsswitchMap, nil
}

func (s *SVM) GetKerberosConfig() (map[string]string, error) {
	var (
		result         []gjson.Result
		svmKerberosMap map[string]string
		err            error
	)

	svmKerberosMap = make(map[string]string)
	query := "api/protocols/nfs/kerberos/interfaces"
	fields := []string{"svm.name", "enabled"}
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(collectors.DefaultBatchSize).
		Build()

	if result, err = collectors.InvokeRestCall(s.client, href); err != nil {
		return nil, err
	}

	for _, kerberosConfig := range result {
		enable := kerberosConfig.Get("enabled").ClonedString()
		svmName := kerberosConfig.Get("svm.name").ClonedString()
		if _, ok := svmKerberosMap[svmName]; !ok {
			svmKerberosMap[svmName] = enable
		} else if enable == "true" {
			// If any interface on the svm has kerberos on, then only set to true
			svmKerberosMap[svmName] = enable
		}
	}

	return svmKerberosMap, nil
}

func (s *SVM) GetFpolicy() (map[string]Fpolicy, error) {
	var (
		result        []gjson.Result
		svmFpolicyMap map[string]Fpolicy
		err           error
	)

	svmFpolicyMap = make(map[string]Fpolicy)
	query := "api/protocols/fpolicy"
	fields := []string{"svm.name", "policies.enabled", "policies.name"}
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(collectors.DefaultBatchSize).
		Build()

	if result, err = collectors.InvokeRestCall(s.client, href); err != nil {
		return nil, err
	}

	for _, fpolicyData := range result {
		fpolicyEnable := fpolicyData.Get("policies.enabled").ClonedString()
		fpolicyName := fpolicyData.Get("policies.name").ClonedString()
		svmName := fpolicyData.Get("svm.name").ClonedString()
		if _, ok := svmFpolicyMap[svmName]; !ok {
			svmFpolicyMap[svmName] = Fpolicy{name: fpolicyName, enable: fpolicyEnable}
		} else if svmFpolicyMap[svmName].enable == "false" {
			// If svm is already present, update the status value only if it is false
			svmFpolicyMap[svmName] = Fpolicy{name: fpolicyName, enable: fpolicyEnable}
		}
	}

	return svmFpolicyMap, nil
}

func (s *SVM) GetIscsiServices() (map[string]string, error) {
	var (
		result             []gjson.Result
		svmIscsiServiceMap map[string]string
		err                error
	)

	svmIscsiServiceMap = make(map[string]string)
	query := "api/protocols/san/iscsi/services"
	fields := []string{"svm.name", "enabled"}
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(collectors.DefaultBatchSize).
		Build()

	if result, err = collectors.InvokeRestCall(s.client, href); err != nil {
		return nil, err
	}

	for _, iscsiData := range result {
		iscsiServiceEnable := iscsiData.Get("enabled").ClonedString()
		svmName := iscsiData.Get("svm.name").ClonedString()
		if _, ok := svmIscsiServiceMap[svmName]; !ok {
			svmIscsiServiceMap[svmName] = iscsiServiceEnable
		} else if svmIscsiServiceMap[svmName] == "false" {
			// If svm is already present, update the map value only if previous value is false
			svmIscsiServiceMap[svmName] = iscsiServiceEnable
		}
	}

	return svmIscsiServiceMap, nil
}

func (s *SVM) GetIscsiCredentials() (map[string]string, error) {
	var (
		result                []gjson.Result
		svmIscsiCredentialMap map[string]string
		err                   error
	)

	svmIscsiCredentialMap = make(map[string]string)
	query := "api/protocols/san/iscsi/credentials"
	fields := []string{"svm.name", "authentication_type"}
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(collectors.DefaultBatchSize).
		Build()

	if result, err = collectors.InvokeRestCall(s.client, href); err != nil {
		return nil, err
	}

	for _, iscsiData := range result {
		authenticationType := iscsiData.Get("authentication_type").ClonedString()
		svmName := iscsiData.Get("svm.name").ClonedString()
		if _, ok := svmIscsiCredentialMap[svmName]; !ok {
			svmIscsiCredentialMap[svmName] = authenticationType
		} else {
			// If svm is already present, update the map value with append this authenticationType to previous value
			svmIscsiCredentialMap[svmName] = svmIscsiCredentialMap[svmName] + "," + authenticationType
		}
	}

	return svmIscsiCredentialMap, nil
}
