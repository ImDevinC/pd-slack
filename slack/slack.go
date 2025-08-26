package slack

import "github.com/slack-go/slack"

type Client struct {
	api *slack.Client
}

func New(token string) *Client {
	api := slack.New(token)
	return &Client{
		api: api,
	}
}
