#!/usr/bin/env bash
set -euo pipefail
cd /home/newbpydev/Development/Xoomby/tusk
fixture=$(mktemp -d /tmp/tusk-u6-visible.XXXXXX)
trap 'rm -rf "$fixture"' EXIT
export TUSK_DB_PATH="$fixture/tasks.db" TUSK_TIMEZONE=UTC TUSK_AUTO_COMPLETE_PARENT=false
unset NO_COLOR
clear
printf 'Tusk U6 — retained terminal verification\n\n'
root=$(bin/tusk add 'Verify tree snapshot' --json | python3 -c 'import json,sys;print(json.load(sys.stdin)["id"])')
child=$(bin/tusk add 'Verify child progress' --parent "$root" --json | python3 -c 'import json,sys;print(json.load(sys.stdin)["id"])')
bin/tusk edit "$child" --progress 010 > /dev/null
printf '$ tusk tree\n'
bin/tusk tree
printf '\n$ tusk done CHILD\n'
bin/tusk done "$child"
printf '\n$ tusk tree\n'
bin/tusk tree
printf '\n$ tusk stats\n'
bin/tusk stats
printf '\nAutomated validation: passed in Codex Bash\nReference latency: unresolved; U6 remains open\nPress Enter to close this owned verification tab.\n'
read -r
