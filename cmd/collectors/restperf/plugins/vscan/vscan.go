package vscan

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"log/slog"
)

type Vscan struct {
	*plugin.AbstractPlugin
	isPerScanner bool
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Vscan{AbstractPlugin: p}
}

func (v *Vscan) Init(conf.Remote) error {
	if err := v.InitAbc(); err != nil {
		return err
	}

	// parsed once at startup rather than on every poll
	v.isPerScanner = v.LoadParam("metricsPerScanner", true)

	return nil
}

func (v *Vscan) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[v.Object]

	v.addSvmAndScannerLabels(data)
	if !v.isPerScanner {
		return nil, nil, nil
	}

	return collectors.AggregatePerScanner(v.SLogger, data, "scan.latency", "scan.request_dispatched_rate")
}

func (v *Vscan) addSvmAndScannerLabels(data *matrix.Matrix) {
	for _, instance := range data.GetInstances() {
		ontapName := instance.GetLabel("id")
		names, ok := collectors.SplitVscanName(ontapName, false)
		if !ok {
			v.SLogger.Warn("Failed to parse svm and scanner labels", slog.String("ontapName", ontapName))
			continue
		}
		instance.SetLabel("svm", names.Svm)
		instance.SetLabel("scanner", names.Scanner)
		instance.SetLabel("node", names.Node)
	}
}
