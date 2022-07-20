package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

func main() {
	timeout := flag.String("timeout", "10s", "set timeout")
	log.SetOutput(os.Stderr)
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		log.Fatal("arguments count mismatch: expecting 2 (host, port)")
	}
	timeoutD, err := time.ParseDuration(*timeout)
	if err != nil {
		log.Fatal(fmt.Errorf("timeout format invalid: %w", err))
	}
	tnClient := NewTelnetClient(net.JoinHostPort(args[0], args[1]), timeoutD, os.Stdin, os.Stdout)
	err = tnClient.Connect()
	if err != nil {
		log.Fatal(fmt.Errorf("connection error: %w", err))
	}
	defer tnClient.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func() {
		if err := tnClient.Send(); err != nil {
			log.Printf("error send operation: %s\n", err)
			cancel()
			return
		}
	}()

	go func() {
		if err := tnClient.Receive(); err != nil {
			log.Printf("error receive operation: %s\n", err)
			cancel()
			return
		}
	}()

	<-ctx.Done()
}
