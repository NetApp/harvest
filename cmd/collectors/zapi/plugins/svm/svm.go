/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package svm

import (
	"errors"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const BatchSize = "500"

var weakCiphers = regexp.MustCompile("(.*)_cbc.*")

type SVM struct {
	*plugin.AbstractPlugin
	currentVal     int
	batchSize      string
	client         *zapi.Client
	auditProtocols map[string]string
	cifsProtocols  map[string]CifsSecurity
	nsswitchInfo   map[string]Nsswitch
	nisInfo        map[string]string
	cifsEnabled    map[string]bool
	nfsEnabled     map[string]string
	sshData        map[string]SSHInfo
	iscsiAuth      map[string]string
	iscsiService   map[string]string
	fpolicyData    map[string]Fpolicy
	ldapData       map[string]string
	kerberosConfig map[string]string
}

type Nsswitch struct {
	nsdb     []string
	nssource []string
}

type Fpolicy struct {
	name   string
	enable string
}

type CifsSecurity struct {
	cifsNtlmEnabled string
	smbEncryption   string
	smbSigning      string
}

type SSHInfo struct {
	ciphers    string
	isInsecure string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SVM{AbstractPlugin: p}
}

func (s *SVM) Init(remote conf.Remote) error {

	var err error

	if err := s.InitAbc(); err != nil {
		return err
	}

	if s.client, err = zapi.New(conf.ZapiPoller(s.ParentParams), s.Auth); err != nil {
		s.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := s.client.Init(5, remote); err != nil {
		return err
	}

	s.auditProtocols = make(map[string]string)
	s.cifsProtocols = make(map[string]CifsSecurity)
	s.nsswitchInfo = make(map[string]Nsswitch)
	s.nisInfo = make(map[string]string)
	s.cifsEnabled = make(map[string]bool)
	s.nfsEnabled = make(map[string]string)
	s.sshData = make(map[string]SSHInfo)
	s.iscsiAuth = make(map[string]string)
	s.iscsiService = make(map[string]string)
	s.fpolicyData = make(map[string]Fpolicy)
	s.ldapData = make(map[string]string)
	s.kerberosConfig = make(map[string]string)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	s.currentVal = s.SetPluginInterval()

	s.batchSize = BatchSize
	if b := s.Params.GetChildContentS("batch_size"); b != "" {
		if _, err := strconv.Atoi(b); err == nil {
			s.batchSize = b
			s.SLogger.Info("using batch-size", slog.String("BatchSize", s.batchSize))
		}
	}

	return nil
}

func (s *SVM) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	var (
		err error
	)

	data := dataMap[s.Object]
	s.client.Metadata.Reset()

	if s.currentVal >= s.PluginInvocationRate {
		s.currentVal = 0

		// invoke fileservice-audit-config-get-iter zapi and get audit protocols
		if s.auditProtocols, err = s.GetAuditProtocols(); err != nil {
			if !errors.Is(err, errs.ErrNoInstance) {
				s.SLogger.Error("Failed to collect audit protocols", slogx.Err(err))
			}
		}

		// invoke cifs-security-get-iter zapi and get cifs protocols
		if s.cifsProtocols, err = s.GetCifsProtocols(); err != nil {
			if !errors.Is(err, errs.ErrNoInstance) {
				s.SLogger.Error("Failed to collect cifs protocols", slogx.Err(err))
			}
		}

		// invoke nameservice-nsswitch-get-iter zapi and get nsswitch info
		if s.nsswitchInfo, err = s.GetNSSwitchInfo(); err != nil {
			if !errors.Is(err, errs.ErrNoInstance) {
				s.SLogger.Error("Failed to collect nsswitch info", slogx.Err(err))
			}
		}

		// invoke nis-get-iter zapi and get nisdomain info
		if s.nisInfo, err = s.GetNisInfo(); err != nil {
			if !errors.Is(err, errs.ErrNoInstance) {
				s.SLogger.Error("Failed to collect nisdomain info", slogx.Err(err))
			}
		}

		// invoke cifs-server-get-iter zapi and get cifsenabled info
		if s.cifsEnabled, err = s.GetCifsEnabled(); err != nil {
			if !errors.Is(err, errs.ErrNoInstance) {
				s.SLogger.Error("Failed to collect cifsenabled info", slogx.Err(err))
			}
		}

		// invoke nfs-service-get-iter zapi and get cifsenabled info
		if s.nfsEnabled, err = s.GetNfsEnabled(); err != nil {
			if !errors.Is(err, errs.ErrNoInstance) {
				s.SLogger.Error("Failed to collect nfsenabled info", slogx.Err(err))
			}
		}

		// invoke security-ssh-get-iter zapi and get ssh data
		if s.sshData, err = s.GetSSHData(); err != nil {
			if !errors.Is(err, errs.ErrNoInstance) {
				s.SLogger.Error("Failed to collect ssh data", slogx.Err(err))
			}
		}

		// invoke iscsi-initiator-auth-get-iter zapi and get iscsi_authentication_type
		if s.iscsiAuth, err = s.GetIscsiInitiatorAuth(); err != nil {
			if !errors.Is(err, errs.ErrNoInstance) {
				s.SLogger.Error("Failed to collect iscsi authentication type", slogx.Err(err))
			}
		}

		// invoke iscsi-service-get-iter zapi and get iscsi_service_enabled
		if s.iscsiService, err = s.GetIscsiService(); err != nil {
			if !errors.Is(err, errs.ErrNoInstance) {
				s.SLogger.Error("Failed to collect iscsi service", slogx.Err(err))
			}
		}

		// invoke fpolicy-policy-status-get-iter zapi and get fpolicy_enabled, fpolicy_name
		if s.fpolicyData, err = s.GetFpolicy(); err != nil {
			if !errors.Is(err, errs.ErrNoInstance) {
				s.SLogger.Error("Failed to collect fpolicy detail", slogx.Err(err))
			}
		}

		// invoke ldap-client-get-iter zapi and get ldap_session_security
		if s.ldapData, err = s.GetLdapData(); err != nil {
			if !errors.Is(err, errs.ErrNoInstance) {
				s.SLogger.Error("Failed to collect ldap session", slogx.Err(err))
			}
		}

		// invoke kerberos-config-get-iter zapi and get nfs_kerberos_protocol_enabled
		if s.kerberosConfig, err = s.GetKerberosConfig(); err != nil {
			if !errors.Is(err, errs.ErrNoInstance) {
				s.SLogger.Error("Failed to collect kerberos config", slogx.Err(err))
			}
		}
	}

	// update svm instance based on the above zapi response
	for _, svmInstance := range data.GetInstances() {
		if !svmInstance.IsExportable() {
			continue
		}
		svmName := svmInstance.GetLabel("svm")

		// Update audit_protocol_enabled label in svm
		svmInstance.SetLabel("audit_protocol_enabled", s.auditProtocols[svmName])

		// Update cifs_ntlm_enabled label in svm
		if cifsData, ok := s.cifsProtocols[svmName]; ok {
			svmInstance.SetLabel("cifs_ntlm_enabled", cifsData.cifsNtlmEnabled)
			svmInstance.SetLabel("smb_encryption_required", cifsData.smbEncryption)
			svmInstance.SetLabel("smb_signing_required", cifsData.smbSigning)
		}

		// Update nis_domain label in svm
		svmInstance.SetLabel("nis_domain", s.nisInfo[svmName])

		// Update nameservice_switch label in svm
		if nsswitchInfo, ok := s.nsswitchInfo[svmName]; ok {
			sort.Strings(nsswitchInfo.nsdb)
			sort.Strings(nsswitchInfo.nssource)
			nsDB := strings.Join(nsswitchInfo.nsdb, ",")
			nsSource := strings.Join(nsswitchInfo.nssource, ",")
			nisDomain := s.nisInfo[svmName]
			svmInstance.SetLabel("ns_source", nsSource)
			svmInstance.SetLabel("ns_db", nsDB)
			collectors.SetNameservice(nsDB, nsSource, nisDomain, svmInstance)
		}

		// Update cifs_protocol_enabled label in svm
		if cifsEnable, ok := s.cifsEnabled[svmName]; ok {
			svmInstance.SetLabel("cifs_protocol_enabled", strconv.FormatBool(cifsEnable))
		}

		// Update nfs_protocol_enabled label in svm
		if nfsEnable, ok := s.nfsEnabled[svmName]; ok {
			svmInstance.SetLabel("nfs_protocol_enabled", nfsEnable)
		}

		// Update ciphers label in svm
		if sshInfoDetail, ok := s.sshData[svmName]; ok {
			svmInstance.SetLabel("ciphers", sshInfoDetail.ciphers)
			svmInstance.SetLabel("insecured", sshInfoDetail.isInsecure)
		}

		// Update iscsi_authentication_type label in svm
		if authType, ok := s.iscsiAuth[svmName]; ok {
			svmInstance.SetLabel("iscsi_authentication_type", authType)
		}

		// Update iscsi_service_enabled label in svm
		if available, ok := s.iscsiService[svmName]; ok {
			svmInstance.SetLabel("iscsi_service_enabled", available)
		}

		// Update fpolicy_enabled, fpolicy_name label in svm
		if fpolicyData, ok := s.fpolicyData[svmName]; ok {
			svmInstance.SetLabel("fpolicy_enabled", fpolicyData.enable)
			svmInstance.SetLabel("fpolicy_name", fpolicyData.name)
		}

		// Update ldap_session_security label in svm
		if ldapSessionSecurity, ok := s.ldapData[svmName]; ok {
			svmInstance.SetLabel("ldap_session_security", ldapSessionSecurity)
		}

		// Update nfs_kerberos_protocol_enabled label in svm
		if kerberosEnabled, ok := s.kerberosConfig[svmName]; ok {
			svmInstance.SetLabel("nfs_kerberos_protocol_enabled", kerberosEnabled)
		}
	}

	s.currentVal++
	return nil, s.client.Metadata, nil
}

func (s *SVM) GetAuditProtocols() (map[string]string, error) {
	var (
		result                []*node.Node
		request               *node.Node
		vserverAuditEnableMap map[string]string
		err                   error
	)

	vserverAuditEnableMap = make(map[string]string)

	request = node.NewXMLS("fileservice-audit-config-get-iter")
	request.NewChildS("max-records", s.batchSize)

	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, fileServiceAuditConfig := range result {
		auditEnabled := fileServiceAuditConfig.GetChildContentS("is-enabled")
		svmName := fileServiceAuditConfig.GetChildContentS("vserver")
		vserverAuditEnableMap[svmName] = auditEnabled
	}
	return vserverAuditEnableMap, nil
}

func (s *SVM) GetCifsProtocols() (map[string]CifsSecurity, error) {
	var (
		result             []*node.Node
		request            *node.Node
		vserverCifsDataMap map[string]CifsSecurity
		err                error
	)

	vserverCifsDataMap = make(map[string]CifsSecurity)

	request = node.NewXMLS("cifs-security-get-iter")
	request.NewChildS("max-records", s.batchSize)

	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, cifsSecurityData := range result {
		lmCompatibilityLevel := cifsSecurityData.GetChildContentS("lm-compatibility-level")
		smbSigning := cifsSecurityData.GetChildContentS("is-signing-required")
		smbEncryption := cifsSecurityData.GetChildContentS("is-smb-encryption-required")
		svmName := cifsSecurityData.GetChildContentS("vserver")
		vserverCifsDataMap[svmName] = CifsSecurity{cifsNtlmEnabled: lmCompatibilityLevel, smbEncryption: smbEncryption, smbSigning: smbSigning}
	}
	return vserverCifsDataMap, nil
}

func (s *SVM) GetNSSwitchInfo() (map[string]Nsswitch, error) {
	var (
		result             []*node.Node
		request            *node.Node
		vserverNsswitchMap map[string]Nsswitch
		ok                 bool
		err                error
	)

	vserverNsswitchMap = make(map[string]Nsswitch)

	request = node.NewXMLS("nameservice-nsswitch-get-iter")
	request.NewChildS("max-records", s.batchSize)

	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, nsswitchConfig := range result {
		var ns Nsswitch
		svmName := nsswitchConfig.GetChildContentS("vserver-name")
		if nssource := nsswitchConfig.GetChildS("nameservice-sources"); nssource != nil {
			nssourcelist := nssource.GetAllChildContentS()
			nsdb := nsswitchConfig.GetChildContentS("nameservice-database")
			if ns, ok = vserverNsswitchMap[svmName]; ok {
				ns.nsdb = append(ns.nsdb, nsdb)
				ns.nssource = append(ns.nssource, nssourcelist...)
			} else {
				ns = Nsswitch{nsdb: []string{nsdb}, nssource: nssourcelist}
			}
			vserverNsswitchMap[svmName] = ns
		}
	}
	return vserverNsswitchMap, nil
}

func (s *SVM) GetNisInfo() (map[string]string, error) {
	var (
		result        []*node.Node
		request       *node.Node
		vserverNisMap map[string]string
		err           error
	)

	vserverNisMap = make(map[string]string)

	request = node.NewXMLS("nis-get-iter")
	request.NewChildS("max-records", s.batchSize)

	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, nisData := range result {
		nisDomain := nisData.GetChildContentS("nis-domain")
		svmName := nisData.GetChildContentS("vserver")
		vserverNisMap[svmName] = nisDomain
	}
	return vserverNisMap, nil
}

func (s *SVM) GetCifsEnabled() (map[string]bool, error) {
	var (
		result         []*node.Node
		request        *node.Node
		vserverCifsMap map[string]bool
		err            error
	)

	vserverCifsMap = make(map[string]bool)

	request = node.NewXMLS("cifs-server-get-iter")
	request.NewChildS("max-records", s.batchSize)
	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, cifsConfig := range result {
		adminStatus := cifsConfig.GetChildContentS("administrative-status") == "up"
		svmName := cifsConfig.GetChildContentS("vserver")
		vserverCifsMap[svmName] = adminStatus
	}
	return vserverCifsMap, nil
}

func (s *SVM) GetNfsEnabled() (map[string]string, error) {
	var (
		result        []*node.Node
		request       *node.Node
		vserverNfsMap map[string]string
		err           error
	)

	vserverNfsMap = make(map[string]string)

	request = node.NewXMLS("nfs-service-get-iter")
	request.NewChildS("max-records", s.batchSize)

	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, nfsConfig := range result {
		adminStatus := nfsConfig.GetChildContentS("is-nfs-access-enabled")
		svmName := nfsConfig.GetChildContentS("vserver")
		vserverNfsMap[svmName] = adminStatus
	}
	return vserverNfsMap, nil
}

func (s *SVM) GetSSHData() (map[string]SSHInfo, error) {
	var (
		result  []*node.Node
		request *node.Node
		sshMap  map[string]SSHInfo
		err     error
	)

	sshMap = make(map[string]SSHInfo)

	request = node.NewXMLS("security-ssh-get-iter")
	request.NewChildS("max-records", s.batchSize)

	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, sshData := range result {
		svmName := sshData.GetChildContentS("vserver-name")
		sshList := sshData.GetChildS("ciphers").GetAllChildContentS()
		sort.Strings(sshList)
		ciphersVal := strings.Join(sshList, ",")
		insecured := weakCiphers.MatchString(ciphersVal)
		sshMap[svmName] = SSHInfo{ciphers: ciphersVal, isInsecure: strconv.FormatBool(insecured)}
	}
	return sshMap, nil
}

func (s *SVM) GetIscsiInitiatorAuth() (map[string]string, error) {
	var (
		result              []*node.Node
		request             *node.Node
		vserverIscsiAuthMap map[string]string
		err                 error
	)

	vserverIscsiAuthMap = make(map[string]string)

	request = node.NewXMLS("iscsi-initiator-auth-get-iter")
	request.NewChildS("max-records", s.batchSize)

	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, iscsiSecurityEntry := range result {
		authType := iscsiSecurityEntry.GetChildContentS("auth-type")
		svmName := iscsiSecurityEntry.GetChildContentS("vserver")
		if _, ok := vserverIscsiAuthMap[svmName]; !ok {
			vserverIscsiAuthMap[svmName] = authType
		} else {
			// If svm is already present, update the map value with append this authenticationType to previous value
			vserverIscsiAuthMap[svmName] = vserverIscsiAuthMap[svmName] + "," + authType
		}
	}
	return vserverIscsiAuthMap, nil
}

func (s *SVM) GetIscsiService() (map[string]string, error) {
	var (
		result                 []*node.Node
		request                *node.Node
		vserverIscsiServiceMap map[string]string
		err                    error
	)

	vserverIscsiServiceMap = make(map[string]string)

	request = node.NewXMLS("iscsi-service-get-iter")
	request.NewChildS("max-records", s.batchSize)

	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, iscsiService := range result {
		available := iscsiService.GetChildContentS("is-available")
		svmName := iscsiService.GetChildContentS("vserver")
		if _, ok := vserverIscsiServiceMap[svmName]; !ok {
			vserverIscsiServiceMap[svmName] = available
		} else if vserverIscsiServiceMap[svmName] == "false" {
			// If svm is already present, update the map value only if previous value is false
			vserverIscsiServiceMap[svmName] = available
		}
	}
	return vserverIscsiServiceMap, nil
}

func (s *SVM) GetFpolicy() (map[string]Fpolicy, error) {
	var (
		result            []*node.Node
		request           *node.Node
		vserverFpolicyMap map[string]Fpolicy
		err               error
	)

	vserverFpolicyMap = make(map[string]Fpolicy)

	request = node.NewXMLS("fpolicy-policy-status-get-iter")
	request.NewChildS("max-records", s.batchSize)

	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, fpolicyData := range result {
		name := fpolicyData.GetChildContentS("policy-name")
		enable := fpolicyData.GetChildContentS("status")
		svmName := fpolicyData.GetChildContentS("vserver")
		if _, ok := vserverFpolicyMap[svmName]; !ok {
			vserverFpolicyMap[svmName] = Fpolicy{name: name, enable: enable}
		} else if vserverFpolicyMap[svmName].enable == "false" {
			// If svm is already present, update the status value only if it is false
			vserverFpolicyMap[svmName] = Fpolicy{name: name, enable: enable}
		}
	}
	return vserverFpolicyMap, nil
}

func (s *SVM) GetLdapData() (map[string]string, error) {
	var (
		result         []*node.Node
		request        *node.Node
		vserverLdapMap map[string]string
		err            error
	)

	vserverLdapMap = make(map[string]string)

	request = node.NewXMLS("ldap-client-get-iter")
	request.NewChildS("max-records", s.batchSize)

	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, ldapData := range result {
		ldapSessionSecurity := ldapData.GetChildContentS("session-security")
		svmName := ldapData.GetChildContentS("vserver")
		vserverLdapMap[svmName] = ldapSessionSecurity
	}
	return vserverLdapMap, nil
}

func (s *SVM) GetKerberosConfig() (map[string]string, error) {
	var (
		result             []*node.Node
		request            *node.Node
		vserverKerberosMap map[string]string
		err                error
	)

	vserverKerberosMap = make(map[string]string)

	request = node.NewXMLS("kerberos-config-get-iter")
	request.NewChildS("max-records", s.batchSize)

	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, kerberosConfig := range result {
		enable := kerberosConfig.GetChildContentS("is-kerberos-enabled")
		svmName := kerberosConfig.GetChildContentS("vserver")
		if _, ok := vserverKerberosMap[svmName]; !ok {
			vserverKerberosMap[svmName] = enable
		} else if enable == "true" {
			// If any interface on the svm has kerberos on, then only set to true
			vserverKerberosMap[svmName] = enable
		}

	}
	return vserverKerberosMap, nil
}
