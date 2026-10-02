#!/usr/bin/env bash
set -euo pipefail
jq_binary_option=''
case "${OSTYPE:-}" in msys*|cygwin*) jq_binary_option=--binary ;; esac
export jq_binary_option
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
fixture="$scratch/checkout ü"
mkdir -p "$fixture/scripts" "$fixture/docs/assets"
cp "$root/go.mod" "$root/go.sum" "$root/LICENSE" "$root/THIRD_PARTY_NOTICES.md" "$fixture/"
cp "$root/scripts/notices.json" "$fixture/scripts/"
cp "$root/docs/assets/tusk-tui.png" "$fixture/docs/assets/"
cp -R "$root/third_party" "$fixture/"
failed=0
# shellcheck disable=SC2329
readable_notice() { [[ -n "$(find "$1" -prune -perm -0644 -print)" ]]; }
expect() {
    local want=$1 name=$2 actual=0
    shift 2
    "$@" >"$scratch/result" 2>&1 || actual=$?
    if [[ "$actual" == "$want" ]]; then printf 'PASS: %s\n' "$name"
    else
        printf 'FAIL: %s (expected %s, got %s)\n' "$name" "$want" "$actual" >&2
        cat "$scratch/result" >&2
        failed=1
    fi
}
expect 0 'accepted full inventory' bash "$root/scripts/notices.sh" check "$fixture"
# Native jq.exe translates LF unless binary output is explicitly selected.
mkdir -p "$scratch/crlf-bin"
cat >"$scratch/crlf-bin/jq" <<'JQ'
#!/usr/bin/env bash
set -euo pipefail
binary=false
args=()
for arg in "$@"; do
    case "$arg" in --binary|-b) binary=true ;; *) args+=("$arg");; esac
done
output=$(mktemp)
trap 'rm -f "$output"' EXIT
status=0
if "$FIXTURE_REAL_JQ" --binary --null-input empty >/dev/null 2>&1; then
    "$FIXTURE_REAL_JQ" --binary "${args[@]}" >"$output" || status=$?
else
    "$FIXTURE_REAL_JQ" "${args[@]}" >"$output" || status=$?
fi
if "$binary"; then cat "$output"; else sed 's/$/\r/' "$output"; fi
exit "$status"
JQ
chmod +x "$scratch/crlf-bin/jq"
printf 'export OSTYPE=msys\n' >"$scratch/windows-json-env"
expect 0 'Windows JSON output preserves the full notice inventory' env BASH_ENV="$scratch/windows-json-env" PATH="$scratch/crlf-bin:$PATH" FIXTURE_REAL_JQ="$(command -v jq)" bash "$root/scripts/notices.sh" check "$fixture"
for mutation in unclassified missing-module wrong-version traversal missing-replacement missing-asset missing-asset-entry missing-sqlite-notice missing-go-notice; do
    cp "$root/scripts/notices.json" "$fixture/scripts/notices.json"
    case "$mutation" in
        unclassified) filter='.modules[0].licenses[0].classification="UNKNOWN"' ;;
        missing-module) filter='.modules |= .[1:]' ;;
        wrong-version) filter='.modules[0].version="v9.9.9"' ;;
        traversal) filter='.modules[0].licenses[0].file="../LICENSE"' ;;
        missing-replacement) filter='(.modules[]|select(.replacement!="")|.licenses[0].sha256)="0000000000000000000000000000000000000000000000000000000000000000"' ;;
        missing-asset) filter='.assets[0].sha256="0000000000000000000000000000000000000000000000000000000000000000"' ;;
        missing-asset-entry) filter='.assets |= .[1:]' ;;
        missing-sqlite-notice) filter='(.modules[]|select(.path=="modernc.org/sqlite")|.licenses) |= map(select(.file!="LICENSE-SQLITE"))' ;;
        missing-go-notice) filter='.extras |= .[1:]' ;;
    esac
    jq ${jq_binary_option:+"--binary"} "$filter" "$root/scripts/notices.json" >"$fixture/scripts/notices.json"
    cp "$root/THIRD_PARTY_NOTICES.md" "$fixture/"
    expect 1 "distribution refuses $mutation" bash "$root/scripts/notices.sh" generate "$fixture"
    expect 0 "failed $mutation preserves accepted notices" cmp "$root/THIRD_PARTY_NOTICES.md" "$fixture/THIRD_PARTY_NOTICES.md"
done
cp "$root/scripts/notices.json" "$fixture/scripts/notices.json"
mv "$fixture/LICENSE" "$scratch/LICENSE"
expect 1 'distribution needs first-party grant' bash "$root/scripts/notices.sh" check "$fixture"
mv "$scratch/LICENSE" "$fixture/LICENSE"
printf '\nstale\n' >>"$fixture/THIRD_PARTY_NOTICES.md"
expect 1 'notice drift fails' bash "$root/scripts/notices.sh" check "$fixture"
expect 0 'regenerate repairs owned notices' bash "$root/scripts/notices.sh" generate "$fixture"
expect 0 'repaired notices match' cmp "$root/THIRD_PARTY_NOTICES.md" "$fixture/THIRD_PARTY_NOTICES.md"
expect 0 'generated notices have portable readable permissions' readable_notice "$fixture/THIRD_PARTY_NOTICES.md"
mkdir -p "$scratch/fail-bin"
cat >"$scratch/fail-bin/chmod" <<'CHMOD'
#!/usr/bin/env bash
case "$2" in */.notices.*) exit 55;;esac
exec "$FIXTURE_REAL_CHMOD" "$@"
CHMOD
chmod +x "$scratch/fail-bin/chmod"
expect 55 'failed notices stage preserves the previous notice' env PATH="$scratch/fail-bin:$PATH" FIXTURE_REAL_CHMOD="$(command -v chmod)" bash "$root/scripts/notices.sh" generate "$fixture"
expect 0 'failed stage leaves accepted bytes intact' cmp "$root/THIRD_PARTY_NOTICES.md" "$fixture/THIRD_PARTY_NOTICES.md"
if [[ -n "$(find "$fixture" -maxdepth 1 -name '.notices.*' -print)" ]]; then echo 'FAIL: incomplete notices stage leaked' >&2; failed=1; fi
exit "$failed"
