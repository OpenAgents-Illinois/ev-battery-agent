package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"ev-battery-agent/internal/agent"
)

// Start launches the Bubble Tea TUI in full-screen mode.
func Start(f *agent.Factory) error {
	p := tea.NewProgram(
		newModel(f),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
