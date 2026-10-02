#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
export GH_HOST=github.com
unset GH_DEBUG DEBUG
die() { printf 'Homebrew: %s\n' "$*" >&2;exit 1; }
case "${1:-check}" in
 check)
  [[ -f "${RELEASE_MANIFEST:-}" && -f "${HOMEBREW_CASK:-}" ]] || die 'RELEASE_MANIFEST and HOMEBREW_CASK required'
  bash "$root/scripts/release_check.sh" verify "$(dirname "$RELEASE_MANIFEST")"
  bash "$root/scripts/release_check.sh" check-cask "$RELEASE_MANIFEST" "$HOMEBREW_CASK" ;;
 render)
  [[ -f "${RELEASE_MANIFEST:-}" && -n "${HOMEBREW_CASK:-}" ]] || die 'manifest and explicit new cask output required'
  parent=$(cd "$(dirname "$HOMEBREW_CASK")" && pwd -P) || die 'existing output parent required'
  output="$parent/$(basename "$HOMEBREW_CASK")"
  candidate=$(cd "$(dirname "$RELEASE_MANIFEST")/.." && pwd -P)
  case "$output" in "$candidate"|"$candidate"/*|"$root"|"$root"/*) die 'render output must be outside candidate and source storage';; esac
  bash "$root/scripts/release_check.sh" verify "$(dirname "$RELEASE_MANIFEST")"
  bash "$root/scripts/release_check.sh" render-cask "$RELEASE_MANIFEST" "$output"
  printf 'Local declarative cask generated; native brew audit/install and tap readiness remain separate.\n' ;;
 destination)
  driver=$(mktemp -d)
  trap 'rm -rf "$driver"' EXIT
  # shellcheck source=scripts/gh_deadline.sh
  source "$root/scripts/gh_deadline.sh"
  gh_deadline_setup "$root" "$driver" || die 'could not prepare bounded GitHub driver'
  # Read-only; a missing/inaccessible tap is a named prerequisite failure.
  if ! gh api https://api.github.com/repos/newbpydev/homebrew-tap > /dev/null;then die '006-ISS-003: owner-controlled newbpydev/homebrew-tap missing or inaccessible';fi
  gh api https://api.github.com/repos/newbpydev/homebrew-tap --jq '.full_name=="newbpydev/homebrew-tap" and .owner.login=="newbpydev" and .private==false and .archived==false and .disabled==false and .permissions.push==true' | jq -e '.==true' >/dev/null || die '006-ISS-003: public owned writable tap required' ;;
 *) die 'expected check, render or destination';;
esac
