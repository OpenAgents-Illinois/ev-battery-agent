package jira

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
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

// FileTicket parses the LLM tool call arguments JSON and creates a Jira issue.
// Returns a human-readable result string (SUCCESS / SKIPPED / FAILED / ERROR).
func (c *Client) FileTicket(argsJSON string) string {
	if c.domain == "" || c.email == "" || c.token == "" || c.projectKey == "" {
		return "ERROR: Jira configuration missing. Set JIRA_TOKEN, JIRA_SPACE_KEY, JIRA_DOMAIN, and JIRA_EMAIL env vars."
	}

	var args struct {
		VIN             string `json:"vin"`
		DefectType      string `json:"defectType"`
		TechnicalReason string `json:"technicalReason"`
		Severity        string `json:"severity"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "ERROR: Failed to parse tool arguments: " + err.Error()
	}

	args.VIN = sanitize(args.VIN)
	args.DefectType = sanitize(args.DefectType)
	args.TechnicalReason = sanitize(args.TechnicalReason)
	args.Severity = strings.ToUpper(sanitize(args.Severity))

	// Deduplication: skip if an open ticket for this VIN already exists
	if existing := c.findExistingTicket(args.VIN); existing != "" {
		return fmt.Sprintf("SKIPPED: Open ticket already exists for VIN %s: %s", args.VIN, existing)
	}

	issueType := severityToIssueType(args.Severity)
	priority := severityToPriority(args.Severity)

	payload := map[string]any{
		"fields": map[string]any{
			"project":   map[string]string{"key": c.projectKey},
			"summary":   fmt.Sprintf("[%s] EV Battery Alert: %s (VIN: %s)", args.Severity, args.DefectType, args.VIN),
			"issuetype": map[string]string{"name": issueType},
			"priority":  map[string]string{"name": priority},
			"description": map[string]any{
				"type":    "doc",
				"version": 1,
				"content": []any{
					map[string]any{
						"type": "paragraph",
						"content": []any{
							map[string]string{"type": "text", "text": "Reasoning: " + args.TechnicalReason},
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "ERROR: Failed to marshal Jira payload: " + err.Error()
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("https://%s/rest/api/3/issue", c.domain),
		strings.NewReader(string(body)),
	)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	req.Header.Set("Authorization", "Basic "+c.encodedAuth())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "ERROR: Could not connect to Jira: " + err.Error()
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 201 {
		var result map[string]any
		ticketKey := "UNKNOWN"
		if err := json.Unmarshal(respBody, &result); err == nil {
			if key, ok := result["key"].(string); ok {
				ticketKey = key
			}
		}
		return fmt.Sprintf("SUCCESS: Ticket created with Key: %s [%s / %s priority]", ticketKey, issueType, priority)
	}
	return fmt.Sprintf("FAILED: Jira API returned %d. Error: %s", resp.StatusCode, string(respBody))
}

// findExistingTicket searches for an open ticket for the given VIN.
// Returns the ticket key if found, empty string otherwise.
func (c *Client) findExistingTicket(vin string) string {
	jql := fmt.Sprintf(`project = "%s" AND summary ~ "%s" AND statusCategory != Done`, c.projectKey, vin)
	apiURL := fmt.Sprintf("https://%s/rest/api/3/search?jql=%s&maxResults=1&fields=key,summary",
		c.domain, url.QueryEscape(jql))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Basic "+c.encodedAuth())
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}
	body, _ := io.ReadAll(resp.Body)

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return ""
	}
	total, _ := result["total"].(float64)
	if total == 0 {
		return ""
	}
	issues, ok := result["issues"].([]any)
	if !ok || len(issues) == 0 {
		return ""
	}
	issue, ok := issues[0].(map[string]any)
	if !ok {
		return ""
	}
	key, _ := issue["key"].(string)
	return key
}

func (c *Client) encodedAuth() string {
	return base64.StdEncoding.EncodeToString([]byte(c.email + ":" + c.token))
}

func severityToIssueType(severity string) string {
	switch severity {
	case "CRITICAL", "EMERGENCY":
		return "Bug"
	default:
		return "Task"
	}
}

func severityToPriority(severity string) string {
	switch severity {
	case "CRITICAL", "EMERGENCY":
		return "Highest"
	case "WARNING":
		return "High"
	default:
		return "Medium"
	}
}

func sanitize(s string) string {
	s = strings.ReplaceAll(s, `"`, "'")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "°", " degrees")
	return strings.TrimSpace(s)
}
