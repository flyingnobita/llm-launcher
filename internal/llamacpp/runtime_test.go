package llamacpp

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestFindLlamaBinary_LLamaCppPathWins(t *testing.T) {
	dir := t.TempDir()
	name := "llama-cli"
	bin := filepath.Join(dir, name)
	if err := os.WriteFile(bin, []byte{}, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvLlamaCppPath, dir)
	t.Setenv("PATH", "/nonexistent")

	got := findLlamaBinary(name)
	if got != bin {
		t.Fatalf("got %q want %q", got, bin)
	}
}

func TestProbeLlamaServerHealth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}
	if !probeLlamaServerHealth(port) {
		t.Fatal("expected health probe success")
	}
}

func TestRuntimeInfo_BinaryPathLines(t *testing.T) {
	r := RuntimeInfo{
		LlamaCLIPath:    "/home/u/.local/bin/llama-cli",
		LlamaServerPath: "/opt/homebrew/bin/llama-server",
		ProbePort:       8080,
	}
	lines := r.BinaryPathLines(80)
	if len(lines) != 3 {
		t.Fatalf("got %d lines", len(lines))
	}
	if !strings.Contains(lines[0], "llama-cli") || !strings.Contains(lines[0], ".local") {
		t.Errorf("cli line: %q", lines[0])
	}
	if !strings.Contains(lines[1], "llama-server") {
		t.Errorf("server line: %q", lines[1])
	}
	if !strings.Contains(lines[2], "8080") || !strings.Contains(lines[2], "listen") {
		t.Errorf("port line: %q", lines[2])
	}
}

func TestRuntimeInfo_Summary(t *testing.T) {
	cases := []struct {
		r    RuntimeInfo
		want string
	}{
		{
			r:    RuntimeInfo{LlamaCLIPath: "/a/llama-cli", LlamaServerPath: "/b/llama-server"},
			want: "llama.cpp: cli ✓ · server ✓",
		},
		{
			r:    RuntimeInfo{ServerRunning: true, ProbePort: 8000},
			want: "llama.cpp: binaries not on PATH — server running :8000",
		},
	}
	for _, tc := range cases {
		if g := tc.r.Summary(); g != tc.want {
			t.Errorf("Summary() = %q want %q", g, tc.want)
		}
	}
}

func TestListenPort_default(t *testing.T) {
	os.Unsetenv(EnvLlamaServerPort)
	if p := ListenPort(); p != defaultLlamaServerPort {
		t.Fatalf("got %d", p)
	}
	t.Setenv(EnvLlamaServerPort, "9000")
	if p := ListenPort(); p != 9000 {
		t.Fatalf("got %d", p)
	}
}
