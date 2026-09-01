# pd-slack

This tool syncs the on-call users from one or more [PagerDuty](https://www.pagerduty.com/) schedules to a [Slack](https://slack.com/) user group.

While there are other tools like this, I couldn't find one that worked how I wanted.

## Features

- Creates a Slack group if it does not exist
- Can sync multiple schedules to a single Slack group (IE: oncall level 1 and oncall level 2 can be synced to `@oncall`)
- Runs as a standalone binary, a Docker image, or a [GitHub Action](#github-action)

## GitHub Action

The easiest way to use this is as a GitHub Action. Add a workflow to your repository:

```yaml
name: Sync on-call

on:
  schedule:
    - cron: '*/15 * * * *'
  workflow_dispatch:

permissions:
  contents: read

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Sync on-call to Slack
        uses: imdevinc/pd-slack@v1
        with:
          slack-bot-token: ${{ secrets.SLACK_BOT_TOKEN }}
          pagerduty-api-token: ${{ secrets.PAGERDUTY_API_TOKEN }}
          slack-group-name: oncall
          pagerduty-schedule-ids: ABCD00E, ABCD00F
```

### Action inputs

| Input | Required | Description |
| --- | --- | --- |
| `slack-bot-token` | Yes | Slack Bot User OAuth token with `usergroups:read` and `usergroups:write` scopes. Store it as a secret. |
| `pagerduty-api-token` | Yes | PagerDuty API token. Store it as a secret. |
| `config-file` | No | Path to a YAML config file (see [below](#configuration)) for syncing multiple groups. When set, this takes precedence over the single-group inputs. Relative paths are resolved against the checkout directory. |
| `slack-group-name` | No | Name of the Slack user group to sync. Ignored when `config-file` is set. |
| `slack-group-description` | No | Description used when the Slack group is created. Ignored when `config-file` is set. |
| `pagerduty-schedule-ids` | No | Comma-separated list of PagerDuty schedule IDs to sync. Ignored when `config-file` is set. |

For more than one Slack group, commit a config file and reference it:

```yaml
steps:
  - name: Sync on-call to Slack
    uses: imdevinc/pd-slack@v1
    with:
      slack-bot-token: ${{ secrets.SLACK_BOT_TOKEN }}
      pagerduty-api-token: ${{ secrets.PAGERDUTY_API_TOKEN }}
      config-file: .github/pd-slack.yaml
```

## Configuration

Copy `config.yaml.sample` to `config.yaml` and provide the details.

- The PagerDuty schedule ID can be found by clicking the schedule in your PagerDuty schedules page and using the ID in the URL.
- The Slack bot needs the following permissions:
  - `usergroups:read`
  - `usergroups:write`

```yaml
slack_bot_token: xoxb-...
pagerduty_api_token: pdus-...
slack_groups:
  - name: oncall
    pagerduty_schedule_ids:
      - ABCD00E
      - ABCD00F
  - name: secondary-oncall
    pagerduty_schedule_ids:
      - ABCD00G
```

The tokens can also be supplied through the `SLACK_BOT_TOKEN` and `PAGERDUTY_API_TOKEN` environment variables, which override the values in the config file.

## Local usage

1. Download the latest release from https://github.com/ImDevinC/pd-slack/releases
1. Create a PagerDuty API token
1. Create a Slack bot with the following permissions:
   - `usergroups:read`
   - `usergroups:write`
1. Copy `config.yaml.sample` to `config.yaml` and provide the details
1. Run `pd-slack`

## Docker

```sh
docker build -t pd-slack .
docker run --rm \
  -e SLACK_BOT_TOKEN=xoxb-... \
  -e PAGERDUTY_API_TOKEN=pdus-... \
  -v "$(pwd)/config.yaml:/config.yaml" \
  pd-slack
```