package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/flyingnobita/llml/internal/llamacpp"
	btable "github.com/flyingnobita/llml/internal/tui/btable"
)

// Model is the root Bubble Tea model.
type Model struct {
	width             int
	height            int
	bodyInnerW        int
	tableBodyH        int
	tableLineWidth    int
	theme             Theme
	themePick         int
	themeToast        string
	styles            styles
	keys              KeyMap
	tbl               btable.Model
	hscroll           viewport.Model
	files             []llamacpp.ModelFile
	runtime           llamacpp.RuntimeInfo
	runtimeScanned    bool
	lastRunNote       string
	loading           bool
	loadErr           error
	runtimeConfigOpen bool
	runtimeFocus      int
	runtimeInputs     [runtimeFieldCount]textinput.Model

	paramPanelOpen        bool
	paramConfirmDelete    int // paramConfirm* (see param_panel.go); 0 = none
	paramModelPath        string
	paramModelDisplayName string
	paramFocus            int
	paramProfileIndex     int
	paramProfiles         []ParameterProfile
	paramEnvCursor        int
	paramArgsCursor       int
	paramEnv              []EnvVar
	paramArgs             []string
	paramEditKind         int
	paramEditInput        textinput.Model
}

// New returns a model with default key bindings and an empty table; Init triggers discovery.
func New() Model {
	pick := initialThemePick()
	th := themeFromPick(pick, lipgloss.HasDarkBackground)
	st := newStyles(th)
	t := btable.New(
		btable.WithColumns(tableColumns(100, nil)),
		btable.WithRows(nil),
		btable.WithFocused(true),
		btable.WithStyles(st.table),
		btable.WithWidth(96),
		btable.WithHeight(defaultTableHeight),
	)
	hv := viewport.New(96, defaultTableHeight)
	hv.SetHorizontalStep(hScrollStep)
	return Model{
		theme:     th,
		themePick: pick,
		styles:    st,
		keys:      DefaultKeyMap(),
		tbl:       t,
		hscroll:   hv,
		runtimeInputs: [runtimeFieldCount]textinput.Model{
			newPathTextInput(),
			newPathTextInput(),
			newPathTextInput(),
			newPortTextInput(),
			newPortTextInput(),
		},
		paramEditInput: newParamLineTextInput(),
		loading:        true,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return startupCmd()
}

// innerWidth returns the usable inner body width for rendering. It falls back
// to a computed value when bodyInnerW has not yet been set by layoutTable.
func (m Model) innerWidth() int {
	if m.bodyInnerW >= 1 {
		return m.bodyInnerW
	}
	if m.width > 0 {
		return max(m.width-appPaddingH*2, minInnerWidth)
	}
	return minInnerWidth
}

func (m Model) layoutTable() Model {
	w := m.width
	if w < minTerminalWidth {
		w = minTerminalWidth
	}
	innerW := m.width - appPaddingH*2
	if innerW < minInnerWidth {
		innerW = w - appPaddingH*2
	}
	m.bodyInnerW = innerW
	// Column widths must use the same budget as the table viewport (inner body
	// width). Using full terminal width here made rows ~4 cells wider than
	// innerW and triggered empty horizontal scrolling.
	cols := tableColumns(innerW, m.files)
	m.tbl.SetColumns(cols)
	m.tbl.SetStyles(m.styles.table)
	minW := tableContentMinWidth(cols)
	m.tbl.SetWidth(max(minW, innerW))

	var h int
	if m.height <= 0 {
		h = defaultTableHeight
	} else {
		// Bubble Tea keeps only the bottom m.height lines if the view is taller;
		// size the table so framed (padding + chrome + body) fits.
		appPad := m.styles.app.GetVerticalFrameSize()
		innerMax := m.height - appPad
		if innerMax < 1 {
			innerMax = 1
		}
		needsHBar := len(m.files) > 0 && minW > innerW
		static := mainChromeLines(m, needsHBar)
		h = innerMax - static
		if h < 1 {
			h = 1
		}
	}

	m.tbl.SetHeight(h)
	m.tbl.SetRows(buildTableRows(m.files, cols))
	tview := m.tbl.View()
	m.tableBodyH = max(1, strings.Count(tview, "\n")+1)
	lines := strings.Split(tview, "\n")
	if len(lines) > 0 {
		m.tableLineWidth = lipgloss.Width(lines[0])
	} else {
		m.tableLineWidth = 0
	}

	// Second pass if scroll bar visibility differs from min-width estimate.
	if m.height > 0 {
		needsHBar := len(m.files) > 0 && m.tableLineWidth > 0 && m.tableLineWidth > innerW
		needsHBarGuess := len(m.files) > 0 && minW > innerW
		if needsHBar != needsHBarGuess {
			appPad := m.styles.app.GetVerticalFrameSize()
			innerMax := m.height - appPad
			if innerMax < 1 {
				innerMax = 1
			}
			static := mainChromeLines(m, needsHBar)
			h2 := innerMax - static
			if h2 < 1 {
				h2 = 1
			}
			if h2 != h {
				h = h2
				m.tbl.SetHeight(h)
				m.tbl.SetRows(buildTableRows(m.files, cols))
				tview = m.tbl.View()
				m.tableBodyH = max(1, strings.Count(tview, "\n")+1)
				lines = strings.Split(tview, "\n")
				if len(lines) > 0 {
					m.tableLineWidth = lipgloss.Width(lines[0])
				}
			}
		}
	}

	m.hscroll.SetContent(tview)
	m.hscroll.Width = innerW
	m.hscroll.Height = m.tableBodyH
	return m
}

// cycleTheme advances dark → light → auto → dark, rebuilds lipgloss styles, and
// shows a short toast on the title row naming the active mode.
func (m Model) cycleTheme() (Model, tea.Cmd) {
	m.themePick = (m.themePick + 1) % themePickCount
	m.theme = themeFromPick(m.themePick, lipgloss.HasDarkBackground)
	m.styles = newStyles(m.theme)
	m.themeToast = themeToastText(m.themePick, m.theme)
	m = m.layoutTable()
	return m, clearThemeToastAfterCmd()
}
