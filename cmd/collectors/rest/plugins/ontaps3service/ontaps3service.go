/*
 * Copyright NetApp Inc, 2023 All rights reserved
 */

package ontaps3service

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

type OntapS3Service struct {
	*plugin.AbstractPlugin
	client *rest.Client
	query  string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &OntapS3Service{AbstractPlugin: p}
}

func (o *OntapS3Service) Init() error {
	var err error

	if err = o.InitAbc(); err != nil {
		return err
	}

	clientTimeout := o.ParentParams.GetChildContentS("client_timeout")
	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	duration, err := time.ParseDuration(clientTimeout)
	if err == nil {
		timeout = duration
	} else {
		o.Logger.Info().Str("timeout", timeout.String()).Msg("Using default timeout")
	}
	if o.client, err = rest.New(conf.ZapiPoller(o.ParentParams), timeout, o.Auth); err != nil {
		o.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = o.client.Init(5); err != nil {
		return err
	}

	o.query = "api/protocols/s3/services"
	return nil
}

func (o *OntapS3Service) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		result      []gjson.Result
		url         string
		err         error
		svmToUrlMap map[string]string
	)

	// reset svmToS3serverMap map
	svmToUrlMap = make(map[string]string)
	data := dataMap[o.Object]

	fields := []string{"svm.name", "name", "is_http_enabled", "is_https_enabled", "secure_port", "port"}
	href := rest.BuildHref("", strings.Join(fields, ","), nil, "", "", "", "", o.query)

	if result, err = collectors.InvokeRestCall(o.client, href, o.Logger); err != nil {
		return nil, err
	}

	for _, ontaps3Service := range result {
		if !ontaps3Service.IsObject() {
			o.Logger.Error().Str("type", ontaps3Service.Type.String()).Msg("Ontap S3 Service is not an object, skipping")
			return nil, errs.New(errs.ErrNoInstance, "Ontap S3 Service is not an object")
		}
		s3ServerName := ontaps3Service.Get("name").String()
		svm := ontaps3Service.Get("svm.name").String()
		isHttpEnabled := ontaps3Service.Get("is_http_enabled").Bool()
		isHttpsEnabled := ontaps3Service.Get("is_https_enabled").Bool()
		securePort := ontaps3Service.Get("secure_port").String()
		port := ontaps3Service.Get("port").String()

		if isHttpsEnabled {
			url = "https://" + s3ServerName
			if securePort != "443" {
				url += ":" + securePort
			}
		} else if isHttpEnabled {
			url = "http://" + s3ServerName
			if port != "80" {
				url += ":" + port
			}
		}
		url += "/"

		// cache url and svm detail
		svmToUrlMap[svm] = url
	}

	for _, ontapS3 := range data.GetInstances() {
		// Example: http://s3server/bucket1
		ontapS3.SetLabel("url", svmToUrlMap[ontapS3.GetLabel("svm")]+ontapS3.GetLabel("bucket"))
	}

	return nil, nil
}
