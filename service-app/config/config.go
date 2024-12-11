package config

import (
	"log"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type (
	Config struct {
		App  App
		Aws  Aws
		Auth Auth
		HTTP HTTP
	}

	App struct {
		Environment     string `envconfig:"ENVIRONMENT"`
		SiriusBaseURL   string `envconfig:"SIRIUS_BASE_URL"`
		SiriusScanURL   string `envconfig:"SIRIUS_SCAN_URL" default:"api/public/v1/scanned-cases"`
		ProjectPath     string `envconfig:"PROJECT_PATH" default:"service-app"`
		ProjectFullPath string
	}

	Aws struct {
		Endpoint string `envconfig:"AWS_ENDPOINT"`
		Region   string `envconfig:"AWS_REGION" default:"eu-west-1"`
	}

	Auth struct {
		ApiUsername   string `envconfig:"API_USERNAME" default:"opg_document_and_d@publicguardian.gsi.gov.uk"`
		JWTSecretARN  string `envconfig:"JWT_SECRET_ARN" default:"local/jwt-key"`
		JWTExpiration int    `envconfig:"JWT_EXPIRATION" default:"3600"`
	}

	HTTP struct {
		Port    string `envconfig:"HTTP_PORT" default:"8081"`
		Timeout int    `envconfig:"HTTP_TIMEOUT" default:"10"`
	}
)

// Loads configuration from .env and environment variables.
func NewConfig() *Config {
	projectRoot, err := util.GetProjectRoot()
	if err != nil {
		log.Fatalf("failed to get project root: %v", err)
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("failed to load environment variables into config: %v", err)
	}

	cfg.App.ProjectFullPath = filepath.Join(projectRoot, cfg.App.ProjectPath)

	return &cfg
}
