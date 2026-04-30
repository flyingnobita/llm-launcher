# Portable Parameter Profile Format

Date: 2026-04-27
Status: Proposed

## Purpose

This document defines the portable parameter profile format for llml. It serves two
purposes:

1. **Human documentation** — describes what a shareable profile file looks like and
   how to write one by hand.
2. **Machine prompt context** — an LLM pointed at this document and a source URL
   (model card, blog post, README) can extract structured profiles without
   additional instructions.

## Background

llml stores parameter profiles per model in `{UserConfigDir}/llml/model-params.json`.
Each profile has a name, a list of environment variables, and a list of command-line
args. This internal format is not portable — it is keyed by local model path and not
designed for sharing.

The portable profile format defined here is a separate, self-contained TOML file
intended for sharing and importing. Import tooling (e.g., the `/llml-import` agent
skill) reads this format and writes the resulting profiles into `model-params.json`.

## Scope

- Covers llama.cpp, vLLM, and Ollama backends.
- One file may contain multiple profiles, for multiple backends, targeting one or
  more model families.
- Cross-backend translation (e.g., converting a llama.cpp profile to an Ollama
  equivalent) is explicitly out of scope for this format version.

## Schema

### Top-level fields

| Field            | Type    | Required | Description                   |
| ---------------- | ------- | -------- | ----------------------------- |
| `schema_version` | integer | yes      | Must be `1` for this version. |

### `[[profiles]]` array

Each entry in `[[profiles]]` is one parameter profile.

| Field         | Type            | Required | Description                                                                                                                                |
| ------------- | --------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `name`        | string          | yes      | Short human-readable name for this profile. Used as the profile name inside llml.                                                          |
| `backend`     | string          | yes      | One of: `llama`, `vllm`, `ollama`. Profiles are backend-specific; do not omit this field.                                                  |
| `model_hint`  | string          | no       | Free-text hint for which model family this profile targets (e.g. `"Llama-3-8B-GGUF"`, `"Qwen2.5-72B"`). Used for display only; not a path. |
| `description` | string          | no       | Human-readable description of what this profile does and when to use it.                                                                   |
| `args`        | array of string | no       | Command-line arguments in panel-row format (see below). Defaults to empty.                                                                 |
| `env`         | array of table  | no       | Environment variables as `{key, value}` pairs. Defaults to empty.                                                                          |

### `args` format

Each element of `args` is one "panel row" — the same string a user would type into
llml's parameter panel. Two formats are accepted:

- `"--flag value"` — a flag and its value separated by a single space. llml splits
  this into two argv tokens at launch time.
- `"--flag"` — a standalone flag (no value).
- `"/path/to/something"` — a bare value (no leading dash).

Do not include the model path or the backend binary name; llml supplies those.

**Storage note:** The portable TOML format stores args as panel-row strings. When
import tooling writes to `model-params.json` it pre-splits each `"--flag value"`
into two separate tokens (`["--flag", "value"]`) to match llml's internal storage
format. This split is the importer's responsibility, not the LLM's — extract args
as panel-row strings in the portable TOML.

### `[[profiles.env]]` format

Each entry is a table with two string fields:

| Field   | Type   | Required | Description                 |
| ------- | ------ | -------- | --------------------------- |
| `key`   | string | yes      | Environment variable name.  |
| `value` | string | yes      | Environment variable value. |

## Example: llama.cpp profiles

```toml
schema_version = 1

[[profiles]]
name = "balanced-q4"
backend = "llama"
model_hint = "Llama-3-8B-GGUF"
description = "Q4_K_M quant, 80 GPU layers, 4096 ctx — M1 Max 32GB"
args = ["--n-gpu-layers 80", "--ctx-size 4096", "--threads 8"]

[[profiles]]
name = "cpu-only"
backend = "llama"
model_hint = "Llama-3-8B-GGUF"
description = "No GPU offload, low memory footprint"
args = ["--n-gpu-layers 0", "--ctx-size 2048", "--threads 4"]

[[profiles]]
name = "max-context"
backend = "llama"
model_hint = "Llama-3-8B-GGUF"
description = "Full 32k context window, requires more VRAM"
args = ["--n-gpu-layers 80", "--ctx-size 32768"]

[[profiles.env]]
key = "LLAMA_CACHE_SIZE"
value = "8192"
```

## Example: vLLM profiles

```toml
schema_version = 1

[[profiles]]
name = "single-gpu"
backend = "vllm"
model_hint = "Qwen2.5-72B-Instruct-AWQ"
description = "Single A100 80GB, AWQ quantization"
args = ["--tensor-parallel-size 1", "--gpu-memory-utilization 0.95", "--max-model-len 8192"]

[[profiles]]
name = "dual-gpu"
backend = "vllm"
model_hint = "Qwen2.5-72B-Instruct-AWQ"
description = "Two A100 80GB, higher throughput"
args = ["--tensor-parallel-size 2", "--gpu-memory-utilization 0.90", "--max-model-len 32768"]
```

## Example: Ollama profiles

```toml
schema_version = 1

[[profiles]]
name = "gpu-full"
backend = "ollama"
model_hint = "llama3.2"
description = "All layers on GPU, 8k context"
args = []

[[profiles.env]]
key = "OLLAMA_NUM_GPU"
value = "999"

[[profiles.env]]
key = "OLLAMA_NUM_CTX"
value = "8192"
```

## LLM extraction instructions

If you are an LLM reading this document in order to extract profiles from a source,
follow these rules:

1. **Backend detection:** Identify whether the source describes llama.cpp, vLLM, or
   Ollama invocations. Use the `backend` field accordingly (`llama`, `vllm`,
   `ollama`). If the source covers multiple backends, emit one `[[profiles]]` entry
   per backend variant.

2. **Args extraction:** Extract only flags and values that appear explicitly in the
   source. Do not infer or invent values. Place each `--flag value` pair as a single
   string in the `args` array. Standalone flags (`--flash-attn`, `--no-mmap`) are
   single-element strings.

3. **Env extraction:** Extract environment variables set in the source (e.g.,
   `export LLAMA_CACHE_SIZE=4096`, `OLLAMA_NUM_GPU=999 ollama run ...`). Place each
   as a `{key, value}` entry under `[[profiles.env]]`.

4. **model_hint:** Set to the model name or family mentioned in the source (e.g.,
   the model card title or the model name in the `ollama run` command). Use the
   short name, not a full path.

5. **name:** Derive from the source context — e.g., `"default"`, `"4-bit-gpu"`,
   `"cpu-only"`, or the section heading the parameters appeared under. Keep it short.

6. **description:** One sentence summarizing what the profile does and what hardware
   or use case it targets, based on information in the source. If the source does not
   provide enough context, omit this field.

7. **Output:** Emit valid TOML matching this schema. Set `schema_version = 1`. Do not
   include fields not listed in this spec. Do not include the model path or binary
   name in `args`.

## Versioning

`schema_version = 1` is the only currently valid value. Future versions will be
documented here. Import tooling should reject files with unrecognized schema versions
with a clear error.

## Relation to internal model-params.json

The portable format and `model-params.json` are separate. The portable format is for
sharing and importing; `model-params.json` is the internal per-model storage keyed
by local model path. Import tooling bridges the two: it reads a portable file, asks
the user which local model to attach the profiles to, and writes into
`model-params.json` under that model's key.
