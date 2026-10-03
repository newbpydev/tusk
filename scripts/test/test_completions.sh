#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
scratch=$(mktemp -d)
cleanup() {
    local status=$?
    if [[ "$status" != 0 ]]; then
        echo 'Completion shell acceptance failed; retained diagnostics:' >&2
        for log in "$scratch/"*diagnostics "$scratch/"*output; do
            if [[ -f "$log" ]]; then cat "$log" >&2; fi
        done
    fi
    rm -rf "$scratch"
}
trap cleanup EXIT
for dependency in bash zsh fish; do
    command -v "$dependency" >/dev/null || { echo "Completion acceptance requires $dependency" >&2; exit 1; }
done
bash_framework="${BASH_COMPLETION_SOURCE:-/usr/share/bash-completion/bash_completion}"
[[ -f "$bash_framework" ]] || { echo 'Set BASH_COMPLETION_SOURCE to the installed Bash completion framework' >&2; exit 1; }
binary="${TUSK_COMPLETION_BINARY:-$root/bin/tusk}"
[[ -x "$binary" ]] || { echo 'Run make build before completion checks' >&2; exit 1; }
mkdir -p "$scratch/bin" "$scratch/home" "$scratch/zsh"
cp "$binary" "$scratch/bin/tusk"
touch "$scratch/blocked"
export PATH="$scratch/bin:$PATH" HOME="$scratch/home" ZDOTDIR="$scratch/zsh"
export XDG_DATA_HOME="$scratch/absent-data" XDG_CONFIG_HOME="$scratch/absent-config"
export TUSK_DB_PATH="$scratch/blocked/never.db" TUSK_TIMEZONE=invalid TUSK_AUTO_COMPLETE_PARENT=invalid
for shell in bash zsh fish; do
    tusk completion "$shell" >"$scratch/tusk.$shell" 2>"$scratch/stderr"
    [[ ! -s "$scratch/stderr" ]]
    cmp "$scratch/tusk.$shell" "$root/docs/completions/tusk.$shell"
    "$shell" -n "$scratch/tusk.$shell"
done

bash --noprofile --norc -c '
    source "$1"
    source "$2"
    COMP_WORDS=(tusk list --status "")
    COMP_CWORD=3 COMP_LINE="tusk list --status " COMP_POINT=19
    __start_tusk
    printf "%s\n" "${COMPREPLY[@]}"
' _ "$bash_framework" "$scratch/tusk.bash" >"$scratch/bash-output" 2>"$scratch/bash-diagnostics"
grep -q '^in-progress' "$scratch/bash-output"

zsh -f -c '
    autoload -Uz compinit
    compinit -D
    source "$1"
    functions _tusk >/dev/null
    tusk __complete list --status ""
' _ "$scratch/tusk.zsh" >"$scratch/zsh-output" 2>"$scratch/zsh-diagnostics"
grep -q '^in-progress' "$scratch/zsh-output"

# Fish, not Bash, expands argv inside this literal program.
# shellcheck disable=SC2016
fish --no-config --private -c '
    source "$argv[1]"
    complete -C "tusk add --priority "
' "$scratch/tusk.fish" >"$scratch/fish-output" 2>"$scratch/fish-diagnostics"
grep -q '^urgent' "$scratch/fish-output"
# Fish can create its own XDG state while completing. Only Tusk storage is ours.
[[ ! -e "$XDG_DATA_HOME/tusk" ]]
[[ ! -e "$TUSK_DB_PATH" ]]
printf 'Real shell syntax/load and Bash/Fish completion plus Zsh protocol checks passed.\n'
printf 'Owned Kitty completion-widget inspection is a separate required receipt.\n'
