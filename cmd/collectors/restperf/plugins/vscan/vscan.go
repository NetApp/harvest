package vscan

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"strconv"
)

type Vscan struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Vscan{AbstractPlugin: p}
}

func (v *Vscan) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[v.Object]
	// defaults plugin options
	isPerScanner := true

	if s := v.Params.GetChildContentS("metricsPerScanner"); s != "" {
		if parseBool, err := strconv.ParseBool(s); err == nil {
			isPerScanner = parseBool
		} else {
			v.SLogger.Error("Failed to parse metricsPerScanner", slogx.Err(err))
		}
	}
	v.SLogger.Debug("Vscan options", slog.Bool("isPerScanner", isPerScanner))

	v.addSvmAndScannerLabels(data)
	if !isPerScanner {
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
