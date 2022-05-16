/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */
package security

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errors"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"time"
)

const DefaultPluginDuration = 30 * time.Minute
const DefaultDataPollDuration = 3 * time.Minute

type Security struct {
	*plugin.AbstractPlugin
	pluginInvocationRate int
	currentVal           int
	batchSize            string
	client               *zapi.Client
	query                string
	fipsEnabled          string
	rshEnabled           string
	telnetEnabled        string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Security{AbstractPlugin: p}
}

func (my *Security) Init() error {

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

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	if my.currentVal, err = collectors.SetPluginInterval(my.ParentParams, my.Params, my.Logger, DefaultDataPollDuration, DefaultPluginDuration); err != nil {
		my.Logger.Error().Err(err).Stack().Msg("Failed while setting the plugin interval")
		return err
	}

	return nil
}

func (my *Security) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		err error
	)

	if my.currentVal >= my.pluginInvocationRate {
		my.currentVal = 0

		// invoke security-config-get zapi with 'ssl' interface and get fips status
		if my.fipsEnabled, err = my.getSecurityConfig(); err != nil {
			my.Logger.Warn().Stack().Err(err).Msg("Failed to collect fips enable status")
			//return nil, nil
		}

		// invoke security-protocol-get zapi with 'telnet' and 'rsh' and get
		if my.telnetEnabled, my.rshEnabled, err = my.getSecurityProtocols(); err != nil {
			my.Logger.Warn().Stack().Err(err).Msg("Failed to collect telnet and rsh enable status")
			//return nil, nil
		}

		// update instance based on the above zapi response
		for _, securityInstance := range data.GetInstances() {
			// Update fips_enabled label in instance
			securityInstance.SetLabel("fips_enabled", my.fipsEnabled)

			// Update telnet_enabled and rsh_enabled label in instance
			securityInstance.SetLabel("telnet_enabled", my.telnetEnabled)
			securityInstance.SetLabel("rsh_enabled", my.rshEnabled)
		}
	}

	my.currentVal++
	return nil, nil
}

func (my *Security) getSecurityConfig() (string, error) {

	var (
		result      []*node.Node
		request     *node.Node
		fipsEnabled string
		err         error
	)

	request = node.NewXMLS("security-config-get")
	request.NewChildS("interface", "ssl")

	if result, _, err = collectors.InvokeZapiCall(my.client, request, my.Logger, ""); err != nil {
		return "", err
	}

	if len(result) == 0 || result == nil {
		return "", errors.New(errors.ErrNoInstance, "no records found")
	}

	for _, securityConfig := range result {
		fipsEnabled = securityConfig.GetChildContentS("is-fips-enabled")
	}
	return fipsEnabled, nil
}

func (my *Security) getSecurityProtocols() (string, string, error) {

	var (
		request       *node.Node
		telnetEnabled string
		rshEnabled    string
		err           error
	)

	// Zapi call for telnet
	request = node.NewXMLS("security-protocol-get")
	request.NewChildS("application", "telnet")
	if telnetEnabled, err = my.getEnabledValue(request); err != nil {
		return "", "", err
	}

	// Zapi call for rsh
	request = node.NewXMLS("security-protocol-get")
	request.NewChildS("application", "rsh")
	if rshEnabled, err = my.getEnabledValue(request); err != nil {
		return "", "", err
	}

	return telnetEnabled, rshEnabled, nil
}

func (my *Security) getEnabledValue(request *node.Node) (string, error) {
	var (
		result  []*node.Node
		enabled string
		err     error
	)

	if result, _, err = collectors.InvokeZapiCall(my.client, request, my.Logger, ""); err != nil {
		return "", err
	}

	if len(result) == 0 || result == nil {
		return "", errors.New(errors.ErrNoInstance, "no records found")
	}

	for _, securityConfig := range result {
		enabled = securityConfig.GetChildContentS("enabled")
	}

	return enabled, nil
}
