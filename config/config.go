package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type (
	Config struct {
		App  app
		Aws  aws
		Auth Auth
		HTTP http
	}

	app struct {
		Environment        string `envconfig:"ENVIRONMENT"`
		SiriusBaseURL      string `envconfig:"SIRIUS_BASE_URL"`
		SiriusCaseStubURL  string `envconfig:"SIRIUS_CASE_STUB_URL" default:"api/public/v1/scanned-cases"`
		SiriusAttachDocURL string `envconfig:"SIRIUS_ATTACH_DOC_URL" default:"api/public/v1/scanned-documents"`
		XSDPath            string `envconfig:"XSD_PATH" default:"xsd"`
	}

	aws struct {
		JobsQueueURL          string `envconfig:"JOBQUEUE_SQS_QUEUE_URL" default:"000000000000/ddc.fifo"`
		JobsQueueBucket       string `envconfig:"JOBQUEUE_S3_BUCKET_NAME" default:"opg-backoffice-jobsqueue-local"`
		JobsQueueBucketKmsKey string `envconfig:"JOBQUEUE_S3_ENCRYPTION_KEY" default:"alias/aws/s3"`
		Endpoint              string `envconfig:"AWS_ENDPOINT"`
		Region                string `envconfig:"AWS_REGION" default:"eu-west-1"`
	}

	Auth struct {
		ApiUsername    string `envconfig:"API_USERNAME" default:"opg_document_and_d@publicguardian.gsi.gov.uk"`
		JWTSecretARN   string `envconfig:"JWT_SECRET_ARN" default:"local/jwt-key"`
		CredentialsARN string `envconfig:"CREDENTIALS_ARN" default:"/local/local-credentials"`
		JWTExpiration  int    `envconfig:"JWT_EXPIRATION" default:"3600"`
	}

	http struct {
		Port    string `envconfig:"HTTP_PORT" default:"8081"`
		Timeout int    `envconfig:"HTTP_TIMEOUT" default:"10"`
	}
)

// Loads configuration from .env and environment variables.
func NewConfig() *Config {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("failed to load environment variables into config: %v", err)
	}

	return &cfg
}
