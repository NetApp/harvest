module mcp-server

go 1.25

replace github.com/netapp/harvest/v2 => ../

require (
	github.com/goccy/go-yaml v1.19.2
	github.com/modelcontextprotocol/go-sdk v1.2.0
	github.com/netapp/harvest/v2 v2.0.0-20260108145919-b9575e5b0cef
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/google/jsonschema-go v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/oauth2 v0.32.0 // indirect
)
