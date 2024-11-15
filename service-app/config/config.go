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
		SiriusBaseURL   string `required:"true" envconfig:"SIRIUS_BASE_URL"`
		ProjectPath     string `required:"true" envconfig:"PROJECT_PATH"`
		ProjectFullPath string
	}

	// HTTP server configuration fields.
	HTTP struct {
		Port string `required:"true" envconfig:"HTTP_PORT"`
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
