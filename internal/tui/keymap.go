package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds key bindings. Add fields here as your TUI grows, and wire them in Update.
type KeyMap struct {
	Quit key.Binding
}

// DefaultKeyMap returns the default global shortcuts.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp satisfies key.KeyMap (optional; use for help overlay later).
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}

// FullHelp satisfies key.KeyMap.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit},
	}
}
