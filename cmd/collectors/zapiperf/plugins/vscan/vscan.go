package vscan

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"strconv"
)

type Vscan struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Vscan{AbstractPlugin: p}
}

func (v *Vscan) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[v.Object]
	// defaults plugin options
	isPerScanner := true

	if s := v.Params.GetChildContentS("metricsPerScanner"); s != "" {
		if parseBool, err := strconv.ParseBool(s); err == nil {
			isPerScanner = parseBool
		} else {
			v.Logger.Error().Err(err).Msg("Failed to parse metricsPerScanner")
		}
	}
	v.Logger.Debug().Bool("isPerScanner", isPerScanner).Msg("Vscan options")

	v.addSvmAndScannerLabels(data)
	if !isPerScanner {
		return nil, nil, nil
	}

	return collectors.AggregatePerScanner(v.Logger, data, "scan_latency", "scan_request_dispatched_rate")
}

func (v *Vscan) addSvmAndScannerLabels(data *matrix.Matrix) {
	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		ontapName := instance.GetLabel("instance_uuid")
		names, ok := collectors.SplitVscanName(ontapName, true)
		if !ok {
			v.Logger.Warn().Str("ontapName", ontapName).Msg("Failed to parse svm and scanner labels")
			continue
		}
		instance.SetLabel("svm", names.Svm)
		instance.SetLabel("scanner", names.Scanner)
		instance.SetLabel("node", names.Node)
	}
}
