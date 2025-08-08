package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ministryofjustice/opg-scanning/internal/api"
	appaws "github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/config"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

func main() {
	// Initialize the tracer provider
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up logging
	logWrapper := logger.New(config.Environment())

	// Initialize configuration
	appConfig, err := config.Read()
	if err != nil {
		logWrapper.Error("Failed to read config", slog.String("error", err.Error()))
		return
	}

	shutdownTracer, err := logger.StartTracerProvider(ctx, logWrapper.SlogLogger, true)
	if err != nil {
		logWrapper.Error("Failed to start tracer provider", slog.String("error", err.Error()))
		return
	}
	defer shutdownTracer()

	// Load AWS configuration
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(appConfig.Aws.Region),
	)

	otelaws.AppendMiddlewares(&cfg.APIOptions)

	if err != nil {
		logWrapper.Error("Failed to load AWS config", "error", err)
		return
	}
	// Initialize AwsClient
	awsClient, err := appaws.NewAwsClient(ctx, cfg, appConfig)
	if err != nil {
		logWrapper.Error("Failed to initialize AWS clients", "error", err)
		return
	}

	if appConfig.Aws.Endpoint != "" {
		cfg.BaseEndpoint = aws.String(appConfig.Aws.Endpoint)
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)

	controller := api.NewIndexController(logWrapper, awsClient, appConfig, dynamoClient)
	logWrapper.Info("Service started...")

	go func() {
		controller.HandleRequests()
	}()

	// Handle graceful shutdown on receiving an interrupt or SIGTERM signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Start shutdown sequence
	logWrapper.Info("Shutting down gracefully...")
	cancel()
	logWrapper.Info("All jobs processed. Exiting.")
}
