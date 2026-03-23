package jira

import "strings"

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
