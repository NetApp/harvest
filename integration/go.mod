module github.com/Netapp/harvest-automation

go 1.24

toolchain go1.24.0

replace github.com/netapp/harvest/v2 => ../

require (
	github.com/carlmjohnson/requests v0.24.3
	github.com/netapp/harvest/v2 v2.0.0-20250811103824-ebd8e38d121e
	golang.org/x/text v0.28.0
)

require (
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.7 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/term v0.34.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
