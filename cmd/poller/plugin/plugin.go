// Copyright NetApp Inc, 2021 All rights reserved
// Package plugin provides abstractions for plugins, as well as
// a number of generic built-in plugins. Plugins allow customizing
// and manipulating data from collectors and sometimes collect additional
// data without changing the sourcecode of collectors. Multiple plugins
// can be put in a pipeline, they are executed in the same order as they
// are defined in the collector's config file.
// Harvest architecture defines three types of plugins:
//
// **built-in**
// 	Statically compiled, generic plugins. "Generic" means
// 	the plugin is collector-agnostic. These plugins are
// 	provided in this package.
//
// **generic**
// 	These are generic plugins as well, but they are compiled
// 	as shared objects and dynamically loaded. These plugins are
// 	living in the directory src/plugins.
//
// **custom**
// 	These plugins are collector-specific. Their source code should
// 	reside inside the plugins/ subdirectory of the collector package.
// 	Custom plugins have access to all the parameters of their parent
// 	collector and should be therefore treated with great care.

package plugin

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"sync"
	"time"
)

const DefaultPluginInterval = 30 * time.Minute
const DefaultPollInterval = 3 * time.Minute

// Plugin defines the methods of a plugin
type Plugin interface {
	GetName() string
	Init(conf.Remote) error
	Run(map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error)
}

var (
	modules   = make(map[string]ModuleInfo)
	modulesMu sync.RWMutex
)

// GetModule returns module information from its ID (full name).
func GetModule(name string) (ModuleInfo, error) {
	modulesMu.RLock()
	defer modulesMu.RUnlock()
	m, ok := modules[name]
	if !ok {
		return ModuleInfo{}, fmt.Errorf("module not registered: %s", name)
	}
	return m, nil
}

func RegisterModule(instance Module) {
	mod := instance.HarvestModule()

	if mod.ID == "" {
		panic("module missing ID")
	}
	if mod.ID == "harvest" || mod.ID == "admin" {
		panic(fmt.Sprintf("module ID '%s' is reserved", mod.ID))
	}
	if mod.New == nil {
		panic("missing ModuleInfo.New")
	}
	if val := mod.New(); val == nil {
		panic("ModuleInfo.New must return a non-nil module instance")
	}
	modulesMu.Lock()
	defer modulesMu.Unlock()

	if _, ok := modules[mod.ID]; ok {
		panic("module already registered: " + mod.ID)
	}
	modules[mod.ID] = mod
}

type Module interface {
	HarvestModule() ModuleInfo
}

type ModuleInfo struct {
	// name of module
	ID string

	New func() Module
}

// AbstractPlugin implements methods of the Plugin interface, except Run()
type AbstractPlugin struct {
	Parent               string           // name of the collector that owns this plugin
	Name                 string           // name of the plugin
	Object               string           // object of the collector, describes what that collector is collecting
	SLogger              *slog.Logger     // logger used for logging
	Options              *options.Options // poller options
	Params               *node.Node       // plugin parameters
	ParentParams         *node.Node       // parent collector parameters
	PluginInvocationRate int
	Auth                 *auth.Credentials
}

// New creates an AbstractPlugin
func New(parent string, o *options.Options, p *node.Node, pp *node.Node, object string, credentials *auth.Credentials) *AbstractPlugin {
	return &AbstractPlugin{
		Parent:       parent,
		Options:      o,
		Params:       p,
		ParentParams: pp,
		Object:       object,
		Auth:         credentials,
	}
}

// GetName returns the name of the plugin
func (p *AbstractPlugin) GetName() string {
	return p.Name
}

// Init initializes the plugin by calling InitAbc
func (p *AbstractPlugin) Init(conf.Remote) error {
	return p.InitAbc()
}

// InitAbc initializes the plugin
func (p *AbstractPlugin) InitAbc() error {

	if p.ParentParams != nil {
		p.ParentParams.NewChildS("poller_name", p.Options.Poller)
	}
	if p.Name = p.Params.GetNameS(); p.Name == "" {
		return errs.New(errs.ErrMissingParam, "plugin name")
	}
	p.SLogger = slog.Default().With(slog.String("plugin", p.Parent+":"+p.Name), slog.String("object", p.Object))
	return nil
}

// Run should run the plugin and return collected data as an array of matrices
// (Since most plugins don't collect data, they will always return nil instead)
func (p *AbstractPlugin) Run(map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	panic(p.Name + " has not implemented Run()")
}

func (p *AbstractPlugin) SetPluginInterval() int {
	pollInterval := p.GetInterval(p.ParentParams, DefaultPollInterval)
	pluginInterval := p.GetInterval(p.Params, DefaultPluginInterval)
	p.PluginInvocationRate = int(pluginInterval / pollInterval)
	p.SLogger.Debug(
		"SetPluginInterval",
		slog.Float64("PollInterval", pollInterval),
		slog.Float64("PluginInterval", pluginInterval),
		slog.Int("PluginInvocationRate", p.PluginInvocationRate),
	)
	return p.PluginInvocationRate
}

func (p *AbstractPlugin) GetInterval(param *node.Node, defaultInterval time.Duration) float64 {
	if param != nil {
		schedule := param.GetChildS("schedule")
		if schedule != nil {
			dataInterval := schedule.GetChildContentS("data")
			if dataInterval != "" {
				if durationVal, err := time.ParseDuration(dataInterval); err == nil {
					return durationVal.Seconds()
				}
			}
		}
	}

	return defaultInterval.Seconds()
}

type DerivedMetric struct {
	Name          string
	Source        string
	IsMax         bool
	HasCustomName bool
}
