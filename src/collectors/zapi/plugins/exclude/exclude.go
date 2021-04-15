package main

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/dict"
	"goharvest2/share/errors"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
)

type Exclude struct {
	*plugin.AbstractPlugin
	labels *dict.Dict
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Exclude{AbstractPlugin: p}
}

func (me *Exclude) Init() error {

	if err := me.AbstractPlugin.Init(); err != nil {
		return err
	}

	me.labels = dict.New()

	if len(me.Params.GetChildren()) == 0 {
		return errors.New(errors.MISSING_PARAM, "exclude parameters")
	}

	for _, x := range me.Params.GetChildren() {
		me.labels.Set(x.GetNameS(), x.GetContentS())
		logger.Debug(me.Prefix, "added exclude for label (%s) = (%s)", x.GetNameS(), x.GetContentS())
	}

	logger.Debug(me.Prefix, "will exclude instances based on %d label values", me.labels.Size())
	return nil
}

func (me *Exclude) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	count := 0

	for _, instance := range data.GetInstances() {
		for label, value := range me.labels.Map() {
			if instance.GetLabel(label); label == value {
				instance.SetExportable(false)
				count += 1
				break
			}
		}
	}

	logger.Debug(me.Prefix, "hahaha, exluded %d instances", count)
	return nil, nil
}
