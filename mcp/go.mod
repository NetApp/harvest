module mcp-server

go 1.25

replace github.com/netapp/harvest/v2 => ../

require (
	github.com/goccy/go-yaml v1.18.0
	github.com/modelcontextprotocol/go-sdk v1.0.0
	github.com/netapp/harvest/v2 v2.0.0-20251010134815-9b0be64ee6ac
	github.com/spf13/cobra v1.10.1
)

require (
	github.com/google/jsonschema-go v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
)
