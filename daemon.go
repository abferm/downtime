package downtime

import (
	"context"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/leekchan/timeutil"
)

func NewDaemon(dataStore DataStore, database EventWriter, sleep time.Duration) *Daemon {
	return NewDaemonWithClock(dataStore, database, sleep, clock.New())
}

func NewDaemonWithClock(dataStore DataStore, database EventWriter, sleep time.Duration, clk clock.Clock) *Daemon {
	return &Daemon{
		dataStore: dataStore,
		database:  database,
		sleep:     sleep,
		clk:       clk,
	}
}

type Daemon struct {
	dataStore DataStore
	database  EventWriter
	sleep     time.Duration
	clk       clock.Clock
}

func (d Daemon) Init(bootTime time.Time, timeFormat string) error {
	err := d.report(bootTime, timeFormat)
	if err != nil {
		return fmt.Errorf("error reporting: %w", err)
	}
	err = d.dataStore.SetBoot(bootTime)
	if err != nil {
		return fmt.Errorf("error updating boot time: %w", err)
	}
	return nil
}

func (d Daemon) Run(ctx context.Context) error {
	d.stamp(false)
	for {
		select {
		case <-d.clk.After(d.sleep):
			d.stamp(false)
		case <-ctx.Done():
			d.stamp(true)
			return ctx.Err()
		}
	}
}

func (d Daemon) stamp(shutdown bool) {
	err := d.dataStore.SetStamp(d.clk.Now())
	if err != nil {
		logger.Errorf("failed to update stamp: %w", err)
	}
	if shutdown {
		err = d.dataStore.SetShutdown(d.clk.Now())
		if err != nil {
			logger.Errorf("failed to update shutdown: %w", err)
		}
	}
}

func (d Daemon) updateDatabase(down, up time.Time, crashed bool) error {
	downEvt := NewEvent(EventTypeShutdown, down)
	if crashed {
		downEvt.What = EventTypeCrash
	}
	upEvt := NewEvent(EventTypeUp, up)

	err := d.database.Append(downEvt)
	if err != nil {
		return err
	}
	return d.database.Append(upEvt)
}

func (d Daemon) report(bootTime time.Time, timeFormat string) error {
	var stamp, shutdown, oldBoot time.Time
	var haveStamp, haveShutdown, haveOldBoot bool
	var oldUptime, downtime time.Duration

	stamp, err := d.dataStore.GetStamp()
	if err != nil {
		logger.Warningf("could not read old stamp: %s", err)
		haveStamp = false
	} else {
		haveStamp = true
	}

	shutdown, err = d.dataStore.GetShutdown()
	if err != nil {
		logger.Warningf("could not read old shutdown: %s", err)
		haveShutdown = false
	} else {
		haveShutdown = true
	}

	oldBoot, err = d.dataStore.GetBoot()
	if err != nil {
		logger.Warningf("could not read old boot: %s", err)
		haveOldBoot = false
	} else {
		haveOldBoot = true
	}

	if !haveStamp && !haveShutdown && !haveOldBoot {
		logger.Infof("starting up first time, no knowledge of downtime")
		return nil
	}

	if !haveStamp {
		return fmt.Errorf("no old run-time stamp")
	}

	if !haveOldBoot {
		return fmt.Errorf("no old boot-time stamp")
	}

	if haveStamp && haveShutdown && shutdown.Before(stamp) {
		haveShutdown = false
	}

	if haveShutdown {
		oldUptime = shutdown.Sub(oldBoot)
		downtime = bootTime.Sub(shutdown)
	} else {
		oldUptime = stamp.Sub(oldBoot)
		downtime = bootTime.Sub(stamp)
	}

	if downtime < 0 {
		/*
			* This happens if we quit and re-start the process (we
				* normally only exit when system goes down.
		*/
		logger.Infof("daemon restarted, no downtime")
		return nil
	}

	if haveShutdown {
		logger.Infof("shutdown at %s", timeutil.Strftime(&shutdown, timeFormat))
		err = d.updateDatabase(shutdown, bootTime, false)
	} else {
		logger.Infof("crashed at %s", stamp.Format(timeFormat))
		err = d.updateDatabase(stamp, bootTime, true)
	}
	logger.Infof("previous uptime was %s (%d seconds)", oldUptime.String(), int(oldUptime.Seconds()))
	logger.Infof("downtime was %s (%d seconds", downtime.String(), int(downtime.Seconds()))
	return err
}
