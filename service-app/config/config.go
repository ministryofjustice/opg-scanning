package config

import (
	"log"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type (
	Config struct {
		App  App
		HTTP HTTP
		Auth Auth
	}

	App struct {
		SiriusBaseURL   string `envconfig:"SIRIUS_BASE_URL" default:"http://localhost:8080"`
		SiriusScanURL   string `envconfig:"SIRIUS_SCAN_URL" default:"api/public/v1/scanned-cases"`
		ProjectPath     string `envconfig:"PROJECT_PATH" default:"service-app"`
		ProjectFullPath string
	}

	Auth struct {
		Email            string        `envconfig:"AUTH_EMAIL" default:"opg_document_and_d@publicguardian.gsi.gov.uk"`
		Password         string        `envconfig:"AUTH_PASSWORD" default:"password"`
		RefreshThreshold time.Duration `envconfig:"AUTH_REFRESH_THRESHOLD" default:"5m"`
	}

	HTTP struct {
		Port    string `envconfig:"HTTP_PORT" default:"8082"`
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
