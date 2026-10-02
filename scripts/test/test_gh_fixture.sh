#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
binary="$scratch/gh"
shell=$(command -v bash)
script="$scratch/fixture script ü"
if command -v cygpath >/dev/null 2>&1; then
    binary="$scratch/gh.exe"
    shell=$(cygpath -m "$shell")
    script=$(cygpath -m "$script")
fi
GH_FIXTURE_OUTPUT="$binary" make --no-print-directory -C "$root" build-gh-fixture
cat >"$script" <<'GH'
#!/usr/bin/env bash
printf '%s\n' "$@"
cat
printf 'owned stderr\n' >&2
exit 47
GH
# These arguments must remain data through native process dispatch.
# shellcheck disable=SC2016
args=('argument with spaces' 'literal$(touch injected)`touch injected`' 'ü')
printf '%s\n' "${args[@]}" 'owned stdin' >"$scratch/want"
status=0
printf 'owned stdin\n' | FIXTURE_GH_BASH="$shell" FIXTURE_GH_SCRIPT="$script" "$binary" "${args[@]}" >"$scratch/stdout" 2>"$scratch/stderr" || status=$?
failed=0
check() {
    local label=$1
    shift
    if "$@"; then printf 'PASS: %s\n' "$label"
    else
        printf 'FAIL: %s\n' "$label" >&2
        cat "$scratch/stdout" "$scratch/stderr" >&2
        failed=1
    fi
}
check 'native fixture preserves exit 47' test "$status" -eq 47
check 'native fixture preserves exact stdout and stdin bytes' cmp "$scratch/want" "$scratch/stdout"
check 'native fixture preserves stderr' grep -Fxq 'owned stderr' "$scratch/stderr"
case "${OSTYPE:-}" in
    msys*|cygwin*) ;;
    *)
        printf '#!/usr/bin/env bash\nkill -TERM "$$"\n' >"$script"
        status=0
        FIXTURE_GH_BASH="$shell" FIXTURE_GH_SCRIPT="$script" "$binary" >"$scratch/stdout" 2>"$scratch/stderr" || status=$?
        check "native fixture maps SIGTERM to 143 (observed $status)" test "$status" -eq 143
        ;;
esac
status=0
FIXTURE_GH_BASH='' FIXTURE_GH_SCRIPT='' "$binary" >"$scratch/stdout" 2>"$scratch/stderr" || status=$?
check 'native fixture refuses missing ownership with exit 127' test "$status" -eq 127
check 'native fixture explains missing ownership' grep -Fq 'owned gh fixture requires FIXTURE_GH_BASH and FIXTURE_GH_SCRIPT' "$scratch/stderr"
exit "$failed"
