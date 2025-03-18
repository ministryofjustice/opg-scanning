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
	Data       *types.BaseDocument
	format     string
	onComplete func(ctx context.Context, processedDoc interface{}, originalDoc *types.BaseDocument) error
}

type JobQueue struct {
	Jobs   chan Job
	Errors chan error
	wg     *sync.WaitGroup
	logger *logger.Logger
}

func NewJobQueue(config *config.Config) *JobQueue {
	queue := &JobQueue{
		Jobs:   make(chan Job, 10), // Buffer size can be adjusted based on needs
		wg:     &sync.WaitGroup{},
		logger: logger.NewLogger(config),
	}
	return queue
}

func NewJobContext(reqCtx context.Context) context.Context {
	enrichedLogger := logger.LoggerFromContext(reqCtx)
	span := trace.SpanFromContext(reqCtx)
	ctx := trace.ContextWithSpan(context.Background(), span)
	return logger.ContextWithLogger(ctx, enrichedLogger)
}

func (q *JobQueue) AddToQueue(ctx context.Context, data *types.BaseDocument, format string, onComplete func(ctx context.Context, processedDoc interface{}, originalDoc *types.BaseDocument) error) {
	jobCtx := NewJobContext(ctx)
	job := Job{ctx: jobCtx, Data: data, format: format, onComplete: onComplete}
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

					// TODO: Load timeout from Config
					// Create a per job timeout context from the jobs context.
					processCtx, cancel := context.WithTimeout(job.ctx, 5*time.Second)
					done := make(chan struct{})

					go func() {
						defer close(done)

						// Initialize document processor
						registry, err := factory.NewRegistry()
						if err != nil {
							q.Errors <- fmt.Errorf("worker %d failed to create registry, job: %v", workerID, err)
							// q.wg.Done()
							// cancel()
							return
						}

						processor, err := factory.NewDocumentProcessor(job.Data, job.Data.Type, job.format, registry, q.logger)
						if err != nil {
							fmt.Print(">>done")
							q.Errors <- fmt.Errorf("worker %d failed to initialize processor for job: %v", workerID, err)
							// q.wg.Done()
							// cancel()
							return
						}

						// Process the document using processCtx to enforce the timeout.
						parsedDoc, err := processor.Process(processCtx)
						if err != nil {
							q.Errors <- fmt.Errorf("worker %d failed to process job: %v", workerID, err)
							// q.wg.Done()
							// cancel()
							return
						}

						if job.onComplete != nil {
							// Pass the jobs original context to the callback.
							err = job.onComplete(job.ctx, parsedDoc, job.Data)

							if err != nil {
								q.Errors <- err
								// q.wg.Done()
								// cancel()
								return
							}
						}
					}()

					select {
					case <-processCtx.Done():
						fmt.Print(">>process timed out")
						q.Errors <- fmt.Errorf("worker %d timed out processing job", workerID)
					case <-done:
						fmt.Print(">>process completed without timeout")
						// Job completed without timing out.
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

func (q *JobQueue) Wait() {
	q.wg.Wait()
}

func (q *JobQueue) Close() {
	q.wg.Wait()
	close(q.Jobs)
	close(q.Errors)
}
