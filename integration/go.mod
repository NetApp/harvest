module github.com/Netapp/harvest-automation

go 1.24

toolchain go1.24.0

replace github.com/netapp/harvest/v2 => ../

require (
	github.com/carlmjohnson/requests v0.24.3
	github.com/netapp/harvest/v2 v2.0.0-20250509140426-b7cfd3aabda0
	golang.org/x/text v0.25.0
)

require (
	github.com/goccy/go-yaml v1.17.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
