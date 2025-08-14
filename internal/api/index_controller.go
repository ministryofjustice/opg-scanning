package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ministryofjustice/opg-scanning/internal/auth"
	"github.com/ministryofjustice/opg-scanning/internal/aws"
	"github.com/ministryofjustice/opg-scanning/internal/config"
	"github.com/ministryofjustice/opg-scanning/internal/ingestion"
	"github.com/ministryofjustice/opg-scanning/internal/logger"
	"github.com/ministryofjustice/opg-scanning/internal/sirius"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Auth interface {
	Authenticate(w http.ResponseWriter, r *http.Request) (auth.AuthenticatedUser, error)
	Check(next http.Handler) http.HandlerFunc
}

type AwsClient interface {
	PersistFormData(ctx context.Context, body io.Reader, docType string) (string, error)
	PersistSetData(ctx context.Context, body []byte) (string, error)
	QueueSetForProcessing(ctx context.Context, scannedCaseResponse *sirius.ScannedCaseResponse, fileName string) (string, error)
}

type worker interface {
	Process(ctx context.Context, bodyStr string) (*sirius.ScannedCaseResponse, error)
}

type IndexController struct {
	config    *config.Config
	logger    *slog.Logger
	auth      Auth
	worker    worker
	awsClient AwsClient
}

type response struct {
	Data responseData `json:"data"`
}

type responseData struct {
	Success          bool     `json:"success"`
	Message          string   `json:"message"`
	Uid              string   `json:"uid,omitempty"`
	ValidationErrors []string `json:"validationErrors,omitempty"`
}

var uidReplacementRegex = regexp.MustCompile(`^7[0-9]{3}-[0-9]{4}-[0-9]{4}$`)

func NewIndexController(logger *slog.Logger, awsClient aws.AwsClientInterface, appConfig *config.Config, dynamoClient *dynamodb.Client) *IndexController {
	siriusService := sirius.NewService(appConfig)

	return &IndexController{
		config:    appConfig,
		logger:    logger,
		auth:      auth.New(appConfig, logger, awsClient),
		worker:    ingestion.NewWorker(logger, appConfig, siriusService, awsClient, dynamoClient),
		awsClient: awsClient,
	}
}

func (c *IndexController) HandleRequests() {
	http.HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte("OK")); err != nil {
			c.logger.ErrorContext(r.Context(), err.Error())
		}
	})

	// Create the route to handle user authentication and issue JWT token
	http.HandleFunc("/auth/sessions", c.authHandler)

	// Protect the route with JWT validation (using the authMiddleware)
	http.Handle("/api/ddc", otelhttp.NewHandler(logger.UseTelemetry(
		c.auth.Check(http.HandlerFunc(c.ingestHandler)),
	), "scanning"))

	c.logger.Info("Starting server on :" + c.config.HTTP.Port)

	server := &http.Server{
		Addr:              ":" + c.config.HTTP.Port,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		c.logger.Error(err.Error())
	}
}

func (c *IndexController) authHandler(w http.ResponseWriter, r *http.Request) {
	// Define response error struct
	type ErrorResponse struct {
		Error string `json:"error"`
	}

	// Authenticate user credentials and issue JWT token
	user, err := c.auth.Authenticate(w, r)
	if err != nil {
		errMsg := fmt.Sprintf("Authentication failed: %v", err)
		c.logger.ErrorContext(r.Context(), errMsg)
		c.authResponse(r.Context(), w, ErrorResponse{Error: errMsg})
		return
	}

	c.authResponse(r.Context(), w, user)
}

func (c *IndexController) authResponse(ctx context.Context, w http.ResponseWriter, resp any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.respondWithError(ctx, w, http.StatusInternalServerError, "Failed to encode response", err)
	}
}

func (c *IndexController) ingestHandler(w http.ResponseWriter, r *http.Request) {
	reqCtx := r.Context()

	if r.Method != http.MethodPost {
		c.respondWithError(reqCtx, w, http.StatusMethodNotAllowed, "Invalid HTTP method", nil)
		return
	}

	c.logger.InfoContext(reqCtx, "Received ingestion request")

	bodyStr, err := c.readRequestBody(r)
	if err != nil {
		c.respondWithError(reqCtx, w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/xml") && !strings.HasPrefix(contentType, "text/xml") {
		c.respondWithError(reqCtx, w, http.StatusBadRequest, "Invalid content type", fmt.Errorf("expected application/xml or text/xml, got %s", contentType))
		return
	}

	// Save Set to S3
	filename, err := c.awsClient.PersistSetData(reqCtx, []byte(bodyStr))
	if err != nil {
		c.respondWithError(reqCtx, w, http.StatusInternalServerError, "Could not persist set to S3", err)
		return
	}

	c.logger.InfoContext(reqCtx, "Stored Set data", slog.String("set_filename", filename))

	scannedCaseResponse, err := c.worker.Process(context.WithoutCancel(reqCtx), bodyStr)
	if scannedCaseResponse == nil {
		scannedCaseResponse = &sirius.ScannedCaseResponse{}
	}

	uid := scannedCaseResponse.UID
	statusCode := http.StatusAccepted

	if err != nil {
		var aperr ingestion.AlreadyProcessedError
		if errors.As(err, &aperr) {
			uid = aperr.CaseNo
			statusCode = http.StatusAlreadyReported
		} else {
			statusCode, message := getPublicError(err, scannedCaseResponse.UID)
			c.respondWithError(reqCtx, w, statusCode, message, err)

			return
		}
	}

	// Send the UID response
	if uidReplacementRegex.MatchString(uid) {
		uid = strings.ReplaceAll(uid, "-", "")
	}

	resp := response{
		Data: responseData{
			Success: true,
			Message: fmt.Sprintf("The document set for case %s has been queued for processing", uid),
			Uid:     uid,
		},
	}

	if statusCode == http.StatusAlreadyReported {
		resp.Data.Message = "Document has already been processed"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.respondWithError(reqCtx, w, http.StatusInternalServerError, "Failed to encode response", err)
	} else {
		c.logger.InfoContext(reqCtx, "Ingestion request processed successfully", slog.String("uid", scannedCaseResponse.UID))
	}
}

func getPublicError(err error, uid string) (int, string) {
	if errors.Is(err, ingestion.ErrScannedCaseResponseUIDMissing) {
		return http.StatusInternalServerError, "Invalid response from Sirius when creating case stub, scannedCaseResponse is nil or missing UID"
	}

	var stubError ingestion.FailedToCreateCaseStubError
	if errors.As(err, &stubError) {
		return http.StatusInternalServerError, "Failed to create case stub in Sirius"
	}

	var setError ingestion.ValidateSetError
	if errors.As(err, &setError) {
		return http.StatusBadRequest, "Validate set failed"
	}

	var sanitizeError ingestion.ValidateAndSanitizeError
	if errors.As(err, &sanitizeError) {
		return http.StatusBadRequest, "Validate and sanitize XML failed"
	}

	var clientError sirius.Error
	if errors.As(err, &clientError) {
		switch clientError.StatusCode {
		case http.StatusBadRequest:
			_, ok := clientError.ValidationErrors["caseReference"]
			if ok {
				return http.StatusBadRequest, fmt.Sprintf("%s is not a valid case UID", uid)
			}
		case http.StatusNotFound:
			return http.StatusBadRequest, fmt.Sprintf("Case not found with UID %s", uid)
		case http.StatusRequestEntityTooLarge:
			return http.StatusRequestEntityTooLarge, "Request content too large: the XML document exceeds the maximum allowed size"
		}
	}

	return http.StatusInternalServerError, "Failed to persist document to Sirius"
}

func (c *IndexController) respondWithError(ctx context.Context, w http.ResponseWriter, statusCode int, message string, err error) {
	if statusCode >= 500 {
		c.logger.ErrorContext(ctx, message, slog.Any("error", err))
	} else {
		c.logger.InfoContext(ctx, message, slog.Any("error", err))
	}

	resp := response{
		Data: responseData{
			Success: false,
			Message: message,
		},
	}

	var problem ingestion.Problem
	if errors.As(err, &problem) {
		resp.Data.Message = problem.Title
		resp.Data.ValidationErrors = problem.ValidationErrors
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.respondWithError(ctx, w, http.StatusInternalServerError, "Failed to encode response", err)
	}
}

// Helper Method: Validate and Sanitize XML
func (c *IndexController) readRequestBody(r *http.Request) (string, error) {
	if r.Body == nil {
		return "", errors.New("request body is empty")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	defer r.Body.Close() //nolint:errcheck // no need to check error when closing body
	return string(body), nil
}
