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
for file in README.md CHANGELOG.md CONTRIBUTING.md SECURITY.md LICENSE THIRD_PARTY_NOTICES.md AGENTS.md MASTERPLAN.md; do
    if [[ -f "$root/$file" ]]; then cp "$root/$file" "$fixture/"; fi
done
expect 0 'complete documentation fixture is a passing control' bash "$root/scripts/docs-check.sh" "$fixture"
cp "$root/README.md" "$fixture/README.md"
printf '\n[![Native CI](https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/newbpydev/tusk/actions/workflows/ci.yml)\n[![MIT](https://img.shields.io/github/license/newbpydev/tusk)](LICENSE)\n' >>"$fixture/README.md"
expect 0 'verified native CI and detected MIT badges are allowed' bash "$root/scripts/docs-check.sh" "$fixture"
cp "$root/README.md" "$fixture/README.md"
cp "$fixture/SECURITY.md" "$scratch/security-saved"
printf '# Security reporting\nGitHub private vulnerability reporting is enabled and verified on 2026-10-03.\nUse https://github.com/newbpydev/tusk/security/advisories/new to report privately.\n' >"$fixture/SECURITY.md"
expect 0 'verified enabled private reporting is allowed' bash "$root/scripts/docs-check.sh" "$fixture"
for state in 'is disabled' 'is currently disabled' 'is not enabled' 'is currently not enabled'; do
    printf '# Security reporting\nGitHub private vulnerability reporting %s.\nUse https://github.com/newbpydev/tusk to contact the maintainer.\n' "$state" >"$fixture/SECURITY.md"
    expect 0 "allow explicit disabled private reporting state: $state" bash "$root/scripts/docs-check.sh" "$fixture"
done
for state in 'was never enabled' 'is no longer enabled' 'is not enabled and verified'; do
    printf '# Security reporting\nGitHub private vulnerability reporting %s\nUse https://github.com/newbpydev/tusk/security/advisories/new.\n' "$state" >"$fixture/SECURITY.md"
    expect 1 "reject ambiguous private reporting state: $state" bash "$root/scripts/docs-check.sh" "$fixture"
done
mv "$scratch/security-saved" "$fixture/SECURITY.md"
for missing in docs/install.md LICENSE docs/assets/tusk-tui.png; do
    if [[ -f "$fixture/$missing" ]]; then
        mv "$fixture/$missing" "$scratch/saved"
        expect 1 "reject missing $missing" bash "$root/scripts/docs-check.sh" "$fixture"
        mv "$scratch/saved" "$fixture/$missing"
    fi
done
for bad in '[Broken](docs/absent.md)' '[Bad anchor](docs/cli.md#absent-heading)' 'go install github.com/newbpydev/tusk/cmd/tusk@v0.3.0' '![Release](https://img.shields.io/github/v/release/newbpydev/tusk)' '![Wrong repo](https://github.com/evil/tusk/actions/workflows/ci.yml/badge.svg)' '![Static pass](https://img.shields.io/badge/CI-passing-green)' 'tusk add sample --due "next week"' 'Guaranteed sub-15ms on every platform'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$bad" >>"$fixture/README.md"
    expect 1 "reject unsupported public content: $bad" bash "$root/scripts/docs-check.sh" "$fixture"
done
cp "$root/README.md" "$fixture/README.md"
for badge in '<img src="https://img.shields.io/github/v/release/newbpydev/tusk">' '[badge]: https://img.shields.io/badge/CI-passing-green' '<img src="https://github.com/evil/tusk/actions/workflows/ci.yml/badge.svg">' '[badge]: https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main-evil'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$badge" >>"$fixture/README.md"
    expect 1 "reject unverified raw badge: $badge" bash "$root/scripts/docs-check.sh" "$fixture"
done
cp "$root/README.md" "$fixture/README.md"
printf '\n<img src="https://img.shields.io/github/license/newbpydev/tusk">\n[ci]: https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main\n' >>"$fixture/README.md"
expect 0 'verified raw HTML and reference badges are allowed' bash "$root/scripts/docs-check.sh" "$fixture"
cp "$root/README.md" "$fixture/README.md"
printf '\n<img src="//img.shields.io/badge/CI-passing-green">\n' >>"$fixture/README.md"
expect 1 'reject protocol-relative unverified badge' bash "$root/scripts/docs-check.sh" "$fixture"
for badge in '<img src="https://IMG.SHIELDS.IO/badge/CI-passing-green">' '<img src="//Img.Shields.Io/badge/CI-passing-green">'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$badge" >>"$fixture/README.md"
    expect 1 "reject mixed-case unverified badge host: $badge" bash "$root/scripts/docs-check.sh" "$fixture"
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
