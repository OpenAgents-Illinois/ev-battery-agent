package jira

import (
	"encoding/base64"
	"os"
)

// Client handles Jira REST API calls. Config is read from environment at construction time.
type Client struct {
	domain     string
	email      string
	token      string
	projectKey string
}

// NewClient reads Jira config from environment variables.
func NewClient() *Client {
	return &Client{
		domain:     os.Getenv("JIRA_DOMAIN"),
		email:      os.Getenv("JIRA_EMAIL"),
		token:      os.Getenv("JIRA_TOKEN"),
		projectKey: os.Getenv("JIRA_SPACE_KEY"),
	}
}

func (c *Client) encodedAuth() string {
	return base64.StdEncoding.EncodeToString([]byte(c.email + ":" + c.token))
}

func (c *Client) isConfigured() bool {
	return c.domain != "" && c.email != "" && c.token != "" && c.projectKey != ""
}
