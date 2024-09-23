/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package security

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
)

type Security struct {
	*plugin.AbstractPlugin
	currentVal    int
	client        *zapi.Client
	fipsEnabled   string
	rshEnabled    string
	telnetEnabled string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Security{AbstractPlugin: p}
}

func (s *Security) Init() error {

	var err error

	if err := s.InitAbc(); err != nil {
		return err
	}

	if s.client, err = zapi.New(conf.ZapiPoller(s.ParentParams), s.Auth); err != nil {
		s.SLogger.Error("connecting", slog.Any("err", err))
		return err
	}

	if err := s.client.Init(5); err != nil {
		return err
	}

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	s.currentVal = s.SetPluginInterval()

	return nil
}

func (s *Security) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

	var (
		err error
	)

	data := dataMap[s.Object]
	s.client.Metadata.Reset()

	if s.currentVal >= s.PluginInvocationRate {
		s.currentVal = 0

		// invoke security-config-get zapi with 'ssl' interface and get fips status
		if s.fipsEnabled, err = s.getSecurityConfig(); err != nil {
			s.SLogger.Warn("Failed to collect fips enable status", slog.Any("err", err))
		}

		// invoke security-protocol-get zapi with 'telnet' and 'rsh' and get
		if s.telnetEnabled, s.rshEnabled, err = s.getSecurityProtocols(); err != nil {
			s.SLogger.Warn("Failed to collect telnet and rsh enable status", slog.Any("err", err))
		}

		// update instance based on the above zapi response
		for _, securityInstance := range data.GetInstances() {
			if !securityInstance.IsExportable() {
				continue
			}
			// Update fips_enabled label in instance
			securityInstance.SetLabel("fips_enabled", s.fipsEnabled)

			// Update telnet_enabled and rsh_enabled label in instance
			securityInstance.SetLabel("telnet_enabled", s.telnetEnabled)
			securityInstance.SetLabel("rsh_enabled", s.rshEnabled)
		}
	}

	s.currentVal++
	return nil, s.client.Metadata, nil
}

func (s *Security) getSecurityConfig() (string, error) {

	var (
		result      []*node.Node
		request     *node.Node
		fipsEnabled string
		err         error
	)

	request = node.NewXMLS("security-config-get")
	request.NewChildS("interface", "ssl")

	// fetching only ssl interface
	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return "", err
	}

	if len(result) == 0 || result == nil {
		return "", errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, securityConfig := range result {
		fipsEnabled = securityConfig.GetChildContentS("is-fips-enabled")
		break
	}
	return fipsEnabled, nil
}

func (s *Security) getSecurityProtocols() (string, string, error) {

	var (
		request       *node.Node
		telnetEnabled string
		rshEnabled    string
		err           error
	)

	// Zapi call for telnet
	request = node.NewXMLS("security-protocol-get")
	request.NewChildS("application", "telnet")
	if telnetEnabled, err = s.getEnabledValue(request); err != nil {
		return "", "", err
	}

	// Zapi call for rsh
	request = node.NewXMLS("security-protocol-get")
	request.NewChildS("application", "rsh")
	if rshEnabled, err = s.getEnabledValue(request); err != nil {
		return "", "", err
	}

	return telnetEnabled, rshEnabled, nil
}

func (s *Security) getEnabledValue(request *node.Node) (string, error) {
	var (
		result  []*node.Node
		enabled string
		err     error
	)

	// fetching only telnet/rsh protocols
	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return "", err
	}

	if len(result) == 0 || result == nil {
		return "", errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, securityConfig := range result {
		enabled = securityConfig.GetChildContentS("enabled")
		break
	}

	return enabled, nil
}
