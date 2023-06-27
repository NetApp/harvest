package admin

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/zekroTJA/timedmap"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Admin struct {
	listen           string
	logger           zerolog.Logger
	localIP          string
	pollerToPromAddr *timedmap.TimedMap
	httpSD           conf.Httpsd
	expireAfter      time.Duration
}

func (a *Admin) startServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sd", a.APISD)

	a.logger.Debug().Str("listen", a.listen).Msg("Admin node starting")
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
		Str("listen", a.listen).
		Bool("TLS", a.httpSD.TLS.KeyFile != "").
		Bool("BasicAuth", a.httpSD.AuthBasic.Username != "").
		Msg("Admin node started")

	if a.httpSD.TLS.KeyFile != "" {
		if err := server.ListenAndServeTLS(a.httpSD.TLS.CertFile, a.httpSD.TLS.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Fatal().Err(err).
				Str("listen", a.listen).
				Str("ssl_cert", a.httpSD.TLS.CertFile).
				Str("ssl_key", a.httpSD.TLS.KeyFile).
				Msg("Admin node could not listen")
		}
	} else {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Fatal().Err(err).
				Str("listen", a.listen).
				Msg("Admin node could not listen")
		}
	}

	<-done
	a.logger.Info().Msg("Admin node stopped")
}

func (a *Admin) APISD(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	if a.httpSD.AuthBasic.Username != "" {
		user, pass, ok := r.BasicAuth()
		if !ok || !a.verifyAuth(user, pass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="api"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}
	if r.Method == "PUT" {
		a.apiPublish(w, r)
	} else if r.Method == "GET" {
		w.WriteHeader(200)
		_, _ = w.Write(a.makeTargets())
	} else {
		w.WriteHeader(400)
	}
}

func (a *Admin) setupLogger() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.ErrorStackMarshaler = logging.MarshalStack

	a.logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Caller().Timestamp().Logger()
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
		a.logger.Err(err).Msg("Unable to parse publish json")
		w.WriteHeader(400)
		return
	}
	a.pollerToPromAddr.Set(publish.Name, publish, a.expireAfter)
	a.logger.Debug().Str("name", publish.Name).Str("ip", publish.IP).Int("port", publish.Port).
		Msg("Published poller")
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
		pd := details.(pollerDetails)
		target := sdTarget{
			Targets: []string{fmt.Sprintf(`%s:%d`, pd.IP, pd.Port)},
			Labels:  labels{MetaPoller: pd.Name},
		}
		targets = append(targets, target)
	}
	a.logger.Debug().Int("size", len(targets)).Msg("makeTargets")
	j, err := json.Marshal(targets)
	if err != nil {
		a.logger.Error().Err(err).Msg("Failed to marshal targets")
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
	err := conf.LoadHarvestConfig(configPath)
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
		a.logger.Fatal().
			Str("config", configPath).
			Msg("Admin.address is empty in config. Must be a valid address")
	}
	if a.httpSD.TLS != (conf.TLS{}) {
		util.CheckCert(a.httpSD.TLS.CertFile, "ssl_cert", configPath, a.logger)
		util.CheckCert(a.httpSD.TLS.KeyFile, "ssl_key", configPath, a.logger)
	}

	a.localIP, _ = util.FindLocalIP()
	a.expireAfter = a.setDuration(a.httpSD.ExpireAfter, 1*time.Minute, "expire_after")
	a.pollerToPromAddr = timedmap.New(a.expireAfter)
	a.logger.Debug().
		Str("expireAfter", a.expireAfter.String()).
		Str("localIP", a.localIP).
		Msg("newAdmin")

	return a
}

func (a *Admin) setDuration(every string, defaultDur time.Duration, name string) time.Duration {
	if every == "" {
		return defaultDur
	}
	everyDur, err := time.ParseDuration(every)
	if err != nil {
		a.logger.Warn().Err(err).Str(name, every).
			Msgf("Failed to parse %s. Using %s", name, defaultDur.String())
		everyDur = defaultDur
	}
	return everyDur
}
