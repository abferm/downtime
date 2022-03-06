//go:build !linux && !darwin && !freebsd && !netbsd && !openbsd && !dragonfly
// +build !linux,!darwin,!freebsd,!netbsd,!openbsd,!dragonfly

package downtime

import (
	"fmt"
	"runtime"
	"time"
)

func bootTime() (time.Time, error) {
	return time.Time{}, fmt.Errorf("os not supported: %s", runtime.GOOS)
}
