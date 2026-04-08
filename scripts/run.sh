#!/usr/bin/env bash

set -euo pipefail

contract=""
output_dir="output"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --contract)
      contract="${2:-}"
      shift 2
      ;;
    --output-dir)
      output_dir="${2:-}"
      shift 2
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 2
      ;;
  esac
done

if [[ -z "$contract" ]]; then
  echo "run.sh requires --contract <path>" >&2
  exit 2
fi

go run ./cmd/service-lasso-harness run --contract "$contract" --output-dir "$output_dir"
