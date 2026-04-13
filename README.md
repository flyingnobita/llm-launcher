# llm-launch

Terminal UI built with [Bubble Tea](https://github.com/charmbracelet/bubbletea),
[Lip Gloss](https://github.com/charmbracelet/lipgloss), and
[Bubbles](https://github.com/charmbracelet/bubbles) (key bindings). Scaffolded from
[flyingnobita/project-template](https://github.com/flyingnobita/project-template).

## Layout

1. `cmd/llm-launch` — `main`, wires `internal/tui.Run()`
2. `internal/tui` — Bubble Tea model: `model.go`, `update.go`, `view.go`, `keymap.go`, `styles.go`, `run.go`

Extend `Model` with your state, handle messages in `Update`, and render in `View`.
Add bindings in `keymap.go` and Lip Gloss styles in `styles.go`.

## Setup

1. Copy `.env.example` to `.env` if you need API keys or tokens
2. Copy `config.toml.example` to `config.toml` for app config (optional)
3. `mise install` — installs Go (latest) and other tools from `mise.toml`
4. `npm install` — Prettier and markdownlint (docs / pre-commit)
5. `go mod download`

## Usage

```bash
mise run run      # go run ./cmd/llm-launch
mise run build    # binary at bin/llm-launch
mise run check    # fmt + vet + prettier + markdownlint + tests
```

Quit the TUI with `q` or `Ctrl+C`.

## Requirements

- Go: see `mise.toml` (`go = "latest"`)
- Node (LTS): for Prettier / markdownlint via `npm install`

## License

MIT
