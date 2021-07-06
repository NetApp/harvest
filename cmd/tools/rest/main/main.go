// Copyright NetApp Inc, 2021 All rights reserved

package main

import (
	"github.com/spf13/cobra"
	"goharvest2/cmd/tools/rest"
)

func main() {
	cobra.CheckErr(rest.Cmd.Execute())
}
