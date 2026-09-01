package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	envConfigFile            = "INPUT_CONFIG-FILE"
	envSlackGroupName        = "INPUT_SLACK-GROUP-NAME"
	envSlackGroupDescription = "INPUT_SLACK-GROUP-DESCRIPTION"
	envPagerdutyScheduleIDs  = "INPUT_PAGERDUTY-SCHEDULE-IDS"
)

type Config struct {
	SlackBotToken     string       `yaml:"slack_bot_token"`
	PagerdutyAPIToken string       `yaml:"pagerduty_api_token"`
	SlackGroups       []slackGroup `yaml:"slack_groups"`
}

type slackGroup struct {
	Name                 string   `yaml:"name"`
	Description          string   `yaml:"description"`
	PagerdutyScheduleIDs []string `yaml:"pagerduty_schedule_ids"`
}

// GroupNames returns the names of the configured slack groups.
func (c *Config) GroupNames() []string {
	names := make([]string, 0, len(c.SlackGroups))
	for _, group := range c.SlackGroups {
		names = append(names, group.Name)
	}
	return names
}

// Get reads the configuration from the specified YAML file and environment variables.
// A missing config file is not an error; the returned config is populated from
// environment variables and GitHub Actions inputs alone.
func Get(file string) (*Config, error) {
	cfg := Config{}
	if data, err := os.ReadFile(file); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	applyEnvironment(&cfg)
	return &cfg, nil
}

// applyEnvironment overlays values from plain environment variables and
// GitHub Actions inputs onto the config parsed from the YAML file.
// A config file explicitly supplied through the action's config-file input
// takes precedence over the single-group inputs.
func applyEnvironment(cfg *Config) {
	if token := os.Getenv("SLACK_BOT_TOKEN"); token != "" {
		cfg.SlackBotToken = token
	}
	if token := os.Getenv("INPUT_SLACK-BOT-TOKEN"); token != "" {
		cfg.SlackBotToken = token
	}
	if token := os.Getenv("PAGERDUTY_API_TOKEN"); token != "" {
		cfg.PagerdutyAPIToken = token
	}
	if token := os.Getenv("INPUT_PAGERDUTY-API-TOKEN"); token != "" {
		cfg.PagerdutyAPIToken = token
	}

	if os.Getenv(envConfigFile) != "" {
		return
	}

	name := os.Getenv(envSlackGroupName)
	scheduleIDs := os.Getenv(envPagerdutyScheduleIDs)
	if name == "" || scheduleIDs == "" {
		return
	}
	cfg.SlackGroups = append(cfg.SlackGroups, slackGroup{
		Name:                 name,
		Description:          os.Getenv(envSlackGroupDescription),
		PagerdutyScheduleIDs: splitScheduleIDs(scheduleIDs),
	})
}

func splitScheduleIDs(raw string) []string {
	parts := strings.Split(raw, ",")
	ids := make([]string, 0, len(parts))
	for _, part := range parts {
		if id := strings.TrimSpace(part); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}
