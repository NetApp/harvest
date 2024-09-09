package collectors

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
)

const (
	shelfPrefix = "shelf"
	aggrPrefix  = "aggr"
)

func GetCommonMetrics() []plugin.CustomMetric {
	return []plugin.CustomMetric{
		{
			Name:         "power",
			Endpoint:     "NA",
			ONTAPCounter: "HarvestGenerated",
			Description:  "Power consumed by a shelf in Watts.",
			Prefix:       shelfPrefix,
		},
		{
			Name:         "minTemperature",
			Endpoint:     "NA",
			ONTAPCounter: "HarvestGenerated",
			Description:  "Minimum temperature of all non-ambient sensors for shelf in Celsius.",
			Prefix:       shelfPrefix,
		},
		{
			Name:         "averageAmbientTemperature",
			Endpoint:     "NA",
			ONTAPCounter: "HarvestGenerated",
			Description:  "Average temperature of all ambient sensors for shelf in Celsius.",
			Prefix:       shelfPrefix,
		},
		{
			Name:         "averageFanSpeed",
			Endpoint:     "NA",
			ONTAPCounter: "HarvestGenerated",
			Description:  "Average fan speed for shelf in rpm.",
			Prefix:       shelfPrefix,
		},
		{
			Name:         "averageTemperature",
			Endpoint:     "NA",
			ONTAPCounter: "HarvestGenerated",
			Description:  "Average temperature of all non-ambient sensors for shelf in Celsius.",
			Prefix:       shelfPrefix,
		},
		{
			Name:         "maxFanSpeed",
			Endpoint:     "NA",
			ONTAPCounter: "HarvestGenerated",
			Description:  "Maximum fan speed for shelf in rpm.",
			Prefix:       shelfPrefix,
		},
		{
			Name:         "maxTemperature",
			Endpoint:     "NA",
			ONTAPCounter: "HarvestGenerated",
			Description:  "Maximum temperature of all non-ambient sensors for shelf in Celsius.",
			Prefix:       shelfPrefix,
		},
		{
			Name:         "minAmbientTemperature",
			Endpoint:     "NA",
			ONTAPCounter: "HarvestGenerated",
			Description:  "Minimum temperature of all ambient sensors for shelf in Celsius.",
			Prefix:       shelfPrefix,
		},
		{
			Name:         "minFanSpeed",
			Endpoint:     "NA",
			ONTAPCounter: "HarvestGenerated",
			Description:  "Minimum fan speed for shelf in rpm.",
			Prefix:       shelfPrefix,
		},
		{
			Name:         "power",
			Endpoint:     "NA",
			ONTAPCounter: "HarvestGenerated",
			Description:  "Power consumed by aggregate in Watts.",
			Prefix:       aggrPrefix,
		},
	}
}
