package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ev-battery-agent/internal/agent"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("51")).
			Bold(true).
			Padding(0, 1)

	subHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238"))

	statusReadyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82")).
				Bold(true)

	statusAnalyzingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("220")).
				Bold(true)

	statusCriticalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)

	statusWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true)

	vehicleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("51"))

	userMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("75")).
			Bold(true)

	agentMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	errMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	inputBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(0, 1)
)

// ── Tea message types ─────────────────────────────────────────────────────────

type agentResultMsg struct {
	text    string
	vehicle string
}

type agentErrMsg struct{ err error }

// ── Model ─────────────────────────────────────────────────────────────────────

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

// ── Init ──────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerH := 4  // title + subtitle + padding
		statusH := 1  // status bar
		inputH := 3   // input box with border
		footerH := 1  // hint line
		vpHeight := m.height - headerH - inputH - statusH - footerH - 4
		if vpHeight < 5 {
			vpHeight = 5
		}
		vpWidth := m.width - 4
		if !m.ready {
			m.viewport = viewport.New(vpWidth, vpHeight)
			m.viewport.SetContent(m.renderLines())
			m.ready = true
		} else {
			m.viewport.Width = vpWidth
			m.viewport.Height = vpHeight
		}
		m.textinput.Width = vpWidth - 4

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if !m.processing {
				return m.submit()
			}
		}

	case agentResultMsg:
		m.processing = false
		m.addLine("")
		m.addLine(agentMsgStyle.Render("Agent › ") + wordWrap(msg.text, m.width-10))
		m.vehicle = msg.vehicle

		lower := strings.ToLower(msg.text)
		switch {
		case strings.Contains(lower, "critical"):
			m.setStatus("CRITICAL — "+ticketStatus(lower), "critical")
		case strings.Contains(lower, "warning"):
			m.setStatus("WARNING — "+ticketStatus(lower), "warning")
		default:
			m.setStatus("Ready", "ready")
		}
		m.viewport.SetContent(m.renderLines())
		m.viewport.GotoBottom()

	case agentErrMsg:
		m.processing = false
		m.addLine(errMsgStyle.Render("Error › ") + msg.err.Error())
		m.setStatus("Error — check output", "error")
		m.viewport.SetContent(m.renderLines())
		m.viewport.GotoBottom()

	case spinner.TickMsg:
		if m.processing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Update sub-components
	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) submit() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(m.textinput.Value())
	if text == "" {
		return m, nil
	}
	m.textinput.SetValue("")
	m.processing = true

	vehicle := agent.DetectModel(text)
	m.vehicle = vehicle
	m.addLine("")
	m.addLine(userMsgStyle.Render("You › ") + text)
	m.setStatus("Analyzing...", "analyzing")
	m.viewport.SetContent(m.renderLines())
	m.viewport.GotoBottom()

	prompt := agent.InteractivePrompt(text)
	factory := m.factory

	return m, tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			result, err := factory.Chat(context.Background(), vehicle, prompt)
			if err != nil {
				return agentErrMsg{err}
			}
			return agentResultMsg{text: result, vehicle: vehicle}
		},
	)
}

func (m *model) addLine(line string) {
	m.lines = append(m.lines, line)
}

func (m *model) setStatus(text, kind string) {
	m.status = text
	m.statusKind = kind
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Header
	header := headerStyle.Render("⚡  EV BATTERY SAFETY AGENT  ⚡")
	subHeader := subHeaderStyle.Render("Gemini 2.0 Flash  ·  Rivian R1S / R1T Manuals  ·  Jira Auto-Ticketing")

	// Conversation viewport
	conversation := borderStyle.
		Width(m.width - 2).
		Render(m.viewport.View())

	// Input box
	inputLabel := dimStyle.Render(" Describe the battery issue — press Enter to analyze ")
	inputContent := m.textinput.View()
	inputBox := inputBorderStyle.Width(m.width - 4).Render(inputContent)

	// Status bar
	statusText := m.renderStatus()
	vehicleText := vehicleStyle.Render(fmt.Sprintf("  Vehicle: %s  ", m.vehicle))
	hint := dimStyle.Render("  Ctrl+C to exit  ")
	sep := dimStyle.Render(" │ ")
	statusBar := vehicleText + sep + statusText + sep + hint

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
	var spinnerText string
	if m.processing {
		spinnerText = m.spinner.View() + " "
	}
	text := spinnerText + m.status
	switch m.statusKind {
	case "critical":
		return statusCriticalStyle.Render("  " + text + "  ")
	case "warning":
		return statusWarningStyle.Render("  " + text + "  ")
	case "analyzing":
		return statusAnalyzingStyle.Render("  " + text + "  ")
	case "error":
		return statusCriticalStyle.Render("  " + text + "  ")
	default:
		return statusReadyStyle.Render("  " + text + "  ")
	}
}

func ticketStatus(lower string) string {
	if strings.Contains(lower, "ticket") || strings.Contains(lower, "kan-") || strings.Contains(lower, "created") {
		return "Ticket filed"
	}
	return "Review needed"
}

// wordWrap wraps text at maxWidth characters.
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

// ── Entry point ───────────────────────────────────────────────────────────────

// Start launches the Bubble Tea TUI.
func Start(f *agent.Factory) error {
	p := tea.NewProgram(
		newModel(f),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
