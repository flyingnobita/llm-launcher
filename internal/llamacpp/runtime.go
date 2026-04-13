package llamacpp

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Environment variables for locating llama.cpp binaries and probing a running server.
const (
	EnvLlamaCppPath    = "LLAMA_CPP_PATH"
	EnvLlamaServerPort = "LLAMA_SERVER_PORT"
)

const defaultLlamaServerPort = 8080

// RuntimeInfo describes detected llama-cli / llama-server binaries and optional running server.
type RuntimeInfo struct {
	LlamaCLIPath    string
	LlamaServerPath string
	ServerRunning   bool
	ProbePort       int // port used when ServerRunning is true (0 if not probed)
}

// Available is true if either binary was found or a llama-server responded on the health probe.
func (r RuntimeInfo) Available() bool {
	return r.LlamaCLIPath != "" || r.LlamaServerPath != "" || r.ServerRunning
}

// Summary is a single-line status for the TUI (no trailing newline).
func (r RuntimeInfo) Summary() string {
	switch {
	case r.LlamaCLIPath != "" && r.LlamaServerPath != "":
		return fmt.Sprintf("llama.cpp: cli %s · server %s", formatBinLabel(r.LlamaCLIPath), formatBinLabel(r.LlamaServerPath))
	case r.LlamaCLIPath != "":
		return fmt.Sprintf("llama.cpp: cli %s · server —", formatBinLabel(r.LlamaCLIPath))
	case r.LlamaServerPath != "":
		return fmt.Sprintf("llama.cpp: cli — · server %s", formatBinLabel(r.LlamaServerPath))
	case r.ServerRunning:
		return fmt.Sprintf("llama.cpp: binaries not on PATH — server running :%d", r.ProbePort)
	default:
		return "llama.cpp: not found — set " + EnvLlamaCppPath + " or install to PATH (Homebrew: ensure /opt/homebrew/bin is on PATH)"
	}
}

func formatBinLabel(abs string) string {
	if abs == "" {
		return "—"
	}
	return "✓"
}

const binaryPathLabelWidth = 12

// BinaryPathLines returns lines for the TUI footer: llama-cli / llama-server paths, listen port for R / health probe.
// Lines are truncated to maxWidth display width.
func (r RuntimeInfo) BinaryPathLines(maxWidth int) []string {
	if maxWidth < 24 {
		maxWidth = 24
	}
	pathW := maxWidth - binaryPathLabelWidth - 1
	if pathW < 8 {
		pathW = 8
	}
	cli := "—"
	if r.LlamaCLIPath != "" {
		cli = TruncateRunes(FormatPathDisplay(r.LlamaCLIPath), pathW)
	}
	srv := "—"
	if r.LlamaServerPath != "" {
		srv = TruncateRunes(FormatPathDisplay(r.LlamaServerPath), pathW)
	} else if r.ServerRunning {
		srv = TruncateRunes(fmt.Sprintf("running · :%d (health OK)", r.ProbePort), pathW)
	}
	port := r.ProbePort
	if port <= 0 {
		port = ListenPort()
	}
	portVal := fmt.Sprintf(":%d (%s)", port, EnvLlamaServerPort)
	line := func(label, value string) string {
		s := fmt.Sprintf("%-*s %s", binaryPathLabelWidth, label, value)
		return TruncateRunes(s, maxWidth)
	}
	return []string{
		line("llama-cli", cli),
		line("llama-server", srv),
		line("listen (R)", TruncateRunes(portVal, pathW)),
	}
}

// DiscoverRuntime locates llama-cli and llama-server using LLAMA_CPP_PATH, common install
// directories (including Homebrew on Apple Silicon), then PATH. If neither binary exists,
// it probes http://127.0.0.1:{LLAMA_SERVER_PORT}/health (default port 8080) with a short timeout.
func DiscoverRuntime() RuntimeInfo {
	cli := findLlamaBinary("llama-cli")
	srv := findLlamaBinary("llama-server")
	port := ListenPort()
	info := RuntimeInfo{
		LlamaCLIPath:    cli,
		LlamaServerPath: srv,
		ProbePort:       port,
	}
	if cli == "" && srv == "" {
		if probeLlamaServerHealth(port) {
			info.ServerRunning = true
		}
	}
	return info
}

// ListenPort returns the TCP port from LLAMA_SERVER_PORT, or 8080 if unset or invalid.
func ListenPort() int {
	if v := os.Getenv(EnvLlamaServerPort); v != "" {
		if p, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && p > 0 && p <= 65535 {
			return p
		}
	}
	return defaultLlamaServerPort
}

// ResolveLlamaServerPath returns the detected llama-server binary path, or the first match on PATH.
func ResolveLlamaServerPath(r RuntimeInfo) string {
	if r.LlamaServerPath != "" {
		return r.LlamaServerPath
	}
	if p, err := exec.LookPath("llama-server"); err == nil {
		return p
	}
	return ""
}

func findLlamaBinary(name string) string {
	// 1. LLAMA_CPP_PATH/<name>
	if dir := os.Getenv(EnvLlamaCppPath); dir != "" {
		candidate := filepath.Join(filepath.Clean(dir), name)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() && st.Mode().IsRegular() {
			return candidate
		}
	}

	// 2. Common install locations (Homebrew Apple Silicon, Linux user local, source build)
	var common []string
	common = append(common,
		"/usr/local/bin",
		"/opt/homebrew/bin",
		"/opt/llama.cpp/build/bin",
	)
	if home, err := os.UserHomeDir(); err == nil {
		common = append(common, filepath.Join(home, ".local", "bin"))
	}
	for _, dir := range common {
		candidate := filepath.Join(dir, name)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() && st.Mode().IsRegular() {
			return candidate
		}
	}

	// 3. PATH
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	return ""
}

// probeLlamaServerHealth GETs /health on 127.0.0.1 (avoids localhost IPv6/IPv4 ambiguity).
func probeLlamaServerHealth(port int) bool {
	url := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
