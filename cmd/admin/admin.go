package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	"goharvest2/pkg/conf"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type Admin struct {
	Addr             string
	logger           zerolog.Logger
	localIP          string
	cachedTargets    string
	pollerToPromAddr map[string]string
}

func (a *Admin) startServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sd", a.ApiPromSD)

	a.logger.Debug().Str("addr", a.Addr).Msgf("Admin node starting")
	server := &http.Server{Addr: a.Addr, Handler: mux}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		a.logger.Info().Msg("Admin node is shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			a.logger.Fatal().
				Err(err).
				Msg("Could not gracefully shutdown the admin node")
		}
		close(done)
	}()

	a.logger.Info().
		Str("addr", a.Addr).
		Msg("Admin node started")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		a.logger.Fatal().Err(err).
			Str("addr", a.Addr).
			Msg("Admin node could not listen")
	}
	<-done
	a.logger.Info().Msg("Admin node stopped")
}

func (a *Admin) ApiPromSD(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
	_, _ = fmt.Fprintf(w, `[{"targets": [%s]}]`, a.cachedTargets)
}

func (a *Admin) setupLogger() {
	zl := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Caller().
		Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	a.logger = zl
}

func (a *Admin) findLocalIP() string {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	defer func(conn net.Conn) { _ = conn.Close() }(conn)
	if err != nil {
		a.logger.Fatal().Err(err).Msg("Unable to find local IP")
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// startTicker never returns - it
// updates the cache of pollers every 30s
// Prometheus polls the SD every minute
func (a *Admin) startTicker() {
	a.findPrometheusPollers()
	tick := time.Tick(30 * time.Second)
	for range tick {
		a.findPrometheusPollers()
	}
}

var pollerRegex = regexp.MustCompile(`--poller\s(.*?)\s.*--promPort (\d+)`)

// mapPollerToPromAddr builds a map of poller name to Prometheus addr.
// This is used to create the SD JSON response in findPrometheusPollers.
// If a Prometheus exporter end-point listens on all interfaces, you
// want to return the local ip, not localhost since localhost wouldn't be
// reachable off-box
func (a *Admin) mapPollerToPromAddr() {
	a.pollerToPromAddr = make(map[string]string)
	localIp := a.findLocalIP()
	pollers, _ := conf.GetPollers2("")

	for pollerName, poller := range pollers {
		for _, exporterName := range poller.Exporters {
			exporter, ok := conf.Config.Exporters[exporterName]
			if ok {
				if exporter.Type != "Prometheus" {
					continue
				}
				if exporter.LocalHttpAddr == "localhost" || exporter.LocalHttpAddr == "127.0.0.1" {
					a.pollerToPromAddr[pollerName] = "127.0.0.1"
				} else {
					a.pollerToPromAddr[pollerName] = localIp
				}
			}
		}
	}
}

// Originally used lsof -Pan -i tcp -sTCP:LISTEN
// but non-root Linux users don't get tcp info. pgrep works well with suid
// and on Mac
func (a *Admin) findPrometheusPollers() {
	pgrepArgs := []string{"-fa", "poller.*promPort"}
	if runtime.GOOS == "darwin" {
		pgrepArgs[0] = "-fL"
	}
	var ee *exec.ExitError
	cmd := exec.Command("pgrep", pgrepArgs...)
	bytes, err := cmd.Output()
	if errors.As(err, &ee) {
		if ee.Stderr != nil {
			a.logger.Warn().Err(err).Str("stderr", string(ee.Stderr)).Msg("pgrep failed")
		}
		return // ran, but non-zero exit code
	} else if err != nil {
		a.logger.Warn().Err(err).Msg("pgrep failed")
		return
	}
	out := string(bytes)
	lines := strings.Split(out, "\n")
	targets := make([]string, 0)
	for _, line := range lines {
		matches := pollerRegex.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}
		if promAddr, ok := a.pollerToPromAddr[matches[1]]; ok {
			targets = append(targets, fmt.Sprintf(`"%s:%s"`, promAddr, matches[2]))
		}
	}
	a.cachedTargets = strings.Join(targets, ",")
}

func Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "admin",
		Short: "Start Harvest admin node",
		Long:  "Start Harvest admin node",
		Run:   doAdmin,
	}
}

func doAdmin(c *cobra.Command, _ []string) {
	var configPath = c.Root().PersistentFlags().Lookup("config").Value.String()
	err := conf.LoadHarvestConfig(configPath)
	if err != nil {
		return
	}

	a := Admin{
		Addr: conf.Config.Admin.Address,
	}
	a.setupLogger()
	if a.Addr == "" {
		a.logger.Fatal().
			Str("config", configPath).
			Str("addr", a.Addr).
			Msg("Admin.address is empty in config. Must be a valid address")
	}

	a.localIP = a.findLocalIP()
	a.mapPollerToPromAddr()
	go a.startTicker()
	a.startServer()
}
