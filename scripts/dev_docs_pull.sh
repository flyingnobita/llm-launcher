#!/usr/bin/env bash
# Update dev-docs submodule and fail loudly if the checkout is empty or broken.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

git submodule sync --recursive
git submodule update --init --remote dev-docs

if [[ ! -e dev-docs/.git ]]; then
	echo "dev-docs-pull: dev-docs/.git missing after submodule update (clone failed?)" >&2
	exit 1
fi

# Count tracked files at HEAD (empty repo or wrong ref surfaces here).
n="$(git -C dev-docs ls-tree -r --name-only HEAD 2>/dev/null | wc -l | tr -d '[:space:]')"
if [[ "${n:-0}" -eq 0 ]]; then
	echo "dev-docs-pull: checkout has zero tracked files at HEAD." >&2
	echo "dev-docs-pull: remote=$(git -C dev-docs remote get-url origin 2>/dev/null || echo '?')" >&2
	echo "dev-docs-pull: HEAD=$(git -C dev-docs rev-parse HEAD 2>/dev/null || echo '?')" >&2
	echo "dev-docs-pull: try: ssh -T git@github.com  (must reach llml-internal); then:" >&2
	echo "  git submodule deinit -f dev-docs && rm -rf dev-docs && mise run dev-docs-pull" >&2
	exit 1
fi

echo "dev-docs-pull: OK (${n} paths at HEAD)" >&2
