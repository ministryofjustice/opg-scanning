package ingestion

import (
	"sync"

	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type Job struct {
	Data    interface{}
	docType string
	format  string
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
	queue.StartWorkerPool(3) // Start with 3 workers
	return queue
}

func (q *JobQueue) AddToQueue(data interface{}, docType string, format string) {
	job := Job{Data: data, docType: docType, format: format}
	q.wg.Add(1)
	q.Jobs <- job
}

func (q *JobQueue) StartWorkerPool(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for job := range q.Jobs {
				q.logger.InfoFormated("Worker %d processing job: %+v\n", workerID, job.Data)

				data, ok := job.Data.([]byte)
				if !ok {
					q.logger.InfoFormated("Worker %d failed on type assertion: %v\n", workerID, job.Data)
					continue
				}

				parsedDoc, err := util.ProcessDocument(data, job.docType, job.format)
				if err != nil {
					q.logger.ErrorFormated("Worker %d failed to process job: %v, error: %v\n", workerID, job.Data, err)
				} else {
					q.logger.InfoFormated("Worker %d successfully processed job: %+v\n", workerID, parsedDoc)
				}

				q.wg.Done()
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
