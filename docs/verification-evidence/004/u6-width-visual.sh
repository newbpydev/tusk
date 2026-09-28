#!/usr/bin/env bash
set -euo pipefail
cd /home/newbpydev/Development/Xoomby/tusk
fixture=$(mktemp -d /tmp/tusk-u6-visual.XXXXXX)
trap 'rm -rf "$fixture"' EXIT
export TUSK_DB_PATH="$fixture/tasks.db" TUSK_TIMEZONE=UTC
unset TUSK_AUTO_COMPLETE_PARENT NO_COLOR
root=$(bin/tusk add $'Unicode 界 👩‍💻 é; control \e[31m and bidi \u202e' --priority high --json | python3 -c 'import json,sys;print(json.load(sys.stdin)["id"])')
child=$(bin/tusk add 'Child checklist' --parent "$root" --json | python3 -c 'import json,sys;print(json.load(sys.stdin)["id"])')
bin/tusk done "$child" >/dev/null
for mode in ${1:-color no-color dumb}; do
 clear
 printf 'U6 Kitty width %s / %s\n\n' "$(tput cols)" "$mode"
 case "$mode" in no-color) export NO_COLOR=1;; dumb) unset NO_COLOR; export TERM=dumb;; esac
 bin/tusk list --all
 printf '\nTree:\n'
 bin/tusk tree
 printf '\nPress Enter for next mode.\n'
 read -r
done
printf 'Shell input remains usable; Enter to close.\n'
read -r
