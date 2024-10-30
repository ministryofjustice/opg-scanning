package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ministryofjustice/opg-scanning/internal/api"
)

func main() {
	controller := api.NewIndexController()
	log.Println("Service started...")

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		controller.HandleRequests()
	}()

	<-stop
	log.Println("Shutting down gracefully...")
	controller.CloseQueue() // Ensure all jobs are processed before shutdown
	log.Println("All jobs processed. Exiting.")
}
