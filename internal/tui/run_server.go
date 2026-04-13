package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/flyingnobita/llm-launch/internal/llamacpp"
)

// shellSingleQuoted returns s wrapped in single quotes for POSIX sh (safe for paths with spaces).
func shellSingleQuoted(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// formatLlamaServerInvocation is a copy-paste style one-liner (with a leading "+ ") printed before launch.
func formatLlamaServerInvocation(bin, modelPath string, port int) string {
	return fmt.Sprintf("+ %s -m %s --port %d", shellSingleQuoted(bin), shellSingleQuoted(modelPath), port)
}

// unixLlamaServerScript echoes the invocation, runs llama-server, then waits for Enter so logs stay readable before the TUI redraws.
func unixLlamaServerScript(bin, modelPath string, port int) string {
	inv := formatLlamaServerInvocation(bin, modelPath, port)
	return fmt.Sprintf(`printf '%%s\n' %s
%s -m %s --port %d
echo
echo 'Press Enter to return to llm-launch...'
read -r _
`, shellSingleQuoted(inv), shellSingleQuoted(bin), shellSingleQuoted(modelPath), port)
}

// runLlamaServerCmd runs llama-server for the selected GGUF in the foreground with the TUI suspended.
// Stdout and stderr go to the terminal (see tea.ExecProcess). Port matches LLAMA_SERVER_PORT / ListenPort.
//
// On Unix, the command is run under `sh -c` with a trailing `read` so the shell stays on the main screen
// until you press Enter after llama-server exits; then Bubble Tea restores the alternate screen.
// Windows runs llama-server directly (no pause); use scrollback or an external terminal if needed.
func runLlamaServerCmd(modelPath string, rt llamacpp.RuntimeInfo) tea.Cmd {
	bin := llamacpp.ResolveLlamaServerPath(rt)
	if bin == "" {
		return func() tea.Msg {
			return runServerErrMsg{
				err: fmt.Errorf("llama-server not found; set %s or install on PATH", llamacpp.EnvLlamaCppPath),
			}
		}
	}
	port := llamacpp.ListenPort()
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command(bin, "-m", modelPath, "--port", fmt.Sprintf("%d", port))
	} else {
		c = exec.Command("sh", "-c", unixLlamaServerScript(bin, modelPath, port))
	}
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return llamaServerExitedMsg{err: err}
	})
}
