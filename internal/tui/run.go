// Package tui implements the Bubble Tea terminal UI for LLM Launcher (llml):
// model discovery, runtime detection, table rendering, and llama-server launch.
package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

// Run starts the full-screen TUI. Use [tea.WithAltScreen] for a dedicated screen buffer;
// omit it if you need to stay in the scrollback (e.g. logging to the terminal).
func Run() error {
	p := tea.NewProgram(New())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}
