package downtime

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
)

type mockDataStore struct {
	stamp, shutdown, boot time.Time
	getErr, setErr        error
}

func (ds *mockDataStore) SetStamp(t time.Time) error {
	if ds.setErr != nil {
		return ds.setErr
	}
	ds.stamp = t
	return nil
}

func (ds *mockDataStore) GetStamp() (time.Time, error) {
	return ds.stamp, ds.getErr
}

func (ds *mockDataStore) SetShutdown(t time.Time) error {
	if ds.setErr != nil {
		return ds.setErr
	}
	ds.shutdown = t
	return nil
}

func (ds *mockDataStore) GetShutdown() (time.Time, error) {
	return ds.shutdown, ds.getErr
}

func (ds *mockDataStore) SetBoot(t time.Time) error {
	if ds.setErr != nil {
		return ds.setErr
	}
	ds.boot = t
	return nil
}

func (ds *mockDataStore) GetBoot() (time.Time, error) {
	return ds.boot, ds.getErr
}

func TestDaemonReporting(t *testing.T) {
	buff := bytes.NewBuffer([]byte{})
	writer := NewDatabaseWriter(buff)

	store := new(mockDataStore)

	clk := clock.NewMock()
	clk.Set(ProcessBootTime())

	expectedEvents := []Event{
		NewEvent(EventTypeCrash, ProcessBootTime().Add(time.Hour)),
		NewEvent(EventTypeUp, ProcessBootTime().Add(time.Hour+time.Minute)),
		NewEvent(EventTypeShutdown, ProcessBootTime().Add(time.Hour*2)),
		NewEvent(EventTypeUp, ProcessBootTime().Add((time.Hour*2)+time.Minute)),
	}

	d := NewDaemonWithClock(store, writer, DefaultSleepSeconds*time.Second, clk)

	store.getErr = fmt.Errorf("test error")
	err := d.Init(ProcessBootTime(), time.Stamp)
	assert.NoError(t, err)
	assert.Len(t, buff.Bytes(), 0, "should not report on first run")

	assert.Equal(t, ProcessBootTime(), store.boot)

	store.getErr = nil

	// generate first pair of events
	clk.Set(expectedEvents[0].When.AsTime())
	d.stamp(false)
	err = d.Init(expectedEvents[1].When.AsTime(), time.Stamp)
	assert.NoError(t, err)
	assert.Len(t, buff.Bytes(), EventSize*2, "there should be 2 events at this point")

	// generate second pair of events
	clk.Set(expectedEvents[2].When.AsTime())
	d.stamp(true)
	err = d.Init(expectedEvents[3].When.AsTime(), time.Stamp)
	assert.NoError(t, err)
	assert.Len(t, buff.Bytes(), EventSize*4, "there should be 4 events at this point")

	// verify events
	r := NewDatabaseReader(bytes.NewReader(buff.Bytes()))
	generatedEvents, err := r.All()
	assert.NoError(t, err)
	assert.Equal(t, expectedEvents, generatedEvents)
}
