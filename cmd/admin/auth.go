package admin

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

func GenerateAdminCerts(opts *tlsOptions, flavor string) {
	certPath := fmt.Sprintf("%s-cert.pem", flavor)
	if _, err := os.Stat(certPath); !os.IsNotExist(err) {
		log.Fatalf("%s already exists. not overwriting\n", certPath)
	}
	keyPath := fmt.Sprintf("%s-key.pem", flavor)
	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		log.Fatalf("%s already exists. not overwriting\n", keyPath)
	}
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Harvest"},
		},
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Duration(opts.Days*24) * time.Hour),

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	for _, ipaddress := range opts.Ipaddress {
		ip := strings.TrimSpace(ipaddress)
		if len(ip) > 0 {
			template.IPAddresses = append(template.IPAddresses, net.ParseIP(ip))
		}
	}

	for _, n := range opts.DnsName {
		name := strings.TrimSpace(n)
		if len(name) > 0 {
			template.DNSNames = append(template.DNSNames, n)
		}
	}

	// Create self-signed certificate.
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if pemCert == nil {
		log.Fatal("Failed to encode certificate to PEM")
	}
	if err := os.WriteFile(certPath, pemCert, 0644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s\n", certPath)

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if pemKey == nil {
		log.Fatal("Failed to encode key to PEM")
	}
	if err := os.WriteFile(keyPath, pemKey, 0600); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s\n", keyPath)
}

func (a *Admin) verifyAuth(user string, pass string) bool {
	return a.httpSD.AuthBasic.Username == user && a.httpSD.AuthBasic.Password == pass
}
