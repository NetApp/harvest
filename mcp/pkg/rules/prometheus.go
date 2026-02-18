package rules

import (
	"errors"
	"fmt"
	"log/slog"
	"mcp-server/pkg/auth"
	"net/http"
	"os"
	"strings"
	"time"
)

// PrometheusClient handles interactions with Prometheus
type PrometheusClient struct {
	baseURL         string
	reloadURL       string
	enableReloadAPI bool
	client          *http.Client
	logger          *slog.Logger
	config          auth.TSDBConfig
}

// NewPrometheusClient creates a new Prometheus client
func NewPrometheusClient(config auth.TSDBConfig, logger *slog.Logger) *PrometheusClient {
	baseURL := config.URL
	if baseURL == "" {
		baseURL = os.Getenv("HARVEST_TSDB_URL")
	}
	if baseURL == "" {
		baseURL = "http://localhost:9090"
	}

	// Remove trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	reloadURL := os.Getenv("HARVEST_TSDB_AUTO_RELOAD_URL")
	if reloadURL == "" {
		reloadURL = baseURL + "/-/reload"
	}

	enableReloadAPI := os.Getenv("HARVEST_TSDB_AUTO_RELOAD") == "true"

	return &PrometheusClient{
		config:          config,
		baseURL:         baseURL,
		reloadURL:       reloadURL,
		enableReloadAPI: enableReloadAPI,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// CanReload returns whether the reload API is enabled
func (pc *PrometheusClient) CanReload() bool {
	return pc.enableReloadAPI
}

// GetReloadURL returns the reload API URL
func (pc *PrometheusClient) GetReloadURL() string {
	return pc.reloadURL
}

// ReloadConfig triggers a Prometheus configuration reload
func (pc *PrometheusClient) ReloadConfig() error {
	if !pc.enableReloadAPI {
		return errors.New("prometheus reload API is disabled. Set HARVEST_TSDB_AUTO_RELOAD=true and restart Prometheus with --web.enable-lifecycle")
	}

	pc.logger.Debug("triggering Prometheus config reload", slog.String("url", pc.reloadURL))

	req, err := http.NewRequest(http.MethodPost, pc.reloadURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create reload request: %w", err)
	}

	resp, err := pc.client.Do(req) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to call reload API: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("reload API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	pc.logger.Info("Prometheus configuration reloaded successfully")
	return nil
}

// TestConnection tests the connection to Prometheus
func (pc *PrometheusClient) TestConnection() error {
	healthURL := pc.baseURL + "/-/healthy"

	pc.logger.Debug("testing Prometheus connection", slog.String("url", healthURL))

	resp, err := pc.client.Get(healthURL)
	if err != nil {
		return fmt.Errorf("failed to connect to Prometheus: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("prometheus health check failed with status %d", resp.StatusCode)
	}

	pc.logger.Debug("Prometheus connection test successful")
	return nil
}

// GetReloadInstructions returns manual reload instructions
func (pc *PrometheusClient) GetReloadInstructions() string {
	if pc.enableReloadAPI {
		return "Automatic reload via API: " + pc.reloadURL
	}

	return `Manual reload required. Choose one of:
1. Restart the Prometheus container/service
2. Send SIGHUP to the Prometheus process
3. Enable reload API by starting Prometheus with --web.enable-lifecycle and set HARVEST_TSDB_AUTO_RELOAD=true`
}
