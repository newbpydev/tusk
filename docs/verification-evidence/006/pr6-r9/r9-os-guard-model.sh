#!/usr/bin/env bash
set -euo pipefail
control=$(cd "$(dirname "${BASH_SOURCE[0]}")/r9-os-guard-model" && pwd)
cd "${TUSK_CHECKOUT:?select the checkout}"
output=$(mktemp)
trap 'rm -f "$output"' EXIT
status=0
PATH="$control/dispatch:$PATH" FIXTURE_CI_MUTANT="$control/ci-check-mutant.sh" GOTOOLCHAIN=go1.27.1 make test-scripts >"$output" 2>&1 || status=$?
cat "$output"
if [[ "$status" == 2 ]] &&
    grep -Fq 'FAIL: wrong native OS refuses cross-build substitution (expected 1, got 0)' "$output" &&
    grep -Fq 'FAIL: native OS refusal reaches the intended guard (expected 0, got 1)' "$output" &&
    grep -Fq 'PASS: native Windows prerequisite fixture' "$output"; then
    echo 'PASS: canonical suite detects a dropped native OS comparison'
else
    echo "FAIL: dropped native OS comparison escaped the suite (Make exit $status)" >&2
    exit 1
fi
