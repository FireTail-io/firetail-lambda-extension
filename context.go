package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Returns a context which will be cancelled whenever a SIGTERM or SIGINT signal is received
func getContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		log.Printf("Received signal '%s'. Exiting...", s.String())
		cancel()
	}()
	return ctx
}
