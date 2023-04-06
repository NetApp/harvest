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
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

type OntapS3Service struct {
	*plugin.AbstractPlugin
	data   *matrix.Matrix
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
		o.Logger.Debug().Str("timeout", timeout.String()).Msg("Using default timeout")
	}
	if o.client, err = rest.New(conf.ZapiPoller(o.ParentParams), timeout, o.Auth); err != nil {
		o.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = o.client.Init(5); err != nil {
		return err
	}

	o.query = "api/protocols/s3/services"
	// new metric would be ontaps3_services_labels
	metricName := "labels"
	o.data = matrix.New(".OntapS3", "ontaps3_services", "ontaps3_services")

	metric, err := o.data.NewMetricFloat64(metricName)
	if err != nil {
		o.Logger.Error().Err(err).Msg("add metric")
		return err
	}
	o.Logger.Trace().Msgf("added metric: (%s) %v", metricName, metric)

	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")

	// apply all instance keys from parent (ontap_s3.yaml) to all ontaps3_services metrics
	if exportOption := o.ParentParams.GetChildS("export_options"); exportOption != nil {
		//parent instancekeys would be added in plugin metrics
		if parentKeys := exportOption.GetChildS("instance_keys"); parentKeys != nil {
			for _, parentKey := range parentKeys.GetAllChildContentS() {
				instanceKeys.NewChildS("", parentKey)
			}
		}
	}

	instanceKeys.NewChildS("", "permission_type")
	instanceKeys.NewChildS("", "permissions")
	instanceKeys.NewChildS("", "user")
	instanceKeys.NewChildS("", "allowed_resources")

	o.data.SetExportOptions(exportOptions)

	return nil
}

func (o *OntapS3Service) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		result            []gjson.Result
		url               string
		err               error
		svmToURLMap       map[string]string
		bucketToPolicyMap map[string]map[string]string
	)

	// reset svmToS3serverMap map
	svmToURLMap = make(map[string]string)
	data := dataMap[o.Object]
	// Purge and reset data
	o.data.PurgeInstances()
	o.data.Reset()

	// Set all global labels from Rest.go if already not exist
	o.data.SetGlobalLabels(data.GetGlobalLabels())

	fields := []string{"svm.name", "name", "is_http_enabled", "is_https_enabled", "secure_port", "port", "buckets"}
	href := rest.BuildHref("", strings.Join(fields, ","), nil, "", "", "", "", o.query)

	if result, err = collectors.InvokeRestCall(o.client, href, o.Logger); err != nil {
		return nil, err
	}

	// Iterate over services API response
	for _, ontaps3Service := range result {
		if !ontaps3Service.IsObject() {
			o.Logger.Error().Str("type", ontaps3Service.Type.String()).Msg("Ontap S3 Service is not an object, skipping")
			return nil, errs.New(errs.ErrNoInstance, "Ontap S3 Service is not an object")
		}
		s3ServerName := ontaps3Service.Get("name").String()
		svm := ontaps3Service.Get("svm.name").String()
		isHTTPEnabled := ontaps3Service.Get("is_http_enabled").Bool()
		isHTTPSEnabled := ontaps3Service.Get("is_https_enabled").Bool()
		securePort := ontaps3Service.Get("secure_port").String()
		port := ontaps3Service.Get("port").String()

		if isHTTPSEnabled {
			url = "https://" + s3ServerName
			if securePort != "443" {
				url += ":" + securePort
			}
		} else if isHTTPEnabled {
			url = "http://" + s3ServerName
			if port != "80" {
				url += ":" + port
			}
		}
		url += "/"

		// cache url and svm detail
		svmToURLMap[svm] = url

		// Handle policy/permissions details
		if buckets := ontaps3Service.Get("buckets"); buckets.Exists() {
			for _, bucket := range buckets.Array() {
				// reset bucketToPolicyMap map
				bucketToPolicyMap = make(map[string]map[string]string)
				bucketUUID := bucket.Get("uuid").String()
				if policy := bucket.Get("policy"); policy.Exists() {
					if statements := policy.Get("statements"); statements.Exists() {
						for i, statement := range statements.Array() {
							policyMap := make(map[string]string)
							policyMap["permission_type"] = statement.Get("effect").String()
							policyMap["permissions"] = convertToString(statement.Get("actions").Array())
							policyMap["user"] = convertToString(statement.Get("principals").Array())
							policyMap["allowed_resources"] = convertToString(statement.Get("resources").Array())
							bucketToPolicyMap[bucketUUID+"-"+strconv.Itoa(i)] = policyMap
						}
					}
				}

				if ontapS3Instance := data.GetInstance(bucketUUID); ontapS3Instance != nil {
					var ontapS3NewInstance *matrix.Instance
					for bucketKey, policyMap := range bucketToPolicyMap {
						ontapS3NewKey := bucketKey
						if ontapS3NewInstance, err = o.data.NewInstance(ontapS3NewKey); err != nil {
							o.Logger.Error().Err(err).Str("add instance failed for instance key", ontapS3NewKey).Msg("")
							return nil, err
						}

						// Add existing labels
						for k, v := range ontapS3Instance.GetLabels().Map() {
							ontapS3NewInstance.SetLabel(k, v)
						}

						// Add policy/permission details
						for key, value := range policyMap {
							ontapS3NewInstance.SetLabel(key, value)
						}

						// Set metric value for instances. For now, only one metric is available
						for metricName, metric := range o.data.GetMetrics() {
							if err = metric.SetValueFloat64(ontapS3NewInstance, 1); err != nil {
								o.Logger.Error().Err(err).Str("metric", metricName).Msg("Unable to set value on metric")
							}
						}
					}
				}
			}
		}
	}

	for _, ontapS3 := range data.GetInstances() {
		// Update url label in ontaps3_labels, Example: http://s3server/bucket1
		ontapS3.SetLabel("url", svmToURLMap[ontapS3.GetLabel("svm")]+ontapS3.GetLabel("bucket"))
	}

	return []*matrix.Matrix{o.data}, nil
}

func convertToString(array []gjson.Result) string {
	var stringArray []string
	for _, value := range array {
		stringArray = append(stringArray, value.String())
	}
	return strings.Join(stringArray, ",")
}
