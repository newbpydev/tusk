#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
failed=0
# These predicates are invoked by expect through its command arguments.
# shellcheck disable=SC2329
is_formatted() { [[ -z "$(gofmt -l "$1")" ]]; }
# shellcheck disable=SC2329
is_accepted_tool() { [[ "$(cat "$1")" == accepted-tool ]]; }
expect() {
    local want="$1" name="$2" actual=0
    shift 2
    "$@" >"$scratch/output" 2>&1 || actual=$?
    if [[ "$actual" == "$want" ]]; then
        printf 'PASS: %s\n' "$name"
    else
        printf 'FAIL: %s (expected %s, got %s)\n' "$name" "$want" "$actual" >&2
        cat "$scratch/output" >&2
        failed=1
    fi
}

fixture="$scratch/tusk ü checkout"
mkdir -p "$fixture/scripts" "$fixture/space ü directory"
cp "$root/scripts/fmt.sh" "$fixture/scripts/fmt.sh"
printf 'package fixture\r\nfunc value( ) int{return 1}\r\n' >"$fixture/space ü directory/fixture file.go"
expect 0 'formatter handles spaced Unicode checkout and CRLF source' bash "$fixture/scripts/fmt.sh"
expect 0 'formatter normalizes the actual spaced file' is_formatted "$fixture/space ü directory/fixture file.go"

expect 0 'CI contract accepts canonical configuration' bash "$root/scripts/ci-check.sh" contract
if [[ ! -f "$root/scripts/tool-versions.json" || ! -f "$root/.github/workflows/ci.yml" ]]; then
    echo 'FAIL: native matrix and immutable tool catalogue are missing' >&2
    exit 1
fi
mkdir -p "$scratch/config/.github/workflows" "$scratch/config/scripts"
cp "$root/scripts/tool-versions.json" "$scratch/config/scripts/"
cp "$root/.github/workflows/ci.yml" "$scratch/config/.github/workflows/"
for mutation in floating write missing-race missing-native wrong-go credentials privileged; do
    case "$mutation" in
        floating) filter='.jobs.native.steps[0].uses = "actions/checkout@main"' ;;
        write) filter='.permissions.contents = "write"' ;;
        missing-race) filter='(.jobs.native.steps[] | select(.name == "Canonical gates").run) = "make test build"' ;;
        missing-native) filter='.jobs.native.strategy.matrix.include |= .[0:4]' ;;
        wrong-go) filter='(.jobs.native.steps[] | select(.uses | strings | startswith("actions/setup-go@")).with["go-version"]) = "1.25.0"' ;;
        credentials) filter='.jobs.native.steps[0].with["persist-credentials"] = true' ;;
        privileged) filter='.on.pull_request_target = {}' ;;
    esac
    jq "$filter" "$root/.github/workflows/ci.yml" >"$scratch/config/.github/workflows/ci.yml"
    expect 1 "CI refuses $mutation" bash "$root/scripts/ci-check.sh" contract "$scratch/config"
done
cp "$root/.github/workflows/ci.yml" "$scratch/config/.github/workflows/ci.yml"
for filter in '.tools.actionlint.assets.linux_amd64.sha256 = "bad"' '.actions.checkout.sha = "main"' 'del(.go.release)' '.tools.goreleaser.version = "latest"'; do
    jq "$filter" "$root/scripts/tool-versions.json" >"$scratch/config/scripts/tool-versions.json"
    expect 1 "CI refuses invalid tool pin: $filter" bash "$root/scripts/ci-check.sh" contract "$scratch/config"
done

# Prerequisite proof uses a fake native identity, never claims native execution.
mkdir -p "$scratch/bin"
cat >"$scratch/bin/go" <<'EOF'
#!/usr/bin/env bash
case "$1" in
    version) echo 'go version go1.27.1 windows/amd64' ;;
    env)
        printf '%s\n' windows amd64 go1.27.1 1
        if [[ "$*" == *GOHOSTOS* ]]; then printf '%s\n' "${FIXTURE_GO_HOST:-windows}" amd64; fi
        ;;
esac
EOF
cat >"$scratch/bin/gcc" <<'EOF'
#!/usr/bin/env bash
if [[ "$1" == -dumpmachine ]]; then echo x86_64-w64-mingw32; else echo 'gcc fixture'; fi
EOF
cat >"$scratch/bin/uname" <<'EOF'
#!/usr/bin/env bash
echo "${FIXTURE_UNAME:-MINGW64_NT-10.0}"
EOF
chmod +x "$scratch/bin/"*
expect 0 'native Windows prerequisite fixture' env PATH="$scratch/bin:$PATH" TUSK_CI_OS=windows TUSK_CI_ARCH=amd64 TUSK_CI_GO=1.27.1 bash "$root/scripts/ci-check.sh" preflight
expect 1 'cross compiler cannot impersonate a native Windows host' env PATH="$scratch/bin:$PATH" FIXTURE_GO_HOST=linux TUSK_CI_OS=windows TUSK_CI_ARCH=amd64 TUSK_CI_GO=1.27.1 bash "$root/scripts/ci-check.sh" preflight
expect 1 'WSL shell cannot close native Windows acceptance' env PATH="$scratch/bin:$PATH" FIXTURE_UNAME=Linux TUSK_CI_OS=windows TUSK_CI_ARCH=amd64 TUSK_CI_GO=1.27.1 bash "$root/scripts/ci-check.sh" preflight
expect 1 'missing Make is actionable' env TUSK_CI_MAKE=missing-tusk-make bash "$root/scripts/ci-check.sh" preflight
expect 1 'missing native compiler is actionable' env TUSK_CI_CC=missing-tusk-cc bash "$root/scripts/ci-check.sh" preflight
expect 1 'wrong native identity refuses cross-build substitution' env TUSK_CI_OS=windows TUSK_CI_ARCH=amd64 bash "$root/scripts/ci-check.sh" preflight

# A digest failure or failed download must preserve the previously accepted tool.
mkdir -p "$scratch/install/bin/tools"
printf 'accepted-tool\n' >"$scratch/install/bin/tools/actionlint"
cat >"$scratch/bin/curl" <<'EOF'
#!/usr/bin/env bash
while [[ $# -gt 0 ]]; do
    if [[ "$1" == --output ]]; then printf 'truncated\n' >"$2"; exit 0; fi
    shift
done
exit 19
EOF
chmod +x "$scratch/bin/curl"
expect 1 'corrupt tool download is rejected' env PATH="$scratch/bin:$PATH" TUSK_CI_TOOL_DIR="$scratch/install/bin/tools" bash "$root/scripts/ci-check.sh" setup
expect 0 'corrupt download preserves accepted executable' is_accepted_tool "$scratch/install/bin/tools/actionlint"
printf '#!/usr/bin/env bash\nexit 19\n' >"$scratch/bin/curl"
expect 1 'failed tool download propagates' env PATH="$scratch/bin:$PATH" TUSK_CI_TOOL_DIR="$scratch/install/bin/tools" bash "$root/scripts/ci-check.sh" setup

mkdir -p "$scratch/drift"
git -C "$scratch/drift" init -q
mkdir -p "$scratch/drift/scripts"
cp "$root/scripts/fmt.sh" "$scratch/drift/scripts/fmt.sh"
printf 'package fixture\nfunc value( )int{return 1}\n' >"$scratch/drift/source.go"
git -C "$scratch/drift" add source.go scripts/fmt.sh
git -C "$scratch/drift" -c user.name=Fixture -c user.email=fixture@example.invalid commit -qm fixture
expect 0 'source gate accepts clean tracked source' bash "$root/scripts/ci-check.sh" drift "$scratch/drift"
expect 0 'formatter repairs tracked negative fixture' bash "$scratch/drift/scripts/fmt.sh"
expect 1 'source gate refuses silent tracked repairs' bash "$root/scripts/ci-check.sh" drift "$scratch/drift"
git -C "$scratch/drift" checkout -- source.go
printf 'unexpected\n' >"$scratch/drift/unexpected.go"
expect 1 'source gate refuses extra untracked generated output' bash "$root/scripts/ci-check.sh" drift "$scratch/drift"

for gate in fmt vet test race coverage test-scripts check-modules; do
    mkdir -p "$scratch/gates/bin"
    cat >"$scratch/gates/bin/go" <<'EOF'
#!/usr/bin/env bash
exit 19
EOF
    chmod +x "$scratch/gates/bin/go"
    # Existing targets already propagate failures; a dry run also checks presence.
    expect 0 "canonical gate remains registered: $gate" make --no-print-directory -n -C "$root" "$gate"
    cp "$root/Makefile" "$scratch/gates/Makefile"
    for other in fmt vet test race coverage test-scripts check-modules; do
        printf '\n%s:\n\t@echo gate-%s\n\t@exit %s\n' "$other" "$other" "$([[ "$gate" == "$other" ]] && echo 19 || echo 0)" >>"$scratch/gates/Makefile"
    done
    expect 2 "aggregate propagates failed $gate" make --no-print-directory -C "$scratch/gates" validate
done
expect 0 'pinned analyzer setup is registered' make --no-print-directory -n -C "$root" setup-vulnerabilities
exit "$failed"
