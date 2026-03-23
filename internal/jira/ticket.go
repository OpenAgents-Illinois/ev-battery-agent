package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// FileTicket parses LLM tool call arguments JSON and creates a Jira issue.
// Returns a human-readable result string (SUCCESS / SKIPPED / FAILED / ERROR).
func (c *Client) FileTicket(argsJSON string) string {
	if !c.isConfigured() {
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
						"type":    "paragraph",
						"content": []any{map[string]string{"type": "text", "text": "Reasoning: " + args.TechnicalReason}},
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "ERROR: Failed to marshal Jira payload: " + err.Error()
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/rest/api/3/issue", c.domain), strings.NewReader(string(body)))
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
	if resp.StatusCode != 201 {
		return fmt.Sprintf("FAILED: Jira API returned %d. Error: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]any
	ticketKey := "UNKNOWN"
	if err := json.Unmarshal(respBody, &result); err == nil {
		if key, ok := result["key"].(string); ok {
			ticketKey = key
		}
	}
	return fmt.Sprintf("SUCCESS: Ticket created with Key: %s [%s / %s priority]", ticketKey, issueType, priority)
}
