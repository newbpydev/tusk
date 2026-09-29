#!/usr/bin/env bash
set -euo pipefail
cd /home/newbpydev/Development/Xoomby/tusk
fixture=$(mktemp -d /tmp/tusk-simplify-visible.XXXXXX)
trap 'rm -rf "$fixture"' EXIT
export TUSK_DB_PATH="$fixture/tasks.db" TUSK_TIMEZONE=UTC
root=$(/tmp/tusk-before-simplify add 'Tree rendering: 界 é 👩‍💻 and wrapped titles' --json | python3 -c 'import json,sys;print(json.load(sys.stdin)["id"])')
child=$(/tmp/tusk-before-simplify add 'First child with a long title that must wrap at narrow terminal widths' --parent "$root" --json | python3 -c 'import json,sys;print(json.load(sys.stdin)["id"])')
/tmp/tusk-before-simplify add 'Nested leaf' --parent "$child" --json >/dev/null
/tmp/tusk-before-simplify add 'Second child' --parent "$root" --json >/dev/null
/tmp/tusk-before-simplify add 'Last root' --json >/dev/null
for binary in /tmp/tusk-before-simplify bin/tusk; do
 clear
 printf '$ tusk tree\n'
 "$binary" tree
 printf '\nPress Enter to continue this owned verification.\n'
 read -r
done
