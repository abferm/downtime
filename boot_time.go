package downtime

import (
	"fmt"
	"time"
)

// capture software start time
var startTime = time.Now()

func SystemBootTime() (time.Time, error) {
	bt, err := bootTime()
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to determine boot time: %w", err)
	}
	return bt, nil
}

func ProcessBootTime() time.Time {
	return startTime
}
