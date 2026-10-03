#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
jq_binary_option=''
case "${OSTYPE:-}" in msys*|cygwin*) jq_binary_option=--binary ;; esac
jq() { command jq ${jq_binary_option:+"--binary"} "$@"; }
failed=0
expect() {
 local want=$1 label=$2 actual=0;shift 2
 "$@" >"$scratch/result" 2>&1 || actual=$?
 if [[ "$actual" == "$want" ]];then printf 'PASS: %s\n' "$label"
 else printf 'FAIL: %s (%s vs %s)\n' "$label" "$actual" "$want" >&2;cat "$scratch/result" >&2;failed=1;fi
}
# Literal child-shell variables expand in the child, not this test harness.
# shellcheck disable=SC2016
history=(bash -c 'source "$1"; release_version_history "$2" "$3" "$4" "$5"' _ "$root/scripts/release_version.sh")
printf '[[]]\n' >"$scratch/releases.json"
printf '[[]]\n' >"$scratch/tags.json"
expect 0 'first stable version has no remote predecessor' "${history[@]}" v0.3.0 candidate "$scratch/releases.json" "$scratch/tags.json"
if [[ ! -f "$root/scripts/release_version.sh" ]];then exit 1;fi
for version in v0.3.0 v0.4.0 v1.0.0;do
 jq -n --arg version "$version" '[[{name:$version}]]' >"$scratch/tags.json"
 expect 1 "candidate rejects reserved tag $version" "${history[@]}" v0.3.0 candidate "$scratch/releases.json" "$scratch/tags.json"
done
printf '[[]]\n' >"$scratch/tags.json"
for version in v0.3.0 v0.4.0;do
 jq -n --arg version "$version" '[[{tag_name:$version,draft:true}]]' >"$scratch/releases.json"
 expect 1 "candidate rejects existing release $version including drafts" "${history[@]}" v0.3.0 candidate "$scratch/releases.json" "$scratch/tags.json"
done
jq -n '[[{tag_name:"v0.3.0",draft:true}]]' >"$scratch/releases.json"
jq -n '[[{name:"v0.3.0"}]]' >"$scratch/tags.json"
expect 0 'exact draft resumption permits selected version' "${history[@]}" v0.3.0 resume "$scratch/releases.json" "$scratch/tags.json"
jq -n '[[{name:"v0.3.0"}],[{name:"v0.4.0"}]]' >"$scratch/tags.json"
expect 1 'resumption refuses newer version on a later page' "${history[@]}" v0.3.0 resume "$scratch/releases.json" "$scratch/tags.json"
printf '[[]]\n' >"$scratch/releases.json"
for pair in 'v1.10.0 v1.9.9' 'v2.0.0 v1.99.99' 'v0.3.10 v0.3.9' 'v999999999999999999999.0.0 v999999999999999999998.0.0';do
 read -r proposed previous <<<"$pair"
 jq -n --arg previous "$previous" '[[{name:$previous}]]' >"$scratch/tags.json"
 expect 0 "numeric SemVer ordering: $proposed after $previous" "${history[@]}" "$proposed" candidate "$scratch/releases.json" "$scratch/tags.json"
 jq -n --arg v "$proposed" '[[{name:$v}]]' >"$scratch/tags.json"
 expect 1 "numeric SemVer ordering refuses $previous before $proposed" "${history[@]}" "$previous" candidate "$scratch/releases.json" "$scratch/tags.json"
done
printf '[[]]\n' >"$scratch/tags.json"
for version in 0.3.0 v01.3.0 v0.03.0 v0.3.00 v0.3.0-rc.1 v0.3.0+build 'v0.3.0;touch marker';do
 expect 1 "reject unsupported version $version" "${history[@]}" "$version" candidate "$scratch/releases.json" "$scratch/tags.json"
done
for bad in '[]' '{}' '[[{}]]' '[[{"tag_name":null}]]' '[[]][[]]';do
 printf '%s\n' "$bad" >"$scratch/releases.json"
 expect 1 "refuse unavailable or malformed release history: $bad" "${history[@]}" v0.3.0 candidate "$scratch/releases.json" "$scratch/tags.json"
done
printf '[[]]\n' >"$scratch/releases.json"
for bad in '[]' '{}' '[[{}]]' '[[{"name":null}]]' '[[]][[]]';do
 printf '%s\n' "$bad" >"$scratch/tags.json"
 expect 1 "refuse unavailable or malformed tag history: $bad" "${history[@]}" v0.3.0 candidate "$scratch/releases.json" "$scratch/tags.json"
done
printf '[[{"name":"legacy-baseline"}]]\n' >"$scratch/tags.json"
expect 0 'unrelated legacy tag does not determine stable version' "${history[@]}" v0.3.0 candidate "$scratch/releases.json" "$scratch/tags.json"
expect 1 'unknown history mode is refused' "${history[@]}" v0.3.0 unknown "$scratch/releases.json" "$scratch/tags.json"
# The remote adapter must read every page and never turn a failed read into an
# empty history. Its caller supplies the separately tested bounded gh driver.
printf '[[]]\n' >"$scratch/tags.json"
cat >"$scratch/remote-driver" <<'REMOTE'
#!/usr/bin/env bash
set -euo pipefail
source "$1"
gh() {
 printf '%s\n' "$*" >>"$FIXTURE_ROOT/calls"
 [[ "$1" == api && "$3" == --paginate && "$4" == --slurp && $# == 4 ]] || return 77
 case "$2" in
  'https://api.github.com/repos/newbpydev/tusk/releases?per_page=100')
   [[ "${FIXTURE_READ_FAIL:-}" != releases ]] || return 49
   cat "$FIXTURE_ROOT/releases.json" ;;
  'https://api.github.com/repos/newbpydev/tusk/tags?per_page=100')
   [[ "${FIXTURE_READ_FAIL:-}" != tags ]] || return 49
   cat "$FIXTURE_ROOT/tags.json" ;;
  *) return 77 ;;
 esac
}
release_version_remote v0.3.0 candidate "$2"
REMOTE
remote=(env FIXTURE_ROOT="$scratch" bash "$scratch/remote-driver" "$root/scripts/release_version.sh")
expect 0 'remote control reads both complete paginated histories' "${remote[@]}" "$scratch/remote-control"
expect 0 'remote control recorded both API reads' test "$(wc -l <"$scratch/calls" | tr -d ' ')" == 2
for endpoint in releases tags;do
 expect 1 "failed $endpoint read refuses an available version" env FIXTURE_READ_FAIL="$endpoint" "${remote[@]}" "$scratch/failed-$endpoint"
done
jq -n '[[{name:"v0.2.0"}],[{name:"v0.4.0"}]]' >"$scratch/tags.json"
expect 1 'remote check inspects newer tag on the second page' "${remote[@]}" "$scratch/later-page"
expect 1 'remote history refuses output reuse' "${remote[@]}" "$scratch/remote-control"
# Guard fidelity: this must be wired into the real workflow before CI/build,
# with only read access, and into the actual promotion implementation.
# The GitHub expressions are literal workflow policy strings.
# shellcheck disable=SC2016
expect 0 'candidate identity gate rejects reused versions before CI' jq -e '.jobs.identity.steps[-1].run=="make check-release-version RELEASE_VERSION_OUTPUT=\"$RUNNER_TEMP/version-history\"" and .jobs.identity.steps[-1].env.GH_TOKEN=="${{ github.token }}" and .jobs.identity.steps[-1].env.RELEASE_VERSION=="${{ inputs.version }}" and .jobs.ci.needs==["identity"] and .permissions=={contents:"read"}' "$root/.github/workflows/release.yml"
expect 0 'promotion uses the shared remote history check' rg -q 'release_version_remote.*resume' "$root/scripts/promote.sh"
exit "$failed"
