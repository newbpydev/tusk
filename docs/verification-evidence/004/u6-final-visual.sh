#!/usr/bin/env bash
set -euo pipefail
cd /home/newbpydev/Development/Xoomby/tusk
fixture=$(mktemp -d /tmp/tusk-u6-final.XXXXXX)
trap 'python3 -c "import shutil,sys;shutil.rmtree(sys.argv[1])" "$fixture"' EXIT
export TUSK_DB_PATH="$fixture/tasks.db" TUSK_TIMEZONE=UTC
unset TUSK_AUTO_COMPLETE_PARENT NO_COLOR
clear
printf 'Tusk U6 — actual CLI review fixes in Kitty\n\n'
id=$(bin/tusk add 'Decimal progress: 010 means ten' --json | python3 -c 'import json,sys;print(json.load(sys.stdin)["id"])')
printf '$ tusk edit ID --progress 010\n'
bin/tusk edit "$id" --progress 010
printf '\n$ tusk edit ID --progress 0x10\n'
set +e
bin/tusk edit "$id" --progress 0x10
code=$?
set -e
printf 'Exit: %s (expected 2)\n\nFresh process readback:\n' "$code"
test "$code" = 2
bin/tusk list --all
printf '\nCanonical validate/build/generated checks: PASS\nReference latency: FAIL; U6 and Phase 4 acceptance pending\nPress Enter to close this owned tab.\n'
read -r
