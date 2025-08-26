# pd-slack
This tools allows you to sync multiple users from a PagerDuty oncall schedule to a Slack group.

While there are other tools like this, I couldn't find one that worked how I wanted.

## Features
- Creates slack group if it does not exist
- Can sync multiple schedules to a single slack group (IE: oncall level 1 and oncall level 2 can be synced to @oncall)

## Usage
1. Download latest release from https://github.com/ImDevinC/pd-slack/releases
1. Create a PagerDuty API token
1. Create a slack bot with the following permissions:
  - usergroups:read
  - usergroups:write
1. Copy `config.yaml.sample` to `config.yaml` and provide details
  - PagerDuty schedule ID can be found by clicking the schedule in your PagerDuty schedules page and using the ID in the URL
1. Run `pd-slack`

