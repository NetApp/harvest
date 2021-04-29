/*
 * Copyright NetApp Inc, 2021 All rights reserved

Package Description:

   Package plugin provides abstractions for plugins, as well as
   a number of generic built-in plugins. Plugins allow to customize
   and manipulate data from collectors and sometimes collect additional
   data without changing the sourcecode of collectors. Multiple plugins
   can be put in a pipeline, they are executed in the same order as they
   are defined in the collector's config file.
   Harvest architecuture defines three types of plugins:

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
	"goharvest2/cmd/poller/options"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logger"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
)

type Plugin interface {
	GetName() string
	GetType() string
	Init() error
	Run(*matrix.Matrix) ([]*matrix.Matrix, error)
}

type AbstractPlugin struct {
	Parent       string
	Name         string
	Prefix       string
	Type         string // @TODO: use or deprecate
	Options      *options.Options
	Params       *node.Node
	ParentParams *node.Node
}

func New(parent string, o *options.Options, p *node.Node, pp *node.Node) *AbstractPlugin {
	pl := AbstractPlugin{Parent: parent, Options: o, Params: p, ParentParams: pp}
	return &pl
}

func (me *AbstractPlugin) Init() error {
	return me.InitAbc()
}

func (me *AbstractPlugin) InitAbc() error {

	if me.Name = me.Params.GetNameS(); me.Name == "" {
		return errors.New(errors.MISSING_PARAM, "plugin name")
	}

	me.Prefix = "(plugin) (" + me.Parent + ":" + me.Name + ")"

	if me.Type = me.Params.GetChildContentS("type"); me.Type == "" {
		//return errors.New(errors.MISSING_PARAM, "plugin type")
	}

	logger.Trace(me.Prefix, "initialized")

	return nil
}

func (p *AbstractPlugin) GetName() string {
	return p.Name
}

func (p *AbstractPlugin) GetType() string {
	return p.Type
}

func (p *AbstractPlugin) Run(*matrix.Matrix) ([]*matrix.Matrix, error) {
	panic(p.Name + " has not implemented Run()")
}
