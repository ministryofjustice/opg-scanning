package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/api"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

func main() {
	// Initialize the tracer provider
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize configuration
	appConfig := config.NewConfig()

	// Set up logging
	logWrapper := logger.NewLogger(appConfig)
	slogLogger := logWrapper.SlogLogger

	shutdownTracer, err := logger.StartTracerProvider(ctx, slogLogger, true)
	if err != nil {
		slogLogger.Error("Failed to start tracer provider", slog.String("error", err.Error()))
		return
	}
	defer shutdownTracer()

	// Load AWS configuration
	cfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(appConfig.Aws.Region),
	)

	otelaws.AppendMiddlewares(&cfg.APIOptions)

	if err != nil {
		slogLogger.Error("Failed to load AWS config", "error", err)
		return
	}
	// Initialize AwsClient
	awsClient, err := aws.NewAwsClient(ctx, cfg, appConfig)
	if err != nil {
		log.Fatalf("failed to initialize AWS clients: %v", err)
	}

	controller := api.NewIndexController(awsClient, appConfig)
	controller.Queue.StartWorkerPool(ctx, 3)
	slogLogger.Info("Service started...")

	go func() {
		controller.HandleRequests()
	}()

	// Handle graceful shutdown on receiving an interrupt or SIGTERM signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Start shutdown sequence
	slogLogger.Info("Shutting down gracefully...")
	cancel()
	controller.Queue.Close()
	slogLogger.Info("All jobs processed. Exiting.")
}
