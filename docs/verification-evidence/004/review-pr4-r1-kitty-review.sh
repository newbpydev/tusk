#!/usr/bin/env bash
set -eu
export HOME=/tmp/compound-engineering-1000/ce-babysit-pr/github.com-newbpydev-tusk-4/kitty-home
export TUSK_DB_PATH="$HOME/review.db" TUSK_TIMEZONE=UTC TERM=xterm-kitty
unset NO_COLOR
cd /home/newbpydev/Development/Xoomby/tusk
printf 'PR #4 — terminal review\n\n'
bin/tusk add $'Directional marks: A\u061cB\u200eC\u200fD\ufeffE' --json > "$HOME/task.json"
TASK_ID=$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["id"])' "$HOME/task.json")
bin/tusk list --all
printf '\nInvalid configuration (value remains private):\n'
TUSK_TIMEZONE=PRIVATE bin/tusk list || true
printf '\nRedirected prompt: safe recovery hint\n'
bin/tusk delete "$TASK_ID" 2> "$HOME/error" || true
cat "$HOME/error"
printf '\nInteractive decline: type n then Enter\n'
bin/tusk delete "$TASK_ID"
printf '\nRemaining task after decline:\n'
TERM= bin/tusk list --all
printf '\nReview complete — waiting for owned-window cleanup.\n'
read -r _
