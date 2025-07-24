package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-scanning/internal/ingestion"
	"github.com/stretchr/testify/assert"
)

const host = "http://localhost:8081"

func TestIntegrationMain(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	token, _ := getToken()

	for _, fileType := range []string{"EP2PG", "LP1F", "LP1H", "LP2"} {
		t.Run(fileType, func(t *testing.T) {
			assert.NoError(t, checkFile(token, fileType, fileType+"-valid", uuid.NewString()))
		})
	}

	for _, fileType := range []string{"Correspondence", "EPA", "LPA002", "LPA114", "LPA115", "LPA116", "LPA117", "LPA120", "LPA-PA", "LPA-PW", "LPC"} {
		t.Run(fileType, func(t *testing.T) {
			assert.NoError(t, checkAttachment(token, fileType, fileType+"-valid", uuid.NewString()))
		})
	}

	t.Run("same file", func(t *testing.T) {
		sameID := uuid.NewString()
		assert.NoError(t, checkFile(token, "LP1F", "LP1F-valid", sameID))
		assert.Equal(t, ingestion.AlreadyProcessedError{}, checkFile(token, "LP1F", "LP1F-valid", sameID))
	})

	t.Run("same attachment", func(t *testing.T) {
		sameAttachmentID := uuid.NewString()
		assert.NoError(t, checkAttachment(token, "LPC", "LPC-valid", sameAttachmentID))
		assert.NoError(t, checkAttachment(token, "LPC", "LPC-valid", sameAttachmentID))
	})
}

func getToken() (string, error) {
	req, err := http.NewRequest(http.MethodPost, host+"/auth/sessions", strings.NewReader(`{"user":{"email":"opg_document_and_d@publicguardian.gsi.gov.uk","password":"password"}}`))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint:errcheck

	var v struct {
		Token string `json:"authentication_token"`
	}
	return v.Token, json.NewDecoder(resp.Body).Decode(&v)
}

func checkFile(token, fileType, fileName, id string) error {
	if err := exec.Command("docker", "compose", "exec", "localstack", "awslocal", "sqs", "purge-queue", "--queue-url=ddc.fifo").Run(); err != nil {
		return err
	}

	uploadResponse, err := upload(token, fileType, fileName+".xml", "", id)
	if err != nil {
		return err
	}
	defer uploadResponse.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(uploadResponse.Body)
	if err != nil {
		return err
	}

	var uv struct {
		Data struct {
			UID              string `json:"uid"`
			Success          bool   `json:"success"`
			Message          string `json:"message"`
			ValidationErrors any    `json:"validationErrors"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &uv); err != nil {
		return err
	}
	if uv.Data.Success != true {
		return fmt.Errorf("file failed (%d): %s", uploadResponse.StatusCode, body)
	}

	uploadUID := uv.Data.UID

	if uploadResponse.StatusCode == http.StatusAlreadyReported {
		return ingestion.AlreadyProcessedError{}
	}

	cmdOut, err := exec.Command("docker", "compose", "exec", "localstack", "awslocal", "sqs", "receive-message", "--queue-url=ddc.fifo").Output()
	if err != nil {
		return err
	}
	var cv struct {
		Messages []struct {
			Body string
		}
	}
	if err := json.Unmarshal(cmdOut, &cv); err != nil {
		return err
	}

	if !strings.Contains(cv.Messages[0].Body, uploadUID) {
		return fmt.Errorf("file failed: body doesn't contain %s (is: %s)", uploadUID, cv.Messages[0].Body)
	}

	return nil
}

func checkAttachment(token, fileType, fileName, id string) error {
	resp, err := upload(token, fileType, fileName+".xml", "7000-1234-1234", id)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var uv struct {
		Data struct {
			UID              string `json:"uid"`
			Success          bool   `json:"success"`
			Message          string `json:"message"`
			ValidationErrors any    `json:"validationErrors"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &uv); err != nil {
		return err
	}

	if uv.Data.Success != true {
		return fmt.Errorf("file failed (%d): %v", resp.StatusCode, string(body))
	}

	return nil
}

func upload(token, fileType, fileName, caseNo, id string) (*http.Response, error) {
	xml, err := os.ReadFile("./testdata/xml/" + fileName)
	if err != nil {
		return nil, err
	}
	set := xmlToSet(fileType, xml, caseNo, id)

	req, err := http.NewRequest(http.MethodPost, host+"/api/ddc", strings.NewReader(set))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "text/xml")
	req.Header.Add("Cookie", "membrane="+token)

	return http.DefaultClient.Do(req)
}

func xmlToSet(fileType string, xml []byte, caseNo, id string) string {
	xmlB64 := base64.StdEncoding.EncodeToString(xml)

	return `<?xml version="1.0" encoding="UTF-8"?>
<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
<Header CaseNo="` + caseNo + `" Scanner="9" ScanTime="2014-09-26 12:38:53" ScannerOperator="Administrator" Schedule="01-0001253-20160909174150" />
<Body>
<Document Type="` + fileType + `" Encoding="UTF-8" NoPages="1" ID="` + id + `">
<XML>` + xmlB64 + `</XML>
<PDF>SUkqAAoAAAAAtBEAAAEDAAEAAAABAAAAAQEDAAEAAAABAAAAAgEDAAEAAAABAAAAAwEDAAEAAAABAAAABgEDAAEAAAABAAAACgEDAAEAAAABAAAADQECAAEAAAAAAAAAEQEEAAEAAAAIAAAAEgEDAAEAAAABAAAAFQEDAAEAAAABAAAAFgEDAAEAAAAAIAAAFwEEAAEAAAABAAAAGgEFAAEAAADcAAAAGwEFAAEAAADkAAAAHAEDAAEAAAABAAAAKAEDAAEAAAACAAAAKQEDAAIAAAAAAAEAAAAAAEgAAAABAAAASAAAAAEAAAA=</PDF>
</Document>
</Body>
</Set>`
}
