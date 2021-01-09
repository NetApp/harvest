module local.host/goharvest2

go 1.15

require (
    local.host/api v0.0.0
    local.host/matrix v0.0.0
    local.host/collector v0.0.0
)

replace (
    local.host/api => ./api
    local.host/matrix => ./matrix
    local.host/collector => ./collector
)
