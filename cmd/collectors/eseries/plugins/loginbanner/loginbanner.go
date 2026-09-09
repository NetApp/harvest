package loginbanner

import (
	"bytes"
	"log/slog"
	"net/http"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

const (
	metricName        = "enabled"
	loginBannerMatrix = "eseries_array_login_banner"
)

// bannerStateFor reports whether a banner is configured. ok is false when the
// state can't be determined - anything other than 200/204, including the
// documented 422 error response.
func bannerStateFor(status int, body []byte) (bool, bool) {
	switch status {
	case http.StatusOK:
		// Empty/whitespace body: treated as not configured.
		return len(bytes.TrimSpace(body)) > 0, true
	case http.StatusNoContent:
		return false, true
	default:
		return false, false
	}
}

type LoginBanner struct {
	*plugin.AbstractPlugin
	client *rest.Client
	data   *matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &LoginBanner{AbstractPlugin: p}
}

func (l *LoginBanner) Init(remote conf.Remote) error {
	if err := l.InitAbc(); err != nil {
		return err
	}

	clientTimeout := l.ParentParams.GetChildContentS("client_timeout")
	if clientTimeout == "" {
		clientTimeout = rest.DefaultTimeout
	}

	duration, err := time.ParseDuration(clientTimeout)
	if err != nil {
		l.SLogger.Info("Using default timeout", slogx.Err(err),
			slog.String("client_timeout", clientTimeout),
			slog.String("timeout", rest.DefaultTimeout))
		duration, _ = time.ParseDuration(rest.DefaultTimeout)
	}

	poller, err := conf.PollerNamed(l.Options.Poller)
	if err != nil {
		return err
	}

	credentials := auth.NewCredentials(poller, l.SLogger)
	if l.client, err = rest.New(poller, duration, credentials, ""); err != nil {
		return err
	}

	if err := l.client.Init(1, remote); err != nil {
		return err
	}

	l.data = matrix.New(l.Parent+".LoginBanner", loginBannerMatrix, loginBannerMatrix)
	l.data.SetExportOptions(matrix.NewExportOptions("wwn"))
	if _, err := l.data.NewMetricFloat64(metricName); err != nil {
		return err
	}

	return nil
}

func (l *LoginBanner) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	l.data.PurgeInstances()
	l.data.Reset()
	l.client.Metadata.Reset()

	parent := dataMap[l.Object]
	if parent != nil {
		l.data.UpdateGlobalLabels(parent.GetGlobalLabels())
	}

	wwn := wwnFromParent(parent)
	if wwn == "" {
		l.SLogger.Warn("wwn not found on Array instance, skipping login banner export")
		return l.export(), l.client.Metadata, nil
	}

	arrayID := l.ParentParams.GetChildContentS("array_id")
	if arrayID == "" {
		l.SLogger.Warn("array_id not found in ParentParams, skipping login banner check", slog.String("wwn", wwn))
		return l.export(), l.client.Metadata, nil
	}

	query := rest.NewURLBuilder().
		APIPath("storage-systems/{array_id}/login-banner").
		ArrayID(arrayID).
		Filter([]string{"asFile=false"}).
		Build()

	status, body, err := l.client.GetRaw(l.client.APIPath + "/" + query)
	if err != nil {
		l.SLogger.Warn("Failed to fetch login banner", slogx.Err(err), slog.String("array_id", arrayID))
		return l.export(), l.client.Metadata, nil
	}

	present, ok := bannerStateFor(status, body)
	if !ok {
		l.SLogger.Warn("Could not determine login banner state, skipping export",
			slog.Int("status", status),
			slog.String("array_id", arrayID),
			slog.String("error", errorMessage(body)))
		return l.export(), l.client.Metadata, nil
	}

	if err := l.applyState(wwn, present); err != nil {
		l.SLogger.Error("Failed to record login banner state", slogx.Err(err), slog.String("wwn", wwn))
	}

	return l.export(), l.client.Metadata, nil
}

func (l *LoginBanner) export() []*matrix.Matrix {
	l.client.Metadata.PluginInstances.Store(uint64(len(l.data.GetInstances())))
	return []*matrix.Matrix{l.data}
}

func wwnFromParent(parent *matrix.Matrix) string {
	if parent == nil {
		return ""
	}
	for _, inst := range parent.GetInstances() {
		if wwn := inst.GetLabel("wwn"); wwn != "" {
			return wwn
		}
	}
	return ""
}

// applyState writes 1 (configured) or 0 (not configured) into the plugin matrix.
func (l *LoginBanner) applyState(wwn string, present bool) error {
	inst, err := l.data.NewInstance(wwn)
	if err != nil {
		return err
	}
	inst.SetLabelTrimmed("wwn", wwn)

	value := 0.0
	if present {
		value = 1.0
	}
	l.data.MustGetMetric(metricName).SetValueFloat64(inst, value)
	return nil
}

func errorMessage(body []byte) string {
	return gjson.GetBytes(body, "errorMessage").ClonedString()
}
