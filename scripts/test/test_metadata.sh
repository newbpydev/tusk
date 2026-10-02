#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
make -n -C "$root" prepare-repository-metadata >/dev/null
make -n -C "$root" apply-repository-metadata >/dev/null
[[ -f "$root/scripts/repository_metadata.sh" ]]
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
failed=0
expect() {
 local want=$1 label=$2 actual=0;shift 2
 "$@" >"$scratch/output" 2>&1 || actual=$?
 if [[ "$actual" == "$want" ]];then printf 'PASS: %s\n' "$label";else printf 'FAIL: %s (%s vs %s)\n' "$label" "$actual" "$want" >&2;cat "$scratch/output" >&2;failed=1;fi
}
# API/approval boundary fixtures, not actual publication or owner approval.
mkdir -p "$scratch/fixture/scripts" "$scratch/fixture/.github" "$scratch/fixture/docs/assets" "$scratch/bin"
cp "$root/scripts/repository_metadata.sh" "$scratch/fixture/scripts/"
 cp "$root/scripts/json_check.sh" "$scratch/fixture/scripts/"
cp "$root/Makefile" "$scratch/fixture/"
cp "$root/scripts/gh_deadline.sh" "$scratch/fixture/scripts/"
cp -r "$root/scripts/ghdeadline" "$scratch/fixture/scripts/"
cp "$root/.github/repository-metadata.json" "$scratch/fixture/.github/"
printf preview-fixture >"$scratch/fixture/docs/assets/tusk-tui.png"
payload_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/fixture/.github/repository-metadata.json")
jq -n '{schema:1,status:"published",repository:"newbpydev/tusk",release_id:8,version:"v0.3.0",source_sha:("a"*40),assets_verified:9,fixture:true}' >"$scratch/release-receipt.json"
release_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/release-receipt.json")
jq -n --arg payload "$payload_hash" --arg release "$release_hash" '{schema:1,status:"owner-approved",action:"metadata",repository:"newbpydev/tusk",documentation_sha:("a"*40),payload_sha256:$payload,release_receipt_sha256:$release,approval_reference:"fixture only; not actual owner consent"}' >"$scratch/approval.json"
printf '{"full_name":"newbpydev/tusk","description":"","homepage":""}' >"$scratch/repository.json"
printf '{"names":[]}' >"$scratch/topics.json"
cat >"$scratch/bin/git" <<'GIT'
#!/usr/bin/env bash
shift 2
case "$1" in rev-parse) printf 'a%.0s' {1..40};;status) ;;*) exit 77;;esac
GIT
cat >"$scratch/bin/gh" <<'GH'
#!/usr/bin/env bash
set -euo pipefail
[[ "${GH_HOST:-}" == github.com && -z "${GH_DEBUG:-}" && -z "${DEBUG:-}" ]] || exit 78
printf '%s\n' "$*" >>"$FIXTURE_ROOT/calls"
[[ "$1" == api ]] || exit 77
[[ "$2" == https://api.github.com/repos/* ]] || exit 79
endpoint=${2#https://api.github.com/};shift 2;method=GET;input=''
while [[ $# -gt 0 ]];do
 case "$1" in --method) method=$2;shift 2;;--input) input=$2;shift 2;;*) shift;;esac
done
case "$endpoint:$method" in
 repos/newbpydev/tusk/releases/8:GET) jq -n '{id:8,draft:false,tag_name:"v0.3.0",target_commitish:("a"*40)}' ;;
 repos/newbpydev/tusk/git/ref/heads/main:GET) jq -n --arg changed "${FIXTURE_MAIN_CHANGED:-0}" '{object:{type:"commit",sha:((if $changed=="1" then "b" else "a" end)*40)}}' ;;
 repos/newbpydev/tusk:GET) cat "$FIXTURE_ROOT/repository.json" ;;
 repos/newbpydev/tusk/topics:GET) cat "$FIXTURE_ROOT/topics.json" ;;
 repos/newbpydev/tusk:PATCH)
  jq -e 'keys==["description","homepage"]' "$input" >/dev/null || exit 77
  jq '.+{full_name:"newbpydev/tusk"}' "$input" >"$FIXTURE_ROOT/repository.json"
  [[ "${FIXTURE_LOST:-0}" != 1 ]] || exit 49 ;;
 repos/newbpydev/tusk/topics:PUT)
  [[ "${FIXTURE_TOPICS_FAIL:-0}" != 1 ]] || exit 47
  cp "$input" "$FIXTURE_ROOT/topics.json"
  [[ "${FIXTURE_LOST:-0}" != 1 ]] || exit 49 ;;
 *) exit 77;;
esac
GH
chmod +x "$scratch/bin/"*
invoke=(env PATH="$scratch/bin:$PATH" GH_HOST=wrong.example GH_DEBUG=api DEBUG=1 FIXTURE_ROOT="$scratch" METADATA_AUTHORIZATION="$scratch/approval.json" METADATA_RELEASE_RECEIPT="$scratch/release-receipt.json")
cat >"$scratch/interrupt-env" <<'INTERRUPT'
trap 'if [[ "$BASH_COMMAND" == gh_deadline_setup* ]]; then trap - DEBUG; kill -TERM "$$"; fi' DEBUG
INTERRUPT
expect 143 'interrupted mutation retains its signal status' "${invoke[@]}" BASH_ENV="$scratch/interrupt-env" METADATA_OUTPUT="$scratch/interrupted" bash "$scratch/fixture/scripts/repository_metadata.sh" apply
expect 0 'interrupted audit receipt records exit 143' jq -e '.exit_status==143' "$scratch/interrupted/result.json"

expect 0 'metadata prepare performs no API request' "${invoke[@]}" METADATA_OUTPUT="$scratch/prepared" bash "$scratch/fixture/scripts/repository_metadata.sh" prepare
[[ ! -e "$scratch/calls" ]] || failed=1
expect 1 'missing owner authorization refuses metadata writes' "${invoke[@]}" METADATA_AUTHORIZATION= METADATA_OUTPUT="$scratch/missing-approval" bash "$scratch/fixture/scripts/repository_metadata.sh" apply
[[ ! -e "$scratch/calls" ]] || failed=1
cp "$scratch/approval.json" "$scratch/original-approval.json"
cp "$scratch/repository.json" "$scratch/original-repository.json";cp "$scratch/topics.json" "$scratch/original-topics.json"
jq '.status="not-approved"' "$scratch/original-approval.json" >"$scratch/approval.json"
cat "$scratch/original-approval.json" >>"$scratch/approval.json"
expect 1 'concatenated metadata approval refuses all API calls' "${invoke[@]}" METADATA_OUTPUT="$scratch/concatenated-approval" bash "$scratch/fixture/scripts/repository_metadata.sh" apply
[[ ! -e "$scratch/calls" ]] || failed=1
mv "$scratch/original-approval.json" "$scratch/approval.json"
mv "$scratch/original-repository.json" "$scratch/repository.json";mv "$scratch/original-topics.json" "$scratch/topics.json"
rm -f "$scratch/calls"
expect 1 'changed main refuses approved stale metadata' "${invoke[@]}" FIXTURE_MAIN_CHANGED=1 METADATA_OUTPUT="$scratch/changed-main" bash "$scratch/fixture/scripts/repository_metadata.sh" apply
expect 1 'partial topic failure preserves acknowledged About state' "${invoke[@]}" FIXTURE_TOPICS_FAIL=1 METADATA_OUTPUT="$scratch/partial" bash "$scratch/fixture/scripts/repository_metadata.sh" apply
expect 0 'lost topic response reconciles without repeating About patch' "${invoke[@]}" FIXTURE_LOST=1 METADATA_OUTPUT="$scratch/retry" bash "$scratch/fixture/scripts/repository_metadata.sh" apply
[[ "$(grep -Ec -- '--method PATCH' "$scratch/calls" || true)" == 1 ]] || failed=1
expect 0 'matching metadata causes no writes' "${invoke[@]}" METADATA_OUTPUT="$scratch/matching" bash "$scratch/fixture/scripts/repository_metadata.sh" apply
[[ "$(grep -Ec -- '--method PATCH' "$scratch/calls" || true)" == 1 && "$(grep -Ec -- '--method PUT' "$scratch/calls" || true)" == 2 ]] || failed=1
expect 1 'retained metadata result not overwritten' "${invoke[@]}" METADATA_OUTPUT="$scratch/matching" bash "$scratch/fixture/scripts/repository_metadata.sh" apply
exit "$failed"
