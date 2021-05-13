package main

import (
	"github.com/spf13/cobra"
	"goharvest2/cmd/tools/zapi"
)

func main() {
	cobra.CheckErr(zapi.ZapiCmd.Execute())
}
