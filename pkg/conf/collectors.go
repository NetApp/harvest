package conf

import (
	"maps"
	"slices"
)

func GetCollectorSlice() []string {
	return slices.Collect(maps.Keys(IsCollector))
}

var IsCollector = map[string]struct{}{
	"AristaRest":  {},
	"CiscoRest":   {},
	"CmPerf":      {},
	"Ems":         {},
	"Eseries":     {},
	"EseriesPerf": {},
	"KeyPerf":     {},
	"Rest":        {},
	"RestPerf":    {},
	"Simple":      {},
	"StatPerf":    {},
	"StorageGrid": {},
	"Unix":        {},
	"Zapi":        {},
	"ZapiPerf":    {},
}

var IsONTAPCollector = map[string]struct{}{
	"ZapiPerf": {},
	"Zapi":     {},
	"Rest":     {},
	"RestPerf": {},
	"CmPerf":   {},
	"StatPerf": {},
	"KeyPerf":  {},
	"Ems":      {},
}

var IsESeriesCollector = map[string]struct{}{
	"Eseries":     {},
	"EseriesPerf": {},
}

var IsNonONTAPCollector = map[string]struct{}{
	"AristaRest":  {},
	"CiscoRest":   {},
	"StorageGrid": {},
	"Eseries":     {},
	"EseriesPerf": {},
}

func IsPingableCollector(collector string) bool {
	switch collector {
	case "Simple", "Unix":
		return false
	}
	return true
}
