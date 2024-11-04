package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ministryofjustice/opg-scanning/internal/api"
)

func main() {
	controller := api.NewIndexController()
	log.Println("Service started...")

	// Create a context for managing worker shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	controller.Queue.StartWorkerPool(ctx, 3)

	go func() {
		controller.HandleRequests()
	}()

	<-stop
	log.Println("Shutting down gracefully...")
	cancel()
	controller.Queue.Close()
	log.Println("All jobs processed. Exiting.")
}
