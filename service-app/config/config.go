package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ministryofjustice/opg-scanning/internal/util"
	"gopkg.in/yaml.v2"
)

type Config struct {
	XSDDirectory string `yaml:"xsd_directory"`
}

var config *Config

func LoadConfig() (*Config, error) {
	projectRoot, err := util.GetProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("could not determine project root: %w", err)
	}

	configPath := filepath.Join(projectRoot, "service-app/config/config.yml")

	data, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	config = &cfg

	cfg.XSDDirectory = filepath.Join(projectRoot, cfg.XSDDirectory)

	return config, nil
}

func GetConfig() *Config {
	if config == nil {
		log.Fatal("configuration not loaded")
	}
	return config
}
