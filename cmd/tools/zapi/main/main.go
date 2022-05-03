package main

import (
	"github.com/netapp/harvest/v2/cmd/tools/zapi"
	"github.com/spf13/cobra"
)

func main() {
	cobra.CheckErr(zapi.Cmd.Execute())
}
