package tui

import (
	"fmt"
	"strings"
)

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	header := headerStyle.Render("⚡  EV BATTERY SAFETY AGENT  ⚡")
	subHeader := subHeaderStyle.Render("Gemini 2.0 Flash  ·  Rivian R1S / R1T Manuals  ·  Jira Auto-Ticketing")

	conversation := borderStyle.
		Width(m.width - 2).
		Render(m.viewport.View())

	inputLabel := dimStyle.Render(" Describe the battery issue — press Enter to analyze ")
	inputBox := inputBorderStyle.Width(m.width - 4).Render(m.textinput.View())

	vehicleText := vehicleStyle.Render(fmt.Sprintf("  Vehicle: %s  ", m.vehicle))
	sep := dimStyle.Render(" │ ")
	hint := dimStyle.Render("  Ctrl+C to exit  ")
	statusBar := vehicleText + sep + m.renderStatus() + sep + hint

	return strings.Join([]string{
		header,
		subHeader,
		"",
		conversation,
		inputLabel,
		inputBox,
		statusBar,
	}, "\n")
}

func (m model) renderLines() string {
	return strings.Join(m.lines, "\n")
}

func (m model) renderStatus() string {
	prefix := ""
	if m.processing {
		prefix = m.spinner.View() + " "
	}
	text := "  " + prefix + m.status + "  "
	switch m.statusKind {
	case "critical", "error":
		return statusCriticalStyle.Render(text)
	case "warning":
		return statusWarningStyle.Render(text)
	case "analyzing":
		return statusAnalyzingStyle.Render(text)
	default:
		return statusReadyStyle.Render(text)
	}
}

func ticketStatus(lower string) string {
	if strings.Contains(lower, "ticket") || strings.Contains(lower, "kan-") || strings.Contains(lower, "created") {
		return "Ticket filed"
	}
	return "Review needed"
}

func wordWrap(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}
	words := strings.Fields(text)
	var lines []string
	var current strings.Builder
	for _, w := range words {
		if current.Len()+len(w)+1 > maxWidth && current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(w)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return strings.Join(lines, "\n      ")
}
