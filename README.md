# downtime
Golang port of [downtimed](ttps://github.com/snabb/downtimed)

# Examples
## Track system downtime
``` golang
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	daemon := downtime.NewDaemon(store, db, sleepDuration)

	boottime, err := downtime.SystemBootTime()
	if err != nil {
		logger.Criticalf(err.Error())
		return err
	}
	err = daemon.Init(boottime, goTimeFormat)
	if err != nil {
		logger.Criticalf("init failed: %s", err.Error())
		return err
	}

	err = daemon.Run(ctx)
	if errors.Is(err, context.Canceled) {
		return nil
	}
```
## Track program downtime
``` golang
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	daemon := downtime.NewDaemon(store, db, sleepDuration)

	boottime := downtime.ProcessBootTime()
	err = daemon.Init(boottime, goTimeFormat)
	if err != nil {
		logger.Criticalf("init failed: %s", err.Error())
		return err
	}

    go func(){
        err = daemon.Run(ctx)
        if errors.Is(err, context.Canceled) {
            return nil
        } else if err != nil{
            panic(err)
        }
    }
```