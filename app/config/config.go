package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Temporal     *TemporalConfig
	IsProduction bool
	API          *APIConfig
}

func MustNewConfig(dir string, env string) *Config {
	cfg, err := NewConfig(dir, env)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize config: %v", err))
	}
	return cfg
}
func NewConfig(dir string, env string) (*Config, error) {

	// Read all environment-specific files
	files := []string{
		fmt.Sprintf("api.%s.yaml", env),
		fmt.Sprintf("temporal.%s.yaml", env),
	}

	var config Config

	for _, filename := range files {
		filePath := filepath.Join(dir, filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filename, err)
		}

		// Unmarshal into the same config struct - YAML will merge
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s: %w", filename, err)
		}
	}
	config.IsProduction = strings.ToLower(env) == "production"
	return &config, nil
}
