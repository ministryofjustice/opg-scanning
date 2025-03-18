package ingestion

import (
	"context"
	"encoding/base64"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-scanning/config"
	"github.com/ministryofjustice/opg-scanning/internal/types"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

func TestJobQueue(t *testing.T) {
	cfg := config.NewConfig()
	queue := NewJobQueue(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queue.StartWorkerPool(ctx, 3)

	var processedJobs int32

	xmlStringData1 := util.LoadXMLFileTesting(t, "../../xml/LP1F-valid.xml")
	xmlStringData2 := util.LoadXMLFileTesting(t, "../../xml/LP1F-alternate.xml")

	xmlData1 := base64.StdEncoding.EncodeToString(xmlStringData1)
	xmlData2 := base64.StdEncoding.EncodeToString(xmlStringData2)

	sampleXMLArray := []string{
		xmlData1,
		xmlData2,
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

		queue.AddToQueue(ctx, doc, "xml", func(ctx context.Context, processedDocument interface{}, doc *types.BaseDocument) error {
			atomic.AddInt32(&processedJobs, 1)
			return nil
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
