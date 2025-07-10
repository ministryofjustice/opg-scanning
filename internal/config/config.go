package config

import (
	"cmp"
	"fmt"
	"os"
	"time"
)

type (
	Config struct {
		App  app
		Aws  aws
		Auth Auth
		HTTP http
	}

	app struct {
		Environment        string
		SiriusBaseURL      string
		SiriusCaseStubURL  string
		SiriusAttachDocURL string
		XSDPath            string
	}

	aws struct {
		JobsQueueURL          string
		JobsQueueBucket       string
		JobsQueueBucketKmsKey string
		Endpoint              string
		Region                string
	}

	Auth struct {
		ApiUsername    string
		JWTSecretARN   string
		CredentialsARN string
		JWTExpiration  time.Duration
	}

	http struct {
		Port    string
		Timeout time.Duration
	}
)

func Environment() string {
	return os.Getenv("ENVIRONMENT")
}

// Loads configuration from environment variables.
func Read() (*Config, error) {
	var (
		httpTimeout, jwtExpiration time.Duration
		err                        error
	)

	if val := os.Getenv("HTTP_TIMEOUT"); val != "" {
		httpTimeout, err = time.ParseDuration(val)
		if err != nil {
			return nil, fmt.Errorf("failed to load environment variables into config 'HTTP_TIMEOUT': %w", err)
		}
	}
	if val := os.Getenv("JWT_EXPIRATION"); val != "" {
		jwtExpiration, err = time.ParseDuration(val)
		if err != nil {
			return nil, fmt.Errorf("failed to load environment variables into config 'JWT_EXPIRATION': %w", err)
		}
	}

	return &Config{
		App: app{
			Environment:        Environment(),
			SiriusBaseURL:      os.Getenv("SIRIUS_BASE_URL"),
			SiriusCaseStubURL:  cmp.Or(os.Getenv("SIRIUS_CASE_STUB_URL"), "api/public/v1/scanned-cases"),
			SiriusAttachDocURL: cmp.Or(os.Getenv("SIRIUS_ATTACH_DOC_URL"), "api/public/v1/scanned-documents"),
			XSDPath:            cmp.Or(os.Getenv("XSD_PATH"), "xsd"),
		},
		Aws: aws{
			JobsQueueURL:          cmp.Or(os.Getenv("JOBQUEUE_SQS_QUEUE_URL"), "000000000000/ddc.fifo"),
			JobsQueueBucket:       cmp.Or(os.Getenv("JOBQUEUE_S3_BUCKET_NAME"), "opg-backoffice-jobsqueue-local"),
			JobsQueueBucketKmsKey: cmp.Or(os.Getenv("JOBQUEUE_S3_ENCRYPTION_KEY"), "alias/aws/s3"),
			Endpoint:              os.Getenv("AWS_ENDPOINT"),
			Region:                cmp.Or(os.Getenv("AWS_REGION"), "eu-west-1"),
		},
		Auth: Auth{
			ApiUsername:    cmp.Or(os.Getenv("API_USERNAME"), "opg_document_and_d@publicguardian.gsi.gov.uk"),
			JWTSecretARN:   cmp.Or(os.Getenv("JWT_SECRET_ARN"), "local/jwt-key"),
			CredentialsARN: cmp.Or(os.Getenv("CREDENTIALS_ARN"), "/local/local-credentials"),
			JWTExpiration:  cmp.Or(jwtExpiration, time.Hour),
		},
		HTTP: http{
			Port:    cmp.Or(os.Getenv("HTTP_PORT"), "8081"),
			Timeout: cmp.Or(httpTimeout, 10*time.Second),
		},
	}, nil
}
