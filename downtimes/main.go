package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/abferm/downtime"
	"github.com/juju/loggo"
)

var logger = loggo.GetLogger("")

func main() {
	err := execute()
	if err != nil {
		os.Exit(1)
	}
}

func execute() error {
	dbPath := flag.String("d", filepath.Join(downtime.DefaultDataDir, downtime.DefaultDBFile), "Use the specified downtime database file instead of the system default.")
	cTimeFormat := flag.String("f", downtime.DefaultTimeFormat, "Specify the time and date format to use when reporting using strftime(3) syntax.")
	num := flag.Int64("n", -1, "Define how many latest downtime records to output. Default is all.")
	sleep := flag.Int("s", downtime.DefaultSleepSeconds, "Calculate the approximate crash time by specifying what was the sleep value of downtimed(8).")
	utc := flag.Bool("u", false, "Display times in UTC")
	version := flag.Bool("v", false, "Display the program version number and copyright message.")
	flag.Parse()

	if *version {
		downtime.PrintVersion()
		return nil
	}

	goTimeFmt, err := downtime.StrftimeToGo(*cTimeFormat)
	if err != nil {
		logger.Criticalf("invalid time format: %s", err.Error())
		return err
	}
	fmt.Println(goTimeFmt)

	dbFile, err := os.Open(*dbPath)
	if err != nil {
		logger.Criticalf("can not open %s: %s", *dbPath, err.Error())
		return err
	}

	fInfo, err := dbFile.Stat()
	if err != nil {
		logger.Criticalf("can not stat %s: %s", *dbPath, err.Error())
		return err
	}

	if (fInfo.Size() % downtime.EventSize) != 0 {
		err := fmt.Errorf("database size is invalid")
		logger.Criticalf(err.Error())
		return err
	}
	if (*num != -1) && (fInfo.Size() > (*num * downtime.EventSize * 2)) {
		_, err := dbFile.Seek(*num*downtime.EventSize*-2, os.SEEK_END)
		if err != nil {
			logger.Criticalf("can not seek: %s", err.Error())
			return err
		}
	}

	db := downtime.NewDatabaseReader(dbFile)

	var tdown time.Time
	var crashed bool
	// adjust crash time assuming we crashed in the middle of our sleep time
	var tadjust = (time.Duration(*sleep) * time.Second) / 2

	var evt downtime.Event
	for evt, err = db.Next(); err == nil; evt, err = db.Next() {
		when := evt.When.AsTime()
		if *utc {
			when = when.UTC()
		} else {
			when = when.Local()
		}
		switch evt.What {
		case downtime.EventTypeShutdown:
			if !tdown.IsZero() {
				// there was a missing up event, report the previous down with unknown duration
				report(tdown, time.Time{}, crashed, goTimeFmt)
			}
			tdown = when
			crashed = false
		case downtime.EventTypeCrash:
			if !tdown.IsZero() {
				// there was a missing up event, report the previous down with unknown duration
				report(tdown, time.Time{}, crashed, goTimeFmt)
			}
			crashed = true
			tdown = when.Add(tadjust)
		case downtime.EventTypeUp:
			report(tdown, when, crashed, goTimeFmt)
			tdown = time.Time{}
		}
	}

	if err != nil && !errors.Is(err, io.EOF) {
		logger.Criticalf(err.Error())
		return err
	}
	return nil
}

func report(tDown, tUp time.Time, crashed bool, timeFormat string) {
	if crashed {
		fmt.Printf("crash %s -> ", tDown.Format(timeFormat))
	} else {
		fmt.Printf("down  %s -> ", tDown.Format(timeFormat))
	}

	fmt.Printf("up %s ", tUp.Format(timeFormat))

	if tDown.IsZero() || tUp.IsZero() {
		fmt.Printf("= %11s (? s)\n", "unknown")
	} else {
		downDuration := tUp.Sub(tDown)
		fmt.Printf("= %11s (%d s)\n", formatDuration(downDuration), int(downDuration.Seconds()))
	}
}

func formatDuration(dur time.Duration) string {
	d := int(dur.Hours()) / 24
	h := int(dur.Hours()) % 24
	m := int(dur.Minutes()) % 60
	s := int(dur.Seconds()) % 60
	hms := fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	if d == 0 {
		return hms
	} else {
		return fmt.Sprintf("%d+%s", d, hms)
	}
}
