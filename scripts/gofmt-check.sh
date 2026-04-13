#!/usr/bin/env bash
# Fail if any Go file needs gofmt. Used by mise `fmt-check`.
set -euo pipefail
mapfile -t files < <(gofmt -l .)
if ((${#files[@]})); then
  printf '%s\n' "${files[@]}"
  echo >&2 "Run: go fmt ./..."
  exit 1
fi
