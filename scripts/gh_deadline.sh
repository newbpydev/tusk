#!/usr/bin/env bash
# Source from maintainer tooling. The caller owns its new output directory.
gh_deadline_setup() {
    local source_root=$1
    gh_deadline_binary="$2/gh-deadline.exe"
    [[ ! -e "$gh_deadline_binary" && ! -L "$gh_deadline_binary" ]] || return 1
    GH_DEADLINE_OUTPUT="$gh_deadline_binary" make -C "$source_root" build-gh-deadline >&2 || return
}
gh() { "$gh_deadline_binary" "$@"; }
