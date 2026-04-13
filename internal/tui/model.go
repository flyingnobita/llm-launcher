package tui

import tea "github.com/charmbracelet/bubbletea"

// Model is the root Bubble Tea model. Extend this struct with your application state
// (navigation, forms, lists, API clients, etc.).
type Model struct {
	width  int
	height int
	keys   KeyMap
}

// New returns a model with default key bindings.
func New() Model {
	return Model{
		keys: DefaultKeyMap(),
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}
