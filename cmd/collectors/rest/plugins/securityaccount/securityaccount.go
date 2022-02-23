/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package securityaccount

import (
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/matrix"
	"strconv"
	"strings"
)

type SecurityAccount struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SecurityAccount{AbstractPlugin: p}
}

func (my *SecurityAccount) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	my.setSecurityAccountLoginMethod(data)
	return nil, nil
}

func (my *SecurityAccount) setSecurityAccountLoginMethod(data *matrix.Matrix) {
	for _, instance := range data.GetInstances() {
		methods := instance.GetLabel("methods")
		my.Logger.Info().Msgf("%s", methods)

		instance.SetLabel("samluser", strconv.FormatBool(strings.Contains(methods, "saml")))
		instance.SetLabel("ldapuser", strconv.FormatBool(strings.Contains(methods, "nsswitch")))
		instance.SetLabel("certificateuser", strconv.FormatBool(strings.Contains(methods, "certificate")))
		instance.SetLabel("localuser", strconv.FormatBool(strings.Contains(methods, "password")))
		instance.SetLabel("activediruser", strconv.FormatBool(strings.Contains(methods, "domain")))
	}

}
