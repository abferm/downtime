package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"log/syslog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/abferm/downtime"
	"github.com/juju/loggo"
	"github.com/juju/loggo/loggocolor"
)

var logger = loggo.GetLogger("")

func main() {
	err := execute()
	if err != nil {
		os.Exit(1)
	}
}

func execute() error {
	noDB := flag.Bool("D", false, "Do not create nor update the downtime database.")
	dataDir := flag.String("d", downtime.DefaultDataDir, "The directory where the time stamp files as well as the downtime database are located.")
	noFork := flag.Bool("F", false, "Do not call daemon(3) to fork(2) to background. Useful with modern system service managers such as systemd(8), launchd(8) and others.")
	cTimeFormat := flag.String("f", downtime.DefaultTimeFormat, "Specify the time and date format to use when reporting using strftime(3) syntax.")
	logDestination := flag.String("l", "daemon", "Logging destination. If the argument contains a slash (/) it is interpreted to be a path name to a log file, which will be created if it does not exist already. Otherwise it is interpreted as a syslog facility name.")
	pidFile := flag.String("p", "/var/run/downtimed.pid", "The location of the file which keeps track of the process ID of the running daemon process. The system default location is determined at compile time. May be disabled by specifying \"none\".")
	flag.Bool("S", false, "Disable fsync (ignored)")
	sleep := flag.Int64("s", 15, "Defines how long to sleep between each update of the on−disk time stamp file. More frequent updates result in more accurate downtime reporting in the case of a system crash. Less frequent updates decrease the amount of disk writes performed.")
	version := flag.Bool("v", false, "Display the program version number, copyright message and the default settings.")
	flag.Parse()

	logger.SetLogLevel(loggo.INFO)

	if *version {
		downtime.PrintVersion()
		return nil
	}

	if !*noFork {
		exe, _ := os.Executable()
		daemonizeArgs := []string{"-p", *pidFile, "-l", *pidFile, "--", exe}
		daemonizeArgs = append(daemonizeArgs, os.Args[1:]...)
		daemonizeArgs = append(daemonizeArgs, "-F") // we're done forking
		cmd := exec.Command("daemonize", daemonizeArgs...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			logger.Criticalf("could not daemonize: %s", err.Error())
			return err
		}
		return nil
	}

	var logDest io.Writer = os.Stdout
	if (*logDestination)[0] == '/' {
		logFile, err := os.OpenFile(*logDestination, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.Criticalf("failed to open log file: %s", err.Error())
			return err
		}
		defer logFile.Close()
		logDest = logFile
	} else if *logDestination != "-" {
		logSyslog, err := syslog.New(syslog.LOG_INFO, *logDestination)
		if err != nil {
			logger.Criticalf("failed to open syslog: %s", err.Error())
			return err
		}
		defer logSyslog.Close()
		logDest = logSyslog
	}
	loggo.ReplaceDefaultWriter(loggocolor.NewColorWriter(logDest))

	store, err := downtime.NewDataDir(*dataDir)
	if err != nil {
		logger.Criticalf(err.Error())
		return err
	}

	var dbWriter io.Writer
	if *noDB {
		dbWriter = bytes.NewBuffer([]byte{})
	} else {
		var err error
		dbWriter, err = os.OpenFile(filepath.Join(*dataDir, downtime.DefaultDBFile), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.Criticalf("could not open downtimedb: %s", err.Error())
			return err
		}
	}

	db := downtime.NewDatabaseWriter(dbWriter)
	defer db.Close()

	daemon := downtime.NewDaemon(store, db, time.Duration(*sleep)*time.Second)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	boottime, err := downtime.SystemBootTime()
	if err != nil {
		logger.Criticalf(err.Error())
		return err
	}

	err = daemon.Init(boottime, downtime.StrftimeToGo(*cTimeFormat))
	if err != nil {
		logger.Criticalf("init failed: %s", err.Error())
		return err
	}

	err = daemon.Run(ctx)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	logger.Criticalf(err.Error())
	return err
}
