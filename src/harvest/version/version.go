package version

import (
	"fmt"
	"runtime"
)

var (
	VERSION = "2.0.2"
	RELEASE = "RC2"
	BUILD   = "SRC"
)

func String() string {
	return fmt.Sprintf("harvest version %s %s/%s - %s/%s",
		VERSION,
		RELEASE,
		BUILD,
		runtime.GOOS,
		runtime.GOARCH,
	)
}
