package ingestion

import (
	"context"
	"encoding/base64"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/types"
)

// Define an array of sample XML documents for testing
var sampleXMLArray = []string{
	`
	<LP1F>
		<Page1>
			<Section1>
				<Title></Title>
				<FirstName>John</FirstName>
				<LastName>Doe</LastName>
			</Section1>
			<BURN>123456789</BURN>
		</Page1>
	</LP1F>
	`,
	`
	<LP1F>
		<Page1>
			<Section1>
				<Title>Ms.</Title>
				<FirstName>Jane</FirstName>
				<LastName>Doe</LastName>
			</Section1>
			<BURN>987654321</BURN>
		</Page1>
	</LP1F>
	`,
}

func TestJobQueue(t *testing.T) {
	queue := NewJobQueue()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queue.StartWorkerPool(ctx, 3)

	var processedJobs int32
	numJobs := len(sampleXMLArray)

	for _, xml := range sampleXMLArray {
		encodedXML := base64.StdEncoding.EncodeToString([]byte(xml))

		doc := &types.BaseDocument{
			Type:        "LP1F",
			Encoding:    "base64",
			NoPages:     1,
			EmbeddedXML: encodedXML,
		}

		queue.AddToQueue(doc, "xml", func() {
			atomic.AddInt32(&processedJobs, 1)
		})
	}

	// Wait for all jobs to be processed with a timeout
	done := make(chan struct{})
	go func() {
		queue.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Ensure the number of processed jobs matches the number of added jobs
		if atomic.LoadInt32(&processedJobs) != int32(numJobs) {
			t.Errorf("Expected %d jobs to be processed, but got %d", numJobs, atomic.LoadInt32(&processedJobs))
		}
	case <-time.After(5 * time.Second):
		t.Error("Test timed out waiting for jobs to be processed")
	}

	// Close the queue and cancel the context to stop workers
	queue.Close()
}
