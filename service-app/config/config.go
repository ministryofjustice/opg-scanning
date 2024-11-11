package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/ministryofjustice/opg-scanning/internal/util"
)

type (
	// Config holds application configuration data.
	Config struct {
		App  `yaml:"app"`
		HTTP `yaml:"http"`
	}

	// App configuration fields.
	App struct {
		Name            string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version         string `env-required:"true" yaml:"version" env:"APP_VERSION"`
		ProjectFullPath string `env-required:"true" env:"PROJECT_FULL_PATH"`
		ProjectPath     string `env-required:"true" yaml:"project_path" env:"PROJECT_PATH"`
	}

	// HTTP server configuration fields.
	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}
)

// NewConfig loads configuration from .env, config.yml, and environment variables.
func NewConfig() *Config {
	// Determine project root
	projectRoot, err := util.GetProjectRoot()
	if err != nil {
		log.Fatalf("failed to get project root: %v", err)
	}

	// Load environment variables from .env if present
	if err := godotenv.Load(filepath.Join(projectRoot, ".env")); err != nil {
		log.Println("No .env file found or could not be loaded. Falling back to OS environment variables.")
	}

	// Ensure PROJECT_PATH is set before using it to set PROJECT_FULL_PATH
	projectPath := os.Getenv("PROJECT_PATH")
	if projectPath == "" {
		log.Fatal("PROJECT_PATH environment variable is required but not set")
	}

	// Set PROJECT_FULL_PATH using projectRoot and PROJECT_PATH
	projectFullPath := filepath.Join(projectRoot, projectPath)
	if err := os.Setenv("PROJECT_FULL_PATH", projectFullPath); err != nil {
		log.Fatalf("failed to set PROJECT_FULL_PATH: %v", err)
	}

	cfg := &Config{}

	// Load configuration from config.yml and environment variables, with env vars taking precedence
	configPath := filepath.Join(projectFullPath, "config/config.yml")
	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		log.Fatalf("failed to load config from %s: %v", configPath, err)
	}

	return cfg
}
