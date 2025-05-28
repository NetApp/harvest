package admin

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/spf13/cobra"
	"github.com/zekroTJA/timedmap/v2"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

type Admin struct {
	listen           string
	logger           *slog.Logger
	localIP          string
	pollerToPromAddr *timedmap.TimedMap[string, pollerDetails]
	httpSD           conf.Httpsd
	expireAfter      time.Duration
}

func (a *Admin) startServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sd", a.APISD)

	a.logger.Debug("Admin node starting", slog.String("listen", a.listen))
	server := &http.Server{
		Addr:              a.listen,
		Handler:           mux,
		ReadHeaderTimeout: 60 * time.Second,
	}
	if a.httpSD.TLS.KeyFile != "" {
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS13,
		}
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		a.logger.Info("Admin node is shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			a.logger.Error("Could not gracefully shutdown the admin node", slogx.Err(err))
			os.Exit(1)
		}
		close(done)
	}()

	a.logger.Info(
		"Admin node started",
		slog.String("listen", a.listen),
		slog.Bool("TLS", a.httpSD.TLS.KeyFile != ""),
		slog.Bool("BasicAuth", a.httpSD.AuthBasic.Username != ""),
	)

	if a.httpSD.TLS.KeyFile != "" {
		if err := server.ListenAndServeTLS(a.httpSD.TLS.CertFile, a.httpSD.TLS.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error(
				"Admin node could not listen",
				slogx.Err(err),
				slog.String("listen", a.listen),
				slog.String("ssl_cert", a.httpSD.TLS.CertFile),
				slog.String("ssl_key", a.httpSD.TLS.KeyFile),
				slog.Bool("BasicAuth", a.httpSD.AuthBasic.Username != ""),
			)
			os.Exit(1)
		}
	} else {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error(
				"Admin node could not listen",
				slogx.Err(err),
				slog.String("listen", a.listen),
			)
			os.Exit(1)
		}
	}

	<-done

	a.logger.Info("Admin node stopped")
}

func (a *Admin) APISD(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if a.httpSD.AuthBasic.Username != "" {
		user, pass, ok := r.BasicAuth()
		if !ok || !a.verifyAuth(user, pass) {
			w.Header().Set("Www-Authenticate", `Basic realm="api"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}
	switch r.Method {
	case http.MethodPut:
		a.apiPublish(w, r)
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(a.makeTargets())
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (a *Admin) setupLogger() {
	handlerOptions := &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				source.File = filepath.Base(source.File)
				a.Value = slog.StringValue(fmt.Sprintf("%s:%d", source.File, source.Line))
			}
			return a
		},
	}
	handler := slog.NewTextHandler(os.Stderr, handlerOptions)
	a.logger = slog.New(handler)
}

type pollerDetails struct {
	Name string `json:"Name,omitempty"`
	IP   string `json:"IP,omitempty"`
	Port int    `json:"Port,omitempty"`
}

func (a *Admin) apiPublish(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var publish pollerDetails
	err := decoder.Decode(&publish)
	if err != nil {
		a.logger.Error("Unable to parse publish json", slogx.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	a.pollerToPromAddr.Set(publish.Name, publish, a.expireAfter)
	a.logger.Debug("Published poller", slog.Any("publish", publish))
	_, _ = fmt.Fprintf(w, "OK")
}

type labels struct {
	MetaPoller string `json:"__meta_poller"`
}

type sdTarget struct {
	Targets []string `json:"targets"`
	Labels  labels   `json:"labels"`
}

func (a *Admin) makeTargets() []byte {
	targets := make([]sdTarget, 0)
	for _, details := range a.pollerToPromAddr.Snapshot() {
		target := sdTarget{
			Targets: []string{fmt.Sprintf(`%s:%d`, details.IP, details.Port)},
			Labels:  labels{MetaPoller: details.Name},
		}
		targets = append(targets, target)
	}
	a.logger.Debug("makeTargets", slog.Int("size", len(targets)))
	j, err := json.Marshal(targets)
	if err != nil {
		a.logger.Error("Failed to marshal targets", slogx.Err(err))
		return []byte{}
	}
	return j
}

type tlsOptions struct {
	DNSName   []string
	Ipaddress []string
	Days      int
}

var opts = &tlsOptions{}

func Cmd() *cobra.Command {
	admin := &cobra.Command{
		Use:   "admin",
		Short: "Harvest admin commands",
	}
	admin.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start Harvest admin node",
		Run:   doAdmin,
	})
	ctls := &cobra.Command{
		Use:   "tls",
		Short: "Builtin helpers for creating certificates",
		Long:  "This command has subcommands for interacting with Harvest TLS.",
	}
	tlsCreate := &cobra.Command{
		Use:   "create",
		Short: "Create ",
	}
	tlsServer := &cobra.Command{
		Use:   "server",
		Short: "Create a new server certificates",
		Run:   doTLS,
	}
	tlsCreate.AddCommand(tlsServer)
	tlsCreate.PersistentFlags().StringSliceVar(
		&opts.DNSName, "dnsname", []string{},
		"Additional dns names for Subject Alternative Names. "+
			"localhost is always included. Comma-separated list or provide flag multiple times",
	)
	tlsCreate.PersistentFlags().StringSliceVar(
		&opts.Ipaddress, "ip", []string{},
		"Additional IP addresses for Subject Alternative Names. "+
			"127.0.0.1 is always included. Comma-separated list or provide flag multiple times",
	)
	tlsCreate.PersistentFlags().IntVarP(
		&opts.Days, "days", "d", 365,
		"Number of days the certificate is valid.",
	)
	ctls.AddCommand(tlsCreate)
	admin.AddCommand(ctls)
	return admin
}

func doTLS(_ *cobra.Command, _ []string) {
	GenerateAdminCerts(opts, "admin")
}

func doAdmin(c *cobra.Command, _ []string) {
	var configPath = c.Root().PersistentFlags().Lookup("config").Value.String()
	_, err := conf.LoadHarvestConfig(configPath)
	if err != nil {
		return
	}

	a := newAdmin(configPath)
	a.startServer()
}

func newAdmin(configPath string) Admin {
	a := Admin{
		httpSD: conf.Config.Admin.Httpsd,
		listen: conf.Config.Admin.Httpsd.Listen,
	}
	a.setupLogger()
	if a.listen == "" {
		a.logger.Error("Admin.address is empty in config. Must be a valid address", slog.String("config", configPath))
		os.Exit(1)
	}
	if a.httpSD.TLS != (conf.TLS{}) {
		requests.CheckCert(a.httpSD.TLS.CertFile, "ssl_cert", configPath, a.logger)
		requests.CheckCert(a.httpSD.TLS.KeyFile, "ssl_key", configPath, a.logger)
	}

	a.localIP, _ = requests.FindLocalIP()
	a.expireAfter = a.setDuration(a.httpSD.ExpireAfter, 1*time.Minute, "expire_after")
	a.pollerToPromAddr = timedmap.New[string, pollerDetails](a.expireAfter)
	a.logger.Debug(
		"newAdmin",
		slog.String("localIP", a.localIP),
		slog.String("expireAfter", a.expireAfter.String()),
	)

	return a
}

func (a *Admin) setDuration(every string, defaultDur time.Duration, name string) time.Duration {
	if every == "" {
		return defaultDur
	}
	everyDur, err := time.ParseDuration(every)
	if err != nil {
		a.logger.Warn(
			"Failed to parse name",
			slogx.Err(err),
			slog.String(name, every),
			slog.String("defaultDur", defaultDur.String()),
		)
		everyDur = defaultDur
	}
	return everyDur
}
