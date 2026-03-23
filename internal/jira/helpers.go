package jira

import (
	"strings"

	jiralib "github.com/andygrunwald/go-jira"
)

// jiraSearchOptions used for deduplication searches.
var jiraSearchOptions = jiralib.SearchOptions{MaxResults: 1, Fields: []string{"key", "summary"}}

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
