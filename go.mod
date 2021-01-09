module local.host/goharvest2

go 1.15

require (
    local.host/lib v0.0.0
)

replace (
    local.host/lib => ./lib
)
