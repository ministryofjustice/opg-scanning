package ingestion

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// Create an array of sample XML
var sampleXMLArray = [][]byte{
	[]byte(`
		<LP1F>
			<Page1>
				<Section1>
					<Title>Mr.</Title>
					<FirstName>John</FirstName>
					<LastName>Doe</LastName>
					<OtherNames>Johnny</OtherNames>
					<DOB>1980-01-01</DOB>
					<Address>123 Main St, Springfield, USA</Address>
					<EmailAddress>john.doe@example.com</EmailAddress>
				</Section1>
				<BURN>123456789</BURN>
				<PhysicalPage>1</PhysicalPage>
			</Page1>
		</LP1F>
		`),
	[]byte(`
		<LP1F>
			<Page1>
				<Section1>
					<Title>Ms.</Title>
					<FirstName>Jane</FirstName>
					<LastName>Doe</LastName>
					<OtherNames>Janey</OtherNames>
					<DOB>1988-01-11</DOB>
					<Address>123 Main St, Springfield, USA</Address>
					<EmailAddress>jane.doe@example.com</EmailAddress>
				</Section1>
				<BURN>123456789</BURN>
				<PhysicalPage>1</PhysicalPage>
			</Page1>
		</LP1F>
`)}

func TestJobQueue(t *testing.T) {
	queue := NewJobQueue()

	// Create a context that will manage the worker lifecycle
	ctx, cancel := context.WithCancel(context.Background())

	// Start worker pool
	queue.StartWorkerPool(ctx, 3)

	var processedJobs int32
	numJobs := 2

	for i := 0; i < numJobs; i++ {
		queue.AddToQueue(sampleXMLArray[i], "LP1F", "xml", func() {
			atomic.AddInt32(&processedJobs, 1)
		})
	}
	done := make(chan struct{})
	go func() {
		queue.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All jobs processed
		if atomic.LoadInt32(&processedJobs) != int32(numJobs) {
			t.Errorf("Expected %d jobs to be processed, but got %d", numJobs, atomic.LoadInt32(&processedJobs))
		}
	case <-time.After(5 * time.Second):
		t.Error("Test timed out waiting for jobs to be processed")
	}

	cancel()
	queue.Close()
}
