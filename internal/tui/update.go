package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"ev-battery-agent/internal/agent"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerH := 3 // title + subtitle + blank line
		inputH := 3  // input box with border (top border + content + bottom border)
		statusH := 1 // status bar
		vpHeight := m.height - headerH - inputH - statusH - 2
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
