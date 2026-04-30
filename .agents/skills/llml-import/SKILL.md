---
name: llml-import
version: 1.4.0
description: |
  Import parameter profiles into llml from a URL or local file. Fetches the content,
  uses the portable profile format spec to extract structured profiles, and writes
  them directly into model-params.json ready to use in llml.
  Usage: /llml-import <url-or-file-path>
triggers:
  - llml-import
  - import profiles
  - import llml profiles
allowed-tools:
  - Bash
  - Read
  - Write
  - WebFetch
  - AskUserQuestion
---

# /llml-import

Import parameter profiles from a URL or local file into llml's model-params.json.

## Step 1: Parse the argument

Extract the URL or file path from the user's invocation. If no argument is provided,
ask the user:

> "Provide a URL or local file path to import profiles from."

## Step 2: Load the profile format spec

Read the canonical spec from the repo:

```bash
cat "$(git rev-parse --show-toplevel 2>/dev/null)/docs/profile-format.md"
```

If the git root is unavailable or the file is missing, use the embedded spec at the
end of this skill file.

## Step 3: Fetch the source content

- **URL** (starts with `http://` or `https://`): use WebFetch to retrieve the page content.
- **File path**: use Read to load the file.

If the fetch fails, stop and report the error clearly.

## Step 4: Extract profiles

You are the extraction agent. Using the format spec from Step 2 and the source
content from Step 3, extract all parameter profiles.

Rules (from the spec's LLM extraction section):
- Extract only parameters that appear explicitly in the source. Do not invent values.
- Each `--flag value` pair becomes one string in `args` (e.g. `"--n-gpu-layers 80"`).
- Standalone flags (e.g. `--flash-attn`) are single-element strings.
- Environment variables become `[[profiles.env]]` entries.
- Set `backend` to `llama`, `vllm`, or `ollama` based on the source context.
- Set `model_hint` to the model name or family from the source.
- Derive `name` from the section heading or context (e.g. `"default"`, `"4-bit-gpu"`,
  `"cpu-only"`).
- Set `description` to one sentence summarizing the profile's purpose and target
  hardware, if the source provides enough context.
- **Do not extract model-location parameters into the portable profile.** Exclude
  flags and env vars that identify where to load the model from, because llml
  supplies the model path itself. Examples to exclude: `LLAMA_CACHE`, `-hf`,
  `--model`, `--lora`, `HF_HOME`, `--tokenizer`, and similar source-location or
  cache-location settings.
- **Import-time stripping remains a defensive backstop.** Step 6 still strips any
  model-location parameters that slip through from messy real-world sources, but the
  extracted portable TOML should not contain them in the first place.

If no valid profiles can be extracted, stop and report: "No recognizable inference
parameters found in the source. Try a model card, README, or a page with llama.cpp,
vLLM, or Ollama launch commands."

Before presenting the list, run this script to classify each distinct `model_hint`
against the local model list:

```bash
python3 << 'PYEOF'
import json, os, re

def load_local_models():
    config_path = os.path.expanduser('~/.config/llml/config.toml')
    try:
        import tomllib
        with open(config_path, 'rb') as f:
            cfg = tomllib.load(f)
        return cfg.get('models', [])
    except Exception:
        try:
            with open(config_path) as f:
                content = f.read()
            models = []
            for sec in re.split(r'\[\[models\]\]', content)[1:]:
                b = (re.search(r'^backend\s*=\s*"([^"]*)"', sec, re.M) or type('',(),{'group':lambda s,n:''})()).group(1)
                n = (re.search(r'^name\s*=\s*"([^"]*)"', sec, re.M) or type('',(),{'group':lambda s,n:''})()).group(1)
                p = (re.search(r'^(?:path|id)\s*=\s*"([^"]*)"', sec, re.M) or type('',(),{'group':lambda s,n:''})()).group(1)
                models.append({'backend': b, 'name': n, 'path': p})
            return models
        except Exception:
            return []

def norm(s):
    return re.sub(r'[-_.\s]', '', s.lower())

def find_match(hint, models):
    h = norm(hint)
    for sfx in ('gguf', 'safetensors', 'ggml'):
        if h.endswith(sfx):
            h = h[:-len(sfx)]
    if not h:
        return None
    for m in models:
        name = m.get('name', '')
        path = m.get('path', m.get('id', ''))
        display = name if name else os.path.basename(path)
        if h in norm(name) or h in norm(path):
            return {'display': display, 'path': path}
    return None

local_models = load_local_models()
# HINTS injected by caller: a JSON list of unique model_hint strings
hints = HINTS
results = {hint: find_match(hint, local_models) for hint in hints}
print(json.dumps(results))
PYEOF
```

Set `HINTS` to a JSON list of the unique `model_hint` values before running (inline
in the heredoc). The script outputs `{ hint: {display, path} | null }`.

Group profiles by `model_hint`, then split into two visually distinct sections based
on the classification. Assign sequential numbers across all groups. Present as:

```
Found N profile(s) across M model variant(s):

✓  In your system
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[Qwen3-35B-A3B-GGUF]  →  Qwen3-35B-A3B-Q4_K_M.gguf
  1. "thinking-fast" — backend: llama
     args: [--n-gpu-layers 80, --ctx-size 8192]
  2. "thinking-slow" — backend: llama
     args: [--n-gpu-layers 80, --ctx-size 32768]

✗  Not found in your system
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[Qwen3-8B-GGUF]
  3. "default" — backend: llama
     args: [--n-gpu-layers 80]
[Qwen3-0.5B-GGUF]
  4. "default" — backend: llama
     args: [--n-gpu-layers 0]
```

Omit a section entirely if it has no entries. If all hints matched, omit the
"Not found" section; if none matched, omit the "In your system" section.

Ask: "Enter profile numbers to import (e.g. `1,3`), `all`, or `none`."

If the user answers `none`, stop and report that nothing was imported.

## Step 5: Resolve target model paths

Read the discovered model list from config.toml once. Use a regex fallback in case the
file has a parse error (e.g., stray characters from a prior edit):

```bash
python3 << 'EOF'
import os, re

config_path = os.path.expanduser('~/.config/llml/config.toml')
try:
    import tomllib
    with open(config_path, 'rb') as f:
        cfg = tomllib.load(f)
    models = cfg.get('models', [])
    for i, m in enumerate(models):
        print(f"{i+1}. [{m.get('backend','')}] {m.get('name','')}")
        print(f"   path: {m.get('path', m.get('id',''))}")
except Exception:
    # Fallback: regex parse [[models]] sections (tolerates corrupt lines elsewhere)
    try:
        with open(config_path) as f:
            content = f.read()
        sections = re.split(r'\[\[models\]\]', content)[1:]
        for i, sec in enumerate(sections):
            b = (re.search(r'^backend\s*=\s*"([^"]*)"', sec, re.M) or type('', (), {'group': lambda s,n: ''})()).group(1)
            n = (re.search(r'^name\s*=\s*"([^"]*)"', sec, re.M) or type('', (), {'group': lambda s,n: ''})()).group(1)
            p = (re.search(r'^(?:path|id)\s*=\s*"([^"]*)"', sec, re.M) or type('', (), {'group': lambda s,n: ''})()).group(1)
            print(f"{i+1}. [{b}] {n}")
            print(f"   path: {p}")
    except Exception as e2:
        print(f"ERROR: {e2}")
EOF
```

For each distinct `model_hint` among the selected profiles (in the order they appear),
use the Step 4 classification result to determine whether a fuzzy match was found:

- **If the hint was in the "In your system" section** (matched a local model): pre-fill
  the answer using the matched path from Step 4 and skip asking the user — unless the
  matched model is ambiguous (multiple hints matched the same file) or the user may want
  a different target. In that case, show the suggested path and confirm:
  > "Attach `<model_hint>` profiles to `<matched_display_name>` (`<matched_path>`)? [Y/n]"
  If the user answers Y or just presses enter, use the matched path.

- **If the hint was in the "Not found in your system" section** (no local match): ask
  once per unmatched hint:
  > "Which local model should the `<model_hint>` profiles be attached to?
  > Enter a number from the list above, or paste a model path directly.
  > (Enter 'skip' to skip this hint's profiles.)"
  If the user answers 'skip', omit those profiles from the import.

If config.toml is missing or has no models, ask the user to paste the model path
directly (file path for llama.cpp/vLLM, `ollama://model-name` for Ollama).

Build a mapping `{ model_hint: local_model_path }` from the answers. Two different
`model_hint` values may map to the same local path — that is valid and handled in
Step 6 by merging their profiles into the same model entry.

## Step 6: Write to model-params.json

Resolve the model-params.json path:

```bash
python3 -c "
import os, json
cfg_dir = os.path.expanduser('~/.config/llml')
path = os.path.join(cfg_dir, 'model-params.json')
print(path)
"
```

Read the existing file (empty `{"version": 2, "models": {}}` if missing). Then merge
the selected profiles:

Write a Python script that:
1. Groups selected profiles by their resolved local model key
2. Strips model-location parameters (env vars and args) that conflict with llml's local-path launch
3. Expands panel-row args to pre-split tokens (matching llml's `saveModelEntry` storage format)
4. Merges all groups into model-params.json in one read/write pass, skipping profiles whose name already exists
5. Writes atomically via temp file + rename

```python
import json, os, pathlib
from collections import defaultdict

MODEL_PARAMS_PATH = os.path.expanduser('~/.config/llml/model-params.json')

# Per-runtime model-location parameters. Stripped at import time because llml supplies
# the model path itself at launch. Ollama is omitted — it uses API discovery.
MODEL_LOCATION_PARAMS = {
    'llama': {
        'env': {
            'LLAMA_CACHE',
            'LLAMA_ARG_MODEL', 'LLAMA_ARG_MODEL_URL', 'LLAMA_ARG_MODEL_DRAFT',
            'LLAMA_ARG_HF_REPO', 'LLAMA_ARG_HF_FILE',
            'LLAMA_ARG_HFD_REPO',
            'LLAMA_ARG_HF_REPO_V', 'LLAMA_ARG_HF_FILE_V',
            'LLAMA_ARG_DOCKER_REPO',
            'LLAMA_ARG_MMPROJ', 'LLAMA_ARG_MMPROJ_URL',
            'LLAMA_ARG_MODELS_DIR', 'LLAMA_ARG_MODELS_PRESET',
            'HF_TOKEN',
        },
        'args': {
            '-m', '--model',
            '-mu', '--model-url',
            '-md', '--model-draft',
            '-mv', '--model-vocoder',
            '-hf', '-hfr', '--hf-repo',
            '-hff', '--hf-file',
            '-hfd', '-hfrd', '--hf-repo-draft',
            '-hfv', '-hfrv', '--hf-repo-v',
            '-hffv', '--hf-file-v',
            '-hft', '--hf-token',
            '-dr', '--docker-repo',
            '-mm', '--mmproj',
            '-mmu', '--mmproj-url',
            '--lora', '--lora-scaled', '--lora-init-without-apply',
            '--control-vector', '--control-vector-scaled',
            '--models-dir', '--models-preset',
            '-lcs', '--lookup-cache-static',
            '-lcd', '--lookup-cache-dynamic',
        },
    },
    'vllm': {
        'env': {
            'HF_HOME', 'HF_TOKEN', 'HF_HUB_TOKEN',
            'HUGGINGFACE_HUB_CACHE', 'HUGGING_FACE_HUB_TOKEN',
            'TRANSFORMERS_CACHE',
            'VLLM_CACHE_ROOT', 'VLLM_ASSETS_CACHE',
            'VLLM_MODEL_REDIRECT_PATH', 'VLLM_XLA_CACHE_PATH',
            'VLLM_USE_MODELSCOPE', 'MODELSCOPE_CACHE',
        },
        'args': {
            '--model', '--tokenizer',
            '--revision', '--code-revision', '--tokenizer-revision',
            '--hf-config-path', '--hf-token', '--hf-overrides',
            '--download-dir', '--load-format',
            '--model-loader-extra-config', '--config',
            '--qlora-adapter-name-or-path',
            '--lora-modules', '--prompt-adapters',
            '--speculative-config', '--speculative-model',
            '--tokenizer-pool-extra-config',
        },
    },
}

def expand_arg_line(line):
    """Mirror of llml's expandArgLine: split '--flag value' into two tokens."""
    line = line.strip()
    if not line:
        return []
    if not line.startswith('-') or ' ' not in line:
        return [line]
    i = line.index(' ')
    return [line[:i], line[i+1:].strip()]

def _arg_first_token(panel_row):
    s = (panel_row or '').strip()
    if not s:
        return ''
    i = s.find(' ')
    return s if i < 0 else s[:i]

def sanitize_profile(p):
    """Strip env/args that specify model-file location for the profile's backend.
    Returns (profile, dropped_env_strs, dropped_arg_strs)."""
    table = MODEL_LOCATION_PARAMS.get(p.get('backend', ''))
    if not table:
        return p, [], []
    kept_env, dropped_env = [], []
    for e in (p.get('env') or []):
        if e.get('key') in table['env']:
            dropped_env.append(f"{e.get('key')}={e.get('value','')}")
        else:
            kept_env.append(e)
    p['env'] = kept_env
    kept_args, dropped_args = [], []
    for a in (p.get('args') or []):
        if _arg_first_token(a) in table['args']:
            dropped_args.append(a)
        else:
            kept_args.append(a)
    p['args'] = kept_args
    return p, dropped_env, dropped_args

def normalize_key(k):
    if not k.startswith('ollama://') and '://' not in k:
        return os.path.normpath(k)
    return k

# HINT_TO_KEY: {model_hint: raw_local_path} — set by caller from Step 5 answers.
# new_profiles: all selected profiles, each retaining their 'model_hint' field.

# Group profiles by their resolved local model key.
key_to_profiles = defaultdict(list)
for p in new_profiles:
    key = normalize_key(HINT_TO_KEY.get(p.get('model_hint', ''), ''))
    if key:
        key_to_profiles[key].append(p)

try:
    with open(MODEL_PARAMS_PATH) as f:
        data = json.load(f)
except FileNotFoundError:
    data = {'version': 2, 'models': {}}
if data.get('models') is None:
    data['models'] = {}
data['version'] = 2

results = {}
for model_key, profiles_for_key in key_to_profiles.items():
    entry = data['models'].get(model_key, {'profiles': [], 'activeIndex': 0})
    existing_names = {p['name'] for p in entry.get('profiles', [])}
    added, skipped, filtered_summary = [], [], []
    for p in profiles_for_key:
        if p['name'] in existing_names:
            skipped.append(p['name'])
            continue
        p, dropped_env, dropped_args = sanitize_profile(p)
        if dropped_env or dropped_args:
            filtered_summary.append({
                'name': p['name'],
                'env': dropped_env,
                'args': dropped_args,
            })
        p['args'] = [tok for a in (p.get('args') or []) for tok in expand_arg_line(a)]
        if p.get('env') is None:
            p['env'] = []
        entry['profiles'].append(p)
        added.append(p['name'])
    if not entry['profiles']:
        entry['profiles'] = [{'name': 'default', 'env': [], 'args': []}]
    data['models'][model_key] = entry
    results[model_key] = {'added': added, 'skipped': skipped, 'filtered': filtered_summary}

os.makedirs(os.path.dirname(MODEL_PARAMS_PATH), exist_ok=True)
tmp = MODEL_PARAMS_PATH + '.tmp'
with open(tmp, 'w') as f:
    json.dump(data, f, indent=2)
os.replace(tmp, MODEL_PARAMS_PATH)

print(json.dumps({'results': results}))
```

Set `HINT_TO_KEY` and `new_profiles` as Python variables before running this block
(inline in a heredoc script, not via sys.argv, to avoid quoting issues with JSON).
`new_profiles` must retain each profile's `model_hint` field so the grouping works.

## Step 7: Report

Show one block per local model key in the results dict:

```
Imported:

[/path/to/qwen3-35b-a3b.gguf]
  Added (2): "thinking-fast", "thinking-slow"
  Skipped — name already exists (1): "default"
  Filtered model-location parameters:
    - "thinking-fast": env: LLAMA_CACHE · args: -hf unsloth/Qwen3-...
  (Filtered entries were stripped because llml supplies the model path at launch.)

[/path/to/qwen3-8b.gguf]
  Added (1): "default"

Press 'p' in llml to view and activate the imported profiles.
```

Omit the "Skipped" line when count is 0. Omit the "Filtered" block when nothing was
stripped. If every profile for a given key was skipped, suggest renaming them in
llml's parameter panel (`p`) before re-importing.

---

## Embedded spec (fallback)

If `docs/profile-format.md` is unavailable, use this summary:

The portable profile format is TOML with `schema_version = 1` and a `[[profiles]]`
array.

Required fields:
- `schema_version = 1`
- `[[profiles]].name`
- `[[profiles]].backend` (`llama`, `vllm`, or `ollama`)

Optional fields:
- `model_hint`
- `description`
- `args` as panel-row strings, for example `"--ctx-size 4096"`
- `[[profiles.env]]` as `{key, value}` pairs

Do not include model-location parameters in `args` or `env`. llml supplies the model
path itself.

Example:
```toml
schema_version = 1

[[profiles]]
name = "balanced"
backend = "llama"
model_hint = "Llama-3-8B"
args = ["--n-gpu-layers 80", "--ctx-size 4096"]
```
