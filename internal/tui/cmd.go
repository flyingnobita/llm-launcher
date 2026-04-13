package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/flyingnobita/llm-launch/internal/llamacpp"
)

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
