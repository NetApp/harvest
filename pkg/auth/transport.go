package auth

import (
	"bufio"
	"bytes"
	"crypto/md5" //nolint:gosec
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
)

type RoundTripFunc func(req *http.Request) (res *http.Response, err error)

func (rtf RoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return rtf(r)
}

func tlsVersion(version string, logger *slog.Logger) uint16 {
	lower := strings.ToLower(version)
	switch lower {
	case "tls10":
		return tls.VersionTLS10
	case "tls11":
		return tls.VersionTLS11
	case "tls12":
		return tls.VersionTLS12
	case "tls13":
		return tls.VersionTLS13
	default:
		logger.Warn("Unknown TLS version, using default", slog.String("version", version))
	}
	return 0
}

func recording(poller *conf.Poller, transport *http.Transport) http.RoundTripper {

	basePath := poller.Recorder.Path

	rtf := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		var (
			err      error
			response *http.Response
		)

		err = os.MkdirAll(basePath, 0750)
		if err != nil {
			return nil, fmt.Errorf("problem while creating directories=%s transport: %w", basePath, err)
		}
		b, err := DumpRequest(req, true)
		if err != nil {
			return nil, err
		}

		requestName, responseName := buildName(b)
		name := filepath.Join(basePath, requestName)
		if err := os.WriteFile(name, b, 0600); err != nil {
			return nil, err
		}
		if response, err = transport.RoundTrip(req); err != nil {
			return nil, err
		}
		b, err = httputil.DumpResponse(response, true)
		if err != nil {
			return nil, err
		}
		name = filepath.Join(basePath, responseName)
		if err := os.WriteFile(name, b, 0600); err != nil {
			return nil, err
		}
		return response, nil
	})

	return rtf
}

func replaying(poller *conf.Poller) http.RoundTripper {

	aFs := os.DirFS(poller.Recorder.Path)

	rtf := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		var (
			err error
		)

		defer func() {
			if err != nil {
				err = fmt.Errorf("problem while replaying transport: %w", err)
			}
		}()

		b, err := DumpRequest(req, true)
		if err != nil {
			return nil, err
		}
		_, name := buildName(b)
		glob := "*" + name
		matches, err := fs.Glob(aFs, glob)
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("%w: no replay file matches %q", errs.ErrResponseNotFound, glob)
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("ambiguous response: multiple replay files match %q", glob)
		}
		b, err = fs.ReadFile(aFs, matches[0])
		if err != nil {
			return nil, err
		}
		r := bufio.NewReader(bytes.NewReader(b))
		return http.ReadResponse(r, req)
	})

	return rtf
}

// return request and response names
func buildName(b []byte) (string, string) {
	h := md5.New() //nolint:gosec
	h.Write(b)
	s := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return s[:8] + ".req.txt", s[:8] + ".res.txt"
}

var reqHeadersToExclude = map[string]bool{
	"Authorization": true,
	"User-Agent":    true,
}

// DumpRequest was copied from httputil.DumpRequest to remove the Host, Authorization, and User-Agent headers
func DumpRequest(req *http.Request, body bool) ([]byte, error) {
	var err error
	save := req.Body
	if !body || req.Body == nil {
		req.Body = nil
	} else {
		save, req.Body, err = drainBody(req.Body)
		if err != nil {
			return nil, err
		}
		//goland:noinspection GoUnhandledErrorResult
		defer req.Body.Close()
		//goland:noinspection GoUnhandledErrorResult
		defer save.Close()
	}

	var b bytes.Buffer

	// By default, print out the unmodified req.RequestURI, which
	// is always set for incoming server requests. But because we
	// previously used req.URL.RequestURI and the docs weren't
	// always so clear about when to use DumpRequest vs
	// DumpRequestOut, fall back to the old way if the caller
	// provides a non-server Request.
	reqURI := req.RequestURI
	if reqURI == "" {
		reqURI = req.URL.RequestURI()
	}

	_, _ = fmt.Fprintf(&b, "%s %s HTTP/%d.%d\r\n", valueOrDefault(req.Method, "GET"),
		reqURI, req.ProtoMajor, req.ProtoMinor)

	chunked := len(req.TransferEncoding) > 0 && req.TransferEncoding[0] == "chunked"
	if len(req.TransferEncoding) > 0 {
		_, _ = fmt.Fprintf(&b, "Transfer-Encoding: %s\r\n", strings.Join(req.TransferEncoding, ","))
	}

	err = req.Header.WriteSubset(&b, reqHeadersToExclude)
	if err != nil {
		return nil, err
	}

	b.WriteString("\r\n")

	if req.Body != nil {
		var dest io.Writer = &b
		if chunked {
			dest = httputil.NewChunkedWriter(dest)
		}
		_, err = io.Copy(dest, req.Body)
		if chunked {
			_ = dest.(io.Closer).Close()
			b.WriteString("\r\n")
		}
	}

	req.Body = save
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// drainBody reads all of b to memory and then returns two equivalent
// ReadClosers yielding the same bytes.
//
// It returns an error if the initial slurp of all bytes fails. It does not attempt
// to make the returned ReadClosers have identical error-matching behavior.
func drainBody(b io.ReadCloser) (io.ReadCloser, io.ReadCloser, error) {
	if b == nil || b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err := b.Close(); err != nil {
		return nil, b, err
	}
	return io.NopCloser(&buf), io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

// Return value if nonempty, def otherwise.
func valueOrDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}
