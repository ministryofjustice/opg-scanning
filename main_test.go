package main

import (
	"bytes"
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

	t.Run("missing auth token", func(t *testing.T) {
		xml, err := os.ReadFile("./testdata/xml/LP1F-valid.xml")
		assert.NoError(t, err)

		set := xmlToSet("LP1F", xml, "7000000", "", "")

		req, err := http.NewRequest(http.MethodPost, host+"/api/ddc", strings.NewReader(set))
		assert.NoError(t, err)

		req.Header.Add("Content-Type", "text/xml")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, "Unauthorized: Missing token\n", readString(resp.Body))

		assertLogsContainMessage(t, "Unauthorized: Missing token: http: named cookie not present")
	})

	t.Run("invalid http method", func(t *testing.T) {
		xml, err := os.ReadFile("./testdata/xml/LP1F-valid.xml")
		assert.NoError(t, err)

		set := xmlToSet("LP1F", xml, "7000000", "", "")

		req, err := http.NewRequest(http.MethodGet, host+"/api/ddc", strings.NewReader(set))
		assert.NoError(t, err)

		req.Header.Add("Content-Type", "text/xml")
		req.Header.Add("Cookie", "membrane="+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		assert.JSONEq(t, `{"data":{"success":false,"message":"Invalid HTTP method"}}`, readString(resp.Body))

		assertLogsContainMessage(t, "Invalid HTTP method")
	})

	t.Run("empty body", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, host+"/api/ddc", strings.NewReader(""))
		assert.NoError(t, err)

		req.Header.Add("Content-Type", "text/xml")
		req.Header.Add("Cookie", "membrane="+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.JSONEq(t, `{"data":{"success":false,"message":"Validate and sanitize XML failed"}}`, readString(resp.Body))

		assertLogsContain(t, "Validate and sanitize XML failed", "failed to parse XML: EOF")
	})

	t.Run("invalid content type", func(t *testing.T) {
		xml, err := os.ReadFile("./testdata/xml/LP1F-valid.xml")
		assert.NoError(t, err)

		set := xmlToSet("LP1F", xml, "7000000", "", "")

		req, err := http.NewRequest(http.MethodPost, host+"/api/ddc", strings.NewReader(set))
		assert.NoError(t, err)

		req.Header.Add("Content-Type", "text/plain")
		req.Header.Add("Cookie", "membrane="+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.JSONEq(t, `{"data":{"success":false,"message":"Invalid content type"}}`, readString(resp.Body))

		assertLogsContain(t, "Invalid content type", "expected application/xml or text/xml, got text/plain")
	})

	t.Run("invalid xml", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, host+"/api/ddc", strings.NewReader("<"))
		assert.NoError(t, err)

		req.Header.Add("Content-Type", "text/xml")
		req.Header.Add("Cookie", "membrane="+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.JSONEq(t, `{"data":{"success":false,"message":"Validate and sanitize XML failed"}}`, readString(resp.Body))

		assertLogsContain(t, "Validate and sanitize XML failed", "failed to parse XML: XML syntax error on line 1: unexpected EOF")
	})

	t.Run("file with caseno", func(t *testing.T) {
		resp, err := upload(token, "LP1F", "LP1F-valid", "7000000", uuid.NewString())
		assert.NoError(t, err)

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.JSONEq(t, `{"data":{"success":false,"message":"Validate set failed"}}`, string(body))

		assertLogsContain(t, "Validate set failed", "must not supply a case number when creating a new case")
	})

	t.Run("attachment with no caseno", func(t *testing.T) {
		resp, err := upload(token, "LPC", "LPC-valid", "", uuid.NewString())
		assert.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.JSONEq(t, `{"data":{"success":false,"message":"Validate set failed"}}`, readString(resp.Body))

		assertLogsContain(t, "Validate set failed", "must supply a case number when not creating a new case")
	})

	t.Run("when sirius fails", func(t *testing.T) {
		xml, err := os.ReadFile("./testdata/xml/LP1F-valid.xml")
		assert.NoError(t, err)

		set := xmlToSet("LP1F", xml, "", "", "bad-batch")

		req, err := http.NewRequest(http.MethodPost, host+"/api/ddc", strings.NewReader(set))
		assert.NoError(t, err)

		req.Header.Add("Content-Type", "text/xml")
		req.Header.Add("Cookie", "membrane="+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.JSONEq(t, `{"data":{"success":false,"message":"Failed to create case stub in Sirius"}}`, readString(resp.Body))

		assertLogsContain(t, "Failed to create case stub in Sirius", "failed to send request to Sirius: client create case stub: received 400 response from Sirius")
	})

	t.Run("with validation errors", func(t *testing.T) {
		xml, err := os.ReadFile("./testdata/xml/LP1F-invalid-dates.xml")
		assert.NoError(t, err)

		set := xmlToSet("LP1F", xml, "", "", "")

		req, err := http.NewRequest(http.MethodPost, host+"/api/ddc", strings.NewReader(set))
		assert.NoError(t, err)

		req.Header.Add("Content-Type", "text/xml")
		req.Header.Add("Cookie", "membrane="+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.JSONEq(t, `{"data":{"success":false,"message":"XML for LP1F failed XSD validation","validationErrors":["Element 'Page10': This element is not expected. Expected is ( Page1 )."]}}`, readString(resp.Body))

		assertLogsContain(t, "Validate and sanitize XML failed", "XML for LP1F failed XSD validation")
	})

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

	t.Run("same file, but initially fails", func(t *testing.T) {
		sameID := uuid.NewString()
		assert.ErrorContains(t, checkFile(token, "LP1F", "LPC-valid", sameID), `"success":false`)
		assert.NoError(t, checkFile(token, "LP1F", "LP1F-valid", sameID))
		assert.Equal(t, ingestion.AlreadyProcessedError{}, checkFile(token, "LP1F", "LP1F-valid", sameID))
	})
}

func readString(r io.Reader) string {
	body, _ := io.ReadAll(r)
	return string(body)
}

func assertLogsContain(t *testing.T, msg, error string) bool {
	cmd := exec.Command("docker", "compose", "logs", "-n", "5", "app")
	logged, err := cmd.Output()
	if !assert.Nil(t, err) {
		return false
	}

	for line := range bytes.SplitSeq(logged, []byte{'\n'}) {
		line = bytes.TrimPrefix(line, []byte("app-1  | "))
		var v map[string]any
		_ = json.Unmarshal(line, &v)

		if v["msg"] == msg {
			return assert.Equal(t, error, v["error"])
		}
	}

	assert.Contains(t, string(logged), "message of '"+msg+"', and error '"+error+"'")
	return false
}

func assertLogsContainMessage(t *testing.T, msg string) bool {
	cmd := exec.Command("docker", "compose", "logs", "-n", "5", "app")
	logged, err := cmd.Output()
	if !assert.Nil(t, err) {
		return false
	}

	for line := range bytes.SplitSeq(logged, []byte{'\n'}) {
		line = bytes.TrimPrefix(line, []byte("app-1  | "))
		var v map[string]any
		_ = json.Unmarshal(line, &v)

		if v["msg"] == msg {
			return assert.Nil(t, v["error"])
		}
	}

	assert.Contains(t, string(logged), "message of '"+msg+"'")
	return false
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

	uploadResponse, err := upload(token, fileType, fileName, "", id)
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
	resp, err := upload(token, fileType, fileName, "7000-1234-1234", id)
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
	xml, err := os.ReadFile("./testdata/xml/" + fileName + ".xml")
	if err != nil {
		return nil, err
	}
	set := xmlToSet(fileType, xml, caseNo, id, "")

	req, err := http.NewRequest(http.MethodPost, host+"/api/ddc", strings.NewReader(set))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "text/xml")
	req.Header.Add("Cookie", "membrane="+token)

	return http.DefaultClient.Do(req)
}

func xmlToSet(fileType string, xml []byte, caseNo, id, scheduleID string) string {
	xmlB64 := base64.StdEncoding.EncodeToString(xml)

	if id == "" {
		id = uuid.NewString()
	}

	if scheduleID == "" {
		scheduleID = "01-0001253-20160909174150"
	}

	return `<?xml version="1.0" encoding="UTF-8"?>
<Set xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="SET.xsd">
<Header CaseNo="` + caseNo + `" Scanner="9" ScanTime="2014-09-26 12:38:53" ScannerOperator="Administrator" Schedule="` + scheduleID + `" />
<Body>
<Document Type="` + fileType + `" Encoding="UTF-8" NoPages="1" ID="` + id + `">
<XML>` + xmlB64 + `</XML>
<PDF>SUkqAAoAAAAAtBEAAAEDAAEAAAABAAAAAQEDAAEAAAABAAAAAgEDAAEAAAABAAAAAwEDAAEAAAABAAAABgEDAAEAAAABAAAACgEDAAEAAAABAAAADQECAAEAAAAAAAAAEQEEAAEAAAAIAAAAEgEDAAEAAAABAAAAFQEDAAEAAAABAAAAFgEDAAEAAAAAIAAAFwEEAAEAAAABAAAAGgEFAAEAAADcAAAAGwEFAAEAAADkAAAAHAEDAAEAAAABAAAAKAEDAAEAAAACAAAAKQEDAAIAAAAAAAEAAAAAAEgAAAABAAAASAAAAAEAAAA=</PDF>
</Document>
</Body>
</Set>`
}
