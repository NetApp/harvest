package collector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"testing"

	"github.com/netapp/harvest/v2/pkg/errs"
)

type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func attrMap(attrs []any) map[string]slog.Value {
	out := map[string]slog.Value{}
	for _, a := range attrs {
		attr := a.(slog.Attr)
		out[attr.Key] = attr.Value
	}
	return out
}

func TestPollErrorAttrs(t *testing.T) {
	restErr := errs.NewRest().StatusCode(404).Message("API not found").API("/api/protocols/san/igroups").Build()
	// the same shapes the REST client builds: RestError wrapped by the collector,
	// a url.Error from client.Do, and a body read that times out
	tests := []struct {
		name string
		err  error
		want map[string]string
	}{
		{"rest status and api", fmt.Errorf("failed to fetch data: %w", restErr),
			map[string]string{"statusCode": "404", "api": "/api/protocols/san/igroups"}},
		{"harvest error status", errs.New(errs.ErrAPIRequestRejected, "rejected", errs.WithStatus(429)),
			map[string]string{"statusCode": "429"}},
		{"connection timeout", fmt.Errorf("connection error: %w", &url.Error{Op: "Get", URL: "https://c/api", Err: timeoutError{}}),
			map[string]string{"timeout": "true"}},
		{"body read deadline", errs.NewRest().StatusCode(200).Error(context.DeadlineExceeded).API("/api/cluster").Build(),
			map[string]string{"statusCode": "200", "api": "/api/cluster", "timeout": "true"}},
		{"plain error", errors.New("something else"), map[string]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := attrMap(pollErrorAttrs(tt.err))
			if len(got) != len(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
			for k, v := range tt.want {
				if got[k].String() != v {
					t.Errorf("%s = %q, want %q", k, got[k].String(), v)
				}
			}
		})
	}
}
