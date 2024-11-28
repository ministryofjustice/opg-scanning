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
		Auth Auth
		HTTP HTTP
	}

	App struct {
		AwsEndpoint     string `envconfig:"AWS_ENDPOINT" default:"http://localstack:4566"`
		SiriusBaseURL   string `envconfig:"SIRIUS_BASE_URL" default:"http://localhost:8080"`
		SiriusScanURL   string `envconfig:"SIRIUS_SCAN_URL" default:"api/public/v1/scanned-cases"`
		ProjectPath     string `envconfig:"PROJECT_PATH" default:"service-app"`
		ProjectFullPath string
	}

	Auth struct {
		JWTSecretARN  string `envconfig:"JWT_SECRET_ARN" default:"local/jwt-key"`
		JWTExpiration int    `envconfig:"JWT_EXPIRATION" default:"300"`
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
