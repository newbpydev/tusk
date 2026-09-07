#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

echo "==> Formatting Go files with gofmt -s..."

# Find all go files excluding docs/research/legacy-audit/snapshots and vendor
GO_FILES=$(find . -name "*.go" -not -path "./docs/research/legacy-audit/snapshots/*" -not -path "./vendor/*" || true)

if [ -n "${GO_FILES}" ]; then
    # Run gofmt -s -w
    echo "${GO_FILES}" | xargs gofmt -s -w
    echo "==> Successfully formatted Go files."
else
    echo "==> No active Go files found to format."
fi

exit 0
