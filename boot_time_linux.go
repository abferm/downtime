package downtime

import (
	"time"

	"github.com/prometheus/procfs"
)

func bootTime() (time.Time, error) {
	// if linux, use procfs to get
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		return time.Time{}, err
	}
	stat, err := fs.Stat()
	return time.Unix(int64(stat.BootTime), 0), err
}
