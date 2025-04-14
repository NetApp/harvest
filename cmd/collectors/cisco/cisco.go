package cisco

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/cisco/plugins/environment"
	"github.com/netapp/harvest/v2/cmd/collectors/cisco/plugins/networkinterface"
	"github.com/netapp/harvest/v2/cmd/collectors/cisco/rest"
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

//goland:noinspection GoNameStartsWithPackageName
type CiscoRest struct { // revive:disable-line exported
	*collector.AbstractCollector
	client *rest.Client
	Prop   *prop
}

func init() {
	plugin.RegisterModule(&CiscoRest{})
}

func (c *CiscoRest) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.ciscorest",
		New: func() plugin.Module { return new(CiscoRest) },
	}
}

func (c *CiscoRest) Init(a *collector.AbstractCollector) error {

	var err error

	c.AbstractCollector = a

	c.Prop = &prop{}

	if err := c.InitClient(); err != nil {
		return err
	}

	if c.Prop.TemplatePath, err = c.LoadTemplate(); err != nil {
		return err
	}

	c.InitVars(a.Params)

	if err := collector.Init(c); err != nil {
		return err
	}

	if err := c.InitCache(); err != nil {
		return err
	}

	if err := c.InitMatrix(); err != nil {
		return err
	}

	c.Logger.Debug("initialized")

	return nil
}

type prop struct {
	Object       string
	Query        string
	TemplatePath string
}

func (c *CiscoRest) InitClient() error {

	var err error
	a := c.AbstractCollector
	if c.client, err = c.getClient(a); err != nil {
		return err
	}

	if c.Options.IsTest {
		return nil
	}

	if err := c.client.Init(5, c.Remote); err != nil {
		return err
	}

	return nil
}

func (c *CiscoRest) InitMatrix() error {
	mat := c.Matrix[c.Object]
	// overwrite from abstract collector
	mat.Object = c.Prop.Object
	// Add system (cluster) name
	mat.SetGlobalLabel("cluster", c.Remote.Name)

	if c.Params.HasChildS("labels") {
		for _, l := range c.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}

	return nil
}

func (c *CiscoRest) LoadTemplate() (string, error) {
	var (
		template *node.Node
		path     string
		err      error
	)

	jitter := c.Params.GetChildContentS("jitter")

	template, path, err = c.ImportSubTemplate(c.Remote.Model, rest2.TemplateFn(c.Params, c.Object), jitter, c.Remote.Version)
	if err != nil {
		return "", err
	}

	c.Params.Union(template)
	return path, nil

}

func (c *CiscoRest) InitVars(config *node.Node) {
	var err error

	clientTimeout := config.GetChildContentS("client_timeout")
	if clientTimeout == "" {
		clientTimeout = rest.DefaultTimeout
	}

	duration, err := time.ParseDuration(clientTimeout)
	if err == nil {
		c.client.Timeout = duration
	} else {
		c.Logger.Info("Using default timeout", slog.String("timeout", rest.DefaultTimeout))
	}
}

func (c *CiscoRest) InitCache() error {

	if x := c.Params.GetChildContentS("object"); x != "" {
		c.Prop.Object = x
	}

	if c.Prop.Query = c.Params.GetChildContentS("query"); c.Prop.Query == "" {
		return errs.New(errs.ErrMissingParam, "query")
	}

	return nil
}

func (c *CiscoRest) getClient(a *collector.AbstractCollector) (*rest.Client, error) {

	var (
		poller *conf.Poller
		client *rest.Client
		err    error
	)

	opt := a.GetOptions()
	if poller, err = conf.PollerNamed(opt.Poller); err != nil {
		c.Logger.Error("", slogx.Err(err), slog.String("poller", opt.Poller))
		return nil, err
	}
	if poller.Addr == "" {
		c.Logger.Error("Address is empty", slog.String("poller", opt.Poller))
		return nil, errs.New(errs.ErrMissingParam, "addr")
	}
	timeout, _ := time.ParseDuration(rest.DefaultTimeout)

	if a.Options.IsTest {
		return nil, nil
	}

	if client, err = rest.New(conf.ZapiPoller(c.Params), timeout, c.Auth); err != nil {
		return nil, fmt.Errorf("error creating new client: %w", err)
	}

	return client, err
}

func (c *CiscoRest) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Environment":
		return environment.New(abc)
	case "Interface":
		return networkinterface.New(abc)
	default:
		c.Logger.Warn("no cisco plugin found", slog.String("kind", kind))
	}
	return nil
}

func (c *CiscoRest) PollData() (map[string]*matrix.Matrix, error) {

	// Unlike the other collectors, the cisco collector does not use a template.
	// The plugins are responsible for collecting, parsing, and storing the data.
	c.client.Metadata.Reset()
	c.Metadata.Reset()

	return c.Matrix, nil
}

// Interface guards
var (
	_ collector.Collector = (*CiscoRest)(nil)
)
