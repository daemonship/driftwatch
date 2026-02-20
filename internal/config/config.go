// Package config handles loading and validating the driftwatch.yml configuration file.
package config

// Config represents the top-level driftwatch.yml configuration.
type Config struct {
	Workspaces   []string `yaml:"workspaces"`
	SlackWebhook string   `yaml:"slack_webhook,omitempty"`
	Binary       string   `yaml:"binary,omitempty"`
}

// Load reads and parses the config file at path.
// Returns an error if the file cannot be read or is malformed.
func Load(path string) (*Config, error) {
	// TODO: implement in Task 2
	return nil, nil
}
