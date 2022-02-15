module github.com/Netapp/harvest-automation

go 1.16

replace goharvest2 => ../

require (
	github.com/docker/docker v20.10.12+incompatible
	github.com/julienroland/usg v0.0.0-20160918114137-cb52eabb3d84
	github.com/rs/zerolog v1.26.1
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.13.0
	goharvest2 v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	github.com/containerd/containerd v1.5.9 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
)
