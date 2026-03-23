package agent

import "strings"

const (
	SeverityCritical = "CRITICAL"
	SeverityWarning  = "WARNING"
	SeverityInfo     = "INFO"
	SeverityUnknown  = "UNKNOWN"
)

// DetectSeverity infers severity from model output text.
func DetectSeverity(text string) string {
	upper := strings.ToUpper(text)
	switch {
	case strings.Contains(upper, "CRITICAL"), strings.Contains(upper, "EMERGENCY"):
		return SeverityCritical
	case strings.Contains(upper, "WARNING"):
		return SeverityWarning
	case strings.Contains(upper, "INFO"):
		return SeverityInfo
	default:
		return SeverityUnknown
	}
}
