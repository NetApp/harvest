package prometheus

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/netapp/harvest/v2/pkg/conf"
)

// clientAuthTypes are the supported values of tls client_auth_type. The names match the
// Prometheus exporter-toolkit web configuration.
var clientAuthTypes = map[string]tls.ClientAuthType{
	"RequireAndVerifyClientCert": tls.RequireAndVerifyClientCert,
	"VerifyClientCertIfGiven":    tls.VerifyClientCertIfGiven,
}

const defaultClientAuthType = "RequireAndVerifyClientCert"

// newTLSConfig returns the server TLS configuration for mutual TLS, or nil when client
// certificate authentication is not configured. The server certificate itself is loaded
// from cert_file and key_file by ListenAndServeTLS.
func newTLSConfig(t conf.PromTLS) (*tls.Config, error) {
	if t.ClientCAFile == "" && t.ClientAuthType == "" {
		return nil, nil
	}
	if t.CertFile == "" || t.KeyFile == "" {
		return nil, errors.New("client_ca_file and client_auth_type require cert_file and key_file")
	}
	if t.ClientCAFile == "" {
		return nil, errors.New("client_auth_type requires client_ca_file")
	}

	authName := t.ClientAuthType
	if authName == "" {
		authName = defaultClientAuthType
	}
	authType, ok := clientAuthTypes[authName]
	if !ok {
		names := make([]string, 0, len(clientAuthTypes))
		for name := range clientAuthTypes {
			names = append(names, name)
		}
		slices.Sort(names)
		return nil, fmt.Errorf("invalid client_auth_type %q, must be one of %s", authName, strings.Join(names, ", "))
	}

	caPEM, err := os.ReadFile(t.ClientCAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read client_ca_file: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("no PEM certificates found in client_ca_file %s", t.ClientCAFile)
	}

	return &tls.Config{
		ClientAuth: authType,
		ClientCAs:  pool,
		MinVersion: tls.VersionTLS12,
	}, nil
}
