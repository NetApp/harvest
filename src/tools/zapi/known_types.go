
package main

import (
	"goharvest2/share/set"
)

var KNOWN_TYPES = set.NewFrom([]string{
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