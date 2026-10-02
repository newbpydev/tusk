#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
failed=0
expect() {
    local want=$1 name=$2 actual=0
    shift 2
    "$@" >"$scratch/output" 2>&1 || actual=$?
    if [[ "$actual" == "$want" ]]; then printf 'PASS: %s\n' "$name"
    else
        printf 'FAIL: %s (expected %s, got %s)\n' "$name" "$want" "$actual" >&2
        cat "$scratch/output" >&2
        failed=1
    fi
}
# Even rejected input must never be re-parsed by Make's recipe shell.
# Intentionally literal attack strings.
# shellcheck disable=SC2016
for injection in 'v0.3.0`touch '"$scratch"'/injected`' 'v0.3.0$(shell touch '"$scratch"'/injected)'; do
    make -C "$root" release-candidate RELEASE_VERSION="$injection" RELEASE_OUTPUT="$scratch/output-dir" >"$scratch/injection-output" 2>&1 || true
    if [[ -e "$scratch/injected" ]]; then
        echo 'FAIL: release version executes code before validation' >&2
        failed=1
        rm "$scratch/injected"
    else
        echo 'PASS: release version cannot execute shell/Make code'
    fi
done
expect 0 'local release configuration contract' bash "$root/scripts/release.sh" contract
mkdir -p "$scratch/config/scripts"
cp "$root/scripts/release.sh" "$scratch/config/scripts/"
for mutation in '.builds[0].hooks = {pre:["touch outside"]}' '.archives[0].files += [{src:".env"}]' '.announce = {slack:{enabled:true}}' '.builds += [.builds[0]]' '.checksum.disable = false'; do
    jq "$mutation" "$root/.goreleaser.yaml" >"$scratch/config/.goreleaser.yaml"
    expect 1 "reject packager mutation: $mutation" bash "$scratch/config/scripts/release.sh" contract
done
# Intentionally literal attack strings.
# shellcheck disable=SC2016
for version in '' v1 v01.2.3 'v1.2.3;exit' 'v1.2.3$(touch marker)' 'v1.2.3";panic()'; do
    expect 1 "reject version: $version" env RELEASE_VERSION="$version" RELEASE_OUTPUT="$scratch/unused" bash "$root/scripts/release.sh" candidate
done
mkdir -p "$scratch/bin" "$scratch/tools"
printf 'accepted-tool\n' >"$scratch/tools/goreleaser"
cat >"$scratch/bin/curl" <<'EOF'
#!/usr/bin/env bash
while [[ $# -gt 0 ]]; do
    if [[ "$1" == --output ]]; then printf 'corrupt\n' >"$2"; exit 0; fi
    shift
done
exit 19
EOF
chmod +x "$scratch/bin/curl"
expect 1 'bad packager digest rejected before execution' env PATH="$scratch/bin:$PATH" TUSK_RELEASE_TOOL_DIR="$scratch/tools" bash "$root/scripts/release.sh" setup
if [[ "$(cat "$scratch/tools/goreleaser")" != accepted-tool ]]; then
    echo 'FAIL: failed download replaced accepted packager' >&2
    failed=1
fi
# Failure injection uses a separate repository and preserves dirty/data files.
mkdir -p "$scratch/checkout/scripts" "$scratch/fail-bin"
cp "$root/scripts/release.sh" "$scratch/checkout/scripts/"
cp "$root/scripts/tool-versions.json" "$scratch/checkout/scripts/"
git -C "$scratch/checkout" init --quiet
git -C "$scratch/checkout" add scripts
git -C "$scratch/checkout" -c user.name=Fixture -c user.email=fixture@example.invalid commit --quiet -m fixture
fixture_sha=$(git -C "$scratch/checkout" rev-parse HEAD)
printf 'user data\n' >"$scratch/checkout/user.db"
printf 'wal\n' >"$scratch/checkout/user.db-wal"
printf 'shm\n' >"$scratch/checkout/user.db-shm"
printf 'unrelated\n' >"$scratch/checkout/unrelated.txt"
git -C "$scratch/checkout" status --porcelain >"$scratch/before"
cat >"$scratch/fail-bin/make" <<'EOF'
#!/usr/bin/env bash
echo 'intentional child gate failure' >&2
exit 55
EOF
chmod +x "$scratch/fail-bin/make"
expect 55 'failed child gate keeps diagnostics' env PATH="$scratch/fail-bin:$PATH" TUSK_RELEASE_TOOL_DIR="$root/bin/tools" RELEASE_VERSION=v0.3.0 RELEASE_SHA="$fixture_sha" RELEASE_OUTPUT="$scratch/failed-run" bash "$scratch/checkout/scripts/release.sh" candidate
if [[ ! -s "$scratch/failed-run/build.log" || ! -s "$scratch/failed-run/result.json" ]]; then
    echo 'FAIL: failed packaging lost diagnostics' >&2
    failed=1
fi
git -C "$scratch/checkout" status --porcelain >"$scratch/after"
if ! cmp -s "$scratch/before" "$scratch/after" || [[ "$(cat "$scratch/checkout/user.db")" != 'user data' ]]; then
    echo 'FAIL: packaging changed checkout/user data' >&2
    failed=1
fi
expect 1 'retained output collision refused' env RELEASE_VERSION=v0.3.0 RELEASE_SHA="$fixture_sha" RELEASE_OUTPUT="$scratch/failed-run" bash "$scratch/checkout/scripts/release.sh" candidate
git -C "$scratch/checkout" tag v0.3.0 "$fixture_sha"
printf 'second\n' >"$scratch/checkout/second"
git -C "$scratch/checkout" add second
git -C "$scratch/checkout" -c user.name=Fixture -c user.email=fixture@example.invalid commit --quiet -m second
expect 1 'conflicting local tag refused before output' env RELEASE_VERSION=v0.3.0 RELEASE_SHA=HEAD RELEASE_OUTPUT="$scratch/tag-conflict" bash "$scratch/checkout/scripts/release.sh" candidate
expect 1 'unsafe SHA data refused' env RELEASE_VERSION=v0.3.0 RELEASE_SHA='HEAD;echo' RELEASE_OUTPUT="$scratch/unsafe" bash "$scratch/checkout/scripts/release.sh" candidate
exit "$failed"
