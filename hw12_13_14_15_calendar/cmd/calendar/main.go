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
	internalconfig "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	internalgrpc "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/server/http"
	internalstorage "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage/factory"
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

	config, err := internalconfig.NewConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	logg := logger.NewLogger(config.Logger.Level, config.Logger.Destination)
	defer logg.Close()

	storage := internalstorage.New(config.Storage, *logg)
	defer storage.Close()
	calendar := app.New(logg, storage)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	// gRPC
	serverGrpc := internalgrpc.NewServer(logg, calendar, config.Server.Grpc.Host, config.Server.Grpc.Port)

	go func() {
		if err := serverGrpc.Start(); err != nil {
			logg.Error("failed to start grpc server: " + err.Error())
		}
	}()

	go func() {
		<-ctx.Done()
		serverGrpc.Stop()
	}()

	// http
	servLogger, err := logger.NewServerLogger("../../server.log")
	if err != nil {
		logg.Error(fmt.Sprintf("failed creating http server: %s", err))
	}

	defer servLogger.Close()
	server := internalhttp.NewServer(*servLogger, calendar, config.Server.HTTP.Host+
		":"+strconv.Itoa(config.Server.HTTP.Port))

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
