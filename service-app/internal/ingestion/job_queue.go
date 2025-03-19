package ingestion

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/factory"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"go.opentelemetry.io/otel/trace"
)

type Job struct {
	ctx        context.Context
	cfg        *config.Config
	Data       *types.BaseDocument
	format     string
	onComplete func(ctx context.Context, processedDoc interface{}, originalDoc *types.BaseDocument) error
}

type JobQueue struct {
	Jobs    chan Job
	wg      *sync.WaitGroup
	logger  *logger.Logger
	errors  []error
	errorMu sync.Mutex
}

func NewJobQueue(config *config.Config) *JobQueue {
	queue := &JobQueue{
		Jobs:   make(chan Job, 10), // Buffer size can be adjusted based on needs
		wg:     &sync.WaitGroup{},
		logger: logger.GetLogger(config),
		errors: make([]error, 0),
	}
	return queue
}

func NewJobContext(reqCtx context.Context) context.Context {
	enrichedLogger := logger.LoggerFromContext(reqCtx)
	span := trace.SpanFromContext(reqCtx)
	ctx := trace.ContextWithSpan(context.Background(), span)
	return logger.ContextWithLogger(ctx, enrichedLogger)
}

func (q *JobQueue) AddToQueue(ctx context.Context, cfg *config.Config, data *types.BaseDocument, format string, onComplete func(ctx context.Context, processedDoc interface{}, originalDoc *types.BaseDocument) error) {
	jobCtx := NewJobContext(ctx)
	job := Job{ctx: jobCtx, cfg: cfg, Data: data, format: format, onComplete: onComplete}
	q.wg.Add(1)
	q.Jobs <- job
}

func (q *JobQueue) StartWorkerPool(ctx context.Context, numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for {
				select {
				case job, ok := <-q.Jobs:
					if !ok {
						return // Exit if the job channel is closed
					}

					// Create a per job timeout context from the jobs context.
					processCtx, cancel := context.WithTimeout(job.ctx, time.Duration(job.cfg.HTTP.Timeout)*time.Second)
					done := make(chan struct{})

					go func() {
						defer close(done)

						// Initialize document processor
						registry, err := factory.NewRegistry()
						if err != nil {
							q.recordError(fmt.Errorf("Worker %d failed to create registry, job: %v", workerID, err))
							return
						}

						processor, err := factory.NewDocumentProcessor(job.Data, job.Data.Type, job.format, registry, q.logger)
						if err != nil {
							q.recordError(fmt.Errorf("Worker %d failed to initialize processor for job: %v", workerID, err))
							return
						}

						// Process the document using processCtx to enforce the timeout.
						parsedDoc, err := processor.Process(processCtx)
						if err != nil {
							q.recordError(fmt.Errorf("Worker %d failed to process job: %v\n", workerID, err))
							return
						}

						if job.onComplete != nil {
							// Pass the jobs original context to the callback.
							err := job.onComplete(job.ctx, parsedDoc, job.Data)
							if err != nil {
								q.recordError(fmt.Errorf("onComplete errors: %v", err.Error()))
							}
						}
					}()

					select {
					case <-processCtx.Done():
						q.recordError(fmt.Errorf("Worker %d timed out processing job\n", workerID))
					case <-done:
						// Job completed without timing out.
						q.logger.Info("Worker completed: %d!\n", nil, workerID)
					}
					// Cancel the timeout context to free resources.
					cancel()
					q.wg.Done()

				case <-ctx.Done():
					q.logger.Info("Worker pool stopped", nil)
					return
				}
			}
		}(i)
	}
}

func (q *JobQueue) recordError(err error) {
	q.errorMu.Lock()
	defer q.errorMu.Unlock()
	q.errors = append(q.errors, err)
}

func (q *JobQueue) GetErrors() []error {
	q.errorMu.Lock()
	defer q.errorMu.Unlock()
	errorsCopy := make([]error, len(q.errors))
	copy(errorsCopy, q.errors)
	return errorsCopy
}

func (q *JobQueue) Wait() {
	q.wg.Wait()
}

func (q *JobQueue) Close() {
	close(q.Jobs)
	q.wg.Wait()
}
