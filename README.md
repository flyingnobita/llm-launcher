# LLM Launcher

[![Go](https://img.shields.io/github/go-mod/go-version/flyingnobita/llml)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**LLM Launcher** (`llml`) is a terminal UI for discovering local **GGUF** and **Hugging Face-style safetensors** models. It supports multiple inference engines including **[llama.cpp](https://github.com/ggerganov/llama.cpp)** (`llama-server`) and **[vLLM](https://github.com/vllm-project/vllm)** (`vllm serve`).

## Features

- **One table** for for all your GGUF files and safetensors models.
- **Parameter profiles** (`p`): per-model named profiles of extra environment variables and CLI arguments.This allows easy model runtime parameter tweaking.
- **Auto-discovery** of for scan roots and binary resolution, overridable via environment variables.

## Requirements

- **Go** [1.26+](go.mod) only if you build from source or work on the project.
- **llama.cpp** binaries (`llama-server`, and optionally `llama-cli`) for GGUF rows, and/or **vLLM** (`vllm`) for safetensors rows, installed where the app can find them (see [Runtime detection](#runtime-detection)).
- **Node.js** (LTS) only if you run project checks (`npm ci` + Prettier / markdownlint via `mise run check`).

## Install

### Pre-built binaries (no Go)

For each [GitHub release](https://github.com/flyingnobita/llml/releases), archives are published for Linux and macOS (`tar.gz`) and Windows (`zip`) on **amd64** and **arm64** (Windows is amd64 only). Download the archive for your OS and CPU, extract the `llml` binary (or `llml.exe` on Windows), and place it on your `PATH`.

Verify the download against `llml_<version>_checksums.txt` on the release page if you rely on checksums.

### From source

```bash
git clone https://github.com/flyingnobita/llml.git
cd llml
go build -o llml ./cmd/llml
```

Install on your `PATH` if you like:

```bash
go install github.com/flyingnobita/llml/cmd/llml@latest
```

(Ensure `$(go env GOPATH)/bin` is on your `PATH`.)

### With mise (optional)

If you use [mise](https://mise.jdx.dev/), the repo includes tasks for run, build, and full checks:

```bash
mise install
mise run build    # binary: bin/llml
```

## Quick start

```bash
./llml
# or: mise run run
```

Place models under default scan locations (see [How it finds models](#how-it-finds-models)) or set `LLM_LAUNCH_LLAMACPP_PATHS`. Point `LLAMA_CPP_PATH` / `VLLM_PATH` at your install dirs if binaries are not on `PATH`.

## Usage

| Key          | Action                                                                            |
| ------------ | --------------------------------------------------------------------------------- |
| `hjkl/↑↓←→`  | Move selection; horizontal scroll when the path column is wider than the terminal |
| `r`          | Rescan filesystem                                                                 |
| **`R`**      | Run server (split view: table + log pane)                                         |
| **ctrl+`R`** | Run server full-screen (`tea.ExecProcess`; same as older behavior)                |
| `c`          | Edit runtime environment (paths, ports)                                           |
| `p`          | Edit parameter profiles for the selected model                                    |
| `t`          | Cycle theme (`dark` → `light` → `auto` → …)                                       |
| `q`          | Quit                                                                              |

### Server output

**`R`** runs the server in a **split layout**: the model table stays in the upper half and **stdout/stderr** stream into a scrollable log pane below (ANSI colors pass through). The model list is always shown with a rounded border; in split mode the **focused** pane uses a brighter border and the other a dimmer one. Focus starts on the **table**; **tab** switches focus between the table and the log. **esc**, **q**, or **ctrl+c** stops the subprocess; **`hjkl/↑↓←→`** moves/scrolls the focused pane (including **PgUp/PgDn** where supported; mouse wheel follows focus).

**ctrl+`R`** uses Bubble Tea’s [`tea.ExecProcess`](https://github.com/charmbracelet/bubbletea): the alternate screen is released and the server process is attached to your terminal so logs print like a normal process until exit. (We use **ctrl+`R`** for full-screen instead of shift+`R` because typing uppercase **R** on common layouts already involves Shift, which would be indistinguishable from a separate “shift+R” binding.) On **Linux/macOS**, a small `sh` wrapper echoes the exact command (line starting with `+`), runs the server, then prompts **Press Enter to return to LLM Launcher…** before restoring the TUI. On **Windows**, the server runs without that echo/pause.

You can also run the printed command manually in another terminal, or redirect output with your shell.

### Parameter profiles (`p`)

Each model path can have **multiple named profiles**. Each profile stores:

- **Environment variables** (`KEY=value` per line).
- **Extra arguments** appended after `--port` (for vLLM, flags and values are separate argv tokens; the UI may show `--flag value` on one line).

**`R`** / **ctrl+`R`** use the **active** profile (the one highlighted in the `p` panel; changes persist automatically). **tab** cycles: profile list → env → extra args. In the list: **`n`** new profile, **`d`** delete (not the last), **`r`** rename. **esc** or **q** closes the panel (**q** on the main screen still quits the app unless a split-pane server is running).

Storage is a single JSON file (not environment variables):

| Platform    | Typical path                                                                    |
| ----------- | ------------------------------------------------------------------------------- |
| Linux (XDG) | `$XDG_CONFIG_HOME/llml/model-params.json` or `~/.config/llml/model-params.json` |
| macOS       | `~/Library/Application Support/llml/model-params.json`                          |
| Windows     | `%AppData%\llml\model-params.json`                                              |

## Configuration

There is **no** runtime `config.toml` and **no** automatic `.env` file. Behavior is driven by **environment variables** and the **parameter profiles** file above.

### Discovery

Default roots include `~/models`, `~/.cache/llama.cpp`, Hugging Face hub cache paths, and `~/.cache/lm-studio/models` (only existing directories are used).

Add extra roots (comma-separated):

```bash
export LLM_LAUNCH_LLAMACPP_PATHS="/data/models,/opt/weights"
```

`HUGGINGFACE_HUB_CACHE` / `HF_HOME` influence Hugging Face cache layout as usual.

### Ports and paths (runtime)

| Variable            | Default   | Role                                                                        |
| ------------------- | --------- | --------------------------------------------------------------------------- |
| `LLAMA_CPP_PATH`    | _(unset)_ | Directory containing `llama-cli` / `llama-server` (checked before `PATH`)   |
| `VLLM_PATH`         | _(unset)_ | Directory where `vllm` or `.venv/bin/vllm` may live                         |
| `VLLM_VENV`         | _(unset)_ | Python venv root; on Unix, **`R`** may `source bin/activate` before `vllm`  |
| `LLAMA_SERVER_PORT` | `8080`    | Port for `llama-server` and `/health` probe                                 |
| `VLLM_SERVER_PORT`  | `8000`    | Port for `vllm serve`                                                       |
| `LLML_THEME`        | `auto`    | Initial TUI palette; **`t`** cycles `dark` → `light` → `auto` while running |

Set these in your shell, or under `[env]` in `mise.toml` for local development.

## How it finds models

- **GGUF**: files ending in `.gguf` under the scan roots.
- **Safetensors (vLLM)**: a directory containing **`config.json`** and at least one **`*.safetensors`** file (Hugging Face-style checkpoint).

The table shows a decoded repo id from `models--*` hub folders when possible; otherwise folder names and paths are shown with `~/` shortened and truncation when the terminal is narrow.

## Runtime detection

On launch and on **`r`**, the app resolves **llama.cpp** and **vLLM** binaries, then scans for models.

### llama.cpp (`llama-cli` / `llama-server`)

1. **`LLAMA_CPP_PATH`** if set: `{LLAMA_CPP_PATH}/<binary>` must exist.
2. Common locations: `/usr/local/bin`, `/opt/homebrew/bin`, `/opt/llama.cpp/build/bin`, `~/.local/bin`.
3. **`PATH`** via `exec.LookPath`.

If both binaries are still missing, the app may probe a **running** llama.cpp server: HTTP `GET` `http://127.0.0.1:<LLAMA_SERVER_PORT>/health` (2s timeout, success = HTTP 200).

### vLLM (`vllm`)

1. **`VLLM_PATH`**: `{VLLM_PATH}/vllm` or `{VLLM_PATH}/.venv/bin/vllm`.
2. **`VLLM_VENV`**: `{VLLM_VENV}/bin/vllm` if present.
3. Common directories as above, then **`PATH`**.

On **Linux/macOS**, if vLLM lives in a venv, **`R`** may source `activate` before `vllm serve` (next to the resolved binary, or via `VLLM_VENV` / `.venv` heuristics). On **Windows**, use an activated shell or put `vllm` on `PATH`.

## Development

Clone the repository and install tooling:

```bash
mise install          # Go + pre-commit (optional)
npm ci                # Prettier + markdownlint
go mod download
```

Common tasks:

```bash
mise run run       # go run ./cmd/llml
mise run build     # bin/llml
mise run check     # fmt + vet + prettier + markdownlint + tests (race)
```

Layout:

- `cmd/llml` — entrypoint.
- `internal/tui` — Bubble Tea UI.
- `internal/llamacpp` — discovery, metadata, runtime detection.

Contributions are welcome. Please run `mise run check` (or equivalent) before opening a pull request.

## License

[MIT](LICENSE)
