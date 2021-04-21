package version

import (
	"fmt"
	"runtime"
)

var (
	VERSION = "2.0.2"
	RELEASE = "rc2"
	BUILD   = "src"
)

func String() string {
	return fmt.Sprintf("harvest version %s %s (%s build) %s/%s",
		VERSION,
		RELEASE,
		BUILD,
		runtime.GOOS,
		runtime.GOARCH,
	)
}
