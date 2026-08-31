/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package certificate

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	ontap "github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"time"
)

type Certificate struct {
	*plugin.AbstractPlugin
	currentVal         int
	client             *rest.Client
	adminVserverSerial string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Certificate{AbstractPlugin: p}
}

func (c *Certificate) Init(remote conf.Remote) error {

	var err error

	if err := c.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if c.client, err = rest.New(conf.ZapiPoller(c.ParentParams), timeout, c.Auth); err != nil {
		c.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if _, err := c.client.Init(5, remote); err != nil {
		return err
	}

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	c.currentVal = c.SetPluginInterval()

	return nil
}

func (c *Certificate) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {

	var (
		expiryTimeMetric *matrix.Metric
		unixTime         time.Time
	)
	data := dataMap[c.Object]
	c.RequestMetadata.Reset()

	if c.currentVal >= c.PluginInvocationRate {
		c.currentVal = 0
		c.adminVserverSerial = ""
		c.refreshAdminSerial()
	}

	// update certificate instance based on admin vserver serial
	for _, certificateInstance := range data.GetInstances() {
		if !certificateInstance.IsExportable() {
			continue
		}
		serialNumber := certificateInstance.GetLabel("serial_number")
		certType := certificateInstance.GetLabel("type")

		if expiryTimeMetric = data.GetMetric("expiration"); expiryTimeMetric == nil {
			c.SLogger.Error("missing expiry time metric")
			continue
		}

		if expiryTime, ok := expiryTimeMetric.GetValueFloat64(certificateInstance); ok {
			// convert expiryTime from float64 to int64 and then to unix Time
			unixTime = time.Unix(int64(expiryTime), 0)
			certificateInstance.SetLabel("expiry_time", unixTime.UTC().Format(time.RFC3339))
		} else {
			// This is fail-safe case
			unixTime = time.Now()
		}

		if c.adminVserverSerial != "" && serialNumber == c.adminVserverSerial && certType == "server" {
			c.setCertificateIssuerType(certificateInstance)
			c.setCertificateValidity(unixTime, certificateInstance)
		}
	}

	c.currentVal++
	return nil, &c.RequestMetadata, nil
}

func (c *Certificate) refreshAdminSerial() {
	adminVserver, err := c.GetAdminVserver()
	if err != nil {
		if ontap.IsRestErr(err, ontap.APINotFound) {
			c.SLogger.Debug("Failed to collect admin SVM", slogx.Err(err))
		} else {
			c.SLogger.Error("Failed to collect admin SVM", slogx.Err(err))
		}
		return
	}
	if adminVserver == "" {
		c.SLogger.Error("Admin SVM is missing in the cluster")
		return
	}
	// invoke private ssl cli rest and get the admin SVM's serial number
	serial, err := c.GetSecuritySsl(adminVserver)
	if err != nil {
		if ontap.IsRestErr(err, ontap.APINotFound) {
			c.SLogger.Debug("Failed to collect admin SVM's serial number", slogx.Err(err))
		} else {
			c.SLogger.Error("Failed to collect admin SVM's serial number", slogx.Err(err))
		}
		return
	}
	c.adminVserverSerial = serial
}

func (c *Certificate) setCertificateIssuerType(instance *matrix.Instance) {
	certificatePEM := instance.GetLabel("certificatePEM")
	certUUID := instance.GetLabel("uuid")

	if certificatePEM == "" {
		c.SLogger.Debug("Certificate PEM is not found", slog.String("uuid", certUUID))
		instance.SetLabel("certificateIssuerType", "unknown")
		return
	}

	issuerType, err := collectors.CertificateIssuerType(certificatePEM)
	if err != nil {
		c.SLogger.Warn(
			"PEM formatted object is not an X.509 certificate. Only PEM formatted X.509 certificate input is allowed",
			slogx.Err(err),
		)
	}
	instance.SetLabel("certificateIssuerType", issuerType)
}

func (c *Certificate) setCertificateValidity(unixTime time.Time, instance *matrix.Instance) {
	instance.SetLabel("certificateExpiryStatus", collectors.CertificateExpiryStatus(unixTime, time.Now()))
}

func (c *Certificate) GetAdminVserver() (string, error) {

	var (
		result       []gjson.Result
		err          error
		adminVserver string
	)

	query := "api/private/cli/vserver"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields([]string{"type"}).
		MaxRecords(collectors.DefaultBatchSize).
		Filter([]string{"type=admin"}).
		Build()

	if result, err = collectors.InvokeRestCall(c.client, href); err != nil {
		return "", err
	}

	// This should be one iteration only as cluster can have one admin vserver
	for _, svm := range result {
		adminVserver = svm.Get("vserver").ClonedString()
	}
	return adminVserver, nil
}

func (c *Certificate) GetSecuritySsl(adminSvm string) (string, error) {

	var (
		result      []gjson.Result
		err         error
		adminSerial string
	)

	query := "api/private/cli/security/ssl"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields([]string{"serial"}).
		MaxRecords(collectors.DefaultBatchSize).
		Filter([]string{"vserver=" + adminSvm}).
		Build()

	if result, err = collectors.InvokeRestCall(c.client, href); err != nil {
		return "", err
	}

	// This should be one iteration only as cluster can have one admin vserver
	for _, ssl := range result {
		adminSerial = ssl.Get("serial").ClonedString()
	}

	return adminSerial, nil
}
