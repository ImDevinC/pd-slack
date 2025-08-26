package pagerduty

import (
	"context"
	"fmt"
	"time"

	"github.com/PagerDuty/go-pagerduty"
)

type Client struct {
	pd *pagerduty.Client
}

func New(token string) *Client {
	pd := pagerduty.NewClient(token)
	return &Client{
		pd: pd,
	}
}

func (c *Client) GetOnCallUsersForSchedule(ctx context.Context, id string) ([]string, error) {
	allUsers := []string{}
	opts := pagerduty.ListOnCallUsersOptions{
		Since: time.Now().UTC().Format(time.RFC3339),
		Until: time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339),
	}
	users, err := c.pd.ListOnCallUsersWithContext(ctx, id, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get oncall users: %w", err)
	}
	for _, user := range users {
		allUsers = append(allUsers, user.Email)
	}
	return allUsers, nil
}
