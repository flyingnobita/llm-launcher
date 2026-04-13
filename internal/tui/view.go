package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

const appTitle = "llm-launch"

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "\n  Initializing…\n"
	}

	title := titleStyle.Render(appTitle)
	sub := subtitleStyle.Render("Bubble Tea + Lip Gloss · edit internal/tui to build your UI")
	body := bodyStyle.Render("Add state to Model, handle messages in Update, render in View.")
	footer := footerStyle.Render(fmt.Sprintf("%s · resize: %d×%d", m.keys.Quit.Help().Key, m.width, m.height))

	block := lipgloss.JoinVertical(lipgloss.Left, title, sub, "", body, footer)
	framed := app.Render(block)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, framed)
}
