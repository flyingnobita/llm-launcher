package llamacpp

import (
	"path/filepath"
	"testing"
)

func TestFormatModelFolderDisplay_hfSnapshots(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	file := filepath.Join(home, ".cache", "huggingface", "hub", "models--unsloth--gemma-GGUF", "snapshots", "8bacec5c8e829a25502cdfe3c3f5b6aabee3218c", "model.gguf")
	got := FormatModelFolderDisplay(file)
	want := filepath.Join("~", ".cache", "huggingface", "hub", "models--unsloth--gemma-GGUF")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatModelFolderDisplay_directInRepo(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	file := filepath.Join(home, ".cache", "huggingface", "hub", "models--unsloth--gemma-GGUF", "model.gguf")
	got := FormatModelFolderDisplay(file)
	want := filepath.Join("~", ".cache", "huggingface", "hub", "models--unsloth--gemma-GGUF")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatModelFolderDisplay_noHFRepoDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	file := filepath.Join(home, "models", "weights", "a.gguf")
	got := FormatModelFolderDisplay(file)
	want := filepath.Join("~", "models", "weights")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatPathDisplay_underHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	p := filepath.Join(home, "models", "x.gguf")
	got := FormatPathDisplay(p)
	want := filepath.Join("~", "models", "x.gguf")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatPathDisplay_homeDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got := FormatPathDisplay(home)
	if got != "~" {
		t.Fatalf("got %q want ~", got)
	}
}

func TestFormatPathDisplay_outsideHome(t *testing.T) {
	t.Setenv("HOME", "/tmp/llm-launch-test-home")

	abs := "/other/mount/model.gguf"
	if got := FormatPathDisplay(abs); got != abs {
		t.Fatalf("got %q want %q", got, abs)
	}
}
