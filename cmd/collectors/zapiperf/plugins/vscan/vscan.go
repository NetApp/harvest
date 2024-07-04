package vscan

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"strconv"
	"strings"
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

	return v.aggregatePerScanner(data)
}

func (v *Vscan) addSvmAndScannerLabels(data *matrix.Matrix) {
	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		ontapName := instance.GetLabel("instance_uuid")
		svm, scanner, node, ok := collectors.SplitVscanName(ontapName)
		if !ok {
			v.Logger.Warn().Str("ontapName", ontapName).Msg("Failed to parse svm and scanner labels")
			continue
		}
		instance.SetLabel("svm", svm)
		instance.SetLabel("scanner", scanner)
		instance.SetLabel("node", node)
	}
}

func (v *Vscan) aggregatePerScanner(data *matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	// When isPerScanner=true, Harvest 1.6 uses this form:
	// netapp.perf.dev.nltl-fas2520.vscan.scanner.10_64_30_62.scanner_stats_pct_mem_used 18 1501765640

	// These counters are per scanner and need averaging:
	// 		scanner_stats_pct_cpu_used
	// 		scanner_stats_pct_mem_used
	// 		scanner_stats_pct_network_used
	// These counters need to be summed:
	// 		scan_request_dispatched_rate

	// create per scanner instance cache
	cache := data.Clone(matrix.With{Data: false, Metrics: true, Instances: false, ExportInstances: true})
	cache.UUID += ".Vscan"

	for _, i := range data.GetInstances() {
		if !i.IsExportable() {
			continue
		}
		scanner := i.GetLabel("scanner")
		if cache.GetInstance(scanner) == nil {
			s, _ := cache.NewInstance(scanner)
			s.SetLabel("scanner", scanner)
		}
		i.SetExportable(false)
	}

	// aggregate per scanner
	counts := make(map[string]map[string]int) // map[scanner][counter] => value

	for _, i := range data.GetInstances() {
		scanner := i.GetLabel("scanner")
		ps := cache.GetInstance(scanner)
		if ps == nil {
			v.Logger.Error().Str("scanner", scanner).Msg("Failed to find scanner instance in cache")
			continue
		}
		_, ok := counts[scanner]
		if !ok {
			counts[scanner] = make(map[string]int)
		}
		for mKey, m := range data.GetMetrics() {
			if !m.IsExportable() && m.GetType() != "float64" {
				continue
			}
			psm := cache.GetMetric(mKey)
			if psm == nil {
				v.Logger.Error().Str("scanner", scanner).Str("metric", mKey).
					Msg("Failed to find metric in scanner cache")
				continue
			}
			if value, ok := m.GetValueFloat64(i); ok {
				fv, _ := psm.GetValueFloat64(ps)

				// sum for scan_request_dispatched_rate
				if mKey == "scan_request_dispatched_rate" {
					err := psm.SetValueFloat64(ps, fv+value)
					if err != nil {
						v.Logger.Error().Err(err).Str("metric", "scan_request_dispatched_rate").
							Msg("Error setting metric value")
					}

					continue
				} else if strings.HasSuffix(mKey, "_used") {
					// these need averaging
					counts[scanner][mKey]++
					runningTotal, _ := psm.GetValueFloat64(ps)
					value, _ := m.GetValueFloat64(ps)
					err := psm.SetValueFloat64(ps, runningTotal+value)
					if err != nil {
						v.Logger.Error().Err(err).Str("mKey", mKey).Msg("Failed to set value")
					}
				}
			}
		}
	}

	// cook averaged values
	for scanner, i := range cache.GetInstances() {
		if !i.IsExportable() {
			continue
		}
		for mKey, m := range cache.GetMetrics() {
			if m.IsExportable() && strings.HasSuffix(m.GetName(), "_used") {
				count := counts[scanner][mKey]
				value, ok := m.GetValueFloat64(i)
				if !ok {
					continue
				}
				if err := m.SetValueFloat64(i, value/float64(count)); err != nil {
					v.Logger.Error().Err(err).
						Str("mKey", mKey).
						Str("name", m.GetName()).
						Msg("Unable to set average")
				}
			}
		}
	}

	return []*matrix.Matrix{cache}, nil, nil
}
