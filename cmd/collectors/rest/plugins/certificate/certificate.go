/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package certificate

import (
	"crypto/x509"
	"encoding/pem"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	ontap "github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"time"
)

type Certificate struct {
	*plugin.AbstractPlugin
	currentVal int
	client     *rest.Client
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

	if err := c.client.Init(5, remote); err != nil {
		return err
	}

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	c.currentVal = c.SetPluginInterval()

	return nil
}

func (c *Certificate) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

	var (
		adminVserver       string
		adminVserverSerial string
		expiryTimeMetric   *matrix.Metric
		unixTime           time.Time
		err                error
	)
	data := dataMap[c.Object]
	c.client.Metadata.Reset()

	if c.currentVal >= c.PluginInvocationRate {
		c.currentVal = 0

		// invoke private vserver cli rest and get admin vserver name
		if adminVserver, err = c.GetAdminVserver(); err != nil {
			if ontap.IsRestErr(err, ontap.APINotFound) {
				c.SLogger.Debug("Failed to collect admin SVM", slogx.Err(err))
			} else {
				c.SLogger.Error("Failed to collect admin SVM", slogx.Err(err))
			}
			return nil, nil, nil
		}

		// invoke private ssl cli rest when admin SVM is non-empty and get the admin SVM's serial number
		if adminVserver == "" {
			c.SLogger.Error("Admin SVM is missing in the cluster")
			return nil, nil, nil
		}

		if adminVserverSerial, err = c.GetSecuritySsl(adminVserver); err != nil {
			if ontap.IsRestErr(err, ontap.APINotFound) {
				c.SLogger.Debug("Failed to collect admin SVM's serial number", slogx.Err(err))
			} else {
				c.SLogger.Error("Failed to collect admin SVM's serial number", slogx.Err(err))
			}
			return nil, nil, nil
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

			if serialNumber == adminVserverSerial && certType == "server" {
				c.setCertificateIssuerType(certificateInstance)
				c.setCertificateValidity(unixTime, certificateInstance)
			}
		}
	}

	c.currentVal++
	return nil, c.client.Metadata, nil
}

func (c *Certificate) setCertificateIssuerType(instance *matrix.Instance) {
	var (
		cert *x509.Certificate
		err  error
	)

	certificatePEM := instance.GetLabel("certificatePEM")
	certUUID := instance.GetLabel("uuid")

	if certificatePEM == "" {
		c.SLogger.Debug("Certificate PEM is not found", slog.String("uuid", certUUID))
		instance.SetLabel("certificateIssuerType", "unknown")
	} else {
		instance.SetLabel("certificateIssuerType", "self_signed")
		certDecoded, _ := pem.Decode([]byte(certificatePEM))
		if certDecoded == nil {
			c.SLogger.Warn("PEM formatted object is not an X.509 certificate. Only PEM formatted X.509 certificate input is allowed")
			instance.SetLabel("certificateIssuerType", "unknown")
			return
		}

		if cert, err = x509.ParseCertificate(certDecoded.Bytes); err != nil {
			c.SLogger.Warn(
				"PEM formatted object is not an X.509 certificate. Only PEM formatted X.509 certificate input is allowed",
				slogx.Err(err),
			)
			instance.SetLabel("certificateIssuerType", "unknown")
			return
		}

		// Verifies if certificate is self-issued. This is true if the subject and issuer are equal.
		if cert.Subject.String() == cert.Issuer.String() {
			// Verifies if certificate is self-signed. This is true if the certificate is signed using its own public key.
			if err = cert.CheckSignature(x509.SHA256WithRSA, cert.RawTBSCertificate, cert.Signature); err != nil {
				// Any verification exception means it is not signed with the give key. i.e. not self-signed
				instance.SetLabel("certificateIssuerType", "ca_signed")
			}
		} else {
			instance.SetLabel("certificateIssuerType", "ca_signed")
		}
	}
}

func (c *Certificate) setCertificateValidity(unixTime time.Time, instance *matrix.Instance) {
	instance.SetLabel("certificateExpiryStatus", "unknown")

	// find difference from unix Time
	timestampDiff := time.Until(unixTime).Hours()

	if timestampDiff <= 0 {
		instance.SetLabel("certificateExpiryStatus", "expired")
	} else {
		// daysRemaining will be more than 0 if it has reached this point, convert to days
		daysRemaining := timestampDiff / 24
		if daysRemaining < 60 {
			instance.SetLabel("certificateExpiryStatus", "expiring")
		} else {
			instance.SetLabel("certificateExpiryStatus", "active")
		}
	}
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
