module github.com/Netapp/harvest-automation

go 1.24.0

replace github.com/netapp/harvest/v2 => ../

require (
	github.com/carlmjohnson/requests v0.25.1
	github.com/netapp/harvest/v2 v2.0.0-20250915081038-7605bc16c55b
	golang.org/x/text v0.29.0
)

require (
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/cobra v1.10.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/term v0.35.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
