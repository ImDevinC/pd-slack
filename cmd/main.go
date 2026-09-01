package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/imdevinc/pd-slack/config"
	"github.com/imdevinc/pd-slack/pagerduty"
	"github.com/imdevinc/pd-slack/slack"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	cfg, err := config.Get(resolveConfigPath())
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	if err := validateConfig(cfg); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}
	slog.Info("config loaded", "groups", cfg.GroupNames())

	pdClient := pagerduty.New(cfg.PagerdutyAPIToken)
	slackClient := slack.New(cfg.SlackBotToken)

	policyCache := make(map[string][]string)
	ctx := context.Background()

	for _, group := range cfg.SlackGroups {
		slog.Info("getting oncall users for group", "group", group.Name)
		oncallUsers := []string{}
		for _, schedule := range group.PagerdutyScheduleIDs {
			if users, ok := policyCache[schedule]; ok {
				oncallUsers = append(oncallUsers, users...)
				continue
			}
			users, err := pdClient.GetOnCallUsersForSchedule(ctx, schedule)
			if err != nil {
				slog.Error("failed to get oncall users", "error", err, "schedule", schedule)
				continue
			}
			oncallUsers = append(oncallUsers, users...)
			policyCache[schedule] = users
			slog.Info("fetched oncall users", "count", len(users), "schedule", schedule)
		}
		slog.Info("total oncall users for group", "group", group.Name, "count", len(oncallUsers))
		groupID, err := slackClient.CheckForGroup(ctx, group.Name)
		if err != nil {
			slog.Error("failed to check for group", "error", err, "group", group.Name)
			continue
		}
		if groupID == "" {
			slog.Info("group not found, creating", "group", group.Name)
			groupID, err = slackClient.CreateGroup(ctx, group.Name, group.Description)
			if err != nil {
				slog.Error("failed to create group", "error", err, "group", group.Name)
				continue
			}
			slog.Info("created group", "group", group.Name, "id", groupID)
		}
		err = slackClient.UpdateUserGroupMembers(ctx, groupID, oncallUsers)
		if err != nil {
			slog.Error("failed to update group members", "error", err, "group", group.Name)
			continue
		}
		slog.Info("updated group members", "group", group.Name, "count", len(oncallUsers))
	}
}

// resolveConfigPath returns the config file to load. When run as a GitHub
// Action, relative paths are resolved against the checked out workspace so
// users can point the config-file input at files in their repo.
func resolveConfigPath() string {
	path := os.Getenv("INPUT_CONFIG-FILE")
	if path == "" {
		path = "config.yaml"
	}
	if !filepath.IsAbs(path) {
		if workspace := os.Getenv("GITHUB_WORKSPACE"); workspace != "" {
			return filepath.Join(workspace, path)
		}
	}
	return path
}

func validateConfig(cfg *config.Config) error {
	if cfg.SlackBotToken == "" {
		return fmt.Errorf("slack bot token is required (set via the slack-bot-token input, the SLACK_BOT_TOKEN environment variable, or the config file)")
	}
	if cfg.PagerdutyAPIToken == "" {
		return fmt.Errorf("pagerduty API token is required (set via the pagerduty-api-token input, the PAGERDUTY_API_TOKEN environment variable, or the config file)")
	}
	if len(cfg.SlackGroups) == 0 {
		return fmt.Errorf("at least one slack group is required (set via the config file or the slack-group-name and pagerduty-schedule-ids inputs)")
	}
	return nil
}
