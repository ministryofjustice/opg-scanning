package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-scanning/internal/api"
)

func main() {
	// Set up logging
	logger := telemetry.NewLogger("opg-scanning-service")

	// Initialize the tracer provider
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTracer, err := telemetry.StartTracerProvider(ctx, logger, true)
	if err != nil {
		logger.Error("Failed to start tracer provider", "error", err)
		return
	}
	defer shutdownTracer()

	controller := api.NewIndexController()
	controller.Queue.StartWorkerPool(ctx, 3)
	logger.Info("Service started...")

	go func() {
		controller.HandleRequests()
	}()

	// Handle graceful shutdown on receiving an interrupt or SIGTERM signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Start shutdown sequence
	logger.Info("Shutting down gracefully...")
	cancel()
	controller.Queue.Close()
	logger.Info("All jobs processed. Exiting.")
}
