/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package certificate

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"strconv"
	"time"
)

const DefaultPluginDuration = 30 * time.Minute
const DefaultDataPollDuration = 3 * time.Minute
const BatchSize = "500"

type Certificate struct {
	*plugin.AbstractPlugin
	pluginInvocationRate int
	currentVal           int
	batchSize            string
	client               *zapi.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Certificate{AbstractPlugin: p}
}

func (my *Certificate) Init() error {

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

	if b := my.Params.GetChildContentS("batch_size"); b != "" {
		if _, err := strconv.Atoi(b); err == nil {
			my.batchSize = b
		}
	} else {
		my.batchSize = BatchSize
	}

	return nil
}

func (my *Certificate) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		adminVserver       string
		adminVserverSerial string
		err                error
	)

	if my.currentVal >= my.pluginInvocationRate {
		my.currentVal = 0

		// invoke vserver-get-iter zapi and get admin vserver name
		if adminVserver, err = my.GetAdminVserver(); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect admin SVM")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect admin SVM")
			}
			return nil, nil
		}

		// invoke security-ssl-get-iter zapi and get admin vserver's serial number
		if adminVserverSerial, err = my.GetSecuritySsl(adminVserver); err != nil {
			if errors.Is(err, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err).Msg("Failed to collect admin SVM's serial number")
			} else {
				my.Logger.Error().Err(err).Msg("Failed to collect admin SVM's serial number")
			}
			return nil, nil
		}

		// update certificate instance based on admin vaserver serial
		for _, certificateInstance := range data.GetInstances() {
			if certificateInstance.IsExportable() {
				certificateInstance.SetExportable(false)
				serialNumber := certificateInstance.GetLabel("serial_number")

				if serialNumber == adminVserverSerial {
					certificateInstance.SetExportable(true)
					my.setCertificateIssuerType(certificateInstance)
					my.setCertificateValidity(data, certificateInstance)
				}
			}
		}

	}

	my.currentVal++
	return nil, nil
}

func (my *Certificate) setCertificateIssuerType(instance *matrix.Instance) {
	var (
		cert *x509.Certificate
		err  error
	)

	certificatePEM := instance.GetLabel("certificatePEM")
	certUUID := instance.GetLabel("uuid")

	if certificatePEM == "" {
		my.Logger.Debug().Str("uuid", certUUID).Msg("Certificate is not found")
		instance.SetLabel("certificateIssuerType", "unknown")
	} else {
		instance.SetLabel("certificateIssuerType", "self_signed")
		certDecoded, _ := pem.Decode([]byte(certificatePEM))
		if certDecoded == nil {
			my.Logger.Warn().Msg("PEM formatted object is not a X.509 certificate. Only PEM formatted X.509 certificate input is allowed")
			instance.SetLabel("certificateIssuerType", "unknown")
			return
		}

		if cert, err = x509.ParseCertificate(certDecoded.Bytes); err != nil {
			my.Logger.Warn().Msg("PEM formatted object is not a X.509 certificate. Only PEM formatted X.509 certificate input is allowed")
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
		}
	}
}

func (my *Certificate) setCertificateValidity(data *matrix.Matrix, instance *matrix.Instance) {
	var (
		expiryTimeMetric matrix.Metric
	)

	instance.SetLabel("certificateExpiryStatus", "unknown")

	if expiryTimeMetric = data.GetMetric("certificate-info.expiration-date"); expiryTimeMetric == nil {
		my.Logger.Error().Msg("missing expiry time metric")
		return
	}

	if expiryTime, ok := expiryTimeMetric.GetValueFloat64(instance); ok {
		// convert expiryTime from float64 to int64 and find difference
		timestampDiff := time.Until(time.Unix(int64(expiryTime), 0)).Hours()

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

}

func (my *Certificate) GetAdminVserver() (string, error) {
	var (
		result       []*node.Node
		request      *node.Node
		err          error
		adminVserver string
	)

	request = node.NewXMLS("vserver-get-iter")
	request.NewChildS("max-records", my.batchSize)
	// Fetching only admin vserver
	query := request.NewChildS("query", "")
	vserverInfo := query.NewChildS("vserver-info", "")
	vserverInfo.NewChildS("vserver-type", "admin")

	if result, _, err = collectors.InvokeZapiCall(my.client, request, my.Logger, ""); err != nil {
		return "", err
	}

	if len(result) == 0 || result == nil {
		return "", errs.New(errs.ErrNoInstance, "no records found")
	}
	// This should be one iteration only as cluster can have one admin vserver
	for _, svm := range result {
		adminVserver = svm.GetChildContentS("vserver-name")
	}
	return adminVserver, nil
}

func (my *Certificate) GetSecuritySsl(adminSvm string) (string, error) {
	var (
		result            []*node.Node
		request           *node.Node
		err               error
		certificateSerial string
	)

	request = node.NewXMLS("security-ssl-get-iter")
	request.NewChildS("max-records", my.batchSize)
	// Fetching only admin vserver
	query := request.NewChildS("query", "")
	vserverInfo := query.NewChildS("vserver-ssl-info", "")
	vserverInfo.NewChildS("vserver", adminSvm)

	if result, _, err = collectors.InvokeZapiCall(my.client, request, my.Logger, ""); err != nil {
		return "", err
	}

	if len(result) == 0 || result == nil {
		return "", errs.New(errs.ErrNoInstance, "no records found")
	}
	// This should be one iteration only as cluster can have one admin vserver
	for _, ssl := range result {
		certificateSerial = ssl.GetChildContentS("certificate-serial-number")
	}
	return certificateSerial, nil
}
