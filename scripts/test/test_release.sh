#!/usr/bin/env bash
set -euo pipefail
jq_binary_option=''
case "${OSTYPE:-}" in msys*|cygwin*) jq_binary_option=--binary ;; esac
export jq_binary_option
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
mkdir -p "$scratch/compiler/scripts" "$scratch/compiler-bin"
cp "$root/Makefile" "$scratch/compiler/"
jq ${jq_binary_option:+"--binary"} '.go.release="1.27.2"' "$root/scripts/tool-versions.json" >"$scratch/compiler/scripts/tool-versions.json"
cat >"$scratch/compiler-bin/go" <<'GO'
#!/usr/bin/env bash
[[ "$GOTOOLCHAIN" == go1.27.2 && "$CGO_ENABLED" == 0 ]] || exit 62
[[ "$1" == build ]] || exit 63
GO
chmod +x "$scratch/compiler-bin/go"
expect 0 'fixture compiler follows the release catalog' env PATH="$scratch/compiler-bin:$PATH" TUSK_RELEASE_OVERLAY=fixture.json TUSK_RELEASE_FIXTURE_BINARY=fixture-binary make -C "$scratch/compiler" build-release-fixture
mkdir -p "$scratch/config/scripts"
cp "$root/scripts/release.sh" "$scratch/config/scripts/"
for mutation in '.builds[0].hooks = {pre:["touch outside"]}' '.archives[0].files += [{src:".env"}]' '.announce = {slack:{enabled:true}}' '.builds += [.builds[0]]' '.checksum.disable = false'; do
    jq ${jq_binary_option:+"--binary"} "$mutation" "$root/.goreleaser.yaml" >"$scratch/config/.goreleaser.yaml"
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
# A partial archive copy must never overwrite the last accepted archive.
mkdir -p "$scratch/atomic/scripts" "$scratch/atomic-bin" "$scratch/atomic-tools" "$scratch/packager"
cp "$root/scripts/release.sh" "$scratch/atomic/scripts/"
printf '#!/usr/bin/env bash\nexit 0\n' >"$scratch/packager/goreleaser"
tar -czf "$scratch/packager.tar.gz" -C "$scratch/packager" goreleaser
archive_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/packager.tar.gz")
jq ${jq_binary_option:+"--binary"} --arg hash "$archive_hash" '.tools.goreleaser.assets.linux_amd64.sha256=$hash' "$root/scripts/tool-versions.json" >"$scratch/atomic/scripts/tool-versions.json"
printf 'accepted archive\n' >"$scratch/atomic-tools/goreleaser.archive"
cat >"$scratch/atomic-bin/go" <<'GO'
#!/usr/bin/env bash
case "$*" in 'env GOHOSTOS') echo linux;;'env GOHOSTARCH') echo amd64;;*) exit 77;;esac
GO
cat >"$scratch/atomic-bin/curl" <<'CURL'
#!/usr/bin/env bash
while [[ $# -gt 0 ]]; do
 if [[ "$1" == --output ]]; then exec "$FIXTURE_REAL_CP" "$FIXTURE_ARCHIVE" "$2"; fi
 shift
done
exit 77
CURL
cat >"$scratch/atomic-bin/cp" <<'COPY'
#!/usr/bin/env bash
case "${*: -1}" in */goreleaser.archive|*/.goreleaser-archive.*) printf partial >"${*: -1}"; exit 55;;esac
exec "$FIXTURE_REAL_CP" "$@"
COPY
chmod +x "$scratch/atomic-bin/"*
expect 55 'partial archive staging copy propagates' env PATH="$scratch/atomic-bin:$PATH" FIXTURE_REAL_CP="$(command -v cp)" FIXTURE_ARCHIVE="$scratch/packager.tar.gz" TUSK_RELEASE_TOOL_DIR="$scratch/atomic-tools" bash "$scratch/atomic/scripts/release.sh" setup
if [[ "$(cat "$scratch/atomic-tools/goreleaser.archive")" != 'accepted archive' ]]; then echo 'FAIL: partial copy overwrote accepted archive' >&2; failed=1; fi
if [[ -n "$(find "$scratch/atomic-tools" -name '.goreleaser*' -print)" ]]; then echo 'FAIL: incomplete packager stage leaked' >&2; failed=1; fi
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
