/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package main

import (
	"github.com/spf13/cobra"
	"goharvest2/cmd/tools/grafana"
)

func main() {
	cobra.CheckErr(grafana.GrafanaCmd.Execute())
}
