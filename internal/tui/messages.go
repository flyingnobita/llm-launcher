package tui

import "github.com/flyingnobita/llm-launch/internal/llamacpp"

type runtimeReadyMsg struct {
	runtime llamacpp.RuntimeInfo
}

type modelsLoadedMsg struct {
	files []llamacpp.ModelFile
}

type modelsErrMsg struct {
	err error
}

type runServerErrMsg struct {
	err error
}

type llamaServerExitedMsg struct {
	err error
}
