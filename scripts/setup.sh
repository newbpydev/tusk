#!/usr/bin/env bash
set -euo pipefail

echo "==> Running Tusk environment setup & verification..."

GO_CMD="${TUSK_GO_BIN:-go}"

# Check if go command exists
if ! command -v "$GO_CMD" >/dev/null 2>&1; then
    echo "ERROR: '${GO_CMD}' binary not found in PATH. Please install Go 1.24+." >&2
    exit 1
fi

# Detect Go version
GO_VERSION_STRING=$("$GO_CMD" version)
echo "Found Go: ${GO_VERSION_STRING}"

# Extract major.minor
if [[ "$GO_VERSION_STRING" =~ go([0-9]+)\.([0-9]+) ]]; then
    MAJOR="${BASH_REMATCH[1]}"
    MINOR="${BASH_REMATCH[2]}"
    if (( MAJOR < 1 )) || (( MAJOR == 1 && MINOR < 24 )); then
        echo "ERROR: Go version 1.24 or higher is required (found ${MAJOR}.${MINOR})." >&2
        exit 1
    fi
else
    echo "WARNING: Could not parse Go version format. Proceeding with caution." >&2
fi

# Ensure bin directory exists
mkdir -p bin

echo "==> Setup complete. Environment is ready for development."
exit 0
