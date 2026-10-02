#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
failed=0
for target in release-prepare release-draft release-publish release-readback;do
 if make -n -C "$root" "$target" >/dev/null 2>&1;then printf 'PASS: %s available\n' "$target";else printf 'FAIL: %s missing\n' "$target" >&2;failed=1;fi
done
[[ -f "$root/scripts/promote.sh" ]] || failed=1
if [[ ! -f "$root/scripts/promote.sh" ]];then exit "$failed";fi
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
expect() {
 local want=$1 label=$2 actual=0;shift 2
 "$@" >"$scratch/output" 2>&1 || actual=$?
 if [[ "$actual" == "$want" ]];then printf 'PASS: %s\n' "$label";else printf 'FAIL: %s (%s vs %s)\n' "$label" "$actual" "$want" >&2;cat "$scratch/output" >&2;failed=1;fi
}
# Only policy/API boundary fixtures. No signatures, native execution or timing proof.
mkdir -p "$scratch/fixture/scripts" "$scratch/candidate/assets" "$scratch/candidate/homebrew/Casks" "$scratch/bin" "$scratch/remote" "$scratch/gates"
cp "$root/scripts/promote.sh" "$scratch/fixture/scripts/"
 cp "$root/scripts/json_check.sh" "$scratch/fixture/scripts/"
cp "$root/Makefile" "$scratch/fixture/"
cp "$root/scripts/gh_deadline.sh" "$scratch/fixture/scripts/"
cp -r "$root/scripts/ghdeadline" "$scratch/fixture/scripts/"
printf '#!/usr/bin/env bash\nexit 0\n' >"$scratch/fixture/scripts/release_check.sh"
sha=$(printf 'a%.0s' {1..40})
for name in tusk_0.3.0_linux_amd64.tar.gz tusk_0.3.0_linux_arm64.tar.gz tusk_0.3.0_darwin_amd64.tar.gz tusk_0.3.0_darwin_arm64.tar.gz tusk_0.3.0_windows_amd64.zip tusk_0.3.0_source.tar.gz THIRD_PARTY_NOTICES.md checksums.txt;do printf 'fixture %s' "$name" >"$scratch/candidate/assets/$name";done
hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/candidate/assets/checksums.txt")
jq -n --arg sha "$sha" --arg hash "$hash" '{schema:1,mode:"candidate",version:"0.3.0",source_sha:$sha,compiler:"1.27.1",checksums_sha256:$hash,files_sha256:{},targets:[{target:"linux/amd64",executable_sha256:("b"*64)}]}' >"$scratch/manifest"
for file in "$scratch/candidate/assets/"*;do
 [[ "$(basename "$file")" != checksums.txt ]] || continue
 hash=$(bash "$root/scripts/test/sha256.sh" "$file")
 jq --arg file "$(basename "$file")" --arg hash "$hash" '.files_sha256[$file]=$hash' "$scratch/manifest" >"$scratch/new";mv "$scratch/new" "$scratch/manifest"
done
jq '. as $m | .targets=[ ["linux/amd64","linux/arm64","darwin/amd64","darwin/arm64","windows/amd64"][] as $t | {target:$t,executable_sha256:("b"*64),archive:("tusk_0.3.0_"+($t|gsub("/";"_"))+(if $t|startswith("windows/") then ".zip" else ".tar.gz" end))}]' "$scratch/manifest" >"$scratch/new";mv "$scratch/new" "$scratch/manifest"
mv "$scratch/manifest" "$scratch/candidate/assets/release-manifest.json"
manifest_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/candidate/assets/release-manifest.json")
printf 'fixture cask' >"$scratch/candidate/homebrew/Casks/tusk.rb"
cask_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/candidate/homebrew/Casks/tusk.rb")
# Intentionally literal shell-looking text, passed only through structured JSON.
# shellcheck disable=SC2016
printf 'Reviewed notes; $(touch marker) and `touch marker` remain literal.\n' >"$scratch/notes.md"
jq -n --arg sha "$sha" --arg hash "$manifest_hash" --arg cask "$cask_hash" '{schema:1,status:"accepted",repository:"newbpydev/tusk",version:"v0.3.0",source_sha:$sha,manifest_sha256:$hash,cask_sha256:$cask,run_id:123,prepublication_scenarios_reviewed:true,gates:[],reports:{}}' >"$scratch/acceptance.json"
for id in native/linux/amd64 native/linux/arm64 native/darwin/amd64 native/darwin/arm64 native/windows/amd64 cli-reference tui-reference cask/darwin/amd64 cask/darwin/arm64;do
 path="$scratch/gates/${id//\//-}.json"
 jq -n --arg id "$id" --arg sha "$sha" --arg hash "$manifest_hash" '{status:"accepted",scope:"trusted-candidate",gate:$id,source_sha:$sha,version:"v0.3.0",manifest_sha256:$hash,run_id:123,fixture:true}' >"$path"
 if [[ "$id" == native/* ]];then
  jq --arg target "${id#native/}" --slurpfile m "$scratch/candidate/assets/release-manifest.json" '. + ($m[0].targets[]|select(.target==$target)|{target:.target,executable_sha256:.executable_sha256,archive_sha256:$m[0].files_sha256[.archive]})' "$path" >"$scratch/new";mv "$scratch/new" "$path"
 fi
 hash=$(bash "$root/scripts/test/sha256.sh" "$path")
 jq --arg id "$id" --arg path "$path" --arg hash "$hash" '.gates += [{id:$id,passed:true,receipt:{path:$path,sha256:$hash}}]' "$scratch/acceptance.json" >"$scratch/new";mv "$scratch/new" "$scratch/acceptance.json"
done
jq -n 'def names: ["--help/invalid-config=false","--help/invalid-config=true","--version/invalid-config=false","--version/invalid-config=true"] + [ [0,100,1000][] as $n | ["list","tree","stats","history"][] as $cmd | [false,true][] | "\($n)/\($cmd)/json=\(.)" ]; {passed:true,manifest:{sha256:("b"*64),go:"go1.27.1",selected_acceptance_profile:"reference",fixture:"true"},cases:[range(1;4) as $run|names[]|{name:.,passed:true,run:$run,output_bytes:1,warmup_ns:[range(5)|1],samples_ns:[range(100)|1],p90_ns:1,p95_ns:1,p99_ns:1,max_ns:1}]}' >"$scratch/cli.json"
jq -n --arg sha "$sha" 'def names: [ [0,100,1000][] as $n | [[80,24],[120,40],[200,60]][] as $s | ["prepare","view"][] | "\(.)/\($n)-tasks/\($s[0])x\($s[1])" ] + [ ["80x24","120x40","200x60"][]|"calendar/1000-tasks/"+.] + ["markdown/32768-bytes","markdown/1048576-bytes","disk-refresh/100-tasks","stress/10000-tasks-1MiB-note-1000-events","stress/read-only-form-1MiB"]; {passed:true,manifest:{binary_sha256:("b"*64),revision:$sha,source_diff:"",go:"go1.27.1",fixture:"true"},cases:[range(1;4) as $run|names[]|{name:.,run:$run,passed:true,preparation_budget:(startswith("prepare/") or startswith("calendar/")),consumed_output_size:1,allocations:[range(100)|0],allocated_bytes:[range(100)|0],warmup_ns:[range(5)|1],samples_ns:[range(100)|1]}]}' >"$scratch/tui.json"
jq -n '{BinarySHA256:("b"*64),Go:"go1.27.1",Runs:[range(1;4)|{run:.,child_max_rss_kib:[range(100)|1],warmup_ns:[range(5)|1],samples_ns:[range(100)|1]}],fixture:true}' >"$scratch/startup.json"
for kind in cli tui startup;do
 hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/$kind.json")
 jq --arg kind "$kind" --arg path "$scratch/$kind.json" --arg hash "$hash" '.reports[$kind]={path:$path,sha256:$hash}' "$scratch/acceptance.json" >"$scratch/new";mv "$scratch/new" "$scratch/acceptance.json"
done
accepted_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/acceptance.json");notes_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/notes.md")
for action in draft publish;do
 jq -n --arg action "$action" --arg sha "$sha" --arg hash "$manifest_hash" --arg cask "$cask_hash" --arg accepted "$accepted_hash" --arg notes "$notes_hash" '{schema:1,status:"owner-approved",action:$action,repository:"newbpydev/tusk",version:"v0.3.0",source_sha:$sha,manifest_sha256:$hash,cask_sha256:$cask,run_id:123,acceptance_sha256:$accepted,notes_sha256:$notes,exclusive_release_window:true,approval_reference:"fixture only; not actual owner consent"}' >"$scratch/$action-approval.json"
done
cat >"$scratch/bin/git" <<'GIT'
#!/usr/bin/env bash
shift 2
case "$1" in
 status) [[ "${FIXTURE_DIRTY:-0}" != 1 ]] || echo ' M scripts/promote.sh' ;;
 merge-base) [[ "${FIXTURE_STALE:-0}" != 1 ]] ;;
 diff) [[ "${FIXTURE_DIFF_FAIL:-0}" != 1 ]] || exit 46;[[ "${FIXTURE_CHANGED:-0}" != 1 ]] || echo scripts/promote.sh ;;
 *) exit 77 ;;
esac
GIT
cat >"$scratch/bin/make" <<'MAKE'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >>"$FIXTURE_ROOT/make-calls"
if [[ "$*" == *build-gh-deadline* ]];then exec "$FIXTURE_REAL_MAKE" "$@";fi
if [[ "$*" == *homebrew-destination* ]];then [[ "${FIXTURE_TAP_MISSING:-0}" != 1 ]] || exit 48;exit 0;fi
[[ "$*" == *verify-candidate* && "${FIXTURE_EXPIRED:-0}" != 1 ]] || exit 47
mkdir "$CANDIDATE_VERIFICATION_DIR"
jq -n --arg hash "$CANDIDATE_MANIFEST_SHA256" --arg cask "$FIXTURE_CASK" '{status:"verified",manifest_sha256:$hash,cask_sha256:$cask,fixture:true}' >"$CANDIDATE_VERIFICATION_DIR/receipt.json"
MAKE
cat >"$scratch/bin/gh" <<'GH'
#!/usr/bin/env bash
set -euo pipefail
[[ "${GH_HOST:-}" == github.com && -z "${GH_DEBUG:-}" && -z "${DEBUG:-}" ]] || exit 78
printf '%s\n' "$*" >>"$FIXTURE_ROOT/gh-calls"
remote="$FIXTURE_ROOT/remote"
[[ "$1" == api ]] || exit 77
[[ "$2" == https://api.github.com/repos/* || "$2" == https://uploads.github.com/repos/* ]] || exit 79
endpoint=${2#https://api.github.com/};endpoint=${endpoint#https://uploads.github.com/};shift 2;method=GET;input=''
while [[ $# -gt 0 ]];do
 case "$1" in --method) method=$2;shift 2;; --input) input=$2;shift 2;; *) shift;; esac
done
case "$endpoint:$method" in
 repos/newbpydev/tusk/releases/8/assets\?name=*:POST)
  name=${endpoint##*name=}
  [[ ! -e "$remote/$name" ]] || exit 33
  cp "$input" "$remote/$name"
  id=$(jq 'length+100' "$remote/assets.json")
  jq --arg name "$name" --argjson id "$id" '.+[{id:$id,name:$name,state:"uploaded"}]' "$remote/assets.json" >"$remote/new";mv "$remote/new" "$remote/assets.json"
  [[ "${FIXTURE_LOST_UPLOAD:-0}" != 1 ]] || exit 49 ;;
 'repos/newbpydev/tusk/git/ref/heads/main:GET') jq -n '{object:{type:"commit",sha:("a"*40)}}' ;;
 repos/newbpydev/tusk/compare/*:GET)
  jq -n --arg changed "${FIXTURE_MAIN_CHANGED:-0}" '{status:"identical",merge_base_commit:{sha:("a"*40)},files:(if $changed=="1" then [{filename:"scripts/promote.sh"}] else [] end)}' ;;
 'repos/newbpydev/tusk/git/ref/tags/v0.3.0:GET')
  if [[ -s "$remote/tag.json" ]];then printf 'HTTP/2.0 200 OK\r\n\r\n';cat "$remote/tag.json";else printf 'HTTP/2.0 404 Not Found\r\n\r\n{}';exit 1;fi ;;
 'repos/newbpydev/tusk/releases?per_page=100:GET') jq -s '.' "$remote/releases.json" ;;
 'repos/newbpydev/tusk/releases:POST')
  jq -e '.draft==true and .target_commitish==("a"*40)' "$input" >/dev/null || exit 77
  jq '.+{id:8}' "$input" | jq -s '.' >"$remote/releases.json"
  [[ "${FIXTURE_LOST_CREATE:-0}" != 1 ]] || exit 49 ;;
 'repos/newbpydev/tusk/releases/8/assets?per_page=100:GET') jq -s '.' "$remote/assets.json" ;;
 repos/newbpydev/tusk/releases/assets/*:GET)
  id=${endpoint##*/};name=$(jq -r --argjson id "$id" '.[]|select(.id==$id)|.name' "$remote/assets.json")
  if [[ "${FIXTURE_BAD_READBACK:-0}" == 1 ]];then printf changed;else cat "$remote/$name";fi ;;
 'repos/newbpydev/tusk/git/refs:POST')
  jq '{object:{type:"commit",sha:.sha}}' "$input" >"$remote/tag.json" ;;
 'repos/newbpydev/tusk/releases/8:PATCH')
  jq -e '.draft==false' "$input" >/dev/null || exit 77
  jq '.[0].draft=false' "$remote/releases.json" >"$remote/new";mv "$remote/new" "$remote/releases.json"
  [[ "${FIXTURE_LOST_PUBLISH:-0}" != 1 ]] || exit 49 ;;
 *) echo "unexpected API $endpoint $method" >&2;exit 77 ;;
esac
GH
chmod +x "$scratch/bin/"*
reset() { rm -f -- "$scratch/remote/"*;printf '[]' >"$scratch/remote/releases.json";printf '[]' >"$scratch/remote/assets.json";: >"$scratch/remote/tag.json";: >"$scratch/gh-calls"; }
reset
invoke=(env PATH="$scratch/bin:$PATH" GH_HOST=wrong.example GH_DEBUG=api DEBUG=1 FIXTURE_REAL_MAKE="$(command -v make)" FIXTURE_ROOT="$scratch" FIXTURE_CASK="$cask_hash" CANDIDATE_DIR="$scratch/candidate" CANDIDATE_RUN_ID=123 CANDIDATE_MANIFEST_SHA256="$manifest_hash" RELEASE_SHA="$sha" RELEASE_VERSION=v0.3.0 RELEASE_ACCEPTANCE="$scratch/acceptance.json" RELEASE_NOTES="$scratch/notes.md" RELEASE_AUTHORIZATION="$scratch/draft-approval.json")
expect 0 'prepare stays unaccepted and read-only' "${invoke[@]}" RELEASE_PROMOTION_OUTPUT="$scratch/prepared" bash "$scratch/fixture/scripts/promote.sh" prepare
[[ ! -s "$scratch/gh-calls" ]] || failed=1
expect 1 'missing owner action authorization refused' "${invoke[@]}" RELEASE_AUTHORIZATION= RELEASE_PROMOTION_OUTPUT="$scratch/no-approval" bash "$scratch/fixture/scripts/promote.sh" draft
expect 1 'publication approval cannot authorize draft' "${invoke[@]}" RELEASE_AUTHORIZATION="$scratch/publish-approval.json" RELEASE_PROMOTION_OUTPUT="$scratch/wrong-action" bash "$scratch/fixture/scripts/promote.sh" draft
expect 1 'dirty tools refused' "${invoke[@]}" FIXTURE_DIRTY=1 RELEASE_PROMOTION_OUTPUT="$scratch/dirty" bash "$scratch/fixture/scripts/promote.sh" draft
expect 1 'changed build tools invalidate source freeze' "${invoke[@]}" FIXTURE_CHANGED=1 RELEASE_PROMOTION_OUTPUT="$scratch/changed" bash "$scratch/fixture/scripts/promote.sh" draft
expect 1 'failed source diff refuses acceptance' "${invoke[@]}" FIXTURE_DIFF_FAIL=1 RELEASE_PROMOTION_OUTPUT="$scratch/diff-failed" bash "$scratch/fixture/scripts/promote.sh" draft
expect 48 'missing tap refuses promotion' "${invoke[@]}" FIXTURE_TAP_MISSING=1 RELEASE_PROMOTION_OUTPUT="$scratch/tap-missing" bash "$scratch/fixture/scripts/promote.sh" draft
expect 47 'expired candidate refuses every remote write' "${invoke[@]}" FIXTURE_EXPIRED=1 RELEASE_PROMOTION_OUTPUT="$scratch/expired" bash "$scratch/fixture/scripts/promote.sh" draft
[[ ! -s "$scratch/gh-calls" ]] || failed=1
expect 1 'fresh main code change invalidates acceptance' "${invoke[@]}" FIXTURE_MAIN_CHANGED=1 RELEASE_PROMOTION_OUTPUT="$scratch/main-changed" bash "$scratch/fixture/scripts/promote.sh" draft
reset
index=0
for mutation in 'del(.executable_sha256)' '.executable_sha256=("c"*64)' '.archive_sha256=("c"*64)' '.target="linux/arm64"';do
 index=$((index+1))
 path="$scratch/gates/native-linux-amd64.json"
 cp "$path" "$scratch/original-gate.json";cp "$scratch/acceptance.json" "$scratch/original-acceptance.json";cp "$scratch/draft-approval.json" "$scratch/original-approval.json"
 jq "$mutation" "$path" >"$scratch/new";mv "$scratch/new" "$path"
 hash=$(bash "$root/scripts/test/sha256.sh" "$path")
 jq --arg hash "$hash" '(.gates[]|select(.id=="native/linux/amd64")|.receipt.sha256)=$hash' "$scratch/acceptance.json" >"$scratch/new";mv "$scratch/new" "$scratch/acceptance.json"
 hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/acceptance.json")
 jq --arg hash "$hash" '.acceptance_sha256=$hash' "$scratch/draft-approval.json" >"$scratch/new";mv "$scratch/new" "$scratch/draft-approval.json"
 expect 1 "reject native candidate binding: $mutation" "${invoke[@]}" RELEASE_PROMOTION_OUTPUT="$scratch/native-binding-$index" bash "$scratch/fixture/scripts/promote.sh" draft
 mv "$scratch/original-gate.json" "$path";mv "$scratch/original-acceptance.json" "$scratch/acceptance.json";mv "$scratch/original-approval.json" "$scratch/draft-approval.json"
 reset
done
cp "$scratch/cli.json" "$scratch/original-cli.json"
jq '.cases[1]=.cases[0]' "$scratch/cli.json" >"$scratch/new";mv "$scratch/new" "$scratch/cli.json"
report_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/cli.json")
cp "$scratch/acceptance.json" "$scratch/original-acceptance.json";cp "$scratch/draft-approval.json" "$scratch/original-approval.json"
jq --arg hash "$report_hash" '.reports.cli.sha256=$hash' "$scratch/acceptance.json" >"$scratch/new";mv "$scratch/new" "$scratch/acceptance.json"
accepted_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/acceptance.json")
jq --arg hash "$accepted_hash" '.acceptance_sha256=$hash' "$scratch/draft-approval.json" >"$scratch/new";mv "$scratch/new" "$scratch/draft-approval.json"
expect 1 'duplicate timing cases cannot satisfy all required runs' "${invoke[@]}" RELEASE_PROMOTION_OUTPUT="$scratch/duplicate-cases" bash "$scratch/fixture/scripts/promote.sh" draft
mv "$scratch/original-cli.json" "$scratch/cli.json";mv "$scratch/original-acceptance.json" "$scratch/acceptance.json";mv "$scratch/original-approval.json" "$scratch/draft-approval.json"
for kind in tui startup;do
 cp "$scratch/$kind.json" "$scratch/original-report.json"
 cp "$scratch/acceptance.json" "$scratch/original-acceptance.json";cp "$scratch/draft-approval.json" "$scratch/original-approval.json"
 if [[ "$kind" == tui ]];then mutation='.cases[0].allocations[0]="unknown"';else mutation='.Runs[0].child_max_rss_kib[0]=-1';fi
 jq "$mutation" "$scratch/$kind.json" >"$scratch/new";mv "$scratch/new" "$scratch/$kind.json"
 report_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/$kind.json")
 jq --arg kind "$kind" --arg hash "$report_hash" '.reports[$kind].sha256=$hash' "$scratch/acceptance.json" >"$scratch/new";mv "$scratch/new" "$scratch/acceptance.json"
 accepted_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/acceptance.json")
 jq --arg hash "$accepted_hash" '.acceptance_sha256=$hash' "$scratch/draft-approval.json" >"$scratch/new";mv "$scratch/new" "$scratch/draft-approval.json"
 expect 1 "$kind nonnumeric/negative memory sample refuses acceptance" "${invoke[@]}" RELEASE_PROMOTION_OUTPUT="$scratch/malformed-$kind" bash "$scratch/fixture/scripts/promote.sh" draft
 mv "$scratch/original-report.json" "$scratch/$kind.json";mv "$scratch/original-acceptance.json" "$scratch/acceptance.json";mv "$scratch/original-approval.json" "$scratch/draft-approval.json"
done
reset
jq -n '{object:{type:"commit",sha:("c"*40)}}' >"$scratch/remote/tag.json"
expect 1 'conflicting tag refuses draft creation' "${invoke[@]}" RELEASE_PROMOTION_OUTPUT="$scratch/tag-conflict" bash "$scratch/fixture/scripts/promote.sh" draft
reset
cp "$scratch/draft-approval.json" "$scratch/original-approval.json"
jq '.status="not-approved"' "$scratch/original-approval.json" >"$scratch/draft-approval.json"
cat "$scratch/original-approval.json" >>"$scratch/draft-approval.json"
expect 1 'concatenated approval refuses all remote calls' "${invoke[@]}" RELEASE_PROMOTION_OUTPUT="$scratch/concatenated-approval" bash "$scratch/fixture/scripts/promote.sh" draft
[[ ! -s "$scratch/gh-calls" ]] || failed=1
mv "$scratch/original-approval.json" "$scratch/draft-approval.json"
reset
for kind in cli tui startup;do
 cp "$scratch/$kind.json" "$scratch/original-report.json"
 cp "$scratch/acceptance.json" "$scratch/original-acceptance.json";cp "$scratch/draft-approval.json" "$scratch/original-approval.json"
 if [[ "$kind" == startup ]];then mutation='.Go="wrong"';else mutation='.passed=false';fi
 jq "$mutation" "$scratch/original-report.json" >"$scratch/$kind.json"
 cat "$scratch/original-report.json" >>"$scratch/$kind.json"
 hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/$kind.json")
 jq --arg kind "$kind" --arg hash "$hash" '.reports[$kind].sha256=$hash' "$scratch/acceptance.json" >"$scratch/new";mv "$scratch/new" "$scratch/acceptance.json"
 hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/acceptance.json")
 jq --arg hash "$hash" '.acceptance_sha256=$hash' "$scratch/draft-approval.json" >"$scratch/new";mv "$scratch/new" "$scratch/draft-approval.json"
 expect 1 "concatenated $kind report refuses all remote calls" "${invoke[@]}" RELEASE_PROMOTION_OUTPUT="$scratch/concatenated-$kind" bash "$scratch/fixture/scripts/promote.sh" draft
 [[ ! -s "$scratch/gh-calls" ]] || failed=1
 mv "$scratch/original-report.json" "$scratch/$kind.json";mv "$scratch/original-acceptance.json" "$scratch/acceptance.json";mv "$scratch/original-approval.json" "$scratch/draft-approval.json"
 reset
done
path="$scratch/gates/native-linux-amd64.json"
cp "$path" "$scratch/original-gate.json"
cp "$scratch/acceptance.json" "$scratch/original-acceptance.json";cp "$scratch/draft-approval.json" "$scratch/original-approval.json"
jq '.archive_sha256=("c"*64)' "$scratch/original-gate.json" >"$path"
cat "$scratch/original-gate.json" >>"$path"
hash=$(bash "$root/scripts/test/sha256.sh" "$path")
jq --arg hash "$hash" '(.gates[]|select(.id=="native/linux/amd64")|.receipt.sha256)=$hash' "$scratch/acceptance.json" >"$scratch/new";mv "$scratch/new" "$scratch/acceptance.json"
hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/acceptance.json")
jq --arg hash "$hash" '.acceptance_sha256=$hash' "$scratch/draft-approval.json" >"$scratch/new";mv "$scratch/new" "$scratch/draft-approval.json"
expect 1 'concatenated native receipt refuses all remote calls' "${invoke[@]}" RELEASE_PROMOTION_OUTPUT="$scratch/concatenated-gate" bash "$scratch/fixture/scripts/promote.sh" draft
[[ ! -s "$scratch/gh-calls" ]] || failed=1
mv "$scratch/original-gate.json" "$path";mv "$scratch/original-acceptance.json" "$scratch/acceptance.json";mv "$scratch/original-approval.json" "$scratch/draft-approval.json"
reset
reset
expect 0 'lost create/upload responses reconcile exact accepted bytes' "${invoke[@]}" FIXTURE_LOST_CREATE=1 FIXTURE_LOST_UPLOAD=1 RELEASE_PROMOTION_OUTPUT="$scratch/draft" bash "$scratch/fixture/scripts/promote.sh" draft
[[ "$(jq -r '.status' "$scratch/draft/receipt.json" 2>/dev/null || true)" == draft-complete ]] || failed=1
if [[ ! -s "$scratch/remote/tag.json" ]];then printf 'FAIL: complete draft lacks accepted tag\n' >&2;failed=1;fi
[[ "$(rg -c '^api https://uploads.github.com/.* --method POST' "$scratch/gh-calls" || true)" == 9 ]] || failed=1
[[ ! -e "$scratch/candidate/assets/marker" ]] || failed=1
expect 0 'matching complete draft resumes without uploads' "${invoke[@]}" RELEASE_PROMOTION_OUTPUT="$scratch/draft-retry" bash "$scratch/fixture/scripts/promote.sh" draft
[[ "$(rg -c '^api https://uploads.github.com/.* --method POST' "$scratch/gh-calls" || true)" == 9 ]] || failed=1
expect 1 'remote tampering refuses publication' "${invoke[@]}" FIXTURE_BAD_READBACK=1 RELEASE_AUTHORIZATION="$scratch/publish-approval.json" RELEASE_PROMOTION_OUTPUT="$scratch/tampered" bash "$scratch/fixture/scripts/promote.sh" publish
expect 0 'lost publication response readback confirms publication' "${invoke[@]}" FIXTURE_LOST_PUBLISH=1 RELEASE_AUTHORIZATION="$scratch/publish-approval.json" RELEASE_PROMOTION_OUTPUT="$scratch/published" bash "$scratch/fixture/scripts/promote.sh" publish
expect 1 'published release refuses all draft writes' "${invoke[@]}" RELEASE_PROMOTION_OUTPUT="$scratch/public-draft" bash "$scratch/fixture/scripts/promote.sh" draft
expect 0 'read-only public readback verifies nine exact assets' "${invoke[@]}" RELEASE_AUTHORIZATION= RELEASE_PROMOTION_OUTPUT="$scratch/readback" bash "$scratch/fixture/scripts/promote.sh" readback
expect 1 'retained output is not overwritten' "${invoke[@]}" RELEASE_PROMOTION_OUTPUT="$scratch/readback" bash "$scratch/fixture/scripts/promote.sh" readback
if rg -q -- '--clobber|--method DELETE|git push' "$scratch/gh-calls";then failed=1;fi
exit "$failed"
