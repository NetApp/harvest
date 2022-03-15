//Copyright NetApp Inc, 2021 All rights reserved

/*
	Package plugin provides abstractions for plugins, as well as
	a number of generic built-in plugins. Plugins allow to customize
	and manipulate data from collectors and sometimes collect additional
	data without changing the sourcecode of collectors. Multiple plugins
	can be put in a pipeline, they are executed in the same order as they
	are defined in the collector's config file.
	Harvest architecture defines three types of plugins:

	**built-in**
    	Statically compiled, generic plugins. "Generic" means
    	the plugin is collector-agnostic. These plugins are
    	provided in this package.

	**generic**
  		These are generic plugins as well, but they are compiled
    	as shared objects and dynamically loaded. These plugins are
    	living in the directory src/plugins.

   **custom**
    	These plugins are collector-specific. Their source code should
    	reside inside the plugins/ subdirectory of the collector package.
    	Custom plugins have access to all the parameters of their parent
    	collector and should be therefore treated with great care.
*/
package plugin

import (
	"fmt"
	"goharvest2/cmd/poller/options"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logging"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"sync"
)

// Plugin defines the methods of a plugin
type Plugin interface {
	GetName() string
	Init() error
	Run(*matrix.Matrix) ([]*matrix.Matrix, error)
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
		panic(fmt.Sprintf("module already registered: %s", mod.ID))
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
	Parent       string
	Name         string
	Logger       *logging.Logger // logger used for logging
	Options      *options.Options
	Params       *node.Node
	ParentParams *node.Node
}

// New creates an AbstractPlugin with arguments:
// @parent	- name of the collector that owns this plugin
// @o		- poller options
// @p		- plugin parameters
// @pp		- parent collector parameters
func New(parent string, o *options.Options, p *node.Node, pp *node.Node) *AbstractPlugin {
	pl := AbstractPlugin{Parent: parent, Options: o, Params: p, ParentParams: pp}
	return &pl
}

// GetName returns the name of the plugin
func (p *AbstractPlugin) GetName() string {
	return p.Name
}

// Init initializes the plugin by calling InitAbc
func (p *AbstractPlugin) Init() error {
	return p.InitAbc()
}

// InitAbc initializes the plugin
func (p *AbstractPlugin) InitAbc() error {

	if p.ParentParams != nil {
		p.ParentParams.NewChildS("poller_name", p.Options.Poller)
	}
	if p.Name = p.Params.GetNameS(); p.Name == "" {
		return errors.New(errors.MISSING_PARAM, "plugin name")
	}
	p.Logger = logging.SubLogger("plugin", p.Parent+":"+p.Name)

	return nil
}

// Run should run the plugin and return collected data as an array of matrices
// (Since most plugins don't collect data, they will always return nil instead)
func (p *AbstractPlugin) Run(*matrix.Matrix) ([]*matrix.Matrix, error) {
	panic(p.Name + " has not implemented Run()")
}
