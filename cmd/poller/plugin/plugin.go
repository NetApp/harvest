/*
 * Copyright NetApp Inc, 2021 All rights reserved
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
