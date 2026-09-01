package collectors

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"
)

// CertificateIssuerType classifies a PEM certificate as self-signed or CA-signed.
func CertificateIssuerType(certificatePEM string) (string, error) {
	if certificatePEM == "" {
		return "unknown", nil
	}

	decoded, _ := pem.Decode([]byte(certificatePEM))
	if decoded == nil {
		return "unknown", errors.New("PEM formatted object is not an X.509 certificate")
	}
	cert, err := x509.ParseCertificate(decoded.Bytes)
	if err != nil {
		return "unknown", err
	}

	if cert.Subject.String() != cert.Issuer.String() {
		return "ca_signed", nil
	}
	if !hasValidSelfSignature(cert) {
		return "ca_signed", nil
	}
	return "self_signed", nil
}

func hasValidSelfSignature(cert *x509.Certificate) bool {
	return cert.CheckSignature(cert.SignatureAlgorithm, cert.RawTBSCertificate, cert.Signature) == nil
}

// CertificateExpiryStatus classifies a certificate expiry relative to now.
func CertificateExpiryStatus(expiry, now time.Time) string {
	hoursRemaining := expiry.Sub(now).Hours()
	if hoursRemaining <= 0 {
		return "expired"
	}
	if hoursRemaining/24 < 60 {
		return "expiring"
	}
	return "active"
}
