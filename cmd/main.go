package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/imdevinc/pd-slack/config"
	"github.com/imdevinc/pd-slack/pagerduty"
	_ "github.com/joho/godotenv/autoload"
)

type app struct {
	pagerdutyClient *pagerduty.Client
	policyCache     map[string][]string
}

func main() {
	cfg, err := config.Get("config.yaml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	slog.Info("config loaded", "config", cfg)

	pd := pagerduty.New(cfg.PagerdutyAPIToken)
	a := app{
		pagerdutyClient: pd,
	}
	if err := a.run(context.Background(), cfg); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func (a *app) run(ctx context.Context, cfg *config.Config) error {
	slog.Info("starting application")
	a.policyCache = make(map[string][]string)

	for _, group := range cfg.SlackGroups {
		slog.Info("processing group", "group", group.Name)
		oncallUsers := []string{}
		for _, schedule := range group.PagerdutyScheduleIDs {
			users, err := a.pagerdutyClient.GetOnCallUsersForSchedule(ctx, schedule)
			if err != nil {
				slog.Error("failed to get oncall users", "error", err, "schedule", schedule)
				continue
			}
			oncallUsers = append(oncallUsers, users...)
			a.policyCache[schedule] = users
			slog.Info("fetched oncall users", "count", len(users), "schedule", schedule)
		}
		slog.Info("total oncall users for group", "group", group.Name, "count", len(oncallUsers))

	}

	return nil
}
