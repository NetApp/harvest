package conf

import (
	"maps"
	"slices"
)

func GetCollectorSlice() []string {
	return slices.Collect(maps.Keys(IsCollector))
}

var IsCollector = map[string]struct{}{
	"CiscoRest":   {},
	"Ems":         {},
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
	"StatPerf": {},
	"KeyPerf":  {},
	"Ems":      {},
}

func IsPingableCollector(collector string) bool {
	switch collector {
	case "Simple", "Unix":
		return false
	}
	return true
}
