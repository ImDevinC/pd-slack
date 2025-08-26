package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SlackToken        string       `yaml:"slack_token"`
	PagerdutyAPIToken string       `yaml:"pagerduty_api_token"`
	SlackGroups       []slackGroup `yaml:"slack_groups"`
}

type slackGroup struct {
	Name                 string   `yaml:"name"`
	Description          string   `yaml:"description"`
	PagerdutyScheduleIDs []string `yaml:"pagerduty_schedule_ids"`
}

// Get reads the configuration from the specified YAML file and environment variables.
func Get(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}
	if token := os.Getenv("SLACK_TOKEN"); token != "" {
		cfg.SlackToken = token
	}
	if token := os.Getenv("PAGERDUTY_API_TOKEN"); token != "" {
		cfg.PagerdutyAPIToken = token
	}
	return &cfg, nil
}
