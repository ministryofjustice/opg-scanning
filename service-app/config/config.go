package config

import (
	"log"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type (
	Config struct {
		App  App
		HTTP HTTP
	}

	// App configuration fields.
	App struct {
		Environment        string `envconfig:"ENVIRONMENT"`
		SiriusBaseURL      string `envconfig:"SIRIUS_BASE_URL"`
		SiriusCaseStubURL  string `envconfig:"SIRIUS_CASE_STUB_URL" default:"api/public/v1/scanned-cases"`
		SiriusAttachDocURL string `envconfig:"SIRIUS_ATTACH_DOC_URL" default:"api/public/v1/scanned-documents"`
		ProjectPath        string `envconfig:"PROJECT_PATH" default:"service-app"`
		ProjectFullPath    string
	}

	Aws struct {
		JobsQueueURL    string `envconfig:"JOBQUEUE_SQS_QUEUE_URL" default:"000000000000/ddc.fifo"`
		JobsQueueBucket string `envconfig:"JOBQUEUE_S3_BUCKET_NAME" default:"opg-backoffice-jobsqueue-local"`
		Endpoint        string `envconfig:"AWS_ENDPOINT"`
		Region          string `envconfig:"AWS_REGION" default:"eu-west-1"`
	}

	Auth struct {
		ApiUsername    string `envconfig:"API_USERNAME" default:"opg_document_and_d@publicguardian.gsi.gov.uk"`
		JWTSecretARN   string `envconfig:"JWT_SECRET_ARN" default:"local/jwt-key"`
		CredentialsARN string `envconfig:"CREDENTIALS_ARN" default:"/local/local-credentials"`
		JWTExpiration  int    `envconfig:"JWT_EXPIRATION" default:"3600"`
	}

	// HTTP server configuration fields.
	HTTP struct {
		Port string `required:"true" envconfig:"HTTP_PORT" default:"8081"`
	}
)

// Loads configuration from .env and environment variables.
func NewConfig() *Config {
	projectRoot, err := util.GetProjectRoot()
	if err != nil {
		log.Fatalf("failed to get project root: %v", err)
	}

	envPath := filepath.Join(projectRoot, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Println("No .env file found or could not be loaded. Falling back to OS environment variables.")
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("failed to load environment variables into config: %v", err)
	}

	cfg.App.ProjectFullPath = filepath.Join(projectRoot, cfg.App.ProjectPath)

	log.Println("Configuration loaded successfully.")

	return &cfg
}
