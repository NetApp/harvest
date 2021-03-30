//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package main

import (
    zapi_collector "goharvest2/cmd/collectors/zapi/collector"
    "goharvest2/cmd/poller/collector"
)

func New(a *collector.AbstractCollector) collector.Collector {
    return zapi_collector.New(a)
}
