package main

import (
	zapi "goharvest2/collectors/zapi/collector"
	"goharvest2/poller/collector"
)

func New(a *collector.AbstractCollector) collector.Collector {
	return zapi.New(a)
}
