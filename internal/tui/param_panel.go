package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	paramFocusProfiles = iota
	paramFocusEnv
	paramFocusArgs
)

// paramConfirmDelete* values for Model.paramConfirmDelete (0 = none).
const (
	paramConfirmNone = iota
	paramConfirmProfile
	paramConfirmEnvRow
	paramConfirmArgRow
)

const (
	paramEditNone = iota
	paramEditEnvLine
	paramEditArgLine
	paramEditProfileName
)

func newParamLineTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 4096
	ti.SetWidth(64)
	ti.Blur()
	return ti
}

func parseEnvLine(s string) EnvVar {
	s = strings.TrimSpace(s)
	if s == "" {
		return EnvVar{}
	}
	i := strings.IndexByte(s, '=')
	if i < 0 {
		return EnvVar{Key: s, Value: ""}
	}
	return EnvVar{Key: strings.TrimSpace(s[:i]), Value: s[i+1:]}
}

func formatEnvVar(e EnvVar) string {
	if e.Key == "" {
		return ""
	}
	return e.Key + "=" + e.Value
}

func (m *Model) syncCurrentProfileOut() {
	if m.paramProfileIndex < 0 || m.paramProfileIndex >= len(m.paramProfiles) {
		return
	}
	m.paramProfiles[m.paramProfileIndex].Env = append([]EnvVar(nil), m.paramEnv...)
	m.paramProfiles[m.paramProfileIndex].Args = flattenArgLines(m.paramArgs)
}

func (m *Model) loadCurrentProfileIn() {
	if m.paramProfileIndex < 0 || m.paramProfileIndex >= len(m.paramProfiles) {
		return
	}
	p := m.paramProfiles[m.paramProfileIndex]
	m.paramEnv = append([]EnvVar(nil), p.Env...)
	m.paramArgs = collapseArgsForDisplay(p.Args)
	m.paramEnvCursor = 0
	m.paramArgsCursor = 0
}

func profileNameTaken(profiles []ParameterProfile, name string, skip int) bool {
	n := strings.TrimSpace(name)
	for i, p := range profiles {
		if i == skip {
			continue
		}
		if strings.TrimSpace(p.Name) == n {
			return true
		}
	}
	return false
}

func nextProfileName(profiles []ParameterProfile) string {
	for n := 1; n < 1000; n++ {
		cand := "Parameter profile"
		if n > 1 {
			cand = fmt.Sprintf("Parameter profile %d", n)
		}
		if !profileNameTaken(profiles, cand, -1) {
			return cand
		}
	}
	return "Parameter profile"
}

func copyProfiles(in []ParameterProfile) []ParameterProfile {
	out := make([]ParameterProfile, len(in))
	for i := range in {
		out[i].Name = in[i].Name
		out[i].Env = append([]EnvVar(nil), in[i].Env...)
		out[i].Args = append([]string(nil), in[i].Args...)
	}
	return out
}

func (m Model) openParamPanel() (Model, tea.Cmd) {
	p := m.SelectedPath()
	if p == "" {
		m.lastRunNote = "Select a model row first."
		return m, nil
	}
	m.paramPanelOpen = true
	m.paramConfirmDelete = paramConfirmNone
	m.paramModelPath = filepath.Clean(p)
	m.paramModelDisplayName = modelDisplayNameForPath(m)
	m.lastRunNote = ""
	m.paramEditKind = paramEditNone
	m.paramEditInput.Blur()
	m.paramEditInput.SetValue("")

	ent, err := loadModelEntry(m.paramModelPath)
	if err != nil {
		m.lastRunNote = err.Error()
		ent = modelEntry{
			Profiles:    []ParameterProfile{{Name: "default", Env: nil, Args: nil}},
			ActiveIndex: 0,
		}
	}
	m.paramProfiles = copyProfiles(ent.Profiles)
	m.paramProfileIndex = clampInt(ent.ActiveIndex, 0, max(0, len(m.paramProfiles)-1))
	m.paramFocus = paramFocusProfiles
	m.loadCurrentProfileIn()
	m.paramEditInput.SetWidth(m.paramEditInnerWidth())
	return m, nil
}

// paramEditInnerWidth is the textinput width for profile/env/argv line edits in the params modal.
func (m Model) paramEditInnerWidth() int {
	cw := m.paramPanelContentWidth()
	frame := m.styles.paramSectionBox.GetHorizontalFrameSize()
	w := cw - frame
	if w < 32 {
		w = 32
	}
	return w
}

func (m Model) closeParamPanel() Model {
	m.paramPanelOpen = false
	m.paramConfirmDelete = paramConfirmNone
	m.paramEditKind = paramEditNone
	m.paramEditInput.Blur()
	m.paramEditInput.SetValue("")
	m.paramEnv = nil
	m.paramArgs = nil
	m.paramProfiles = nil
	m.paramModelPath = ""
	m.paramModelDisplayName = ""
	return m
}

// modelDisplayNameForPath returns the table display name for the row whose path is selected, or a basename fallback.
func modelDisplayNameForPath(m Model) string {
	p := m.SelectedPath()
	if p == "" {
		return ""
	}
	p = filepath.Clean(p)
	for i := range m.files {
		if filepath.Clean(m.files[i].Path) == p {
			if n := strings.TrimSpace(m.files[i].Name); n != "" {
				return n
			}
			break
		}
	}
	return filepath.Base(p)
}

func (m Model) focusParamEdit() (Model, tea.Cmd) {
	return m, m.paramEditInput.Focus()
}

func (m Model) blurParamEdit() Model {
	m.paramEditInput.Blur()
	return m
}

func (m Model) paramEnvLen() int { return len(m.paramEnv) }
func (m Model) paramArgsLen() int {
	return len(m.paramArgs)
}

func (m Model) commitParamLineEdit() Model {
	line := m.paramEditInput.Value()
	kind := m.paramEditKind

	switch kind {
	case paramEditEnvLine:
		if strings.TrimSpace(line) == "" {
			m = m.cancelParamLineEdit()
			if m.paramEnvCursor >= 0 && m.paramEnvCursor < m.paramEnvLen() {
				e := m.paramEnv[m.paramEnvCursor]
				if strings.TrimSpace(e.Key) == "" && strings.TrimSpace(e.Value) == "" {
					m = m.deleteParamRow()
				}
			}
			return m
		}
	case paramEditArgLine:
		if strings.TrimSpace(line) == "" {
			m = m.cancelParamLineEdit()
			if m.paramArgsCursor >= 0 && m.paramArgsCursor < m.paramArgsLen() &&
				strings.TrimSpace(m.paramArgs[m.paramArgsCursor]) == "" {
				m = m.deleteParamRow()
			}
			return m
		}
	}

	m.paramEditKind = paramEditNone
	m = m.blurParamEdit()
	switch kind {
	case paramEditProfileName:
		if m.paramProfileIndex >= 0 && m.paramProfileIndex < len(m.paramProfiles) {
			name := strings.TrimSpace(line)
			if name == "" {
				name = fmt.Sprintf("parameter profile %d", m.paramProfileIndex+1)
			}
			if profileNameTaken(m.paramProfiles, name, m.paramProfileIndex) {
				name = nextProfileName(m.paramProfiles)
			}
			m.paramProfiles[m.paramProfileIndex].Name = name
		}
	case paramEditEnvLine:
		if m.paramEnvCursor >= 0 && m.paramEnvCursor < m.paramEnvLen() {
			m.paramEnv[m.paramEnvCursor] = parseEnvLine(line)
		}
	case paramEditArgLine:
		if m.paramArgsCursor >= 0 && m.paramArgsCursor < m.paramArgsLen() {
			m.paramArgs[m.paramArgsCursor] = line
		}
	}
	m.paramEditInput.SetValue("")
	return m
}

func (m Model) cancelParamLineEdit() Model {
	m.paramEditKind = paramEditNone
	m = m.blurParamEdit()
	m.paramEditInput.SetValue("")
	return m
}

func (m Model) startParamLineEdit() (Model, tea.Cmd) {
	switch m.paramFocus {
	case paramFocusEnv:
		if m.paramEnvLen() == 0 {
			return m, nil
		}
		m.paramEditKind = paramEditEnvLine
		m.paramEditInput.SetValue(formatEnvVar(m.paramEnv[m.paramEnvCursor]))
	case paramFocusArgs:
		if m.paramArgsLen() == 0 {
			return m, nil
		}
		m.paramEditKind = paramEditArgLine
		m.paramEditInput.SetValue(m.paramArgs[m.paramArgsCursor])
	default:
		return m, nil
	}
	return m.focusParamEdit()
}

func (m Model) startProfileNameEdit() (Model, tea.Cmd) {
	if m.paramProfileIndex < 0 || m.paramProfileIndex >= len(m.paramProfiles) {
		return m, nil
	}
	m.paramEditKind = paramEditProfileName
	m.paramEditInput.SetValue(m.paramProfiles[m.paramProfileIndex].Name)
	return m.focusParamEdit()
}

func (m Model) addParamRow() (Model, tea.Cmd) {
	(&m).syncCurrentProfileOut()
	switch m.paramFocus {
	case paramFocusEnv:
		m.paramEnv = append(m.paramEnv, EnvVar{})
		m.paramEnvCursor = m.paramEnvLen() - 1
		m.paramEditKind = paramEditEnvLine
		m.paramEditInput.SetValue("")
	case paramFocusArgs:
		m.paramArgs = append(m.paramArgs, "")
		m.paramArgsCursor = m.paramArgsLen() - 1
		m.paramEditKind = paramEditArgLine
		m.paramEditInput.SetValue("")
	default:
		return m, nil
	}
	return m.focusParamEdit()
}

func (m Model) deleteParamRow() Model {
	(&m).syncCurrentProfileOut()
	switch m.paramFocus {
	case paramFocusEnv:
		if m.paramEnvLen() == 0 || m.paramEnvCursor < 0 || m.paramEnvCursor >= m.paramEnvLen() {
			return m
		}
		m.paramEnv = append(m.paramEnv[:m.paramEnvCursor], m.paramEnv[m.paramEnvCursor+1:]...)
		if m.paramEnvCursor >= m.paramEnvLen() && m.paramEnvLen() > 0 {
			m.paramEnvCursor = m.paramEnvLen() - 1
		}
	case paramFocusArgs:
		if m.paramArgsLen() == 0 || m.paramArgsCursor < 0 || m.paramArgsCursor >= m.paramArgsLen() {
			return m
		}
		m.paramArgs = append(m.paramArgs[:m.paramArgsCursor], m.paramArgs[m.paramArgsCursor+1:]...)
		if m.paramArgsCursor >= m.paramArgsLen() && m.paramArgsLen() > 0 {
			m.paramArgsCursor = m.paramArgsLen() - 1
		}
	default:
		return m
	}
	return m
}

func (m Model) addProfile() Model {
	(&m).syncCurrentProfileOut()
	nm := nextProfileName(m.paramProfiles)
	m.paramProfiles = append(m.paramProfiles, ParameterProfile{Name: nm, Env: nil, Args: nil})
	m.paramProfileIndex = len(m.paramProfiles) - 1
	m.loadCurrentProfileIn()
	m.paramEnvCursor = 0
	m.paramArgsCursor = 0
	return m
}

func (m Model) deleteProfile() Model {
	if len(m.paramProfiles) <= 1 {
		return m
	}
	(&m).syncCurrentProfileOut()
	m.paramProfiles = append(m.paramProfiles[:m.paramProfileIndex], m.paramProfiles[m.paramProfileIndex+1:]...)
	if m.paramProfileIndex >= len(m.paramProfiles) {
		m.paramProfileIndex = len(m.paramProfiles) - 1
	}
	m.loadCurrentProfileIn()
	m.paramEnvCursor = 0
	m.paramArgsCursor = 0
	return m
}

func (m Model) cycleParamFocus(delta int) Model {
	(&m).syncCurrentProfileOut()
	m.paramFocus = (m.paramFocus + delta + 3) % 3
	return m
}

func (m Model) moveProfile(delta int) Model {
	(&m).syncCurrentProfileOut()
	n := len(m.paramProfiles)
	if n == 0 {
		return m
	}
	next := m.paramProfileIndex + delta
	if next < 0 || next >= n {
		return m
	}
	m.paramProfileIndex = next
	m.loadCurrentProfileIn()
	m.paramEnvCursor = 0
	m.paramArgsCursor = 0
	return m
}

// persistParamPanel writes the current parameter profiles to disk without closing the panel.
func (m Model) persistParamPanel() (Model, tea.Cmd) {
	(&m).syncCurrentProfileOut()
	ent := modelEntry{
		Profiles:    copyProfiles(m.paramProfiles),
		ActiveIndex: m.paramProfileIndex,
	}
	if err := saveModelEntry(m.paramModelPath, ent); err != nil {
		m.lastRunNote = err.Error()
		return m, nil
	}
	m.lastRunNote = ""
	return m, nil
}

// closeParamPanelWithPersist saves first; on error the panel stays open and lastRunNote is set.
func (m Model) closeParamPanelWithPersist() (Model, tea.Cmd) {
	m, cmd := m.persistParamPanel()
	if m.lastRunNote != "" {
		return m, cmd
	}
	m = m.closeParamPanel()
	return m, cmd
}

// updateParamPanelKey handles keys while the parameters panel is open.
func (m Model) updateParamPanelKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if m.paramConfirmDelete != paramConfirmNone {
		switch msg.String() {
		case "y", "Y":
			k := m.paramConfirmDelete
			m.paramConfirmDelete = paramConfirmNone
			switch k {
			case paramConfirmProfile:
				m = m.deleteProfile()
			case paramConfirmEnvRow, paramConfirmArgRow:
				m = m.deleteParamRow()
			}
			return m.persistParamPanel()
		case "n", "N":
			m.paramConfirmDelete = paramConfirmNone
			return m, nil
		default:
			return m, nil
		}
	}

	if m.paramEditKind != paramEditNone {
		switch msg.String() {
		case "esc":
			m = m.cancelParamLineEdit()
			return m, nil
		case "enter":
			m = m.commitParamLineEdit()
			return m.persistParamPanel()
		case "tab":
			m = m.commitParamLineEdit()
			m = m.cycleParamFocus(1)
			return m.persistParamPanel()
		case "shift+tab":
			m = m.commitParamLineEdit()
			m = m.cycleParamFocus(-1)
			return m.persistParamPanel()
		default:
			var cmd tea.Cmd
			m.paramEditInput, cmd = m.paramEditInput.Update(msg)
			return m, cmd
		}
	}

	switch msg.String() {
	case "esc", "q":
		return m.closeParamPanelWithPersist()
	case "t":
		var cmd tea.Cmd
		m, cmd = m.cycleTheme()
		return m, cmd
	case "tab":
		m = m.cycleParamFocus(1)
		return m, nil
	case "shift+tab":
		m = m.cycleParamFocus(-1)
		return m, nil
	case "up", "k":
		switch m.paramFocus {
		case paramFocusProfiles:
			m = m.moveProfile(-1)
			return m.persistParamPanel()
		case paramFocusEnv:
			if m.paramEnvCursor > 0 {
				m.paramEnvCursor--
			}
		case paramFocusArgs:
			if m.paramArgsCursor > 0 {
				m.paramArgsCursor--
			}
		}
		return m, nil
	case "down", "j":
		switch m.paramFocus {
		case paramFocusProfiles:
			m = m.moveProfile(1)
			return m.persistParamPanel()
		case paramFocusEnv:
			if m.paramEnvCursor < m.paramEnvLen()-1 {
				m.paramEnvCursor++
			}
		case paramFocusArgs:
			if m.paramArgsCursor < m.paramArgsLen()-1 {
				m.paramArgsCursor++
			}
		}
		return m, nil
	case "n":
		if m.paramFocus == paramFocusProfiles {
			m = m.addProfile()
			return m.persistParamPanel()
		}
		return m, nil
	case "a":
		if m.paramFocus == paramFocusEnv || m.paramFocus == paramFocusArgs {
			var cmd tea.Cmd
			m, cmd = m.addParamRow()
			m, pcmd := m.persistParamPanel()
			return m, tea.Batch(cmd, pcmd)
		}
		return m, nil
	case "d":
		switch m.paramFocus {
		case paramFocusProfiles:
			if len(m.paramProfiles) <= 1 {
				return m, nil
			}
			m.paramConfirmDelete = paramConfirmProfile
			return m, nil
		case paramFocusEnv:
			if m.paramEnvLen() == 0 {
				return m, nil
			}
			m.paramConfirmDelete = paramConfirmEnvRow
			return m, nil
		case paramFocusArgs:
			if m.paramArgsLen() == 0 {
				return m, nil
			}
			m.paramConfirmDelete = paramConfirmArgRow
			return m, nil
		}
		return m, nil
	case "r", "R":
		if m.paramFocus == paramFocusProfiles {
			return m.startProfileNameEdit()
		}
		return m, nil
	case "enter":
		if m.paramFocus == paramFocusEnv || m.paramFocus == paramFocusArgs {
			return m.startParamLineEdit()
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) paramPanelView() string {
	cw := m.paramPanelContentWidth()
	maxLine := cw
	if maxLine < 24 {
		maxLine = 24
	}
	secBox := m.styles.paramSectionBox
	maxSec := cw - secBox.GetHorizontalFrameSize()
	if maxSec < 24 {
		maxSec = 24
	}

	title := m.modalTitleRow(cw, m.styles.portConfigTitle, "Parameters — "+m.paramModelDisplayName)
	rows := []string{title, ""}

	if k := m.paramConfirmDelete; k != paramConfirmNone {
		confirmBox := m.styles.paramConfirmDialog
		confirmInner := cw - confirmBox.GetHorizontalFrameSize()
		if confirmInner < 24 {
			confirmInner = 24
		}
		var confirmRows []string
		switch k {
		case paramConfirmProfile:
			pName := ""
			if m.paramProfileIndex >= 0 && m.paramProfileIndex < len(m.paramProfiles) {
				pName = m.paramProfiles[m.paramProfileIndex].Name
			}
			if pName == "" {
				pName = "(unnamed)"
			}
			nameLine := lipgloss.JoinHorizontal(lipgloss.Top,
				m.styles.body.Render("  "),
				m.styles.paramProfileName.Render(truncateParamLine(pName, confirmInner-2)),
			)
			confirmRows = []string{
				m.styles.body.Render("Delete this parameter profile?"),
				nameLine,
			}
		case paramConfirmEnvRow:
			line := ""
			if m.paramEnvCursor >= 0 && m.paramEnvCursor < m.paramEnvLen() {
				line = formatEnvVar(m.paramEnv[m.paramEnvCursor])
			}
			confirmRows = []string{
				m.styles.body.Render("Delete this environment variable line?"),
				m.styles.body.Render("  " + truncateParamLine(line, max(confirmInner-2, 8))),
			}
		case paramConfirmArgRow:
			line := ""
			if m.paramArgsCursor >= 0 && m.paramArgsCursor < m.paramArgsLen() {
				line = m.paramArgs[m.paramArgsCursor]
			}
			confirmRows = []string{
				m.styles.body.Render("Delete this extra argument line?"),
				m.styles.body.Render("  " + truncateParamLine(line, max(confirmInner-2, 8))),
			}
		}
		if len(confirmRows) > 0 {
			confirmRows = append(confirmRows, "",
				m.styles.footer.Render("y: yes · n: no"),
			)
			rows = append(rows, confirmBox.Width(cw).Render(lipgloss.JoinVertical(lipgloss.Left, confirmRows...)))
			rows = append(rows, "")
		}
	}

	rows = append(rows, m.styles.body.Render("  Profiles"))
	rows = append(rows, "")
	for i := range m.paramProfiles {
		name := m.paramProfiles[i].Name
		if name == "" {
			name = "(unnamed)"
		}
		focused := m.paramFocus == paramFocusProfiles && i == m.paramProfileIndex
		switch {
		case focused && m.paramEditKind == paramEditProfileName:
			rows = append(rows, m.paramEditInput.View())
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
				m.styles.body.Render(prefix),
				m.styles.paramProfileName.Render(truncateParamLine(name, nameW)),
			)
			rows = append(rows, row)
		}
	}
	if len(m.paramProfiles) == 0 {
		rows = append(rows, m.styles.body.Render("  (none)"))
	}

	rows = append(rows, "")
	var detailRows []string
	const sectionHeadingIndent = "  "
	envHeading := "Environment Variables (e.g. PYTORCH_CUDA_ALLOC_CONF=expandable_segments:True)"
	detailRows = append(detailRows, lipgloss.JoinHorizontal(lipgloss.Top,
		m.styles.body.Render(sectionHeadingIndent),
		m.styles.paramSectionHeading.Render(truncateParamLine(envHeading, maxSec-lipgloss.Width(sectionHeadingIndent))),
	))
	detailRows = append(detailRows, "")
	if m.paramEnvLen() == 0 && !(m.paramFocus == paramFocusEnv && m.paramEditKind == paramEditEnvLine) {
		prefix := "  "
		if m.paramFocus == paramFocusEnv {
			prefix = "› "
		}
		detailRows = append(detailRows, m.styles.body.Render(prefix+"(none)"))
	}
	for i := range m.paramEnv {
		line := formatEnvVar(m.paramEnv[i])
		focused := m.paramFocus == paramFocusEnv && m.paramEnvCursor == i
		switch {
		case focused && m.paramEditKind == paramEditEnvLine:
			detailRows = append(detailRows, m.paramEditInput.View())
		default:
			prefix := "  "
			if focused {
				prefix = "› "
			}
			detailRows = append(detailRows, m.styles.body.Render(prefix+truncateParamLine(line, maxSec)))
		}
	}

	detailRows = append(detailRows, "")

	argHeading := "Extra arguments (e.g. --max-model-len 131072)"
	detailRows = append(detailRows, lipgloss.JoinHorizontal(lipgloss.Top,
		m.styles.body.Render(sectionHeadingIndent),
		m.styles.paramSectionHeading.Render(truncateParamLine(argHeading, maxSec-lipgloss.Width(sectionHeadingIndent))),
	))
	detailRows = append(detailRows, "")
	if m.paramArgsLen() == 0 && !(m.paramFocus == paramFocusArgs && m.paramEditKind == paramEditArgLine) {
		prefix := "  "
		if m.paramFocus == paramFocusArgs {
			prefix = "› "
		}
		detailRows = append(detailRows, m.styles.body.Render(prefix+"(none)"))
	}
	for i := range m.paramArgs {
		line := m.paramArgs[i]
		focused := m.paramFocus == paramFocusArgs && m.paramArgsCursor == i
		switch {
		case focused && m.paramEditKind == paramEditArgLine:
			detailRows = append(detailRows, m.paramEditInput.View())
		default:
			prefix := "  "
			if focused {
				prefix = "› "
			}
			detailRows = append(detailRows, m.styles.body.Render(prefix+truncateParamLine(line, maxSec)))
		}
	}
	rows = append(rows, secBox.Width(cw).Render(lipgloss.JoinVertical(lipgloss.Left, detailRows...)))

	var footerHelp string
	switch m.paramFocus {
	case paramFocusProfiles:
		footerHelp = "tab: sections · hjkl: nav · n: new · d: delete · r: rename · esc/q: back"
	case paramFocusEnv:
		if m.paramEnvLen() == 0 {
			footerHelp = "tab: sections · hjkl: nav · a: add row · d: delete · esc/q: back"
		} else {
			footerHelp = "tab: sections · hjkl: nav · enter: edit · a: add row · d: delete · esc/q: back"
		}
	case paramFocusArgs:
		if m.paramArgsLen() == 0 {
			footerHelp = "tab: sections · hjkl: nav · a: add row · d: delete · esc/q: back"
		} else {
			footerHelp = "tab: sections · hjkl: nav · enter: edit · a: add row · d: delete · esc/q: back"
		}
	}
	if m.paramConfirmDelete == paramConfirmNone {
		rows = append(rows, "", m.styles.footer.Render(footerHelp))
	}
	block := lipgloss.JoinVertical(lipgloss.Left, rows...)
	if m.lastRunNote != "" {
		block = lipgloss.JoinVertical(lipgloss.Left, block, "", m.styles.errLine.Render(m.lastRunNote))
	}
	framed := m.styles.portConfigBox.Render(block)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, framed)
}

func truncateParamLine(s string, maxW int) string {
	if maxW < 8 {
		return s
	}
	if lipgloss.Width(s) <= maxW {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r)) > maxW {
		r = r[:len(r)-1]
	}
	return string(r)
}
