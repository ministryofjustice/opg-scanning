package ingestion

import (
	"context"
	"fmt"
	"sync"

	"github.com/ministryofjustice/opg-scanning/internal/config"
	"github.com/ministryofjustice/opg-scanning/internal/constants"
	"github.com/ministryofjustice/opg-scanning/internal/factory"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"go.opentelemetry.io/otel/trace"
)

type JobQueue struct {
	wg     *sync.WaitGroup
	logger *logger.Logger
	errors []error
}

func NewJobQueue(config *config.Config) *JobQueue {
	queue := &JobQueue{
		wg:     &sync.WaitGroup{},
		logger: logger.GetLogger(config.App.Environment),
		errors: make([]error, 0),
	}
	return queue
}

func newJobContext(reqCtx context.Context) context.Context {
	enrichedLogger := logger.LoggerFromContext(reqCtx)
	span := trace.SpanFromContext(reqCtx)
	ctx := trace.ContextWithSpan(context.Background(), span)
	return logger.ContextWithLogger(ctx, enrichedLogger)
}

func (q *JobQueue) AddToQueueSequentially(ctx context.Context, cfg *config.Config, data *types.BaseDocument, format string, onComplete func(ctx context.Context, processedDoc interface{}, originalDoc *types.BaseDocument) error) error {
	// Create a job context
	jobCtx := newJobContext(ctx)

	// Initialize the registry and processor synchronously
	registry, err := factory.NewRegistry()
	if err != nil {
		return fmt.Errorf("failed to create registry: %v", err)
	}

	processor, err := factory.NewDocumentProcessor(data, data.Type, format, registry, q.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize processor: %v", err)
	}

	// Use a per-job timeout context.
	processCtx, cancel := context.WithTimeout(jobCtx, cfg.HTTP.Timeout)
	processCtx = context.WithValue(processCtx, constants.TokenContextKey, ctx.Value(constants.TokenContextKey))
	defer cancel()

	parsedDoc, err := processor.Process(processCtx)
	if err != nil {
		return fmt.Errorf("failed to process job: %v", err)
	}

	if onComplete != nil {
		return onComplete(processCtx, parsedDoc, data)
	}

	return nil
}
