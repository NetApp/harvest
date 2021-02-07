package plugin

import (
	"goharvest2/poller/struct/matrix"
	"goharvest2/poller/struct/options"
	"goharvest2/poller/struct/yaml"
	"goharvest2/poller/errors"
)

type Plugin interface {
	GetName() string
	GetType() string
	Init() error
	Run(*matrix.Matrix) (*matrix.Matrix, error)
}


type AbstractPlugin struct {
	Parent string
	Name string
	Type string
	Options *options.Options
	Params *yaml.Node
}

func New(parent string, o *options.Options, p *yaml.Node) *AbstractPlugin {
	pl := AbstractPlugin{Parent: parent, Options: o, Params: p}
	return &pl
}

func (p *AbstractPlugin) Init() error {
	return p.InitAbc()
}


func (p *AbstractPlugin) InitAbc() error {

	if p.Name = p.Params.Name; p.Name == "" {
		return errors.New(errors.MISSING_PARAM, "plugin name")
	}

	if p.Type = p.Params.GetChildValue("type"); p.Type == "" {
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

func (p *AbstractPlugin) Run(*matrix.Matrix) (*matrix.Matrix, error) {
	panic(p.Name + " has not implemented Run()")
}