package requests

import (
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func FindLocalIP() (string, error) {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		return "", err
	}
	defer func(conn net.Conn) { _ = conn.Close() }(conn)
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func CheckCert(certPath string, name string, configPath string, logger *slog.Logger) {
	if certPath == "" {
		logger.Error("TLS is enabled but cert path is empty",
			slog.String("config", configPath),
			slog.String(name, certPath),
		)
		os.Exit(1)
	}
	absPath := certPath
	if _, err := os.Stat(absPath); err != nil {
		logger.Error("TLS is enabled but cert path is invalid",
			slogx.Err(err),
			slog.String("config", configPath),
			slog.String(name, certPath),
		)
		os.Exit(1)
	}
}

func GetQueryParam(href string, query string) (string, error) {
	u, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	v := u.Query()
	mr := v.Get(query)
	return mr, nil
}

func EncodeURL(href string) (string, error) {
	u, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	u.RawQuery = u.Query().Encode()
	return u.RequestURI(), nil
}

func GetURLWithoutHost(r *http.Request) string {
	urlWithoutHost := r.URL.Path
	if r.URL.RawQuery != "" {
		urlWithoutHost += "?" + r.URL.RawQuery
	}
	return urlWithoutHost
}

// IsPublicAPI returns false if api endpoint has private keyword in it else true
func IsPublicAPI(query string) bool {
	return !strings.Contains(query, "private")
}
