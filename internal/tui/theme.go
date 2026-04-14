package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// EnvLLMLTheme is the environment variable that selects the color theme.
// Values: "dark", "light", or "auto" (default). When "auto" or unset, the
// terminal background is queried via lipgloss.HasDarkBackground().
const EnvLLMLTheme = "LLML_THEME"

// themePick selects which palette mode is active (including auto). The user
// cycles these with the t key: dark → light → auto → dark …
const (
	themePickDark = iota
	themePickLight
	themePickAuto
)

const themePickCount = 3

// themeToastText is the transient message after cycling themes (pick + resolved palette for auto).
func themeToastText(pick int, resolved Theme) string {
	switch pick {
	case themePickDark:
		return "Theme: dark"
	case themePickLight:
		return "Theme: light"
	default:
		if resolved == DarkTheme() {
			return "Theme: auto (dark)"
		}
		return "Theme: auto (light)"
	}
}

// Theme holds semantic colors for the TUI. All values are ANSI 256 or hex
// strings suitable for lipgloss.Color.
type Theme struct {
	Title        lipgloss.Color
	Subtitle     lipgloss.Color
	Body         lipgloss.Color
	Footer       lipgloss.Color
	Error        lipgloss.Color
	Border       lipgloss.Color
	RuntimePanel lipgloss.Color
	ModalTitle   lipgloss.Color
	// ParamSectionHeading labels nested blocks inside the parameters modal (env / argv).
	ParamSectionHeading lipgloss.Color
	ModalBody           lipgloss.Color
	TableHeader         lipgloss.Color
	TableCell           lipgloss.Color
	TableSelected       lipgloss.Color
	// ParamProfileName highlights parameter profile names in the params modal.
	ParamProfileName lipgloss.Color
}

// DarkTheme returns the default dark-terminal palette (original llml colors).
func DarkTheme() Theme {
	return Theme{
		Title:        lipgloss.Color("99"),
		Subtitle:     lipgloss.Color("241"),
		Body:         lipgloss.Color("252"),
		Footer:       lipgloss.Color("240"),
		Error:        lipgloss.Color("203"),
		Border:       lipgloss.Color("240"),
		RuntimePanel: lipgloss.Color("246"),
		// Brighter orchid than main Title (99); modal chrome reads as its own layer.
		ModalTitle: lipgloss.Color("183"),
		// Muted slate below ModalTitle (183); env/argv section captions.
		ParamSectionHeading: lipgloss.Color("109"),
		ModalBody:           lipgloss.Color("252"),
		TableHeader:         lipgloss.Color("252"),
		TableCell:           lipgloss.Color("252"),
		TableSelected:       lipgloss.Color("51"),
		// Distinct from ModalTitle / ParamSectionHeading; warm vs purple chrome.
		ParamProfileName: lipgloss.Color("178"),
	}
}

// LightTheme returns a palette tuned for light terminal backgrounds.
func LightTheme() Theme {
	return Theme{
		Title:        lipgloss.Color("55"),
		Subtitle:     lipgloss.Color("243"),
		Body:         lipgloss.Color("235"),
		Footer:       lipgloss.Color("249"),
		Error:        lipgloss.Color("160"),
		Border:       lipgloss.Color("249"),
		RuntimePanel: lipgloss.Color("238"),
		// Richer purple than main Title (55); dialogs stand out from the header.
		ModalTitle: lipgloss.Color("99"),
		// Steel blue; secondary to ModalTitle (99) for env/argv section captions.
		ParamSectionHeading: lipgloss.Color("61"),
		ModalBody:           lipgloss.Color("235"),
		TableHeader:         lipgloss.Color("235"),
		TableCell:           lipgloss.Color("235"),
		TableSelected:       lipgloss.Color("27"),
		// Distinct from ModalTitle / ParamSectionHeading; green accent on light bg.
		ParamProfileName: lipgloss.Color("30"),
	}
}

// initialThemePick maps LLML_THEME to the starting cycle index.
func initialThemePick() int {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(EnvLLMLTheme)))
	switch v {
	case "dark":
		return themePickDark
	case "light":
		return themePickLight
	default:
		return themePickAuto
	}
}

// themeFromPick returns the palette for a pick value. themePickAuto uses
// detectDark to choose dark vs light (same rules as LLML_THEME=auto).
func themeFromPick(pick int, detectDark func() bool) Theme {
	switch pick {
	case themePickDark:
		return DarkTheme()
	case themePickLight:
		return LightTheme()
	default:
		if detectDark() {
			return DarkTheme()
		}
		return LightTheme()
	}
}

// resolveTheme picks a theme from LLML_THEME and terminal background detection.
func resolveTheme() Theme {
	return resolveThemeWithDetector(lipgloss.HasDarkBackground)
}

// resolveThemeWithDetector is like resolveTheme but uses detectDark for the auto path
// (including unknown env values), so tests do not depend on the real terminal.
func resolveThemeWithDetector(detectDark func() bool) Theme {
	return themeFromPick(initialThemePick(), detectDark)
}
