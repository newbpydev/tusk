#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
printf abc >"$scratch/spaced ü input"
expected=ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
[[ "$(bash "$root/scripts/test/sha256.sh" "$scratch/spaced ü input")" == "$expected" ]]
# Exercise the fallback dispatch without requiring Perl on a GNU-only host.
mkdir "$scratch/bin"
if command -v shasum >/dev/null 2>&1;then
 ln -s "$(command -v shasum)" "$scratch/bin/shasum"
else
 cat >"$scratch/bin/shasum" <<'SHASUM'
#!/bin/bash
[[ "$#" == 3 && "$1" == -a && "$2" == 256 && "$(<"$3")" == abc ]] || exit 1
printf 'ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad  %s\n' "$3"
SHASUM
 chmod +x "$scratch/bin/shasum"
fi
ln -s "$(command -v awk)" "$scratch/bin/awk"
[[ "$(PATH="$scratch/bin" "$(command -v bash)" "$root/scripts/test/sha256.sh" "$scratch/spaced ü input")" == "$expected" ]]
if bash "$root/scripts/test/sha256.sh" "$scratch/absent" >/dev/null 2>&1;then
 printf 'FAIL: hash helper accepted absent input\n' >&2;exit 1
fi
printf 'PASS: ordinary and shasum-only fixture hashes preserve spaced Unicode paths and refuse missing input\n'
