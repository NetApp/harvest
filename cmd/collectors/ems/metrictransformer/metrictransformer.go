package metrictransformer

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/util"
	"sort"
	"strings"
)

type MetricTransformer struct {
	*plugin.AbstractPlugin
	excludeKeys *set.Set
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &MetricTransformer{AbstractPlugin: p}
}

func (m *MetricTransformer) Init(conf.Remote) error {
	if err := m.InitAbc(); err != nil {
		return fmt.Errorf("failed to initialize AbstractPlugin: %w", err)
	}
	m.excludeKeys = set.New()

	if exportOption := m.ParentParams.GetChildS("exports"); exportOption != nil {
		for _, c := range exportOption.GetAllChildContentS() {
			if c != "" {
				_, display, _, _ := util.ParseMetric(c)
				m.excludeKeys.Add(display)
			}
		}
	}
	return nil
}

func (m *MetricTransformer) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	for _, data := range dataMap {
		for _, instance := range data.GetInstances() {
			if !instance.IsExportable() {
				continue
			}
			var keys []string
			for key := range instance.GetLabels() {
				if !m.excludeKeys.Has(key) {
					keys = append(keys, key)
				}
			}
			sort.Strings(keys)
			var parameters []string
			for _, key := range keys {
				value := instance.GetLabel(key)
				if value != "" {
					parameters = append(parameters, key+"="+value)
				}
				instance.RemoveLabel(key)
			}
			parametersLabel := strings.Join(parameters, ", ")
			if parametersLabel != "" {
				instance.SetLabel("parameters", parametersLabel)
			}
		}
	}
	return nil, nil, nil
}
