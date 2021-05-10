/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package main

import (
	"goharvest2/pkg/set"
)

var knownTypes = set.NewFrom([]string{
	"string",
	"integer",
	"boolean",
	"node-name",
	"aggr-name",
	"vserver-name",
	"volume-name",
	"uuid", "size",
	"cache-policy",
	"junction-path",
	"volstyle",
	"repos-constituent-role",
	"language-code",
	"snaplocktype",
	"space-slo-enum",
})
