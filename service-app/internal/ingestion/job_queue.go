package ingestion

import (
	"context"
	"sync"
	"time"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/factory"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
)

type Job struct {
	Data       *types.BaseDocument
	format     string
	onComplete func(processedDoc interface{}, originalDoc *types.BaseDocument)
}

type JobQueue struct {
	Jobs   chan Job
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

func (q *JobQueue) AddToQueue(data *types.BaseDocument, format string, onComplete func(interface{}, *types.BaseDocument)) {
	job := Job{Data: data, format: format, onComplete: onComplete}
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

					processCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
					defer cancel()

					done := make(chan struct{})
					go func() {
						defer close(done)

						// Initialize document processor
						registry := factory.NewRegistry()
						processor, err := factory.NewDocumentProcessor(job.Data, job.Data.Type, job.format, registry, q.logger)
						if err != nil {
							q.logger.Error("Worker %d failed to initialize processor for job: %v\n", nil, workerID, err)
							return
						}

						// Process the document
						parsedDoc, err := processor.Process()
						if err != nil {
							q.logger.Error("Worker %d failed to process job: %v\n", nil, workerID, err)
							return
						}

						if job.onComplete != nil {
							job.onComplete(parsedDoc, job.Data)
						}
					}()

					select {
					case <-processCtx.Done():
						q.logger.Error("Worker %d timed out processing job\n", nil, workerID)
					case <-done:
						// Job completed without timing out
					}

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
	close(q.Jobs)
	q.wg.Wait()
}
