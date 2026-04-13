#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -eq 0 ]; then
  exit 0
fi

md_files=()
prettier_files=()
restage_files=()
files_to_stage=()

for file in "$@"; do
  [ -f "$file" ] || continue

  case "$file" in
    *.md)
      md_files+=("$file")
      prettier_files+=("$file")
      restage_files+=("$file")
      ;;
    *.yaml|*.yml|*.json|.prettierrc)
      prettier_files+=("$file")
      restage_files+=("$file")
      ;;
  esac
done

if [ "${#prettier_files[@]}" -gt 0 ]; then
  npx prettier --write "${prettier_files[@]}"
fi

if [ "${#md_files[@]}" -gt 0 ]; then
  npx markdownlint-cli2 --fix "${md_files[@]}"
fi

if [ "${#restage_files[@]}" -gt 0 ]; then
  for file in "${restage_files[@]}"; do
    if git diff --cached --name-only -- "$file" | grep -Fqx "$file"; then
      files_to_stage+=("$file")
    fi
  done
fi

if [ "${#files_to_stage[@]}" -gt 0 ]; then
  git add -- "${files_to_stage[@]}"
fi
