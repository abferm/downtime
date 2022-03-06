//go:generate go-enum -f=$GOFILE --marshal

package downtime

import (
	"fmt"
	"time"
)

/*ENUM(
None = 0
Up = 1
Shutdown = 2
Crash = 3
)
*/
type EventType uint8

type UnixTimestamp int64

func (ut UnixTimestamp) AsTime() time.Time {
	return time.Unix(int64(ut), 0)
}

func (ut UnixTimestamp) String() string {
	return ut.AsTime().String()
}

func NewEvent(what EventType, when time.Time) Event {
	return Event{
		What: what,
		When: UnixTimestamp(when.Unix()),
	}
}

type Event struct {
	What EventType
	_    [7]uint8 // padding
	When UnixTimestamp
}

func (e Event) String() string {
	return fmt.Sprintf("%s at %s", e.What, e.When)
}
