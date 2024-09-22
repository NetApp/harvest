/*
 * Copyright NetApp Inc, 2023 All rights reserved
 */

package securityaccount

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"log/slog"
	"time"
)

type SecurityAccount struct {
	*plugin.AbstractPlugin
	client *rest.Client
	query  string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SecurityAccount{AbstractPlugin: p}
}

func (s *SecurityAccount) Init() error {
	var err error

	if err := s.InitAbc(); err != nil {
		return err
	}

	clientTimeout := s.ParentParams.GetChildContentS("client_timeout")
	if clientTimeout == "" {
		clientTimeout = rest.DefaultTimeout
	}
	timeout, err := time.ParseDuration(clientTimeout)
	if err != nil {
		s.SLogger.Info("Using default timeout", slog.String("timeout", rest.DefaultTimeout))
	}
	if s.client, err = rest.New(conf.ZapiPoller(s.ParentParams), timeout, s.Auth); err != nil {
		return fmt.Errorf("failed to connect err=%w", err)
	}

	if err := s.client.Init(5); err != nil {
		return err
	}

	s.query = "api/security/accounts"
	return nil
}

func (s *SecurityAccount) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	var (
		result                 []gjson.Result
		err                    error
		applicationToMethodMap map[string][]string
	)

	data := dataMap[s.Object]
	href := rest.NewHrefBuilder().
		APIPath(s.query).
		Fields([]string{"applications"}).
		Build()

	s.client.Metadata.Reset()
	if result, err = collectors.InvokeRestCall(s.client, href, s.SLogger); err != nil {
		return nil, nil, err
	}

	for _, securityAccount := range result {
		username := securityAccount.Get("name").String()
		svm := securityAccount.Get("owner.name").String()

		if !securityAccount.IsObject() {
			return nil, nil, errs.New(errs.ErrNoInstance, "security account is not an object. type="+securityAccount.Type.String())
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

		securityAccountKey := svm + username
		if securityAccountInstance := data.GetInstance(securityAccountKey); securityAccountInstance != nil {
			securityAccountInstance.SetExportable(false)

			for application, methods := range applicationToMethodMap {
				for _, method := range methods {
					var securityAccountNewInstance *matrix.Instance
					securityAccountNewKey := securityAccountKey + application + method
					if securityAccountNewInstance, err = data.NewInstance(securityAccountNewKey); err != nil {
						return nil, nil, fmt.Errorf("add instance failed for instance key %s err: %w", securityAccountNewKey, err)
					}

					for k, v := range securityAccountInstance.GetLabels() {
						securityAccountNewInstance.SetLabel(k, v)
					}
					securityAccountNewInstance.SetLabel("applications", application)
					securityAccountNewInstance.SetLabel("methods", method)
				}
			}
		}
	}

	return []*matrix.Matrix{data}, s.client.Metadata, nil
}
