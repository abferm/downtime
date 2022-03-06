//go:build darwin || freebsd || netbsd || openbsd || dragonfly
// +build darwin freebsd netbsd openbsd dragonfly

package downtime

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

func bootTime() (time.Time, error) {
	out, err := unix.SysctlRaw("kern.boottime")
	if err != nil {
		return time.Time{}, err
	}
	var timeval syscall.Timeval
	if len(out) != int(unsafe.Sizeof(timeval)) {
		return time.Time{}, fmt.Errorf("unexpected output of sysctl kern.boottime: %v (len: %d)", out, len(out))
	}
	timeval = *(*syscall.Timeval)(unsafe.Pointer(&out[0]))
	sec, nano := timeval.Unix()
	return time.Unix(sec, nano), nil
}
