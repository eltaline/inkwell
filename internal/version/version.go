package version

import (
	"fmt"
	"runtime"
)

// Set via -ldflags at build time.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func String() string {
	return fmt.Sprintf("inkwell %s (commit %s, built %s, %s/%s)",
		Version, Commit, BuildDate, runtime.GOOS, runtime.GOARCH)
}
