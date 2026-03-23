package agent

import (
	"fmt"
	"strings"
)

const systemPrompt = `You are an EV Battery Specialist. Analyze battery telemetry and determine if safety thresholds are violated.
Always analyze the telemetry data provided, regardless of its format.
When calling the fileEngineeringTicket tool, use only plain text — no special characters, quotes, or newlines in the vin, defectType, or technicalReason arguments.
Keep technicalReason under 80 characters.
Choose severity based on risk: CRITICAL for immediate safety hazards (thermal runaway, fire risk), WARNING for out-of-range but non-immediate issues, INFO for anomalies that need monitoring.`

const interactivePromptTmpl = `A Rivian EV owner has reported the following battery issue. Analyze it, determine if any safety thresholds are violated, and if so file a Jira ticket using the fileEngineeringTicket tool with severity CRITICAL, WARNING, or INFO.

Report: %s`

// DetectModel detects R1S or R1T from free text. Defaults to R1S.
func DetectModel(text string) string {
	if strings.Contains(strings.ToUpper(text), "R1T") {
		return "R1T"
	}
	return "R1S"
}

// InteractivePrompt formats a plain-English battery report for the agent.
func InteractivePrompt(report string) string {
	return fmt.Sprintf(interactivePromptTmpl, report)
}
