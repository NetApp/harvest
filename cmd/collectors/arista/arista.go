package arista

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/arista/plugins/environment"
	"github.com/netapp/harvest/v2/cmd/collectors/arista/plugins/lldp"
	"github.com/netapp/harvest/v2/cmd/collectors/arista/plugins/networkinterface"
	"github.com/netapp/harvest/v2/cmd/collectors/arista/plugins/optic"
	"github.com/netapp/harvest/v2/cmd/collectors/arista/plugins/version"
	"github.com/netapp/harvest/v2/cmd/collectors/arista/rest"
	rest2 "github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"time"
)

type Rest struct {
	*collector.AbstractCollector
	client *rest.Client
	Prop   *prop
}

func init() {
	plugin.RegisterModule(&Rest{})
}

func (r *Rest) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.aristarest",
		New: func() plugin.Module { return new(Rest) },
	}
}

func (r *Rest) Init(a *collector.AbstractCollector) error {

	var err error

	r.AbstractCollector = a

	r.Prop = &prop{}

	if err := r.InitClient(); err != nil {
		return err
	}

	if r.Prop.TemplatePath, err = r.LoadTemplate(); err != nil {
		return err
	}

	r.InitVars(a.Params)

	if err := collector.Init(r); err != nil {
		return err
	}

	if err := r.InitCache(); err != nil {
		return err
	}

	if err := r.InitMatrix(); err != nil {
		return err
	}

	r.Logger.Debug("initialized")

	return nil
}

type prop struct {
	Object       string
	Query        string
	TemplatePath string
}

func (r *Rest) InitClient() error {

	var err error
	a := r.AbstractCollector
	if r.client, err = r.getClient(a); err != nil {
		return err
	}

	if r.Options.IsTest {
		return nil
	}

	if err := r.client.Init(5, r.Remote); err != nil {
		return err
	}

	return nil
}

func (r *Rest) InitMatrix() error {
	mat := r.Matrix[r.Object]
	// overwrite from abstract collector
	mat.Object = r.Prop.Object
	// Add system (switch) name
	mat.SetGlobalLabel("switch", r.Remote.Name)

	if r.Params.HasChildS("labels") {
		for _, l := range r.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}

	return nil
}

func (r *Rest) LoadTemplate() (string, error) {
	var (
		template *node.Node
		path     string
		err      error
	)

	jitter := r.Params.GetChildContentS("jitter")
	models := []string{r.Remote.Model}
	template, path, err = r.ImportSubTemplate(models, rest2.TemplateFn(r.Params, r.Object), jitter, r.Remote.Version)
	if err != nil {
		return "", err
	}

	r.Params.Union(template)
	return path, nil

}

func (r *Rest) InitVars(config *node.Node) {
	var err error

	clientTimeout := config.GetChildContentS("client_timeout")
	if clientTimeout == "" {
		clientTimeout = rest.DefaultTimeout
	}

	duration, err := time.ParseDuration(clientTimeout)
	if err == nil {
		r.client.Timeout = duration
	} else {
		r.Logger.Info("Using default timeout", slog.String("timeout", rest.DefaultTimeout))
	}
}

func (r *Rest) InitCache() error {

	if x := r.Params.GetChildContentS("object"); x != "" {
		r.Prop.Object = x
	}

	if r.Prop.Query = r.Params.GetChildContentS("query"); r.Prop.Query == "" {
		return errs.New(errs.ErrMissingParam, "query")
	}

	return nil
}

func (r *Rest) getClient(a *collector.AbstractCollector) (*rest.Client, error) {

	var (
		poller *conf.Poller
		client *rest.Client
		err    error
	)

	opt := a.GetOptions()
	if poller, err = conf.PollerNamed(opt.Poller); err != nil {
		r.Logger.Error("", slogx.Err(err), slog.String("poller", opt.Poller))
		return nil, err
	}
	if poller.Addr == "" {
		r.Logger.Error("Address is empty", slog.String("poller", opt.Poller))
		return nil, errs.New(errs.ErrMissingParam, "addr")
	}

	if a.Options.IsTest {
		return nil, nil
	}

	if client, err = rest.New(conf.ZapiPoller(r.Params), r.Auth); err != nil {
		return nil, fmt.Errorf("error creating new client: %w", err)
	}

	return client, err
}

func (r *Rest) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Environment":
		return environment.New(abc)
	case "Interface":
		return networkinterface.New(abc)
	case "LLDP":
		return lldp.New(abc)
	case "Optic":
		return optic.New(abc)
	case "Version":
		return version.New(abc)
	default:
		r.Logger.Warn("no arista plugin found", slog.String("kind", kind))
	}
	return nil
}

func (r *Rest) PollData() (map[string]*matrix.Matrix, error) {

	// Unlike the other collectors, the arista collector does not use a template.
	// The plugins are responsible for collecting, parsing, and storing the data.
	r.client.Metadata.Reset()
	r.Metadata.Reset()

	return r.Matrix, nil
}

// Interface guards
var (
	_ collector.Collector = (*Rest)(nil)
)
