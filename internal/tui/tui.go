package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/OpenAgents-Illinois/ev-battery-agent/internal/agent"
)

// Start launches the Bubble Tea TUI in full-screen mode.
func Start(f *agent.Factory) error {
	p := tea.NewProgram(
		newModel(f),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	return err
}
