package ingestion

import (
	"context"
	"encoding/base64"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/stretchr/testify/require"
)

func TestJobQueue(t *testing.T) {
	queue := NewJobQueue()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queue.StartWorkerPool(ctx, 3)

	var processedJobs int32

	sampleXMLArray := []string{
		loadXMLFile(t, "../../xml/LP1F-valid.xml"),
		loadXMLFile(t, "../../xml/LP1F-alternate.xml"),
	}

	numJobs := len(sampleXMLArray)

	// Add each XML as a job in the queue
	for _, xml := range sampleXMLArray {
		doc := &types.BaseDocument{
			Type:        "LP1F",
			Encoding:    "base64",
			NoPages:     1,
			EmbeddedXML: xml,
		}

		queue.AddToQueue(doc, "xml", func() {
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
		if atomic.LoadInt32(&processedJobs) != int32(numJobs) {
			t.Errorf("Expected %d jobs to be processed, but got %d", numJobs, atomic.LoadInt32(&processedJobs))
		}
	case <-time.After(5 * time.Second):
		t.Error("Test timed out waiting for jobs to be processed")
	}

	queue.Close()
}

func loadXMLFile(t *testing.T, filepath string) string {
	data, err := os.ReadFile(filepath)
	if err != nil {
		require.FailNow(t, "Failed to read XML file", err.Error())
	}
	return base64.StdEncoding.EncodeToString(data)
}
