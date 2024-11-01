package ingestion

import (
	"context"
	"sync"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type Job struct {
	Data       *types.Document
	format     string
	onComplete func()
}

type JobQueue struct {
	Jobs   chan Job
	wg     *sync.WaitGroup
	logger *logger.Logger
}

func NewJobQueue() *JobQueue {
	queue := &JobQueue{
		Jobs: make(chan Job, 10), // Buffer size can be adjusted based on needs
		wg:   &sync.WaitGroup{},
	}
	return queue
}

func (q *JobQueue) AddToQueue(data *types.Document, format string, onComplete func()) {
	job := Job{Data: data, format: format, onComplete: onComplete}
	q.wg.Add(1)
	q.Jobs <- job
}

func (q *JobQueue) StartWorkerPool(ctx context.Context, numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			select {
			case job, ok := <-q.Jobs:
				if !ok {
					return
				}

				processCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				done := make(chan struct{})
				go func() {
					defer close(done)
					_, err := util.ProcessDocument(job.Data, job.Data.Type, job.format)
					if err != nil {
						q.logger.ErrorFormated("Worker %d failed to process job: %v, error: %v\n", workerID, job.Data, err)
					}
				}()

				select {
				case <-processCtx.Done():
					q.logger.ErrorFormated("Worker %d timed out processing job: %v\n", workerID, job.Data)
				case <-done:
					if job.onComplete != nil {
						job.onComplete()
					}
				}

				q.wg.Done()

			case <-ctx.Done():
				q.logger.Info("Worker pool stopped")
				return
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
