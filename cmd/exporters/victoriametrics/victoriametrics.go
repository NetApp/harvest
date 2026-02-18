package victoriametrics

import (
	"bytes"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/exporters"
	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultPort          = 8428
	defaultTimeout       = 5
	defaultAPIVersion    = "1"
	globalPrefix         = ""
	expectedResponseCode = 204
)

type VictoriaMetrics struct {
	*exporter.AbstractExporter
	client       *http.Client
	url          string
	addMetaTags  bool
	globalPrefix string
	bufferPool   *sync.Pool
}

func New(abc *exporter.AbstractExporter) exporter.Exporter {
	return &VictoriaMetrics{AbstractExporter: abc}
}

func (v *VictoriaMetrics) Init() error {

	if err := v.InitAbc(); err != nil {
		return err
	}

	// Initialize buffer pool
	v.bufferPool = &sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}

	var (
		url, addr, version *string
		port               *int
	)

	if instance, err := v.Metadata.NewInstance("http"); err == nil {
		instance.SetLabel("task", "http")
	} else {
		return err
	}

	if instance, err := v.Metadata.NewInstance("info"); err == nil {
		instance.SetLabel("task", "info")
	} else {
		return err
	}

	if x := v.Params.GlobalPrefix; x != nil {
		v.Logger.Debug("use global prefix", slog.String("prefix", *x))
		v.globalPrefix = *x
		if !strings.HasSuffix(v.globalPrefix, "_") {
			v.globalPrefix += "_"
		}
	} else {
		v.globalPrefix = globalPrefix
	}

	// Checking the required/optional params
	// customer should either provide url or addr
	// url is expected to be the full write URL api/v1/import/prometheus
	// when url is defined, addr and port are ignored

	// addr is expected to include host only (no port)
	// when addr is defined, port is required

	dbEndpoint := "addr"
	if url = v.Params.URL; url != nil {
		v.url = *url
		dbEndpoint = "url"
	} else {
		if addr = v.Params.Addr; addr == nil {
			v.Logger.Error("missing url or addr")
			return errs.New(errs.ErrMissingParam, "url or addr")
		}
		if port = v.Params.Port; port == nil {
			v.Logger.Debug("using default port", slog.Int("default", defaultPort))
			port = new(defaultPort)
		}
		if version = v.Params.Version; version == nil {
			version = new(defaultAPIVersion)
		}
		v.Logger.Debug("using api version", slog.String("version", *version))

		//goland:noinspection HttpUrlsUsage
		url = new("http://" + *addr + ":" + strconv.Itoa(*port))
		v.url = fmt.Sprintf("%s/api/v%s/import/prometheus", *url, *version)
	}

	// timeout parameter
	timeout := time.Duration(defaultTimeout) * time.Second
	if ct := v.Params.ClientTimeout; ct != nil {
		if t, err := strconv.Atoi(*ct); err == nil {
			timeout = time.Duration(t) * time.Second
		} else {
			v.Logger.Warn(
				"invalid client_timeout, using default",
				slog.String("client_timeout", *ct),
				slog.Int("default", defaultTimeout),
			)
		}
	} else {
		v.Logger.Debug("using default client_timeout", slog.Int("default", defaultTimeout))
	}

	v.Logger.Debug("initializing exporter", slog.String("endpoint", dbEndpoint), slog.String("url", v.url))

	// construct HTTP client
	v.client = &http.Client{Timeout: timeout}

	return nil
}

func (v *VictoriaMetrics) Export(data *matrix.Matrix) (exporter.Stats, error) {

	var (
		metrics [][]byte
		err     error
		s       time.Time
		stats   exporter.Stats
	)

	v.Lock()
	defer v.Unlock()

	s = time.Now()

	// update timestamp when backfill historical data, Ex: time.Now().Add(-1*24*time.Hour)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10) // Ex: "1762933202"

	// render metrics into open metrics format with timestamp
	metrics, stats, _ = exporters.Render(data, v.addMetaTags, v.Params.SortLabels, v.globalPrefix, v.Logger, timestamp)

	// fix render time
	if err = v.Metadata.LazyAddValueInt64("time", "render", time.Since(s).Microseconds()); err != nil {
		v.Logger.Error("metadata render time", slogx.Err(err))
	}
	// in test mode, don't emit metrics
	if v.Options.IsTest {
		return stats, nil
		// otherwise, to the actual export: send to the DB
	} else if err = v.Emit(metrics); err != nil {
		return stats, fmt.Errorf("unable to emit object: %s, uuid: %s, err=%w", data.Object, data.UUID, err)
	}

	v.Logger.Debug(
		"exported",
		slog.String("object", data.Object),
		slog.String("uuid", data.UUID),
		slog.Int("numMetric", len(metrics)),
	)

	// update metadata
	if err = v.Metadata.LazySetValueInt64("time", "export", time.Since(s).Microseconds()); err != nil {
		v.Logger.Error("metadata export time", slogx.Err(err))
	}

	// render metadata metrics into open metrics format with timestamp
	metrics, stats, _ = exporters.Render(v.Metadata, v.addMetaTags, v.Params.SortLabels, v.globalPrefix, v.Logger, timestamp)
	if err = v.Emit(metrics); err != nil {
		v.Logger.Error("emit metadata", slogx.Err(err))
	}

	return stats, nil
}

func (v *VictoriaMetrics) Emit(data [][]byte) error {
	var buffer *bytes.Buffer
	var request *http.Request
	var response *http.Response
	var err error

	buffer = v.bufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	_, _ = buffer.Write(bytes.Join(data, []byte("\n")))

	defer v.bufferPool.Put(buffer)

	if request, err = requests.New("POST", v.url, buffer); err != nil {
		return err
	}

	if response, err = v.client.Do(request); err != nil {
		return err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer response.Body.Close()
	if response.StatusCode != expectedResponseCode {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return errs.New(errs.ErrAPIResponse, err.Error())
		}
		return fmt.Errorf("%w: %s", errs.ErrAPIRequestRejected, string(body))
	}
	return nil
}
