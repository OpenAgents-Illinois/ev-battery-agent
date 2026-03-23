package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/OpenAgents-Illinois/ev-battery-agent/internal/agent"
)

type model struct {
	factory    *agent.Factory
	viewport   viewport.Model
	textinput  textinput.Model
	spinner    spinner.Model
	lines      []string
	status     string
	statusKind string // "ready" | "analyzing" | "critical" | "warning" | "error"
	vehicle    string
	processing bool
	width      int
	height     int
	ready      bool
}

func newModel(f *agent.Factory) model {
	ti := textinput.New()
	ti.Placeholder = "Describe the battery issue (e.g. R1S VIN ABC123 temp 65C voltage 2.8V)..."
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("230"))
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 60

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))

	welcome := []string{
		"Welcome! Describe any EV battery issue in plain English.",
		"",
		"  Examples:",
		"    My R1S VIN ABC123 battery is at 62 degrees celsius, voltage 2.9V, only 15% charge while driving",
		"    R1T warning light on — temp showing 58C, VIN XYZ789",
		"    Battery degraded significantly, R1S VIN DEF456, only charging to 70% max",
	}

	return model{
		factory:    f,
		textinput:  ti,
		spinner:    sp,
		lines:      welcome,
		status:     "Ready",
		statusKind: "ready",
		vehicle:    "R1S",
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}
