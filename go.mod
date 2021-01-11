module local.host/goharvest2

go 1.15

require (
	gopkg.in/yaml.v2 v2.4.0 // indirect
	local.host/api v0.0.0
	local.host/collector v0.0.0
	local.host/matrix v0.0.0
    local.host/share v0.0.0
)

replace (
	local.host/api => ./api
	local.host/collector => ./collector
	local.host/matrix => ./matrix
    loca.host/share => ./share
)
