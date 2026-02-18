package auth

import (
	"cmp"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"mcp-server/cmd/version"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/netapp/harvest/v2/pkg/slogx"
)

// Type represents the type of authentication
type Type string

const (
	None  Type = "none"
	Basic Type = "basic"
	Cert  Type = "cert"
)

// Config holds authentication configuration
type Config struct {
	Type            Type
	Username        string
	Password        string //nolint:gosec
	CertFile        string
	KeyFile         string
	CAFile          string
	InsecureSkipTLS bool // Skip TLS certificate verification
}

// TSDBConfig holds Time Series Database configuration
type TSDBConfig struct {
	URL       string
	Auth      Config
	Timeout   time.Duration
	RulesPath string
}

var logger = slog.Default()

// SetLogger sets the logger for the auth package
func SetLogger(l *slog.Logger) {
	logger = l
}

// LoadAuthConfig loads authentication configuration from environment variables
func LoadAuthConfig() Config {
	authType := cmp.Or(Type(strings.ToLower(os.Getenv("HARVEST_TSDB_AUTH_TYPE"))), None)

	config := Config{Type: authType}

	switch authType {
	case Basic:
		config.Username = os.Getenv("HARVEST_TSDB_USERNAME")
		config.Password = os.Getenv("HARVEST_TSDB_PASSWORD")

		if config.Username == "" || config.Password == "" {
			logger.Warn("basic auth configured but username or password missing",
				slog.String("username_set", strconv.FormatBool(config.Username != "")),
				slog.String("password_set", strconv.FormatBool(config.Password != "")))
		}

	case Cert:
		config.CertFile = os.Getenv("HARVEST_TSDB_CERT_FILE")
		config.KeyFile = os.Getenv("HARVEST_TSDB_KEY_FILE")
		config.CAFile = os.Getenv("HARVEST_TSDB_CA_FILE")

		if config.CertFile == "" || config.KeyFile == "" {
			logger.Warn("certificate auth configured but cert file or key file missing",
				slog.String("cert_file_set", strconv.FormatBool(config.CertFile != "")),
				slog.String("key_file_set", strconv.FormatBool(config.KeyFile != "")))
		} else {
			if _, err := os.Stat(config.CertFile); os.IsNotExist(err) {
				logger.Warn("certificate file does not exist", slog.String("cert_file", config.CertFile))
			}
			if _, err := os.Stat(config.KeyFile); os.IsNotExist(err) {
				logger.Warn("key file does not exist", slog.String("key_file", config.KeyFile))
			}
			if config.CAFile != "" {
				if _, err := os.Stat(config.CAFile); os.IsNotExist(err) {
					logger.Warn("CA file does not exist", slog.String("ca_file", config.CAFile))
				}
			}
		}

	case None:
		// No authentication needed

	default:
		logger.Warn("unknown auth type, defaulting to none", slog.String("auth_type", string(authType)))
		config.Type = None
	}

	if insecureStr := os.Getenv("HARVEST_TSDB_TLS_INSECURE"); insecureStr != "" {
		config.InsecureSkipTLS = strings.EqualFold(insecureStr, "true")
		if config.InsecureSkipTLS {
			logger.Warn("TLS certificate verification disabled - use only for development/testing")
		}
	}

	return config
}

// createTLSConfig creates a TLS configuration for certificate authentication
func createTLSConfig(auth Config) (*tls.Config, error) {
	if auth.CertFile == "" || auth.KeyFile == "" {
		return nil, errors.New("certificate auth configured but cert file or key file missing")
	}

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(auth.CertFile, auth.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Load CA certificate if provided
	if auth.CAFile != "" {
		caCert, err := os.ReadFile(auth.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}

// GetTSDBConfig loads the complete TSDB configuration
func GetTSDBConfig() TSDBConfig {
	// Parse timeout from environment variable, default to 30 seconds
	timeout := 30 * time.Second
	if timeoutStr := os.Getenv("HARVEST_TSDB_TIMEOUT"); timeoutStr != "" {
		if parsedTimeout, err := time.ParseDuration(timeoutStr); err != nil {
			logger.Warn("invalid HARVEST_TSDB_TIMEOUT format, using default",
				slog.String("timeout", timeoutStr),
				slog.String("default", timeout.String()))
		} else {
			timeout = parsedTimeout
		}
	}

	// Rules path configuration
	rulesPath := os.Getenv("HARVEST_RULES_PATH")
	if rulesPath == "" {
		// Try to infer from common patterns
		if workDir, err := os.Getwd(); err == nil {
			if strings.Contains(workDir, "harvest") || strings.Contains(workDir, "Harvest") {
				// Look for common rule file locations
				commonPaths := []string{
					filepath.Join(workDir, "rules"),
					filepath.Join(workDir, "conf", "rules"),
					filepath.Join(workDir, "prometheus", "rules"),
					filepath.Join(workDir, "grafana", "rules"),
				}

				for _, path := range commonPaths {
					if info, err := os.Stat(path); err == nil && info.IsDir() {
						rulesPath = path
						break
					}
				}
			}
		}
	}

	return TSDBConfig{
		URL:       os.Getenv("HARVEST_TSDB_URL"),
		Auth:      LoadAuthConfig(),
		Timeout:   timeout,
		RulesPath: rulesPath,
	}
}

func createHTTPClient(config TSDBConfig) *http.Client {
	transport := &http.Transport{
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}

	// Configure TLS settings
	var tlsConfig *tls.Config

	if config.Auth.Type == Cert {
		certTLSConfig, err := createTLSConfig(config.Auth)
		if err != nil {
			logger.Error("failed to create TLS config for certificate auth", slogx.Err(err))
			tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		} else {
			tlsConfig = certTLSConfig
		}
	} else {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	// Apply insecure TLS setting if configured (overrides certificate verification)
	if config.Auth.InsecureSkipTLS {
		tlsConfig.InsecureSkipVerify = true
		logger.Debug("TLS certificate verification disabled")
	}

	transport.TLSClientConfig = tlsConfig

	return &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}
}

// addAuthToRequest adds authentication to an HTTP request
func addAuthToRequest(req *http.Request, auth Config) error {
	switch auth.Type {
	case None:
		// No authentication needed

	case Basic:
		if auth.Username == "" || auth.Password == "" {
			return errors.New("basic auth configured but username or password missing")
		}
		req.SetBasicAuth(auth.Username, auth.Password)

	case Cert:
		if auth.CertFile == "" || auth.KeyFile == "" {
			return errors.New("certificate auth configured but cert file or key file missing")
		}

	default:
		return fmt.Errorf("unsupported auth type: %s", auth.Type)
	}

	return nil
}

// MakeRequest makes an HTTP GET request with authentication
func MakeRequest(config TSDBConfig, url string) (*http.Response, error) {
	client := createHTTPClient(config)

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if err := addAuthToRequest(req, config.Auth); err != nil {
		return nil, fmt.Errorf("failed to add auth to request: %w", err)
	}

	req.Header.Set("User-Agent", "harvest-mcp-server/"+version.Info())

	resp, err := client.Do(req) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to make request to %s: %w", url, err)
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn("failed to close response body after auth error", slog.Any("error", closeErr))
		}
		return nil, errors.New("authentication failed (401 Unauthorized) - check credentials")
	case http.StatusForbidden:
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn("failed to close response body after forbidden error", slog.Any("error", closeErr))
		}
		return nil, errors.New("access forbidden (403 Forbidden) - insufficient permissions")
	case http.StatusNotFound:
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn("failed to close response body after not found error", slog.Any("error", closeErr))
		}
		return nil, fmt.Errorf("endpoint not found (404) - check TSDB_URL: %s", config.URL)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn("failed to close response body after HTTP error", slog.Any("error", closeErr))
		}
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	return resp, nil
}
