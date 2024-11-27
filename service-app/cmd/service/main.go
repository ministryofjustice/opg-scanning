package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-scanning/internal/api"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
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

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-west-1"),
		config.WithBaseEndpoint(os.Getenv("AWS_ENDPOINT")),
	)
	if err != nil {
		logger.Error("Failed to load AWS config", "error", err)
		return
	}
	// Initialize AwsClient
	awsClient, err := aws.NewAwsClient(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize AWS clients: %v", err)
	}

	controller := api.NewIndexController(awsClient)
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
