/*
 * Copyright NetApp Inc, 2023 All rights reserved
 */

package ontaps3service

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
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

func (o *OntapS3Service) Init(remote conf.Remote) error {
	var err error

	if err := o.InitAbc(); err != nil {
		return err
	}

	clientTimeout := o.ParentParams.GetChildContentS("client_timeout")
	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	duration, err := time.ParseDuration(clientTimeout)
	if err == nil {
		timeout = duration
	} else {
		o.SLogger.Debug("Using default timeout", slog.String("timeout", timeout.String()))
	}
	if o.client, err = rest.New(conf.ZapiPoller(o.ParentParams), timeout, o.Auth); err != nil {
		o.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := o.client.Init(5, remote); err != nil {
		return err
	}

	o.query = "api/protocols/s3/services"
	return nil
}

func (o *OntapS3Service) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	var (
		result      []gjson.Result
		err         error
		svmToURLMap map[string][]string
		urlValue    []string
	)

	// reset svmToS3serverMap map
	svmToURLMap = make(map[string][]string)
	data := dataMap[o.Object]
	o.client.Metadata.Reset()

	fields := []string{"svm.name", "name", "is_http_enabled", "is_https_enabled", "secure_port", "port"}
	href := rest.NewHrefBuilder().
		APIPath(o.query).
		Fields(fields).
		MaxRecords(collectors.DefaultBatchSize).
		Build()

	if result, err = collectors.InvokeRestCall(o.client, href); err != nil {
		return nil, nil, err
	}

	// Iterate over services API response
	for _, ontaps3Service := range result {
		if !ontaps3Service.IsObject() {
			o.SLogger.Error("Ontap S3 Service is not an object, skipping", slog.String("type", ontaps3Service.Type.String()))
			return nil, nil, errs.New(errs.ErrNoInstance, "Ontap S3 Service is not an object")
		}
		s3ServerName := ontaps3Service.Get("name").ClonedString()
		svm := ontaps3Service.Get("svm.name").ClonedString()
		isHTTPEnabled := ontaps3Service.Get("is_http_enabled").Bool()
		isHTTPSEnabled := ontaps3Service.Get("is_https_enabled").Bool()
		securePort := ontaps3Service.Get("secure_port").ClonedString()
		port := ontaps3Service.Get("port").ClonedString()

		httpsURL := ""
		httpURL := ""
		if isHTTPSEnabled {
			httpsURL = "https://" + s3ServerName
			if securePort != "443" {
				httpsURL += ":" + securePort
			}
			httpsURL += "/"
		}
		if isHTTPEnabled {
			httpURL = "http://" + s3ServerName
			if port != "80" {
				httpURL += ":" + port
			}
			httpURL += "/"
		}

		// cache url and svm detail
		svmToURLMap[svm] = []string{httpsURL, httpURL}
	}

	for _, ontapS3 := range data.GetInstances() {
		urlValue = make([]string, 0)
		for _, url := range svmToURLMap[ontapS3.GetLabel("svm")] {
			if url != "" {
				urlValue = append(urlValue, url+ontapS3.GetLabel("bucket"))
			}
		}
		// Update url label in ontaps3_labels, Example: http://s3server/bucket1
		// If http and https both are enabled, then url label in ontaps3_labels, https://s3server/bucket1, http://s3server/bucket1
		ontapS3.SetLabel("url", strings.Join(urlValue, ","))
	}

	return nil, o.client.Metadata, nil
}
