package tui

import (
	"github.com/charmbracelet/lipgloss"

	btable "github.com/flyingnobita/llml/internal/tui/btable"
)

// styles holds all lipgloss styles for one resolved theme.
type styles struct {
	app                 lipgloss.Style
	title               lipgloss.Style
	subtitle            lipgloss.Style
	body                lipgloss.Style
	footer              lipgloss.Style
	errLine             lipgloss.Style
	runtimePanel        lipgloss.Style
	portConfigTitle     lipgloss.Style
	portConfigBox       lipgloss.Style
	paramSectionBox     lipgloss.Style
	paramConfirmDialog  lipgloss.Style
	paramSectionHeading lipgloss.Style
	themeToastInline    lipgloss.Style
	paramProfileName    lipgloss.Style
	table               btable.Styles
}

// newStyles builds lipgloss styles from a Theme. Table Header, Cell, and
// Selected use PaddingRight(1) so columns align (Selected must match Cell).
func newStyles(theme Theme) styles {
	return styles{
		app: lipgloss.NewStyle().Padding(1, appPaddingH),
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Title).
			MarginBottom(1),
		subtitle: lipgloss.NewStyle().
			Foreground(theme.Subtitle).
			MarginBottom(1),
		body: lipgloss.NewStyle().
			Foreground(theme.Body),
		footer: lipgloss.NewStyle().
			Foreground(theme.Footer).
			MarginTop(1),
		errLine: lipgloss.NewStyle().Foreground(theme.Error),
		runtimePanel: lipgloss.NewStyle().
			BorderTop(true).
			BorderForeground(theme.Border).
			Foreground(theme.RuntimePanel).
			Padding(1, 0).
			MarginTop(1),
		portConfigTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.ModalTitle),
		portConfigBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Padding(1, 2).
			Foreground(theme.ModalBody),
		// Nested sections inside the parameters modal (env + argv).
		paramSectionBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Padding(0, 1),
		paramConfirmDialog: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Error).
			Padding(0, 1),
		paramSectionHeading: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.ParamSectionHeading),
		// Compact reversed chip on the title row (no extra viewport row).
		themeToastInline: lipgloss.NewStyle().
			Bold(true).
			Reverse(true).
			Padding(0, 1),
		paramProfileName: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.ParamProfileName),
		table: btable.Styles{
			Header: lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.TableHeader).
				PaddingRight(1),
			Cell: lipgloss.NewStyle().
				Foreground(theme.TableCell).
				PaddingRight(1),
			Selected: lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.TableSelected).
				PaddingRight(1),
		},
	}
}
