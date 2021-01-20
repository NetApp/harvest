module local.host/poller

go 1.15

require (
    local.host/args v0.0.0
	local.host/client v0.0.0
	local.host/collector v0.0.0
	local.host/exporter v0.0.0
    local.host/schedule v0.0.0
    local.host/xml v0.0.0
    local.host/yaml v0.0.0
    local.host/matrix v0.0.0
    local.host/share v0.0.0
    local.host/logger v0.0.0
)

replace (
    local.host/args => ./args
	local.host/client => ./client
	local.host/collector => ./collector
	local.host/exporter => ./exporter
	local.host/schedule => ./schedule
	local.host/xml => ./xml
    local.host/yaml => ./yaml
	local.host/matrix => ./matrix
    local.host/share => ./share
    local.host/logger => ./logger
)
