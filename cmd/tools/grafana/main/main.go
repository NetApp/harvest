/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package main

import (
	"github.com/netapp/harvest/v2/cmd/tools/grafana"
	"github.com/spf13/cobra"
)

func main() {
	cobra.CheckErr(grafana.Cmd.Execute())
}
