#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

echo "==> Formatting Go files with gofmt -s..."

# find -exec preserves filenames on GNU, BSD and MSYS without word splitting.
find . -name '*.go' -not -path './docs/research/legacy-audit/snapshots/*' \
    -not -path './vendor/*' -exec gofmt -s -w {} +
echo "==> Successfully formatted Go files."

exit 0
