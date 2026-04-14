package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/flyingnobita/llml/internal/llamacpp"
)

const themeToastVisibleDuration = 2 * time.Second

func discoverRuntimeCmd() tea.Cmd {
	return func() tea.Msg {
		return runtimeReadyMsg{runtime: llamacpp.DiscoverRuntime()}
	}
}

func loadModelsCmd() tea.Cmd {
	return func() tea.Msg {
		files, err := llamacpp.Discover(llamacpp.Options{})
		if err != nil {
			return modelsErrMsg{err: err}
		}
		return modelsLoadedMsg{files: files}
	}
}

// startupCmd runs llama.cpp binary detection first, then GGUF discovery.
func startupCmd() tea.Cmd {
	return tea.Sequence(discoverRuntimeCmd(), loadModelsCmd())
}

// clearThemeToastAfterCmd schedules removal of the theme banner.
func clearThemeToastAfterCmd() tea.Cmd {
	return tea.Tick(themeToastVisibleDuration, func(time.Time) tea.Msg {
		return themeToastClearMsg{}
	})
}
