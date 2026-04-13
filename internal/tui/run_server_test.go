package tui

import (
	"strings"
	"testing"
)

func TestShellSingleQuoted(t *testing.T) {
	if g := shellSingleQuoted(`a'b`); g != `'a'"'"'b'` {
		t.Fatalf("got %q", g)
	}
	if g := shellSingleQuoted("/opt/bin/llama-server"); g != "'/opt/bin/llama-server'" {
		t.Fatalf("got %q", g)
	}
}

func TestFormatLlamaServerInvocation(t *testing.T) {
	got := formatLlamaServerInvocation("/bin/llama-server", "/m/a.gguf", 9090)
	want := "+ '/bin/llama-server' -m '/m/a.gguf' --port 9090"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestUnixLlamaServerScript_containsRead(t *testing.T) {
	s := unixLlamaServerScript("/bin/llama-server", "/m/model.gguf", 8080)
	if !strings.Contains(s, "read -r _") {
		t.Fatalf("expected read pause: %q", s)
	}
	if !strings.Contains(s, "'/bin/llama-server'") {
		t.Fatalf("expected quoted bin: %q", s)
	}
	if !strings.Contains(s, "printf") {
		t.Fatalf("expected echo of invocation: %q", s)
	}
}
