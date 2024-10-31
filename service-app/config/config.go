package config

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	XSDDirectory string `yaml:"xsd_directory"`
}

var config *Config

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	config = &cfg
	return config, nil
}

func GetConfig() *Config {
	if config == nil {
		log.Fatal("configuration not loaded")
	}
	return config
}
