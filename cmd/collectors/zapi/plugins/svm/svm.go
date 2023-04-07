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
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"sort"
	"strconv"
	"strings"
)

const BatchSize = "500"

type SVM struct {
	*plugin.AbstractPlugin
	currentVal     int
	batchSize      string
	client         *zapi.Client
	auditProtocols map[string]string
	cifsProtocols  map[string]cifsSecurity
	nsswitchInfo   map[string]nsswitch
	nisInfo        map[string]string
	cifsEnabled    map[string]bool
	nfsEnabled     map[string]string
	sshData        map[string]string
	iscsiAuth      map[string]string
	iscsiService   map[string]string
	fpolicyData    map[string]fpolicy
	ldapData       map[string]string
	kerberosConfig map[string]string
}

type nsswitch struct {
	nsdb     []string
	nssource []string
}

type fpolicy struct {
	name   string
	enable string
}

type cifsSecurity struct {
	cifsNtlmEnabled string
	smbEncryption   string
	smbSigning      string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SVM{AbstractPlugin: p}
}

func (my *SVM) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	if my.client, err = zapi.New(conf.ZapiPoller(my.ParentParams), my.Auth); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	my.auditProtocols = make(map[string]string)
	my.cifsProtocols = make(map[string]cifsSecurity)
	my.nsswitchInfo = make(map[string]nsswitch)
	my.nisInfo = make(map[string]string)
	my.cifsEnabled = make(map[string]bool)
	my.nfsEnabled = make(map[string]string)
	my.sshData = make(map[string]string)
	my.iscsiAuth = make(map[string]string)
	my.iscsiService = make(map[string]string)
	my.fpolicyData = make(map[string]fpolicy)
	my.ldapData = make(map[string]string)
	my.kerberosConfig = make(map[string]string)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	my.currentVal = my.SetPluginInterval()

	my.batchSize = BatchSize
	if b := my.Params.GetChildContentS("batch_size"); b != "" {
		if _, err := strconv.Atoi(b); err == nil {
			my.batchSize = b
			my.Logger.Info().Str("BatchSize", my.batchSize).Msg("using batch-size")
		}
	} else {
		my.Logger.Trace().Str("BatchSize", BatchSize).Msg("Using default batch-size")
	}

	return nil
}

func (my *SVM) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		err error
	)

	data := dataMap[my.Object]
	if my.currentVal >= my.PluginInvocationRate {
		my.currentVal = 0

		// invoke fileservice-audit-config-get-iter zapi and get audit protocols
		if my.auditProtocols, err = my.GetAuditProtocols(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect audit protocols")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect audit protocols")
			}
		}

		// invoke cifs-security-get-iter zapi and get cifs protocols
		if my.cifsProtocols, err = my.GetCifsProtocols(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect cifs protocols")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect cifs protocols")
			}
		}

		// invoke nameservice-nsswitch-get-iter zapi and get nsswitch info
		if my.nsswitchInfo, err = my.GetNSSwitchInfo(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect nsswitch info")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect nsswitch info")
			}
		}

		// invoke nis-get-iter zapi and get nisdomain info
		if my.nisInfo, err = my.GetNisInfo(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect nisdomain info")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect nisdomain info")
			}
		}

		// invoke cifs-server-get-iter zapi and get cifsenabled info
		if my.cifsEnabled, err = my.GetCifsEnabled(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect cifsenabled info")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect cifsenabled info")
			}
		}

		// invoke nfs-service-get-iter zapi and get cifsenabled info
		if my.nfsEnabled, err = my.GetNfsEnabled(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect nfsenabled info")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect nfsenabled info")
			}
		}

		// invoke security-ssh-get-iter zapi and get ssh data
		if my.sshData, err = my.GetSSHData(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect ssh data")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect ssh data")
			}
		}

		// invoke iscsi-initiator-auth-get-iter zapi and get iscsi_authentication_type
		if my.iscsiAuth, err = my.GetIscsiInitiatorAuth(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect iscsi authentication type")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect iscsi authentication type")
			}
		}

		// invoke iscsi-service-get-iter zapi and get iscsi_service_enabled
		if my.iscsiService, err = my.GetIscsiService(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect iscsi service")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect iscsi service")
			}
		}

		// invoke fpolicy-policy-status-get-iter zapi and get fpolicy_enabled, fpolicy_name
		if my.fpolicyData, err = my.GetFpolicy(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect fpolicy detail")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect fpolicy detail")
			}
		}

		// invoke ldap-client-get-iter zapi and get ldap_session_security
		if my.ldapData, err = my.GetLdapData(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect ldap session")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect ldap session")
			}
		}

		// invoke kerberos-config-get-iter zapi and get nfs_kerberos_protocol_enabled
		if my.kerberosConfig, err = my.GetKerberosConfig(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect kerberos config")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect kerberos config")
			}
		}

		// update svm instance based on the above zapi response
		for _, svmInstance := range data.GetInstances() {
			svmName := svmInstance.GetLabel("svm")

			// Update audit_protocol_enabled label in svm
			svmInstance.SetLabel("audit_protocol_enabled", my.auditProtocols[svmName])

			// Update cifs_ntlm_enabled label in svm
			if cifsData, ok := my.cifsProtocols[svmName]; ok {
				svmInstance.SetLabel("cifs_ntlm_enabled", cifsData.cifsNtlmEnabled)
				svmInstance.SetLabel("smb_encryption_required", cifsData.smbEncryption)
				svmInstance.SetLabel("smb_signing_required", cifsData.smbSigning)
			}

			// Update nis_domain label in svm
			svmInstance.SetLabel("nis_domain", my.nisInfo[svmName])

			// Update nameservice_switch label in svm
			if nsswitchInfo, ok := my.nsswitchInfo[svmName]; ok {
				sort.Strings(nsswitchInfo.nsdb)
				sort.Strings(nsswitchInfo.nssource)
				nsDB := strings.Join(nsswitchInfo.nsdb, ",")
				nsSource := strings.Join(nsswitchInfo.nssource, ",")
				nisDomain := my.nisInfo[svmName]
				svmInstance.SetLabel("ns_source", nsSource)
				svmInstance.SetLabel("ns_db", nsDB)
				collectors.SetNameservice(nsDB, nsSource, nisDomain, svmInstance)
			}

			// Update cifs_protocol_enabled label in svm
			if cifsEnable, ok := my.cifsEnabled[svmName]; ok {
				svmInstance.SetLabel("cifs_protocol_enabled", strconv.FormatBool(cifsEnable))
			}

			// Update nfs_protocol_enabled label in svm
			if nfsEnable, ok := my.nfsEnabled[svmName]; ok {
				svmInstance.SetLabel("nfs_protocol_enabled", nfsEnable)
			}

			// Update ciphers label in svm
			if sshInfo, ok := my.sshData[svmName]; ok {
				svmInstance.SetLabel("ciphers", sshInfo)
			}

			// Update iscsi_authentication_type label in svm
			if authType, ok := my.iscsiAuth[svmName]; ok {
				svmInstance.SetLabel("iscsi_authentication_type", authType)
			}

			// Update iscsi_service_enabled label in svm
			if available, ok := my.iscsiService[svmName]; ok {
				svmInstance.SetLabel("iscsi_service_enabled", available)
			}

			// Update fpolicy_enabled, fpolicy_name label in svm
			if fpolicyData, ok := my.fpolicyData[svmName]; ok {
				svmInstance.SetLabel("fpolicy_enabled", fpolicyData.enable)
				svmInstance.SetLabel("fpolicy_name", fpolicyData.name)
			}

			// Update ldap_session_security label in svm
			if ldapSessionSecurity, ok := my.ldapData[svmName]; ok {
				svmInstance.SetLabel("ldap_session_security", ldapSessionSecurity)
			}

			// Update nfs_kerberos_protocol_enabled label in svm
			if kerberosEnabled, ok := my.kerberosConfig[svmName]; ok {
				svmInstance.SetLabel("nfs_kerberos_protocol_enabled", kerberosEnabled)
			}

		}
	}

	my.currentVal++
	return nil, nil
}

func (my *SVM) GetAuditProtocols() (map[string]string, error) {
	var (
		result                []*node.Node
		request               *node.Node
		vserverAuditEnableMap map[string]string
		err                   error
	)

	vserverAuditEnableMap = make(map[string]string)

	request = node.NewXMLS("fileservice-audit-config-get-iter")
	request.NewChildS("max-records", my.batchSize)

	if result, err = my.client.InvokeZapiCall(request); err != nil {
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

func (my *SVM) GetCifsProtocols() (map[string]cifsSecurity, error) {
	var (
		result             []*node.Node
		request            *node.Node
		vserverCifsDataMap map[string]cifsSecurity
		err                error
	)

	vserverCifsDataMap = make(map[string]cifsSecurity)

	request = node.NewXMLS("cifs-security-get-iter")
	request.NewChildS("max-records", my.batchSize)

	if result, err = my.client.InvokeZapiCall(request); err != nil {
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
		vserverCifsDataMap[svmName] = cifsSecurity{cifsNtlmEnabled: lmCompatibilityLevel, smbEncryption: smbEncryption, smbSigning: smbSigning}
	}
	return vserverCifsDataMap, nil
}

func (my *SVM) GetNSSwitchInfo() (map[string]nsswitch, error) {
	var (
		result             []*node.Node
		request            *node.Node
		vserverNsswitchMap map[string]nsswitch
		ns                 nsswitch
		ok                 bool
		err                error
	)

	vserverNsswitchMap = make(map[string]nsswitch)

	request = node.NewXMLS("nameservice-nsswitch-get-iter")
	request.NewChildS("max-records", my.batchSize)

	if result, err = my.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, nsswitchConfig := range result {
		nsdb := nsswitchConfig.GetChildContentS("nameservice-database")
		svmName := nsswitchConfig.GetChildContentS("vserver-name")
		nssource := nsswitchConfig.GetChildS("nameservice-sources")
		nssourcelist := nssource.GetAllChildContentS()

		if ns, ok = vserverNsswitchMap[svmName]; ok {
			ns.nsdb = append(ns.nsdb, nsdb)
			ns.nssource = append(ns.nssource, nssourcelist...)
		} else {
			ns = nsswitch{nsdb: []string{nsdb}, nssource: nssourcelist}
		}
		vserverNsswitchMap[svmName] = ns
	}
	return vserverNsswitchMap, nil
}

func (my *SVM) GetNisInfo() (map[string]string, error) {
	var (
		result        []*node.Node
		request       *node.Node
		vserverNisMap map[string]string
		err           error
	)

	vserverNisMap = make(map[string]string)

	request = node.NewXMLS("nis-get-iter")
	request.NewChildS("max-records", my.batchSize)

	if result, err = my.client.InvokeZapiCall(request); err != nil {
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

func (my *SVM) GetCifsEnabled() (map[string]bool, error) {
	var (
		result         []*node.Node
		request        *node.Node
		vserverCifsMap map[string]bool
		err            error
	)

	vserverCifsMap = make(map[string]bool)

	request = node.NewXMLS("cifs-server-get-iter")
	request.NewChildS("max-records", my.batchSize)
	if result, err = my.client.InvokeZapiCall(request); err != nil {
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

func (my *SVM) GetNfsEnabled() (map[string]string, error) {
	var (
		result        []*node.Node
		request       *node.Node
		vserverNfsMap map[string]string
		err           error
	)

	vserverNfsMap = make(map[string]string)

	request = node.NewXMLS("nfs-service-get-iter")
	request.NewChildS("max-records", my.batchSize)

	if result, err = my.client.InvokeZapiCall(request); err != nil {
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

func (my *SVM) GetSSHData() (map[string]string, error) {
	var (
		result  []*node.Node
		request *node.Node
		sshMap  map[string]string
		err     error
	)

	sshMap = make(map[string]string)

	request = node.NewXMLS("security-ssh-get-iter")
	request.NewChildS("max-records", my.batchSize)

	if result, err = my.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, sshData := range result {
		svmName := sshData.GetChildContentS("vserver-name")
		sshList := sshData.GetChildS("ciphers").GetAllChildContentS()
		sshMap[svmName] = strings.Join(sshList, ",")
	}
	return sshMap, nil
}

func (my *SVM) GetIscsiInitiatorAuth() (map[string]string, error) {
	var (
		result              []*node.Node
		request             *node.Node
		vserverIscsiAuthMap map[string]string
		err                 error
	)

	vserverIscsiAuthMap = make(map[string]string)

	request = node.NewXMLS("iscsi-initiator-auth-get-iter")
	request.NewChildS("max-records", my.batchSize)

	if result, err = my.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, iscsiSecurityEntry := range result {
		authType := iscsiSecurityEntry.GetChildContentS("auth-type")
		svmName := iscsiSecurityEntry.GetChildContentS("vserver")
		vserverIscsiAuthMap[svmName] = authType
	}
	return vserverIscsiAuthMap, nil
}

func (my *SVM) GetIscsiService() (map[string]string, error) {
	var (
		result                 []*node.Node
		request                *node.Node
		vserverIscsiServiceMap map[string]string
		err                    error
	)

	vserverIscsiServiceMap = make(map[string]string)

	request = node.NewXMLS("iscsi-service-get-iter")
	request.NewChildS("max-records", my.batchSize)

	if result, err = my.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, iscsiService := range result {
		available := iscsiService.GetChildContentS("is-available")
		svmName := iscsiService.GetChildContentS("vserver")
		vserverIscsiServiceMap[svmName] = available
	}
	return vserverIscsiServiceMap, nil
}

func (my *SVM) GetFpolicy() (map[string]fpolicy, error) {
	var (
		result            []*node.Node
		request           *node.Node
		vserverFpolicyMap map[string]fpolicy
		err               error
	)

	vserverFpolicyMap = make(map[string]fpolicy)

	request = node.NewXMLS("fpolicy-policy-status-get-iter")
	request.NewChildS("max-records", my.batchSize)

	if result, err = my.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, fpolicyData := range result {
		name := fpolicyData.GetChildContentS("policy-name")
		enable := fpolicyData.GetChildContentS("status")
		svmName := fpolicyData.GetChildContentS("vserver")
		vserverFpolicyMap[svmName] = fpolicy{name: name, enable: enable}
	}
	return vserverFpolicyMap, nil
}

func (my *SVM) GetLdapData() (map[string]string, error) {
	var (
		result         []*node.Node
		request        *node.Node
		vserverLdapMap map[string]string
		err            error
	)

	vserverLdapMap = make(map[string]string)

	request = node.NewXMLS("ldap-client-get-iter")
	request.NewChildS("max-records", my.batchSize)

	if result, err = my.client.InvokeZapiCall(request); err != nil {
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

func (my *SVM) GetKerberosConfig() (map[string]string, error) {
	var (
		result             []*node.Node
		request            *node.Node
		vserverKerberosMap map[string]string
		err                error
	)

	vserverKerberosMap = make(map[string]string)

	request = node.NewXMLS("kerberos-config-get-iter")
	request.NewChildS("max-records", my.batchSize)

	if result, err = my.client.InvokeZapiCall(request); err != nil {
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
		} else {
			// If any interface on the svm has kerberos on, then only set to true
			if enable == "true" {
				vserverKerberosMap[svmName] = enable
			}
		}

	}
	return vserverKerberosMap, nil
}
