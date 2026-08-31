package collectors

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func TestCertificateIssuerType(t *testing.T) {
	certs := testCertificates(t)
	tests := []struct {
		name    string
		pem     string
		want    string
		wantErr bool
	}{
		{name: "empty", want: "unknown"},
		{name: "malformed PEM", pem: "not a certificate", want: "unknown", wantErr: true},
		{name: "self signed", pem: certs.selfSignedRSA, want: "self_signed"},
		{name: "self signed ECDSA", pem: certs.selfSignedECDSA, want: "self_signed"},
		{name: "self signed RSA SHA-384", pem: certs.selfSignedRSA384, want: "self_signed"},
		{name: "CA signed", pem: certs.caSigned, want: "ca_signed"},
		{name: "self issued but signed by another key", pem: certs.selfIssuedForeignSigned, want: "ca_signed"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := CertificateIssuerType(test.pem)
			if (err != nil) != test.wantErr {
				t.Fatalf("CertificateIssuerType() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Fatalf("CertificateIssuerType() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestCertificateExpiryStatus(t *testing.T) {
	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		expiry time.Time
		want   string
	}{
		{name: "expired", expiry: now.Add(-time.Second), want: "expired"},
		{name: "expires now", expiry: now, want: "expired"},
		{name: "under sixty days", expiry: now.Add(60*24*time.Hour - time.Second), want: "expiring"},
		{name: "sixty days", expiry: now.Add(60 * 24 * time.Hour), want: "active"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := CertificateExpiryStatus(test.expiry, now); got != test.want {
				t.Fatalf("CertificateExpiryStatus() = %q, want %q", got, test.want)
			}
		})
	}
}

type testCerts struct {
	selfSignedRSA           string
	selfSignedRSA384        string
	selfSignedECDSA         string
	caSigned                string
	selfIssuedForeignSigned string
}

func testCertificates(t *testing.T) testCerts {
	t.Helper()

	caKey := rsaKey(t)
	ca := certTemplate("Harvest Test CA", 1, x509.SHA256WithRSA, true)

	leafKey := rsaKey(t)
	leaf := certTemplate("Harvest Test Leaf", 2, x509.SHA256WithRSA, false)

	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ecSelfSigned := certTemplate("Harvest Test ECDSA", 3, x509.ECDSAWithSHA256, true)

	sha384Key := rsaKey(t)
	sha384SelfSigned := certTemplate("Harvest Test SHA384", 4, x509.SHA384WithRSA, true)

	// Subject equals Issuer, but the signature is made by rogueKey while the
	// certificate carries impostorKey's public key: self-issued, not self-signed.
	const sharedCN = "Harvest Test Impostor"
	impostorKey := rsaKey(t)
	impostor := certTemplate(sharedCN, 5, x509.SHA256WithRSA, false)
	rogueKey := rsaKey(t)
	rogueCA := certTemplate(sharedCN, 6, x509.SHA256WithRSA, true)

	return testCerts{
		selfSignedRSA:           issueCert(t, ca, nil, &caKey.PublicKey, caKey),
		selfSignedRSA384:        issueCert(t, sha384SelfSigned, nil, &sha384Key.PublicKey, sha384Key),
		selfSignedECDSA:         issueCert(t, ecSelfSigned, nil, &ecKey.PublicKey, ecKey),
		caSigned:                issueCert(t, leaf, ca, &leafKey.PublicKey, caKey),
		selfIssuedForeignSigned: issueCert(t, impostor, rogueCA, &impostorKey.PublicKey, rogueKey),
	}
}

// certTemplate returns a certificate template valid for the next hour.
func certTemplate(commonName string, serial int64, algo x509.SignatureAlgorithm, isCA bool) *x509.Certificate {
	tmpl := &x509.Certificate{
		SerialNumber:       big.NewInt(serial),
		Subject:            pkix.Name{CommonName: commonName},
		NotBefore:          time.Now().Add(-time.Hour),
		NotAfter:           time.Now().Add(time.Hour),
		KeyUsage:           x509.KeyUsageDigitalSignature,
		SignatureAlgorithm: algo,
	}
	if isCA {
		tmpl.KeyUsage |= x509.KeyUsageCertSign
		tmpl.BasicConstraintsValid = true
		tmpl.IsCA = true
	}
	return tmpl
}

// issueCert signs tmpl with signerKey and PEM encodes the result. A nil parent
// self-issues the certificate, making its Subject and Issuer identical.
func issueCert(t *testing.T, tmpl, parent *x509.Certificate, pub any, signerKey crypto.Signer) string {
	t.Helper()
	if parent == nil {
		parent = tmpl
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, parent, pub, signerKey)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func rsaKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return key
}
