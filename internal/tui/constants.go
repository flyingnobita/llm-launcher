package tui

// Layout constants used across model, view, and table_layout.
const (
	// minTerminalWidth is the minimum terminal width we attempt to render into.
	minTerminalWidth = 56

	// minInnerWidth is the minimum inner body width (after app padding).
	minInnerWidth = 40

	// defaultTableHeight is the fallback table row-area height before the first
	// WindowSizeMsg arrives.
	defaultTableHeight = 18

	// appPaddingH is the Lip Gloss horizontal padding per side (app style uses
	// Padding(1, 2), so 2 on each side = 4 total consumed columns).
	appPaddingH = 2

	// hScrollStep is the number of columns scrolled per arrow/key press.
	hScrollStep = 4

	// appSubtitle is the subtitle line shown below the app title.
	appSubtitle = "llama.cpp (GGUF) · vLLM (config.json + safetensors) — filesystem scan · Last modified = file mtime"

	// paramPanelMaxInnerWidth caps the parameters modal inner width on wide
	// terminals so the panel does not stretch edge-to-edge.
	paramPanelMaxInnerWidth = 88
)

// Column-width defaults for the model table.
const (
	defaultNameColW = 36
	runtimeColW     = 11 // "llama.cpp", "vllm"
	sizeColW        = 9
	modTimeColW     = 17
	maxNameColW     = 72
	minPathColW     = 14
	maxPathColW     = 400
	colPaddingExtra = 8 // extra padding bubbles/table adds across 5 columns
)
