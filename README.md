# llm-launch

Terminal UI built with [Bubble Tea](https://github.com/charmbracelet/bubbletea),
[Lip Gloss](https://github.com/charmbracelet/lipgloss), and
[Bubbles](https://github.com/charmbracelet/bubbles) (table + key bindings). Scaffolded from
[flyingnobita/project-template](https://github.com/flyingnobita/project-template).

## What it does (v1)

Scans common directories for **llama.cpp**-style **GGUF** model files and shows them in a table:

| Column        | Meaning                                                                                                                                                      |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Name          | File basename                                                                                                                                                |
| Path          | Model folder: Hugging Face hub cache stops at `models--*` (no `snapshots/<hash>/`); elsewhere, parent dir of the file; `~/` shortened; truncated when narrow |
| Size          | File size on disk                                                                                                                                            |
| Last modified | File `mtime` (not “last inference run”; there is no OS-wide run log)                                                                                         |
| Parameters    | Summary from GGUF metadata (architecture, context length when present)                                                                                       |

Press `r` to rescan, **`R`** (shift+r) to run **`llama-server`** with the selected GGUF, `q` or `Ctrl+C` to quit. Use arrow keys or `j` / `k` to move in the table.

### Seeing `llama-server` output

**Default (`R` in the app):** the TUI uses Bubble Tea’s [`tea.ExecProcess`](https://github.com/charmbracelet/bubbletea): it **releases the alternate screen** and attaches **`llama-server` to your terminal**, so **logs print normally** until the process exits. On **Linux/macOS**, a **`sh` script** first **echoes the exact command** (line starting with **`+`**, shell-quoted paths), then runs **`llama-server`**, then prints **“Press Enter to return to llm-launch…”** and waits for **Enter** before the alternate screen comes back. On **Windows**, the server runs directly without that echo/pause (use scrollback or another window if needed).

**Alternatives:**

1. **Another terminal** — run the same command the app would use: `llama-server -m <path/to/model.gguf> --port <port>` (port from `LLAMA_SERVER_PORT` or `8080`).
2. **Save logs** — shell redirect or `tee`, e.g. `llama-server ... 2>&1 | tee llama-server.log`.
3. **In-TUI log pane** — possible by piping stdout/stderr into a scrollable viewport; not implemented here (more moving parts than `ExecProcess`).

## Layout

1. `cmd/llm-launch` — `main`, calls `internal/tui.Run()`
2. `internal/tui` — Bubble Tea UI; table via `internal/tui/btable` (fork of Bubbles table)
3. `internal/llamacpp` — GGUF discovery and metadata

## Discovery paths

Default roots include `~/models`, `~/.cache/llama.cpp`, `~/.cache/huggingface/hub`,
`~/.cache/lm-studio/models` (only existing directories are scanned).

Add more roots (comma-separated):

```bash
export LLM_LAUNCH_LLAMACPP_PATHS="/data/models,/opt/weights"
```

## llama.cpp binary detection (runtime)

The TUI locates **`llama-cli`** and **`llama-server`** independently so it can show paths in the bottom panel. Implementation: `internal/llamacpp` (`DiscoverRuntime`).

**Startup order:** on launch (and on `r` refresh), the app runs **`llama.cpp` binary detection first**, then scans for GGUF files (`tea.Sequence` in the TUI).

**Per binary** (`llama-cli` and `llama-server` each use this sequence):

1. **`LLAMA_CPP_PATH`** — if set, the file `{LLAMA_CPP_PATH}/<binary-name>` must exist as a regular file.
2. **Common install directories** (first match wins): `/usr/local/bin`, `/opt/homebrew/bin`, `/opt/llama.cpp/build/bin`, `~/.local/bin`.
3. **`PATH`** — `exec.LookPath` (same as a normal shell lookup).

**If both binaries are still missing:** the app probes a **running server** with HTTP **GET** `http://127.0.0.1:<port>/health` (2s timeout, success = HTTP 200). Port comes from **`LLAMA_SERVER_PORT`**, or **8080** if unset.

### Port (`LLAMA_SERVER_PORT`)

Use **one environment variable** for every feature that needs a listen port:

| Use                                    | Behavior                                    |
| -------------------------------------- | ------------------------------------------- |
| **`R` (run server)**                   | `llama-server ... --port <n>`               |
| **Health probe** (no binaries on disk) | `GET http://127.0.0.1:<n>/health`           |
| **Bottom panel**                       | Shows `listen (R) :<n> (LLAMA_SERVER_PORT)` |

Set it for the whole session, e.g. `export LLAMA_SERVER_PORT=9090`, or in **`mise.toml`** under **`[env]`** (same as `LLAMA_CPP_PATH`). If something already binds **8080** (another `llama-server`, Ollama, etc.), pick a free port (e.g. **9090**) and restart the app so detection and **`R`** stay in sync.

| Variable            | Default  | Role                                                                      |
| ------------------- | -------- | ------------------------------------------------------------------------- |
| `LLAMA_CPP_PATH`    | _(none)_ | Directory containing `llama-cli` / `llama-server`; checked before `PATH`. |
| `LLAMA_SERVER_PORT` | `8080`   | Single listen port: **`R`**, `/health` probe, and UI label (see above).   |

## Setup

1. `mise install` — Go (see `mise.toml`) and tools (e.g. pre-commit via pipx)
2. `npm ci` or `npm install` — Prettier and markdownlint (formatting / CI)
3. `go mod download`

**Configuration:** the app has **no** `config.toml` and does not load a `.env` file. It reads **environment variables only** (e.g. `LLAMA_CPP_PATH`, `LLAMA_SERVER_PORT`, `LLM_LAUNCH_LLAMACPP_PATHS`). Set them in your shell or under `[env]` in `mise.toml` (see **Discovery paths** and **llama.cpp binary detection** above in this file).

## Usage

```bash
mise run run      # go run ./cmd/llm-launch
mise run build    # binary at bin/llm-launch
mise run check    # fmt + vet + prettier + markdownlint + tests
```

## Requirements

- Go: see `mise.toml` (`go = "latest"`)
- Node (LTS): for Prettier / markdownlint via `npm install`

## License

MIT
