/*
 * Copyright NetApp Inc, 2023 All rights reserved
 */

package securityaccount

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"time"
)

type SecurityAccount struct {
	*plugin.AbstractPlugin
	data   *matrix.Matrix
	client *rest.Client
	query  string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SecurityAccount{AbstractPlugin: p}
}

func (s *SecurityAccount) Init() error {
	var err error

	if err = s.InitAbc(); err != nil {
		return err
	}

	clientTimeout := s.ParentParams.GetChildContentS("client_timeout")
	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	duration, err := time.ParseDuration(clientTimeout)
	if err == nil {
		timeout = duration
	} else {
		s.Logger.Info().Str("timeout", timeout.String()).Msg("Using default timeout")
	}
	if s.client, err = rest.New(conf.ZapiPoller(s.ParentParams), timeout); err != nil {
		s.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = s.client.Init(5); err != nil {
		return err
	}

	s.query = "api/security/accounts"
	s.data = matrix.New(s.Parent+".SecurityAccount", "security_account", "security_account")

	exportOptions := node.NewS("export_options")
	instanceLabels := exportOptions.NewChildS("instance_labels", "")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")

	if exportOption := s.ParentParams.GetChildS("export_options"); exportOption != nil {
		if exportedLabels := exportOption.GetChildS("instance_labels"); exportedLabels != nil {
			for _, label := range exportedLabels.GetAllChildContentS() {
				instanceLabels.NewChildS("", label)
			}
		}
		if exportedKeys := exportOption.GetChildS("instance_keys"); exportedKeys != nil {
			for _, key := range exportedKeys.GetAllChildContentS() {
				instanceKeys.NewChildS("", key)
			}
		}
	}
	s.data.SetExportOptions(exportOptions)
	return nil
}

func (s *SecurityAccount) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		result                 []gjson.Result
		err                    error
		applicationToMethodMap map[string][]string
	)

	// Purge and reset data
	s.data.PurgeInstances()
	s.data.Reset()

	// Set all global labels from Rest.go if already not exist
	s.data.SetGlobalLabels(data.GetGlobalLabels())

	href := rest.BuildHref("", "*", nil, "", "", "", "", s.query)

	if result, err = collectors.InvokeRestCall(s.client, href, s.Logger); err != nil {
		return nil, err
	}

	for _, securityAccount := range result {
		username := securityAccount.Get("name").String()
		svm := securityAccount.Get("owner.name").String()

		if !securityAccount.IsObject() {
			s.Logger.Error().Str("type", securityAccount.Type.String()).Msg("Security Account is not an object, skipping")
			return nil, errs.New(errs.ErrNoInstance, "security account is not an object")
		}

		// reset applicationToMethodMap map
		applicationToMethodMap = make(map[string][]string)
		// Parse application object and cache the details
		if applications := securityAccount.Get("applications"); applications.Exists() {
			for _, applicationDetail := range applications.Array() {
				application := applicationDetail.Get("application").String()
				if methodList := applicationDetail.Get("authentication_methods"); methodList.Exists() {
					for _, method := range applicationDetail.Get("authentication_methods").Array() {
						applicationToMethodMap[application] = append(applicationToMethodMap[application], method.String())
					}
				}
			}
		}

		securityAccountKey := username + svm
		if securityAccountInstance := data.GetInstance(securityAccountKey); securityAccountInstance != nil {
			securityAccountInstance.SetExportable(false)

			for application, methods := range applicationToMethodMap {
				for _, method := range methods {
					var securityAccountNewInstance *matrix.Instance
					securityAccountNewKey := securityAccountKey + application + method
					if securityAccountNewInstance, err = s.data.NewInstance(securityAccountNewKey); err != nil {
						s.Logger.Debug().Msgf("add instance failed for key %s: %v", securityAccountNewKey, err)
						s.Logger.Error().Err(err).Str("Instance key", securityAccountNewKey).Msg("")
						return nil, err
					}

					for k, v := range securityAccountInstance.GetLabels().Map() {
						securityAccountNewInstance.SetLabel(k, v)
					}
					securityAccountNewInstance.SetLabel("applications", application)
					securityAccountNewInstance.SetLabel("methods", method)
				}
			}
		}
	}

	return []*matrix.Matrix{s.data}, nil
}
