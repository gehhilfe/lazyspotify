#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

command -v vhs >/dev/null 2>&1 || {
  echo "vhs is required to render demo tapes" >&2
  exit 1
}

mkdir -p docs/assets/demos

vhs validate 'docs/vhs/*.tape'

for tape in docs/vhs/*.tape; do
  vhs "$tape"
done
