#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
failed=0
expect() {
    local want=$1 name=$2 actual=0
    shift 2
    "$@" >"$scratch/result" 2>&1 || actual=$?
    if [[ "$actual" == "$want" ]]; then
        printf 'PASS: %s\n' "$name"
    else
        printf 'FAIL: %s (expected %s, got %s)\n' "$name" "$want" "$actual" >&2
        cat "$scratch/result" >&2
        failed=1
    fi
}
expect 0 'public documentation contract' bash "$root/scripts/docs-check.sh" "$root"
fixture="$scratch/checkout ü"
mkdir -p "$fixture"
cp -R "$root/docs" "$fixture/docs"
for file in README.md CONTRIBUTING.md SECURITY.md LICENSE THIRD_PARTY_NOTICES.md AGENTS.md MASTERPLAN.md; do
    if [[ -f "$root/$file" ]]; then cp "$root/$file" "$fixture/"; fi
done
expect 0 'complete documentation fixture is a passing control' bash "$root/scripts/docs-check.sh" "$fixture"
for missing in docs/install.md LICENSE docs/assets/tusk-tui.png; do
    if [[ -f "$fixture/$missing" ]]; then
        mv "$fixture/$missing" "$scratch/saved"
        expect 1 "reject missing $missing" bash "$root/scripts/docs-check.sh" "$fixture"
        mv "$scratch/saved" "$fixture/$missing"
    fi
done
for bad in '[Broken](docs/absent.md)' '[Bad anchor](docs/cli.md#absent-heading)' 'go install github.com/newbpydev/tusk/cmd/tusk@v0.3.0' '![Release](https://img.shields.io/github/v/release/newbpydev/tusk)' 'tusk add sample --due "next week"' 'Guaranteed sub-15ms on every platform'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$bad" >>"$fixture/README.md"
    expect 1 "reject unsupported public content: $bad" bash "$root/scripts/docs-check.sh" "$fixture"
done
cp "$root/README.md" "$fixture/README.md"
printf "\ntusk add sample --due='next week'\n" >>"$fixture/README.md"
expect 1 'reject single-quoted unsupported due-date claim' bash "$root/scripts/docs-check.sh" "$fixture"
cp "$root/README.md" "$fixture/README.md"
if [[ -f "$fixture/docs/install.md" ]]; then
    sed '/^Install Git,/s/\*\*Bash\*\*/a POSIX shell/' "$root/docs/install.md" >"$fixture/docs/install.md"
    expect 1 'source route declares Bash prerequisite' bash "$root/scripts/docs-check.sh" "$fixture"
    if ! grep -Fq 'lacks source Bash prerequisite' "$scratch/result"; then
        echo 'FAIL: missing Bash did not reach its prerequisite boundary' >&2
        cat "$scratch/result" >&2
        failed=1
    fi
fi
exit "$failed"
