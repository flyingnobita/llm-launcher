package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// runtimePanelView renders the bottom llama.cpp binary path section. contentWidth is the inner width for wrapping.
func runtimePanelView(m Model, contentWidth int) string {
	if m.width == 0 {
		return ""
	}
	if contentWidth < 24 {
		contentWidth = 24
	}
	var block string
	if !m.runtimeScanned && m.loading {
		block = "Detecting llama.cpp runtime…"
	} else {
		lines := m.runtime.BinaryPathLines(contentWidth)
		block = strings.Join(lines, "\n")
	}
	inner := "llama.cpp binaries\n" + block
	return runtimePanelStyle.Width(contentWidth).Render(inner)
}

const appTitle = "llm-launch"

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "\n  Initializing…\n"
	}

	title := titleStyle.Render(appTitle)
	sub := subtitleStyle.Render("llama.cpp · GGUF models (filesystem scan · Last modified = file mtime, not inference)")

	var body string
	switch {
	case m.loading:
		body = bodyStyle.Render("Scanning for models…")
	case m.loadErr != nil:
		body = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render("Error: " + m.loadErr.Error())
	case len(m.files) == 0:
		body = bodyStyle.Render("No GGUF files found. Set HUGGINGFACE_HUB_CACHE or HF_HOME if your Hub cache is non-default; add paths via LLM_LAUNCH_LLAMACPP_PATHS or place models under ~/models, ~/.cache/huggingface/hub, etc.")
	default:
		m.hscroll.SetContent(m.tbl.View())
		iw := m.bodyInnerW
		if iw < 1 && m.width > 0 {
			iw = max(m.width-4, 40)
		}
		th := m.tableBodyH
		if th < 1 {
			th = 18
		}
		m.hscroll.Width = iw
		m.hscroll.Height = th
		body = m.hscroll.View()
	}

	var hBar string
	if len(m.files) > 0 && m.tableLineWidth > 0 {
		iw := m.bodyInnerW
		if iw < 1 && m.width > 0 {
			iw = max(m.width-4, 40)
		}
		if m.tableLineWidth > iw {
			pct := m.hscroll.HorizontalScrollPercent()
			hBar = footerStyle.Render(horizontalScrollBarLine(pct, iw))
		}
	}

	help := fmt.Sprintf(
		"%s · %s · %s · ↑/↓ select · wheel · enter copy · ←/→ · %d×%d",
		m.keys.Refresh.Help().Key,
		m.keys.RunServer.Help().Key,
		m.keys.Quit.Help().Key,
		m.width,
		m.height,
	)
	footer := footerStyle.Render(help)

	iw := m.bodyInnerW
	if iw < 1 && m.width > 0 {
		iw = max(m.width-4, 40)
	}
	runtimePanel := runtimePanelView(m, iw)

	rows := []string{title, sub, "", body}
	if hBar != "" {
		rows = append(rows, hBar)
	}
	if runtimePanel != "" {
		rows = append(rows, runtimePanel)
	}
	rows = append(rows, "", footer)
	if m.lastRunNote != "" {
		noteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
		rows = append(rows, noteStyle.Render(m.lastRunNote))
	}
	block := lipgloss.JoinVertical(lipgloss.Left, rows...)
	framed := app.Render(block)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, framed)
}

// SelectedPath returns the full path of the highlighted row, or empty if none.
func (m Model) SelectedPath() string {
	if len(m.tbl.Rows()) == 0 || m.tbl.Cursor() < 0 {
		return ""
	}
	row := m.tbl.SelectedRow()
	if len(row) < 2 {
		return ""
	}
	// Path column is second cell; cells may be truncated — use backing slice.
	i := m.tbl.Cursor()
	if i < 0 || i >= len(m.files) {
		return ""
	}
	return m.files[i].Path
}

// horizontalScrollBarLine renders a filled track (█) and remainder (░) for horizontal scroll position.
func horizontalScrollBarLine(pct float64, maxWidth int) string {
	if maxWidth < 14 {
		return ""
	}
	inner := maxWidth - 4
	if inner < 8 {
		return ""
	}
	filled := int(pct * float64(inner))
	if filled > inner {
		filled = inner
	}
	if filled < 0 {
		filled = 0
	}
	return "  " + strings.Repeat("█", filled) + strings.Repeat("░", inner-filled) + "  "
}
