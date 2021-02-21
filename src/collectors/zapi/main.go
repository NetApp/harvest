package main

import (
    "goharvest2/poller/collector"
	zapi_collector "goharvest2/collectors/zapi/collector"
)

func New(a *collector.AbstractCollector) collector.Collector {
	return zapi_collector.New(a)
}