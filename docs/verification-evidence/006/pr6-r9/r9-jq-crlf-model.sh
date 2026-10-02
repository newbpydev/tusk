#!/usr/bin/env bash
set -euo pipefail
control=$(cd "$(dirname "${BASH_SOURCE[0]}")/r9-jq-crlf-model" && pwd)
real_jq=$(command -v jq)
cd "${TUSK_CHECKOUT:?select the checkout}"
PATH="$control/bin:$PATH" MODEL_REAL_JQ="$real_jq" GOTOOLCHAIN=go1.27.1 make test-scripts
