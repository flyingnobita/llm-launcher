package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func isEscapeKey(msg tea.KeyPressMsg) bool {
	switch msg.String() {
	case "esc", "escape":
		return true
	}
	return msg.Key().Code == tea.KeyEscape
}

func isCtrlC(msg tea.KeyPressMsg) bool {
	s := strings.ToLower(strings.TrimSpace(msg.String()))
	if s == "ctrl+c" || s == "^c" {
		return true
	}
	k := msg.Key()
	return k.Mod.Contains(tea.ModCtrl) && (k.Code == 'c' || k.Text == "c" || k.Text == "C")
}

func isEnterKey(msg tea.KeyPressMsg) bool {
	if msg.String() == "enter" {
		return true
	}
	return msg.Key().Code == tea.KeyEnter
}

// updateServerSplitKeys handles input while a split-pane server is running.
// Tab switches focus between the model table and the log viewport; see [Model.server.splitFocused].
func (m Model) updateServerSplitKeys(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if m.server.exited {
		switch {
		case isTabKey(msg):
			m.server.splitFocused = !m.server.splitFocused
			if m.server.splitFocused {
				m.table.tbl.Blur()
			} else {
				m.table.tbl.Focus()
			}
			m = m.applySplitPaneFocusStyles()
			return m, nil
		case isEnterKey(msg), key.Matches(msg, m.keys.Quit), isEscapeKey(msg), isCtrlC(msg):
			m = m.dismissSplitServer()
			return m, nil
		}
		if !m.server.splitFocused {
			return m.updateServerSplitTableKeys(msg)
		}
		var cmd tea.Cmd
		m.server.viewport, cmd = m.server.viewport.Update(msg)
		return m, cmd
	}
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m.stopSplitServer()
	case isEscapeKey(msg):
		return m.stopSplitServer()
	case isCtrlC(msg):
		return m.stopSplitServer()
	case isTabKey(msg):
		m.server.splitFocused = !m.server.splitFocused
		if m.server.splitFocused {
			m.table.tbl.Blur()
		} else {
			m.table.tbl.Focus()
		}
		m = m.applySplitPaneFocusStyles()
		return m, nil
	}
	if !m.server.splitFocused {
		return m.updateServerSplitTableKeys(msg)
	}
	var cmd tea.Cmd
	m.server.viewport, cmd = m.server.viewport.Update(msg)
	return m, cmd
}

func isTabKey(msg tea.KeyPressMsg) bool {
	if msg.String() == "tab" {
		return true
	}
	return msg.Key().Code == tea.KeyTab
}

func (m Model) stopSplitServer() (Model, tea.Cmd) {
	if m.server.exited {
		m = m.dismissSplitServer()
		return m, nil
	}
	_ = interruptServerProcess(m.server.cmd)
	return m, nil
}
