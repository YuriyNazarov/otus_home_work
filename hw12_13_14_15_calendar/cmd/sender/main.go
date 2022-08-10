package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/app"
	internalconfig "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	internalmq "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/rabbit"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "../../configs/sender_config.example.json", "Path to configuration file")
}

func main() {
	flag.Parse()

	config, err := internalconfig.NewSenderConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}
	logg := logger.NewLogger(config.Logger.Level, config.Logger.Destination)
	defer logg.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	rabbit := internalmq.NewRabbit(ctx, config.Queue.DSN, config.Queue.Exchange, config.Queue.Queue, logg)

	sender := app.NewSender(rabbit, logg)
	sender.Run()
	<-ctx.Done()
}
