package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	return path
}

func TestGetFromFile(t *testing.T) {
	path := writeTempFile(t, "config.yaml", `
slack_bot_token: slack-token
pagerduty_api_token: pd-token
slack_groups:
  - name: oncall
    description: Primary oncall
    pagerduty_schedule_ids:
      - ABCD00E
      - ABCD00F
  - name: secondary-oncall
    pagerduty_schedule_ids:
      - ABCD00G
`)

	cfg, err := Get(path)
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	if cfg.SlackBotToken != "slack-token" {
		t.Errorf("SlackBotToken = %q, want %q", cfg.SlackBotToken, "slack-token")
	}
	if cfg.PagerdutyAPIToken != "pd-token" {
		t.Errorf("PagerdutyAPIToken = %q, want %q", cfg.PagerdutyAPIToken, "pd-token")
	}
	if len(cfg.SlackGroups) != 2 {
		t.Fatalf("got %d slack groups, want 2", len(cfg.SlackGroups))
	}
	if cfg.SlackGroups[0].Name != "oncall" || cfg.SlackGroups[0].Description != "Primary oncall" {
		t.Errorf("first group = %+v, want oncall with description", cfg.SlackGroups[0])
	}
	if len(cfg.SlackGroups[0].PagerdutyScheduleIDs) != 2 {
		t.Errorf("first group has %d schedule ids, want 2", len(cfg.SlackGroups[0].PagerdutyScheduleIDs))
	}
}

func TestGetMissingFileIsNotAnError(t *testing.T) {
	cfg, err := Get(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err != nil {
		t.Fatalf("Get returned an error for a missing file: %v", err)
	}
	if cfg == nil {
		t.Fatal("Get returned a nil config")
	}
}

func TestGetEnvironmentTokensOverrideFile(t *testing.T) {
	path := writeTempFile(t, "config.yaml", "slack_bot_token: from-file\npagerduty_api_token: from-file\n")

	t.Setenv("SLACK_BOT_TOKEN", "from-env")
	t.Setenv("PAGERDUTY_API_TOKEN", "from-env")

	cfg, err := Get(path)
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	if cfg.SlackBotToken != "from-env" {
		t.Errorf("SlackBotToken = %q, want %q", cfg.SlackBotToken, "from-env")
	}
	if cfg.PagerdutyAPIToken != "from-env" {
		t.Errorf("PagerdutyAPIToken = %q, want %q", cfg.PagerdutyAPIToken, "from-env")
	}
}

func TestGetSingleGroupInputs(t *testing.T) {
	t.Setenv("INPUT_SLACK_BOT_TOKEN", "input-token")
	t.Setenv("INPUT_PAGERDUTY_API_TOKEN", "input-pd-token")
	t.Setenv("INPUT_SLACK_GROUP_NAME", "oncall")
	t.Setenv("INPUT_SLACK_GROUP_DESCRIPTION", "Level 1 oncall")
	t.Setenv("INPUT_PAGERDUTY_SCHEDULE_IDS", "ABCD00E, ABCD00F, ")

	cfg, err := Get(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	if cfg.SlackBotToken != "input-token" {
		t.Errorf("SlackBotToken = %q, want %q", cfg.SlackBotToken, "input-token")
	}
	if cfg.PagerdutyAPIToken != "input-pd-token" {
		t.Errorf("PagerdutyAPIToken = %q, want %q", cfg.PagerdutyAPIToken, "input-pd-token")
	}
	if len(cfg.SlackGroups) != 1 {
		t.Fatalf("got %d slack groups, want 1", len(cfg.SlackGroups))
	}
	group := cfg.SlackGroups[0]
	if group.Name != "oncall" || group.Description != "Level 1 oncall" {
		t.Errorf("group = %+v, want oncall with description", group)
	}
	want := []string{"ABCD00E", "ABCD00F"}
	if len(group.PagerdutyScheduleIDs) != len(want) {
		t.Fatalf("got schedule ids %v, want %v", group.PagerdutyScheduleIDs, want)
	}
	for i := range want {
		if group.PagerdutyScheduleIDs[i] != want[i] {
			t.Errorf("schedule ids = %v, want %v", group.PagerdutyScheduleIDs, want)
			break
		}
	}
}

func TestGetConfigFileInputTakesPrecedenceOverSingleGroup(t *testing.T) {
	t.Setenv("INPUT_CONFIG_FILE", "config.yaml")
	t.Setenv("INPUT_SLACK_GROUP_NAME", "ignored")
	t.Setenv("INPUT_PAGERDUTY_SCHEDULE_IDS", "ABCD00E")

	cfg, err := Get(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	if len(cfg.SlackGroups) != 0 {
		t.Errorf("got %d slack groups, want 0", len(cfg.SlackGroups))
	}
}

func TestGetPartialSingleGroupInputsAreIgnored(t *testing.T) {
	t.Setenv("INPUT_SLACK_GROUP_NAME", "oncall")
	// INPUT_PAGERDUTY_SCHEDULE_IDS intentionally left unset.

	cfg, err := Get(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	if len(cfg.SlackGroups) != 0 {
		t.Errorf("got %d slack groups, want 0", len(cfg.SlackGroups))
	}
}

func TestGroupNames(t *testing.T) {
	path := writeTempFile(t, "config.yaml", `
slack_groups:
  - name: oncall
    pagerduty_schedule_ids: [ABCD00E]
  - name: secondary-oncall
    pagerduty_schedule_ids: [ABCD00G]
`)

	cfg, err := Get(path)
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	got := cfg.GroupNames()
	want := []string{"oncall", "secondary-oncall"}
	if len(got) != len(want) {
		t.Fatalf("GroupNames = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("GroupNames = %v, want %v", got, want)
			break
		}
	}
}
