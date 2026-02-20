// Package config handles loading and validating the driftwatch.yml configuration file.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level driftwatch.yml configuration.
type Config struct {
	Workspaces   []string `yaml:"workspaces"`
	SlackWebhook string   `yaml:"slack_webhook,omitempty"`
	Binary       string   `yaml:"binary,omitempty"`
}

// Load reads and parses the config file at path.
// Returns an error if the file cannot be read or is malformed.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Ensure Workspaces is at least an empty slice, not nil
	if cfg.Workspaces == nil {
		cfg.Workspaces = []string{}
	}

	return &cfg, nil
}
