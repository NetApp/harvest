/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package certificate

import (
	"crypto/x509"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/matrix"
)

type Certificate struct {
	*plugin.AbstractPlugin
	data                 *matrix.Matrix
	instanceKeys         map[string]string
	instanceLabels       map[string]*dict.Dict
	pluginInvocationRate int
	currentVal           int
	client               *rest.Client
	query                string
	snapmirrorFields     []string
	outgoingSM           map[string][]string
	incomingSM           map[string]string
	isHealthySM          map[string]bool
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Certificate{AbstractPlugin: p}
}

// [TODO] https://opengrok-prd.eng.netapp.com/source/xref/bard_main/modules/dfm-data-access/src/main/java/com/netapp/dfm/entity/platform/event/builtin/ClusterSelfSignedCertificateCheckListener.java#67
func (my *Certificate) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	// Purge and reset data
	my.data.PurgeInstances()
	my.data.Reset()

	// Set all global labels from zapi.go if already not exist
	my.data.SetGlobalLabels(data.GetGlobalLabels())

	my.setCertificateIssuerType(data)
	return nil, nil
}

func (my *Certificate) setCertificateIssuerType(data *matrix.Matrix) {
	var (
		certificatePEM string
		cert           *x509.Certificate
		err            error
	)

	for _, instance := range data.GetInstances() {
		certificatePEM = instance.GetLabel("certificatePEM")
		my.Logger.Info().Msgf("%s", certificatePEM)

		if certificatePEM == "" {
			instance.SetLabel("certificateIssuerType", "unknown")
		} else {
			instance.SetLabel("certificateIssuerType", "self_signed")
			if cert, err = x509.ParseCertificate([]byte(certificatePEM)); err != nil {
				my.Logger.Error().Err(err).Msg("error")
				instance.SetLabel("certificateIssuerType", "unknown")
			}
			my.Logger.Info().Msgf("%v", cert)
			if cert.Subject.String() == cert.Issuer.String() {
				//opts := x509.VerifyOptions{}
				//cert.VerifyHostname(cert.RawTBSCertificate)
				//rsa.VerifyPKCS1v15()
				if err = cert.CheckSignature(x509.SHA1WithRSA, cert.RawTBSCertificate, cert.Signature); err != nil {
					my.Logger.Error().Err(err).Msg("error")
					instance.SetLabel("certificateIssuerType", "unknown")
				}
				instance.SetLabel("certificateIssuerType", "ca_signed")
			}
		}

	}

}
