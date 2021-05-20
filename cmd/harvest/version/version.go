/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package version

import (
	"fmt"
	"runtime"
)

var (
	VERSION   = "2.0.2"
	Release   = "rc2"
	Commit    = "HEAD"
	BuildDate = "undefined"
)

func String() string {
	return fmt.Sprintf("harvest version %s %s (commit %s) (build date %s) %s/%s\n",
		VERSION,
		Release,
		Commit,
		BuildDate,
		runtime.GOOS,
		runtime.GOARCH,
	)
}
