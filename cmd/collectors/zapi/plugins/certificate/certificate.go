/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package certificate

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
	"strconv"
	"time"
)

const BatchSize = "500"

type Certificate struct {
	*plugin.AbstractPlugin
	currentVal int
	batchSize  string
	client     *zapi.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Certificate{AbstractPlugin: p}
}

func (c *Certificate) Init(remote conf.Remote) error {

	var err error

	if err := c.InitAbc(); err != nil {
		return err
	}

	if c.client, err = zapi.New(conf.ZapiPoller(c.ParentParams), c.Auth); err != nil {
		c.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := c.client.Init(5, remote); err != nil {
		return err
	}

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	c.currentVal = c.SetPluginInterval()

	c.batchSize = BatchSize
	if b := c.Params.GetChildContentS("batch_size"); b != "" {
		if _, err := strconv.Atoi(b); err == nil {
			c.batchSize = b
		}
	}

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

		// invoke vserver-get-iter zapi and get admin vserver name
		if adminVserver, err = c.GetAdminVserver(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				c.SLogger.Debug("Failed to collect admin SVM", slogx.Err(err))
			} else {
				c.SLogger.Error("Failed to collect admin SVM", slogx.Err(err))
			}
			return nil, nil, nil
		}

		// invoke security-ssl-get-iter zapi and get admin vserver's serial number
		if adminVserverSerial, err = c.GetSecuritySsl(adminVserver); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				c.SLogger.Debug("Failed to collect admin SVM's serial number", slogx.Err(err))
			} else {
				c.SLogger.Error("Failed to collect admin SVM's serial number", slogx.Err(err))
			}
			return nil, nil, nil
		}

		// update certificate instance based on admin vserver serial
		for certificateInstanceKey, certificateInstance := range data.GetInstances() {
			if !certificateInstance.IsExportable() {
				continue
			}
			name := certificateInstance.GetLabel("name")
			serialNumber := certificateInstance.GetLabel("serial_number")
			svm := certificateInstance.GetLabel("svm")
			certType := certificateInstance.GetLabel("type")
			certificateInstance.SetLabel("uuid", name+serialNumber+svm)

			if expiryTimeMetric = data.GetMetric("certificate-info.expiration-date"); expiryTimeMetric == nil {
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
				c.setCertificateIssuerType(certificateInstance, certificateInstanceKey)
				c.setCertificateValidity(unixTime, certificateInstance)
			}
		}
	}

	c.currentVal++
	return nil, c.client.Metadata, nil
}

func (c *Certificate) setCertificateIssuerType(instance *matrix.Instance, certificateInstanceKey string) {
	var (
		cert *x509.Certificate
		err  error
	)

	certificatePEM := instance.GetLabel("certificatePEM")

	if certificatePEM == "" {
		c.SLogger.Debug("Certificate is not found", slog.String("certificateInstanceKey", certificateInstanceKey))
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
			c.SLogger.Warn("PEM formatted object is not an X.509 certificate. Only PEM formatted X.509 certificate input is allowed", slogx.Err(err))
			instance.SetLabel("certificateIssuerType", "unknown")
			return
		}

		// Verifies if certificate is self-issued. This is true if the subject and issuer are equal.
		if cert.Subject.String() == cert.Issuer.String() {
			// Verifies if the certificate is self-signed. This is true if the certificate is signed using its own public key.
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
		result       []*node.Node
		request      *node.Node
		err          error
		adminVserver string
	)

	request = node.NewXMLS("vserver-get-iter")
	request.NewChildS("max-records", c.batchSize)
	// Fetching only admin vserver
	query := request.NewChildS("query", "")
	vserverInfo := query.NewChildS("vserver-info", "")
	vserverInfo.NewChildS("vserver-type", "admin")

	// Fetching only admin SVMs
	if result, err = c.client.InvokeZapiCall(request); err != nil {
		return "", err
	}

	if len(result) == 0 || result == nil {
		return "", errs.New(errs.ErrNoInstance, "no records found")
	}
	// This should be one iteration only as cluster can have one admin vserver
	for _, svm := range result {
		adminVserver = svm.GetChildContentS("vserver-name")
		break
	}
	return adminVserver, nil
}

func (c *Certificate) GetSecuritySsl(adminSvm string) (string, error) {
	var (
		result            []*node.Node
		request           *node.Node
		err               error
		certificateSerial string
	)

	request = node.NewXMLS("security-ssl-get-iter")
	request.NewChildS("max-records", c.batchSize)
	// Fetching only admin vserver
	query := request.NewChildS("query", "")
	vserverInfo := query.NewChildS("vserver-ssl-info", "")
	vserverInfo.NewChildS("vserver", adminSvm)

	// fetching data of only admin vservers
	if result, err = c.client.InvokeZapiCall(request); err != nil {
		return "", err
	}

	if len(result) == 0 || result == nil {
		return "", errs.New(errs.ErrNoInstance, "no records found")
	}
	// This should be one iteration only as cluster can have one admin vserver
	for _, ssl := range result {
		certificateSerial = ssl.GetChildContentS("certificate-serial-number")
		break
	}
	return certificateSerial, nil
}
