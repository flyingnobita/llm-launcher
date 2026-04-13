package tui

import "testing"

func TestNew_zeroSize(t *testing.T) {
	m := New()
	if m.width != 0 || m.height != 0 {
		t.Fatalf("expected zero dimensions, got %dx%d", m.width, m.height)
	}
}
