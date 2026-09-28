#!/usr/bin/env bash
set -uo pipefail
cd /home/newbpydev/Development/Xoomby/tusk || exit 1
export GOCACHE=/tmp/tusk-004-go-cache
printf 'Tusk U1 verification in Kitty\n'
printf 'Terminal: %s; stdin/stdout/stderr are terminals: ' "${TERM-}"
if [[ -t 0 && -t 1 && -t 2 ]]; then printf 'yes\n'; else printf 'no\n'; exit 1; fi
make validate build test-cli 2>&1 | tee /tmp/tusk-004-u1-kitty-gates.log
status=${PIPESTATUS[0]}
printf '%s\n' "$status" > /tmp/tusk-004-u1-kitty-gates.exit
if ((status)); then exit "$status"; fi
printf '\n--- Help ---\n'
bin/tusk --help
printf '\n--- Version ---\n'
bin/tusk --version
printf '\n--- Unknown command (expected exit 2) ---\n'
bin/tusk unknown-command
status=$?
printf 'Observed exit: %s\n' "$status"
[[ $status == 2 ]]
