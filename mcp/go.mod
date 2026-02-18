module mcp-server

go 1.26

replace github.com/netapp/harvest/v2 => ../

require (
	github.com/goccy/go-yaml v1.19.2
	github.com/modelcontextprotocol/go-sdk v1.3.0
	github.com/netapp/harvest/v2 v2.0.0-20260211133735-682c93ad66b7
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/google/jsonschema-go v0.4.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/oauth2 v0.35.0 // indirect
)
