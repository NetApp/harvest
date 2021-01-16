module local.host/goharvest2

go 1.15

require (
	local.host/params v0.0.0
	local.host/client v0.0.0
	local.host/collector v0.0.0
	local.host/exporter v0.0.0
    local.host/xmltree v0.0.0
    local.host/matrix v0.0.0
    local.host/template v0.0.0
    local.host/share v0.0.0
)

replace (
	local.host/params => ./params
	local.host/client => ./client
	local.host/collector => ./collector
	local.host/exporter => ./exporter
	local.host/xmltree => ./xmltree
	local.host/matrix => ./matrix
    local.host/template => ./template
    local.host/share => ./share
)
