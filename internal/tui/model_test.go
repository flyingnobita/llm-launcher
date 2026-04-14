package tui

import (
	"testing"
	"time"

	"github.com/flyingnobita/llml/internal/llamacpp"
)

func TestLayoutTable_wideTerminalFitsViewport(t *testing.T) {
	m := New()
	m.width = 203
	m.height = 80
	m.files = []llamacpp.ModelFile{
		{
			Backend: llamacpp.BackendLlama,
			Path:    "/x",
			Name:    "m",
			Size:    1,
			ModTime: time.Unix(0, 0),
		},
	}
	m.loading = false
	m = m.layoutTable()
	innerW := m.bodyInnerW
	if m.tableLineWidth > innerW {
		t.Fatalf("table line width %d > inner width %d (spurious horizontal scroll)", m.tableLineWidth, innerW)
	}
}

func TestNew_zeroSize(t *testing.T) {
	m := New()
	if m.width != 0 || m.height != 0 {
		t.Fatalf("expected zero dimensions, got %dx%d", m.width, m.height)
	}
	if !m.loading {
		t.Fatal("expected loading true before first frame")
	}
}
