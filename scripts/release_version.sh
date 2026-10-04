#!/usr/bin/env bash
# Shared stable-version policy for candidate construction and promotion.
release_version_valid() { [[ "$1" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; }
release_version_error() { printf 'Release version: %s\n' "$*" >&2;return 1; }
release_version_newer() {
 local proposed=() previous=() i left right LC_ALL=C
 IFS=. read -r -a proposed <<<"${1#v}"
 IFS=. read -r -a previous <<<"${2#v}"
 # Compare digit lengths then ASCII digits, without integer overflow or jq's
 # floating-point rounding. SemVer components have no leading zeroes.
 for i in 0 1 2;do
  left=${proposed[$i]};right=${previous[$i]}
  if [[ ${#left} -ne ${#right} ]];then [[ ${#left} -gt ${#right} ]];return;fi
  if [[ "$left" != "$right" ]];then [[ "$left" > "$right" ]];return;fi
 done
 return 1
}
release_version_history() {
 local version=$1 mode=$2 releases=$3 tags=$4 binary='' existing
 case "${OSTYPE:-}" in msys*|cygwin*) binary=--binary ;; esac
 release_version_valid "$version" || { release_version_error 'expected stable vMAJOR.MINOR.PATCH; prereleases are not supported';return 1; }
 [[ "$mode" == candidate || "$mode" == resume ]] || { release_version_error 'invalid history mode';return 1; }
 for existing in "$releases" "$tags";do
  jq ${binary:+"--binary"} -e -s 'length==1 and (.[0]|type=="array" and length>0 and all(.[];type=="array"))' "$existing" >/dev/null || { release_version_error 'complete paginated history required';return 1; }
 done
 if ! jq ${binary:+"--binary"} -e 'all(.[][];type=="object" and (.tag_name|type=="string" and length>0))' "$releases" >/dev/null ||
    ! jq ${binary:+"--binary"} -e 'all(.[][];type=="object" and (.name|type=="string" and length>0))' "$tags" >/dev/null;then
  release_version_error 'malformed version history';return 1
 fi
 while IFS= read -r existing;do
  release_version_valid "$existing" || continue
  [[ "$mode" != resume || "$version" != "$existing" ]] || continue
  release_version_newer "$version" "$existing" || { release_version_error "$version must be newer than reserved version $existing";return 1; }
 done < <(jq ${binary:+"--binary"} -r '.[][]|.tag_name' "$releases";jq ${binary:+"--binary"} -r '.[][]|.name' "$tags")
}
release_version_remote() {
 local version=$1 mode=$2 directory=$3
 release_version_valid "$version" || { release_version_error 'expected stable vMAJOR.MINOR.PATCH';return 1; }
 mkdir "$directory" || { release_version_error 'history output already exists';return 1; }
 gh api 'https://api.github.com/repos/newbpydev/tusk/releases?per_page=100' --paginate --slurp >"$directory/releases.json" || { release_version_error 'release history unavailable; no release permitted';return 1; }
 gh api 'https://api.github.com/repos/newbpydev/tusk/tags?per_page=100' --paginate --slurp >"$directory/tags.json" || { release_version_error 'tag history unavailable; no release permitted';return 1; }
 release_version_history "$version" "$mode" "$directory/releases.json" "$directory/tags.json"
}
if [[ "${BASH_SOURCE[0]}" == "$0" ]];then
 set -euo pipefail
 root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
 export GH_HOST=github.com
 unset GH_DEBUG DEBUG
 release_version_valid "${RELEASE_VERSION:-}" || { release_version_error 'expected stable vMAJOR.MINOR.PATCH';exit 1; }
 [[ -n "${RELEASE_VERSION_OUTPUT:-}" ]] || { release_version_error 'new retained output required';exit 1; }
 parent=$(cd "$(dirname "$RELEASE_VERSION_OUTPUT")" && pwd -P)
 output="$parent/$(basename "$RELEASE_VERSION_OUTPUT")"
 case "$output" in "$root"|"$root"/*) release_version_error 'output must be outside source storage';exit 1;; esac
 mkdir "$output" || { release_version_error 'output collision';exit 1; }
 # shellcheck source=scripts/gh_deadline.sh
 source "$root/scripts/gh_deadline.sh"
 gh_deadline_setup "$root" "$output"
 release_version_remote "$RELEASE_VERSION" candidate "$output/history"
 binary=''
 case "${OSTYPE:-}" in msys*|cygwin*) binary=--binary ;; esac
 jq ${binary:+"--binary"} -n --arg version "$RELEASE_VERSION" '{schema:1,status:"available",repository:"newbpydev/tusk",version:$version,scope:"live-history-readback; no reservation or publication"}' >"$output/receipt.json"
 printf 'Version %s is available above existing stable tags and releases.\n' "$RELEASE_VERSION"
fi
