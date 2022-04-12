/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package svm

import (
	"goharvest2/cmd/collectors"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/matrix"
)

type SVM struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SVM{AbstractPlugin: p}
}

func (my *SVM) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	collectors.SetNameservice(data)
	return nil, nil
}
