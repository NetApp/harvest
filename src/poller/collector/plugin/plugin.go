package plugin

import (
	"goharvest2/share/matrix"
	"goharvest2/poller/options"
	"goharvest2/share/tree/node"
	"goharvest2/share/errors"
)

type Plugin interface {
	GetName() string
	GetType() string
	Init() error
	Run(*matrix.Matrix) ([]*matrix.Matrix, error)
}


type AbstractPlugin struct {
	Parent string
	Name string
	Prefix string
	Type string
	Options *options.Options
	Params *node.Node
	ParentParams *node.Node
}

func New(parent string, o *options.Options, p *node.Node, pp *node.Node) *AbstractPlugin {
	pl := AbstractPlugin{Parent: parent, Options: o, Params: p, ParentParams: pp}
	return &pl
}

func (p *AbstractPlugin) Init() error {
	return p.InitAbc()
}


func (p *AbstractPlugin) InitAbc() error {

	if p.Name = p.Params.GetNameS(); p.Name == "" {
		return errors.New(errors.MISSING_PARAM, "plugin name")
	}

	p.Prefix = "(plugin) (" + p.Parent + ":" + p.Name + ")"

	if p.Type = p.Params.GetChildContentS("type"); p.Type == "" {
		return errors.New(errors.MISSING_PARAM, "plugin type")
	}

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
