package main

import (
	zapi_collector "goharvest2/collectors/zapi/collector"
	"goharvest2/poller/collector"
)

func New(a *collector.AbstractCollector) collector.Collector {
	return zapi_collector.New(a)
}
