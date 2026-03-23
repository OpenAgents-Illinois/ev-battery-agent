package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Keep these bars at a stable height.
	// If they wrap, Bubble Tea may leave "ghost" lines from previous renders.
	availableWidth := m.width - 2
	if availableWidth < 1 {
		availableWidth = 1
	}

	// Header/subheader should be 1-line; if the terminal is too narrow they may wrap,
	// but the primary ghosting issue is the bottom bar, which we clip separately.
	header := headerStyle.Render("EV Battery Agent")
	subHeader := subHeaderStyle.Render("Gemini 2.0 Flash  ·  Rivian R1S / R1T Manuals  ·  Jira Auto-Ticketing")
	mainBoxWidth := m.width - 2
	if mainBoxWidth < 3 {
		mainBoxWidth = 3
	}

	conversation := borderStyle.
		Width(mainBoxWidth - borderStyle.GetHorizontalFrameSize()).
		Render(m.viewport.View())

	inputInnerWidth := mainBoxWidth - inputBorderStyle.GetHorizontalFrameSize()
	if inputInnerWidth < 1 {
		inputInnerWidth = 1
	}
	inputView := strings.SplitN(m.textinput.View(), "\n", 2)[0]
	inputView = lipgloss.NewStyle().
		MaxWidth(inputInnerWidth).
		Width(inputInnerWidth).
		Render(inputView)

	inputBox := inputBorderStyle.
		Width(inputInnerWidth).
		Height(1).
		Render(inputView)

	vehicleText := vehicleStyle.Render(fmt.Sprintf("  Vehicle: %s  ", m.vehicle))
	sep := dimStyle.Render(" │ ")
	hint := dimStyle.Render("  Ctrl+C to exit  ")

	vehicleRaw := fmt.Sprintf("  Vehicle: %s  ", m.vehicle)
	sepRaw := " │ "
	hintRaw := "  Ctrl+C to exit  "
	// Subtract an extra cell as a safety margin to avoid accidental wrapping.
	statusMaxWidth := availableWidth - lipgloss.Width(vehicleRaw) - (2 * lipgloss.Width(sepRaw)) - lipgloss.Width(hintRaw) - 1
	if statusMaxWidth < 1 {
		statusMaxWidth = 1
	}

	statusBar := vehicleText + sep + m.renderStatus(statusMaxWidth) + sep + hint

	return strings.Join([]string{
		header,
		subHeader,
		"",
		conversation,
		inputBox,
		statusBar,
	}, "\n")
}

func (m model) renderLines() string {
	return strings.Join(m.lines, "\n")
}

func (m model) renderStatus(maxWidth int) string {
	prefix := ""
	if m.processing {
		prefix = m.spinner.View() + " "
	}
	text := "  " + prefix + m.status + "  "
	text = clipTextToWidth(text, maxWidth)
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

func clipTextToWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxWidth {
		return s
	}

	// Simple rune-based truncation that respects the visible width.
	// Input strings are small (header/status), so O(n^2) is fine here.
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r)) > maxWidth {
		r = r[:len(r)-1]
	}
	return string(r)
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

var inlineBulletRe = regexp.MustCompile(`\s\*\s+`)

func formatAgentResponse(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}

	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	// Turn inline markdown bullets into actual lines.
	normalized = inlineBulletRe.ReplaceAllString(normalized, "\n* ")
	normalized = strings.ReplaceAll(normalized, "**", "")
	normalized = strings.ReplaceAll(normalized, "`", "")

	var out []string
	lines := strings.Split(normalized, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "- ") {
			item := strings.TrimSpace(line[2:])
			wrapped := wordWrap(item, maxWidth-2)
			wrapped = strings.ReplaceAll(wrapped, "\n      ", "\n  ")
			out = append(out, "• "+wrapped)
			continue
		}

		out = append(out, wordWrap(line, maxWidth))
	}

	return strings.Join(out, "\n      ")
}
