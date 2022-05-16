/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package svm

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errors"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"strconv"
	"strings"
	"time"
)

const DefaultPluginDuration = 30 * time.Minute
const DefaultDataPollDuration = 3 * time.Minute
const BatchSize = "500"

type SVM struct {
	*plugin.AbstractPlugin
	pluginInvocationRate int
	currentVal           int
	batchSize            string
	client               *zapi.Client
	auditProtocols       map[string]string
	cifsProtocols        map[string]string
	nsswitchInfo         map[string]nsswitch
	nisInfo              map[string]string
	cifsEnabled          map[string]bool
	nfsEnabled           map[string]string
	sshData              map[string]string
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

	if my.client, err = zapi.New(conf.ZapiPoller(my.ParentParams)); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	my.auditProtocols = make(map[string]string)
	my.cifsProtocols = make(map[string]string)
	my.nsswitchInfo = make(map[string]nsswitch)
	my.nisInfo = make(map[string]string)
	my.cifsEnabled = make(map[string]bool)
	my.nfsEnabled = make(map[string]string)
	my.sshData = make(map[string]string)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	if my.currentVal, err = collectors.SetPluginInterval(my.ParentParams, my.Params, my.Logger, DefaultDataPollDuration, DefaultPluginDuration); err != nil {
		my.Logger.Error().Err(err).Stack().Msg("Failed while setting the plugin interval")
		return err
	}

	if b := my.Params.GetChildContentS("batch_size"); b != "" {
		if _, err := strconv.Atoi(b); err == nil {
			my.batchSize = b
			my.Logger.Info().Str("BatchSize", my.batchSize).Msg("using batch-size")
		}
	} else {
		my.batchSize = BatchSize
		my.Logger.Trace().Str("BatchSize", BatchSize).Msg("Using default batch-size")
	}

	return nil
}

func (my *SVM) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		err error
	)

	if my.currentVal >= my.pluginInvocationRate {
		my.currentVal = 0

		// invoke fileservice-audit-config-get-iter zapi and get audit protocols
		if my.auditProtocols, err = my.GetAuditProtocols(); err != nil {
			my.Logger.Warn().Stack().Err(err).Msg("Failed to collect audit protocols")
			//return nil, nil
		}

		// invoke cifs-security-get-iter zapi and get cifs protocols
		if my.cifsProtocols, err = my.GetCifsProtocols(); err != nil {
			my.Logger.Warn().Stack().Err(err).Msg("Failed to collect cifs protocols")
			//return nil, nil
		}

		// invoke nameservice-nsswitch-get-iter zapi and get nsswitch info
		if my.nsswitchInfo, err = my.GetNSSwitchInfo(); err != nil {
			my.Logger.Warn().Stack().Err(err).Msg("Failed to collect nsswitch info")
			//return nil, nil
		}

		// invoke nis-get-iter zapi and get nisdomain info
		if my.nisInfo, err = my.GetNisInfo(); err != nil {
			my.Logger.Warn().Stack().Err(err).Msg("Failed to collect nisdomain info")
			//return nil, nil
		}

		// invoke cifs-server-get-iter zapi and get cifsenabled info
		if my.cifsEnabled, err = my.GetCifsEnabled(); err != nil {
			my.Logger.Warn().Stack().Err(err).Msg("Failed to collect cifsenabled info")
			//return nil, nil
		}

		// invoke nfs-service-get-iter zapi and get cifsenabled info
		if my.nfsEnabled, err = my.GetNfsEnabled(); err != nil {
			my.Logger.Warn().Stack().Err(err).Msg("Failed to collect nfsenabled info")
			//return nil, nil
		}

		// invoke security-ssh-get-iter zapi and get ssh data
		if my.sshData, err = my.GetSSHData(); err != nil {
			my.Logger.Warn().Stack().Err(err).Msg("Failed to collect ssh data")
			//return nil, nil
		}

		// update svm instance based on the above zapi response
		for _, svmInstance := range data.GetInstances() {
			svmName := svmInstance.GetLabel("svm")

			// Update audit_protocol_enabled label in svm
			svmInstance.SetLabel("audit_protocol_enabled", my.auditProtocols[svmName])

			// Update cifs_ntlm_enabled label in svm
			if ntmlEnabled, ok := my.cifsProtocols[svmName]; ok {
				svmInstance.SetLabel("cifs_ntlm_enabled", ntmlEnabled)
			}

			// Update nis_domain label in svm
			svmInstance.SetLabel("nis_domain", my.nisInfo[svmName])

			// Update nameservice_switch label in svm
			if nsswitchInfo, ok := my.nsswitchInfo[svmName]; ok {
				nsDb := strings.Join(nsswitchInfo.nsdb, ",")
				nsSource := strings.Join(nsswitchInfo.nssource, ",")
				nisDomain := my.nisInfo[svmName]
				svmInstance.SetLabel("ns_source", nsSource)
				svmInstance.SetLabel("ns_db", nsDb)
				collectors.SetNameservice(nsDb, nsSource, nisDomain, svmInstance)
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

	if result, _, err = collectors.InvokeZapiCall(my.client, request, my.Logger, ""); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errors.New(errors.ErrNoInstance, "no records found")
	}

	for _, fileServiceAuditConfig := range result {
		auditEnabled := fileServiceAuditConfig.GetChildContentS("is-enabled")
		svmName := fileServiceAuditConfig.GetChildContentS("vserver")
		vserverAuditEnableMap[svmName] = auditEnabled
	}
	return vserverAuditEnableMap, nil
}

func (my *SVM) GetCifsProtocols() (map[string]string, error) {
	var (
		result               []*node.Node
		request              *node.Node
		vserverCifsEnableMap map[string]string
		err                  error
	)

	vserverCifsEnableMap = make(map[string]string)

	request = node.NewXMLS("cifs-security-get-iter")
	request.NewChildS("max-records", my.batchSize)

	if result, _, err = collectors.InvokeZapiCall(my.client, request, my.Logger, ""); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errors.New(errors.ErrNoInstance, "no records found")
	}

	for _, cifsSecurity := range result {
		lmCompatibilityLevel := cifsSecurity.GetChildContentS("lm-compatibility-level")
		svmName := cifsSecurity.GetChildContentS("vserver")
		vserverCifsEnableMap[svmName] = lmCompatibilityLevel
	}
	return vserverCifsEnableMap, nil
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

	if result, _, err = collectors.InvokeZapiCall(my.client, request, my.Logger, ""); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errors.New(errors.ErrNoInstance, "no records found")
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

	if result, _, err = collectors.InvokeZapiCall(my.client, request, my.Logger, ""); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errors.New(errors.ErrNoInstance, "no records found")
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

	if result, _, err = collectors.InvokeZapiCall(my.client, request, my.Logger, ""); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errors.New(errors.ErrNoInstance, "no records found")
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

	if result, _, err = collectors.InvokeZapiCall(my.client, request, my.Logger, ""); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errors.New(errors.ErrNoInstance, "no records found")
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

	if result, _, err = collectors.InvokeZapiCall(my.client, request, my.Logger, ""); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errors.New(errors.ErrNoInstance, "no records found")
	}

	for _, sshData := range result {
		svmName := sshData.GetChildContentS("vserver-name")
		sshList := sshData.GetChildS("ciphers").GetAllChildContentS()
		sshMap[svmName] = strings.Join(sshList, ",")
	}
	return sshMap, nil
}
