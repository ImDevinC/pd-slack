package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
)

type Client struct {
	api *slack.Client
}

func New(token string, options ...slack.Option) *Client {
	api := slack.New(token, options...)
	return &Client{
		api: api,
	}
}

// CheckForGroup checks if a Slack user group with the given name exists
// and returns its ID if found, or an empty string if not found
func (c *Client) CheckForGroup(ctx context.Context, name string) (string, error) {
	groups, err := c.api.GetUserGroupsContext(ctx)
	if err != nil {
		return "", err
	}
	for _, group := range groups {
		if group.Handle == name {
			return group.ID, nil
		}
	}
	return "", nil
}

// CreateGroup creates a new Slack user group with the given name and description
func (c *Client) CreateGroup(ctx context.Context, name string, description string) (string, error) {
	group, err := c.api.CreateUserGroupContext(ctx, slack.UserGroup{
		Name:        name,
		Description: description,
	})
	if err != nil {
		return "", err
	}
	return group.ID, nil
}

// UpdateUserGroupMembers updates a Slack user group from member email addresses.
func (c *Client) UpdateUserGroupMembers(ctx context.Context, groupID string, emails []string) error {
	if len(emails) == 0 {
		return fmt.Errorf("cannot update Slack user group %q without members", groupID)
	}

	memberIDs := make([]string, 0, len(emails))
	seen := make(map[string]struct{}, len(emails))
	for _, email := range emails {
		user, err := c.api.GetUserByEmailContext(ctx, email)
		if err != nil {
			return fmt.Errorf("failed to find Slack user for %q: %w", email, err)
		}
		if user.ID == "" {
			return fmt.Errorf("Slack user for %q has no ID", email)
		}
		if _, ok := seen[user.ID]; ok {
			continue
		}
		seen[user.ID] = struct{}{}
		memberIDs = append(memberIDs, user.ID)
	}

	memberList := strings.Join(memberIDs, ",")
	_, err := c.api.UpdateUserGroupMembersContext(ctx, groupID, memberList)
	if err != nil {
		return fmt.Errorf("failed to update Slack user group %q: %w", groupID, err)
	}
	return nil
}
