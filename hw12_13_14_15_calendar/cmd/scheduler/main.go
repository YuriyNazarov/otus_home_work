package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	internalapp "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/app"
	internalconfig "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	internalmq "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/rabbit"
	internalstorage "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage/factory"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "../../configs/scheduler_config.example.json", "Path to configuration file")
}

func main() {
	flag.Parse()

	config, err := internalconfig.NewSchedulerConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	logg := logger.NewLogger(config.Logger.Level, config.Logger.Destination)
	defer logg.Close()

	storage := internalstorage.New(config.Storage, *logg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	rabbit := internalmq.NewRabbit(ctx, config.Queue.DSN, config.Queue.Exchange, config.Queue.Queue, logg)
	app := internalapp.NewScheduler(logg, storage, rabbit)
	app.Notify(ctx)
	remindTimer := time.Tick(config.Remind.RemindPeriod)
	clearTimer := time.Tick(config.Remind.ClearPeriod)
	go func() {
		for {
			select {
			case <-remindTimer:
				app.Notify(ctx)
			case <-clearTimer:
				app.RemoveOldEvents()
			case <-ctx.Done():
				return
			}
		}
	}()

	logg.Info("scheduler is running...")

	<-ctx.Done()
}
