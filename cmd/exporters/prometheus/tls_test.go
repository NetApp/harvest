package prometheus

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
)

type testCert struct {
	cert     *x509.Certificate
	key      *ecdsa.PrivateKey
	certFile string
	keyFile  string
}

// newTestCert creates a certificate signed by parent, or a self-signed CA when parent is nil.
func newTestCert(t *testing.T, dir string, name string, parent *testCert, usage x509.ExtKeyUsage) *testCert {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 64))
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	signer, signerKey := tmpl, key
	if parent == nil {
		tmpl.IsCA = true
		tmpl.KeyUsage |= x509.KeyUsageCertSign
	} else {
		tmpl.ExtKeyUsage = []x509.ExtKeyUsage{usage}
		tmpl.DNSNames = []string{"localhost"}
		tmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
		signer, signerKey = parent.cert, parent.key
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, signer, &key.PublicKey, signerKey)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	tc := &testCert{
		cert:     cert,
		key:      key,
		certFile: filepath.Join(dir, name+"-cert.pem"),
		keyFile:  filepath.Join(dir, name+"-key.pem"),
	}
	writePEM(t, tc.certFile, "CERTIFICATE", der)
	writePEM(t, tc.keyFile, "PRIVATE KEY", keyDER)
	return tc
}

func writePEM(t *testing.T, path string, blockType string, der []byte) {
	t.Helper()
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
}

type mtlsFixture struct {
	dir         string
	ca          *testCert
	server      *testCert
	client      *testCert
	rogueClient *testCert
}

func newMTLSFixture(t *testing.T) mtlsFixture {
	t.Helper()
	dir := t.TempDir()
	ca := newTestCert(t, dir, "ca", nil, 0)
	rogueCA := newTestCert(t, dir, "rogue-ca", nil, 0)
	return mtlsFixture{
		dir:         dir,
		ca:          ca,
		server:      newTestCert(t, dir, "server", ca, x509.ExtKeyUsageServerAuth),
		client:      newTestCert(t, dir, "client", ca, x509.ExtKeyUsageClientAuth),
		rogueClient: newTestCert(t, dir, "rogue-client", rogueCA, x509.ExtKeyUsageClientAuth),
	}
}

// startTestServer serves the exporter's real HTTP server over TLS on a random loopback port.
func startTestServer(t *testing.T, tlsConf conf.TLS) string {
	t.Helper()
	absExp := exporter.New("Prometheus", "prom-mtls", &options.Options{PromPort: 1},
		conf.Exporter{IsTest: true, TLS: tlsConf}, nil)
	p := New(absExp).(*Prometheus)
	if err := p.Init(); err != nil {
		t.Fatal(err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := p.newServer("127.0.0.1", 0)
	go func() {
		_ = server.ServeTLS(ln, tlsConf.CertFile, tlsConf.KeyFile)
	}()
	t.Cleanup(func() { _ = server.Close() })
	return "https://" + ln.Addr().String() + "/health"
}

func newTestClient(t *testing.T, f mtlsFixture, clientCert *testCert) *http.Client {
	t.Helper()
	roots := x509.NewCertPool()
	roots.AddCert(f.ca.cert)
	tlsConf := &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
	if clientCert != nil {
		pair, err := tls.LoadX509KeyPair(clientCert.certFile, clientCert.keyFile)
		if err != nil {
			t.Fatal(err)
		}
		// Always present the certificate, even when it is not issued by one of the CAs the
		// server advertises, so the server's verification is what accepts or rejects it.
		tlsConf.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return &pair, nil
		}
	}
	return &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConf}, Timeout: 5 * time.Second}
}

func get(client *http.Client, url string) (int, error) {
	resp, err := client.Get(url) //nolint:noctx
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}

func TestMutualTLS(t *testing.T) {
	f := newMTLSFixture(t)

	tests := []struct {
		name       string
		authType   string
		clientCert func() *testCert
		wantOK     bool
	}{
		{name: "default requires a client cert, trusted cert", clientCert: func() *testCert { return f.client }, wantOK: true},
		{name: "default requires a client cert, no cert", clientCert: func() *testCert { return nil }, wantOK: false},
		{name: "default requires a client cert, untrusted cert", clientCert: func() *testCert { return f.rogueClient }, wantOK: false},
		{name: "verify if given, no cert", authType: "VerifyClientCertIfGiven", clientCert: func() *testCert { return nil }, wantOK: true},
		{name: "verify if given, trusted cert", authType: "VerifyClientCertIfGiven", clientCert: func() *testCert { return f.client }, wantOK: true},
		{name: "verify if given, untrusted cert", authType: "VerifyClientCertIfGiven", clientCert: func() *testCert { return f.rogueClient }, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := startTestServer(t, conf.TLS{
				CertFile:       f.server.certFile,
				KeyFile:        f.server.keyFile,
				ClientCAFile:   f.ca.certFile,
				ClientAuthType: tt.authType,
			})
			code, err := get(newTestClient(t, f, tt.clientCert()), url)
			if tt.wantOK {
				assert.Nil(t, err)
				assert.Equal(t, code, http.StatusOK)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestTLSWithoutClientCAIsUnchanged(t *testing.T) {
	f := newMTLSFixture(t)
	url := startTestServer(t, conf.TLS{CertFile: f.server.certFile, KeyFile: f.server.keyFile})

	code, err := get(newTestClient(t, f, nil), url)
	assert.Nil(t, err)
	assert.Equal(t, code, http.StatusOK)
}

func TestNewTLSConfig(t *testing.T) {
	f := newMTLSFixture(t)
	notPEM := filepath.Join(f.dir, "not-pem.txt")
	if err := os.WriteFile(notPEM, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	server := conf.TLS{CertFile: f.server.certFile, KeyFile: f.server.keyFile}
	withCA := func(t conf.TLS, ca string, auth string) conf.TLS {
		t.ClientCAFile = ca
		t.ClientAuthType = auth
		return t
	}

	tests := []struct {
		name     string
		tls      conf.TLS
		wantNil  bool
		wantAuth tls.ClientAuthType
		wantErr  string
	}{
		{name: "no tls", tls: conf.TLS{}, wantNil: true},
		{name: "server tls only", tls: server, wantNil: true},
		{name: "client ca defaults to require and verify", tls: withCA(server, f.ca.certFile, ""), wantAuth: tls.RequireAndVerifyClientCert},
		{name: "verify if given", tls: withCA(server, f.ca.certFile, "VerifyClientCertIfGiven"), wantAuth: tls.VerifyClientCertIfGiven},
		{name: "client ca without server cert", tls: withCA(conf.TLS{}, f.ca.certFile, ""), wantErr: "require cert_file and key_file"},
		{name: "auth type without client ca", tls: withCA(server, "", "RequireAndVerifyClientCert"), wantErr: "requires client_ca_file"},
		{name: "unknown auth type", tls: withCA(server, f.ca.certFile, "RequestClientCert"), wantErr: "invalid client_auth_type"},
		{name: "missing ca file", tls: withCA(server, filepath.Join(f.dir, "missing.pem"), ""), wantErr: "failed to read client_ca_file"},
		{name: "ca file without certificates", tls: withCA(server, notPEM, ""), wantErr: "no PEM certificates"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newTLSConfig(tt.tls)
			if tt.wantErr != "" {
				assert.NotNil(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.wantErr))
				return
			}
			assert.Nil(t, err)
			if tt.wantNil {
				assert.True(t, got == nil)
				return
			}
			assert.Equal(t, got.ClientAuth, tt.wantAuth)
			assert.Equal(t, got.MinVersion, uint16(tls.VersionTLS12))
			assert.True(t, got.ClientCAs != nil)
		})
	}
}

func TestInitRejectsInvalidTLS(t *testing.T) {
	absExp := exporter.New("Prometheus", "prom-bad-tls", &options.Options{PromPort: 1},
		conf.Exporter{IsTest: true, TLS: conf.TLS{ClientCAFile: "ca.pem"}}, nil)
	err := New(absExp).Init()
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "tls"))
}
