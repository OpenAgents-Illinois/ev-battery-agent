package jira

import (
	"encoding/json"
	"fmt"
	"strings"

	jiralib "github.com/andygrunwald/go-jira"
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
	labels := []string{"ev-battery-agent", "ev-battery", severityToLabel(args.Severity)}

	issue := &jiralib.Issue{
		Fields: &jiralib.IssueFields{
			Project:     jiralib.Project{Key: c.projectKey},
			Summary:     fmt.Sprintf("[%s] EV Battery Alert: %s (VIN: %s)", args.Severity, args.DefectType, args.VIN),
			Type:        jiralib.IssueType{Name: issueType},
			Priority:    &jiralib.Priority{Name: priority},
			Description: "Reasoning: " + args.TechnicalReason,
			Labels:      labels,
		},
	}

	created, _, err := c.api.Issue.Create(issue)
	if err != nil {
		return "FAILED: Could not create Jira issue: " + err.Error()
	}
	return fmt.Sprintf("SUCCESS: Ticket created with Key: %s [%s / %s priority, labels=%s]", created.Key, issueType, priority, strings.Join(labels, ","))
}
