package jira

import (
	"fmt"
	"os"

	jiralib "github.com/andygrunwald/go-jira"
)

// Client wraps the go-jira client with project-specific config.
type Client struct {
	api        *jiralib.Client
	projectKey string
}

// NewClient reads Jira config from environment variables and returns a ready client.
// Returns a no-op client if config is missing (FileTicket will return an error message).
func NewClient() *Client {
	domain := os.Getenv("JIRA_DOMAIN")
	email := os.Getenv("JIRA_EMAIL")
	token := os.Getenv("JIRA_TOKEN")
	projectKey := os.Getenv("JIRA_SPACE_KEY")

	if domain == "" || email == "" || token == "" || projectKey == "" {
		return &Client{}
	}

	tp := jiralib.BasicAuthTransport{Username: email, Password: token}
	api, err := jiralib.NewClient(tp.Client(), fmt.Sprintf("https://%s", domain))
	if err != nil {
		return &Client{}
	}
	return &Client{api: api, projectKey: projectKey}
}

func (c *Client) isConfigured() bool {
	return c.api != nil && c.projectKey != ""
}
