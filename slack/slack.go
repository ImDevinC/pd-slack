package slack

import (
	"context"
	"strings"

	"github.com/slack-go/slack"
)

type Client struct {
	api *slack.Client
}

func New(token string) *Client {
	api := slack.New(token)
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

// UpdateUserGroupMembers updates the members of a Slack user group with the given ID
func (c *Client) UpdateUserGroupMembers(ctx context.Context, groupID string, members []string) error {
	memberList := strings.Join(members, ",")
	_, err := c.api.UpdateUserGroupMembersContext(ctx, groupID, memberList)
	return err
}
