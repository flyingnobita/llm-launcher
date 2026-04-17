package tui

import (
	"charm.land/lipgloss/v2"
)

func (m Model) paramPanelModalBlock() string {
	cw := m.paramPanelContentWidth()
	maxLine := cw
	if maxLine < 24 {
		maxLine = 24
	}
	secBox := m.ui.styles.paramSectionBox
	maxSec := cw - secBox.GetHorizontalFrameSize()
	if maxSec < 24 {
		maxSec = 24
	}

	title := m.modalTitleRow(cw, m.ui.styles.portConfigTitle, "Parameters — "+m.params.modelDisplayName)
	rows := []string{title, ""}

	if k := m.params.confirmDelete; k != paramConfirmNone {
		confirmBox := m.ui.styles.paramConfirmDialog
		confirmInner := cw - confirmBox.GetHorizontalFrameSize()
		if confirmInner < 24 {
			confirmInner = 24
		}
		var confirmRows []string
		switch k {
		case paramConfirmProfile:
			pName := ""
			if m.params.profileIndex >= 0 && m.params.profileIndex < len(m.params.profiles) {
				pName = m.params.profiles[m.params.profileIndex].Name
			}
			if pName == "" {
				pName = "(unnamed)"
			}
			nameLine := lipgloss.JoinHorizontal(lipgloss.Top,
				m.ui.styles.body.Render("  "),
				m.ui.styles.paramProfileName.Render(truncateParamLine(pName, confirmInner-2)),
			)
			confirmRows = []string{
				m.ui.styles.body.Render("Delete this parameter profile?"),
				nameLine,
			}
		case paramConfirmEnvRow:
			line := ""
			if m.params.envCursor >= 0 && m.params.envCursor < m.paramEnvLen() {
				line = formatEnvVar(m.params.env[m.params.envCursor])
			}
			confirmRows = []string{
				m.ui.styles.body.Render("Delete this environment variable line?"),
				m.ui.styles.body.Render("  " + truncateParamLine(line, max(confirmInner-2, 8))),
			}
		case paramConfirmArgRow:
			line := ""
			if m.params.argsCursor >= 0 && m.params.argsCursor < m.paramArgsLen() {
				line = m.params.args[m.params.argsCursor]
			}
			confirmRows = []string{
				m.ui.styles.body.Render("Delete this extra argument line?"),
				m.ui.styles.body.Render("  " + truncateParamLine(line, max(confirmInner-2, 8))),
			}
		}
		if len(confirmRows) > 0 {
			confirmRows = append(confirmRows, "",
				m.ui.styles.footer.Render(FooterParamConfirmYN),
			)
			rows = append(rows, confirmBox.Width(cw).Render(lipgloss.JoinVertical(lipgloss.Left, confirmRows...)))
			rows = append(rows, "")
		}
	}

	rows = append(rows, m.ui.styles.body.Render("  Profiles"))
	rows = append(rows, "")
	for i := range m.params.profiles {
		name := m.params.profiles[i].Name
		if name == "" {
			name = "(unnamed)"
		}
		focused := m.params.focus == paramFocusProfiles && i == m.params.profileIndex
		switch {
		case focused && m.params.editKind == paramEditProfileName:
			rows = append(rows, m.params.editInput.View())
		default:
			prefix := "  "
			if focused {
				prefix = "› "
			}
			pw := lipgloss.Width(prefix)
			nameW := maxLine - pw
			if nameW < 8 {
				nameW = maxLine
			}
			row := lipgloss.JoinHorizontal(lipgloss.Top,
				m.ui.styles.body.Render(prefix),
				m.ui.styles.paramProfileName.Render(truncateParamLine(name, nameW)),
			)
			rows = append(rows, row)
		}
	}
	if len(m.params.profiles) == 0 {
		rows = append(rows, m.ui.styles.body.Render("  (none)"))
	}

	rows = append(rows, "")
	var detailRows []string
	const sectionHeadingIndent = "  "
	envHeading := "Environment Variables (e.g. PYTORCH_CUDA_ALLOC_CONF=expandable_segments:True)"
	detailRows = append(detailRows, lipgloss.JoinHorizontal(lipgloss.Top,
		m.ui.styles.body.Render(sectionHeadingIndent),
		m.ui.styles.paramSectionHeading.Render(truncateParamLine(envHeading, maxSec-lipgloss.Width(sectionHeadingIndent))),
	))
	detailRows = append(detailRows, "")
	if m.paramEnvLen() == 0 && !(m.params.focus == paramFocusEnv && m.params.editKind == paramEditEnvLine) {
		prefix := "  "
		if m.params.focus == paramFocusEnv {
			prefix = "› "
		}
		detailRows = append(detailRows, m.ui.styles.body.Render(prefix+"(none)"))
	}
	for i := range m.params.env {
		line := formatEnvVar(m.params.env[i])
		focused := m.params.focus == paramFocusEnv && m.params.envCursor == i
		switch {
		case focused && m.params.editKind == paramEditEnvLine:
			detailRows = append(detailRows, m.params.editInput.View())
		default:
			prefix := "  "
			if focused {
				prefix = "› "
			}
			detailRows = append(detailRows, m.ui.styles.body.Render(prefix+truncateParamLine(line, maxSec)))
		}
	}

	detailRows = append(detailRows, "")

	argHeading := "Extra arguments (e.g. --max-model-len 131072)"
	detailRows = append(detailRows, lipgloss.JoinHorizontal(lipgloss.Top,
		m.ui.styles.body.Render(sectionHeadingIndent),
		m.ui.styles.paramSectionHeading.Render(truncateParamLine(argHeading, maxSec-lipgloss.Width(sectionHeadingIndent))),
	))
	detailRows = append(detailRows, "")
	if m.paramArgsLen() == 0 && !(m.params.focus == paramFocusArgs && m.params.editKind == paramEditArgLine) {
		prefix := "  "
		if m.params.focus == paramFocusArgs {
			prefix = "› "
		}
		detailRows = append(detailRows, m.ui.styles.body.Render(prefix+"(none)"))
	}
	for i := range m.params.args {
		line := m.params.args[i]
		focused := m.params.focus == paramFocusArgs && m.params.argsCursor == i
		switch {
		case focused && m.params.editKind == paramEditArgLine:
			detailRows = append(detailRows, m.params.editInput.View())
		default:
			prefix := "  "
			if focused {
				prefix = "› "
			}
			detailRows = append(detailRows, m.ui.styles.body.Render(prefix+truncateParamLine(line, maxSec)))
		}
	}
	rows = append(rows, secBox.Width(cw).Render(lipgloss.JoinVertical(lipgloss.Left, detailRows...)))

	var footerHelp string
	switch m.params.focus {
	case paramFocusProfiles:
		footerHelp = FooterParamFooterProfiles
	case paramFocusEnv:
		if m.paramEnvLen() == 0 {
			footerHelp = FooterParamFooterDetailEmpty
		} else {
			footerHelp = FooterParamFooterDetailRows
		}
	case paramFocusArgs:
		if m.paramArgsLen() == 0 {
			footerHelp = FooterParamFooterDetailEmpty
		} else {
			footerHelp = FooterParamFooterDetailRows
		}
	}
	if m.params.confirmDelete == paramConfirmNone {
		rows = append(rows, "", m.ui.styles.footer.Render(footerHelp))
	}
	block := lipgloss.JoinVertical(lipgloss.Left, rows...)
	if m.lastRunNote != "" {
		block = lipgloss.JoinVertical(lipgloss.Left, block, "", m.lastRunNoteView())
	}
	return m.ui.styles.portConfigBox.Render(block)
}
