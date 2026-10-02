#!/usr/bin/env bash
set -euo pipefail
control=$(cd "$(dirname "${BASH_SOURCE[0]}")/r8-windows-identity-model" && pwd)
cd "${TUSK_CHECKOUT:?select the checkout}"
PATH="$control/dispatch:$PATH" FIXTURE_NATIVE_BIN="$control/native" GOTOOLCHAIN=go1.27.1 make test-scripts
