package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/app"
	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage/memory"
	dbstorage "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "../../configs/config.example.json", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := NewConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	logg := logger.NewLogger(config.Logger.Level, config.Logger.Destination)
	defer logg.Close()

	var storage app.Storage
	if config.MemoryStorage {
		storage = memorystorage.New(*logg)
	} else {
		storage = dbstorage.New(
			*logg,
			config.Database.Host,
			config.Database.Port,
			config.Database.User,
			config.Database.Password,
			config.Database.Name,
		)
	}
	defer storage.Close()
	calendar := app.New(logg, storage)

	servLogger, err := logger.NewServerLogger("../../server.log")
	fmt.Println(err)
	defer servLogger.Close()
	server := internalhttp.NewServer(*servLogger, calendar, config.Server.Host+":"+strconv.Itoa(config.Server.Port))

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
