#!/usr/bin/env bash

# Synced relative paths (keep aligned with .gitignore "# Agents" + mise.local.toml).
AGENT_PATHS=(.claude .codex .cursor .gemini .gstack .opencode TODOS.md mise.local.toml)

set -euo pipefail

# Sync gitignored local paths between this repo and a peer directory: entries under
# "# Agents" in .gitignore plus mise.local.toml (Environment section).
#
# Default peer (when LLML_AGENTS_PEER is unset):
#   - In a linked git worktree: the primary (first) checkout from "git worktree list",
#     so import pulls ignored agent files from the main tree into the worktree.
#   - Otherwise: parent of the repo root (shared agent dirs next to the project).
#
# Direction (default: import peer -> repo):
#   LLML_AGENTS_SYNC=import  copy from peer into repo (default)
#   LLML_AGENTS_SYNC=export copy from repo into peer
#   LLML_AGENTS_SYNC=none    no-op
#
# Optional peer directory (overrides default above):
#   LLML_AGENTS_PEER=/path/to/directory

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
if [[ ! -e "$ROOT/.git" ]]; then
	echo "sync-gitignore-agents: expected git repo at $ROOT" >&2
	exit 1
fi

root_abs="$ROOT"
main_wt="$(git -C "$ROOT" worktree list --porcelain 2>/dev/null | awk '/^worktree /{print $2; exit}' || true)"
if [[ -n "$main_wt" ]]; then
	main_abs="$(cd "$main_wt" && pwd -P)"
	if [[ "$root_abs" != "$main_abs" ]]; then
		default_peer="$main_abs"
	else
		default_peer="$(dirname "$root_abs")"
	fi
else
	default_peer="$(dirname "$root_abs")"
fi
PEER="${LLML_AGENTS_PEER:-$default_peer}"

MODE="${LLML_AGENTS_SYNC:-import}"

if [[ "$MODE" == "none" ]]; then
	exit 0
fi

if [[ "$MODE" != "import" && "$MODE" != "export" ]]; then
	echo "sync-gitignore-agents: unknown LLML_AGENTS_SYNC=$MODE (use import, export, or none)" >&2
	exit 1
fi

sync_one() {
	local from_root="$1"
	local to_root="$2"
	local p="$3"
	local from="$from_root/$p"
	local to="$to_root/$p"
	if [[ ! -e "$from" ]]; then
		return 0
	fi
	if [[ -d "$from" ]]; then
		mkdir -p "$to"
		rsync -a "$from/" "$to/"
	else
		mkdir -p "$(dirname "$to")"
		cp -f "$from" "$to"
	fi
}

synced=0
skipped=0
for p in "${AGENT_PATHS[@]}"; do
	if [[ "$MODE" == "import" ]]; then
		from="$PEER/$p"
		sync_one "$PEER" "$ROOT" "$p"
	else
		from="$ROOT/$p"
		sync_one "$ROOT" "$PEER" "$p"
	fi
	if [[ -n "$from" && -e "$from" ]]; then
		((synced++)) || true
	else
		((skipped++)) || true
	fi
done
echo "sync-gitignore-agents: mode=${MODE} peer=${PEER} repo=${ROOT} copied=${synced} missing_source=${skipped}" >&2
