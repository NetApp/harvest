package prometheus

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/netapp/harvest/v2/assert"
)

func TestPprofLoopbackOnly(t *testing.T) {
	mux := (&Prometheus{}).newMux()

	tests := []struct {
		name       string
		remoteAddr string
		host       string
		path       string
		want       int
	}{
		{name: "remote with spoofed localhost Host", remoteAddr: "203.0.113.9:5555", host: "localhost:12990", path: "/debug/pprof/", want: http.StatusNotFound},
		{name: "remote cmdline", remoteAddr: "203.0.113.9:5555", host: "localhost", path: "/debug/pprof/cmdline", want: http.StatusNotFound},
		{name: "remote heap", remoteAddr: "203.0.113.9:5555", host: "localhost", path: "/debug/pprof/heap", want: http.StatusNotFound},
		{name: "remote goroutine", remoteAddr: "203.0.113.9:5555", host: "localhost", path: "/debug/pprof/goroutine", want: http.StatusNotFound},
		{name: "remote health is unaffected", remoteAddr: "203.0.113.9:5555", host: "10.1.2.3:12990", path: "/health", want: http.StatusOK},
		{name: "remote IPv6",remoteAddr: "[2001:db8::1]:5555", host: "localhost", path: "/debug/pprof/", want: http.StatusNotFound},
		{name: "malformed remote addr", remoteAddr: "garbage", host: "localhost", path: "/debug/pprof/", want: http.StatusNotFound},
		{name: "loopback IPv4", remoteAddr: "127.0.0.1:5555", host: "10.1.2.3:12990", path: "/debug/pprof/", want: http.StatusOK},
		{name: "loopback heap", remoteAddr: "127.0.0.1:5555", host: "localhost", path: "/debug/pprof/heap", want: http.StatusOK},
		{name: "loopback IPv6",remoteAddr: "[::1]:5555", host: "localhost", path: "/debug/pprof/cmdline", want: http.StatusOK},
		{name: "IPv4-mapped loopback", remoteAddr: "[::ffff:127.0.0.1]:5555", host: "localhost", path: "/debug/pprof/", want: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "http://"+tt.host+tt.path, http.NoBody)
			r.RemoteAddr = tt.remoteAddr
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			assert.Equal(t, w.Code, tt.want)
		})
	}
}
