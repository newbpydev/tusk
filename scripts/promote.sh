#!/usr/bin/env bash
jq_binary_option=''
case "${OSTYPE:-}" in msys*|cygwin*) jq_binary_option=--binary ;; esac
# Maintainer tooling only. APIs receive structured JSON and literal argument arrays.
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
# shellcheck source=scripts/json_check.sh
source "$root/scripts/json_check.sh"
repo=newbpydev/tusk
export GH_HOST=github.com
unset GH_DEBUG DEBUG
action=${1:-prepare}
die() { printf 'Promotion: %s\n' "$*" >&2;exit 1; }
digest() { if command -v sha256sum >/dev/null 2>&1;then sha256sum "$1"|awk '{print $1}';else shasum -a 256 "$1"|awk '{print $1}';fi; }
regular() { [[ -f "$1" && ! -L "$1" ]]; }
[[ "$action" == prepare || "$action" == draft || "$action" == publish || "$action" == readback ]] || die 'expected prepare, draft, publish or readback'
[[ "${RELEASE_VERSION:-}" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ && "${RELEASE_SHA:-}" =~ ^[0-9a-f]{40}$ && "${CANDIDATE_RUN_ID:-}" =~ ^[1-9][0-9]*$ && "${CANDIDATE_MANIFEST_SHA256:-}" =~ ^[0-9a-f]{64}$ ]] || die 'explicit version/source/run/manifest identity required'
candidate=$(cd "${CANDIDATE_DIR:?}" && pwd -P)
manifest="$candidate/assets/release-manifest.json"
json_object "$manifest" && [[ "$(digest "$manifest")" == "$CANDIDATE_MANIFEST_SHA256" ]] || die 'selected manifest changed'
jq ${jq_binary_option:+"--binary"} -e --arg sha "$RELEASE_SHA" --arg version "${RELEASE_VERSION#v}" '.mode=="candidate" and .source_sha==$sha and .version==$version' "$manifest" >/dev/null || die 'candidate identity differs'
bash "$root/scripts/release_check.sh" verify "$candidate/assets"
bash "$root/scripts/release_check.sh" check-cask "$manifest" "$candidate/homebrew/Casks/tusk.rb"
cask_hash=$(digest "$candidate/homebrew/Casks/tusk.rb")
[[ -n "${RELEASE_PROMOTION_OUTPUT:-}" ]] || die 'new retained output required'
parent=$(cd "$(dirname "$RELEASE_PROMOTION_OUTPUT")" && pwd -P) || die 'output parent missing'
output="$parent/$(basename "$RELEASE_PROMOTION_OUTPUT")"
case "$output" in "$candidate"|"$candidate"/*|"$root"|"$root"/*) die 'output must be outside candidate/source storage';; esac
mkdir "$output" || die 'output exists; use a fresh readback directory'
# Failures keep their diagnostics; no deletion or automatic HTTP write retry.
trap 'printf "{\"action\":\"%s\",\"exit_status\":%s}\n" "$action" "$?" >"$output/result.json"' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
jq ${jq_binary_option:+"--binary"} --arg manifest "$CANDIDATE_MANIFEST_SHA256" '.files_sha256 + {"checksums.txt":.checksums_sha256,"release-manifest.json":$manifest}' "$manifest" >"$output/files.json"
[[ "$(jq ${jq_binary_option:+"--binary"} 'length' "$output/files.json")" == 9 ]] || die 'exact nine public assets required'
if [[ "$action" == prepare ]];then
 jq ${jq_binary_option:+"--binary"} -n --slurpfile files "$output/files.json" --arg repo "$repo" --arg version "$RELEASE_VERSION" --arg sha "$RELEASE_SHA" --arg hash "$CANDIDATE_MANIFEST_SHA256" --arg cask "$cask_hash" --argjson run "$CANDIDATE_RUN_ID" '{schema:1,status:"prepared-unaccepted",repository:$repo,version:$version,source_sha:$sha,manifest_sha256:$hash,cask_sha256:$cask,run_id:$run,assets:$files[0],trusted_verification:"pending",native_terminal_performance_cask:"pending",owner_authorization:"pending"}' >"$output/plan.json"
 printf 'Prepared reviewable asset plan; no hosted/native acceptance or mutation.\n';exit 0
fi
# Human acceptance is hash-bound in a separate record of explicit owner approval.
if ! json_object "${RELEASE_ACCEPTANCE:-}" || ! regular "${RELEASE_NOTES:-}";then die 'accepted gate inventory and reviewed notes required';fi
acceptance_hash=$(digest "$RELEASE_ACCEPTANCE");notes_hash=$(digest "$RELEASE_NOTES")
jq ${jq_binary_option:+"--binary"} -e --arg sha "$RELEASE_SHA" --arg version "$RELEASE_VERSION" --arg hash "$CANDIDATE_MANIFEST_SHA256" --arg cask "$cask_hash" --argjson run "$CANDIDATE_RUN_ID" '
 .schema==1 and .status=="accepted" and .repository=="newbpydev/tusk" and .source_sha==$sha and .version==$version and .manifest_sha256==$hash and .cask_sha256==$cask and .run_id==$run and .prepublication_scenarios_reviewed==true and
 ([.gates[].id]|sort)==(["native/linux/amd64","native/linux/arm64","native/darwin/amd64","native/darwin/arm64","native/windows/amd64","cli-reference","tui-reference","cask/darwin/amd64","cask/darwin/arm64"]|sort) and
 all(.gates[];.passed==true and (.receipt.path|type=="string" and length>0) and (.receipt.sha256|test("^[0-9a-f]{64}$")))
' "$RELEASE_ACCEPTANCE" >/dev/null || die 'native/terminal/performance/cask gates incomplete or stale'
while IFS= read -r gate;do
 path=$(jq ${jq_binary_option:+"--binary"} -r '.receipt.path' <<<"$gate");hash=$(jq ${jq_binary_option:+"--binary"} -r '.receipt.sha256' <<<"$gate")
 json_object "$path" && [[ "$(digest "$path")" == "$hash" ]] || die 'gate receipt changed or absent'
 jq ${jq_binary_option:+"--binary"} -e --arg id "$(jq ${jq_binary_option:+"--binary"} -r '.id' <<<"$gate")" --arg sha "$RELEASE_SHA" --arg version "$RELEASE_VERSION" --arg hash "$CANDIDATE_MANIFEST_SHA256" --argjson run "$CANDIDATE_RUN_ID" '.status=="accepted" and .scope=="trusted-candidate" and .gate==$id and .source_sha==$sha and .version==$version and .manifest_sha256==$hash and .run_id==$run' "$path" >/dev/null || die 'gate receipt does not describe this trusted candidate'
 id=$(jq ${jq_binary_option:+"--binary"} -r '.id' <<<"$gate")
 if [[ "$id" == native/* ]];then
  jq ${jq_binary_option:+"--binary"} -e --arg target "${id#native/}" --slurpfile m "$manifest" '
   . as $r | [$m[0].targets[]|select(.target==$target)] as $t |
   ($t|length)==1 and $r.target==$target and $r.executable_sha256==$t[0].executable_sha256 and
   $r.archive_sha256==$m[0].files_sha256[$t[0].archive] and $r.observed_version==$m[0].version
  ' "$path" >/dev/null || die 'native receipt executable/archive differs from candidate target'
 fi
done < <(jq ${jq_binary_option:+"--binary"} -c '.gates[]' "$RELEASE_ACCEPTANCE")
if [[ "$action" != readback ]];then
 json_object "${RELEASE_AUTHORIZATION:-}" || die 'record of explicit owner authorization required'
 jq ${jq_binary_option:+"--binary"} -e --arg action "$action" --arg sha "$RELEASE_SHA" --arg version "$RELEASE_VERSION" --arg hash "$CANDIDATE_MANIFEST_SHA256" --arg cask "$cask_hash" --arg accepted "$acceptance_hash" --arg notes "$notes_hash" --argjson run "$CANDIDATE_RUN_ID" '
 keys==(["schema","status","action","repository","source_sha","version","manifest_sha256","cask_sha256","run_id","acceptance_sha256","notes_sha256","exclusive_release_window","approval_reference"]|sort) and
 .schema==1 and .status=="owner-approved" and .action==$action and .repository=="newbpydev/tusk" and .source_sha==$sha and .version==$version and .manifest_sha256==$hash and .cask_sha256==$cask and .run_id==$run and .acceptance_sha256==$accepted and .notes_sha256==$notes and .exclusive_release_window==true and (.approval_reference|type=="string" and length>0)
 ' "$RELEASE_AUTHORIZATION" >/dev/null || die 'approval does not authorize this exact action/acceptance/notes'
fi
# Clean tools must be at the frozen source, or contain only later governance prose.
[[ -z "$(git -C "$root" status --porcelain)" ]] || die 'promotion tooling checkout must be clean'
git -C "$root" merge-base --is-ancestor "$RELEASE_SHA" HEAD || die 'candidate source is outside current checkout history'
git -C "$root" diff --name-only "$RELEASE_SHA" HEAD >"$output/source-diff.txt" || die 'source history could not be verified'
while IFS= read -r changed;do
 case "$changed" in README.md|MASTERPLAN.md|CONCEPTS.md|docs/*.md|docs/verification-evidence/*) ;; '') ;; *) die "build/acceptance tooling changed after source freeze: $changed";; esac
done <"$output/source-diff.txt"
CANDIDATE_VERIFICATION_DIR="$output/verification" make -C "$root" verify-candidate >"$output/verification.log" 2>&1
json_object "$output/verification/receipt.json" || die 'invalid verification receipt document'
jq ${jq_binary_option:+"--binary"} -e --arg hash "$CANDIDATE_MANIFEST_SHA256" --arg cask "$cask_hash" '.status=="verified" and .manifest_sha256==$hash and .cask_sha256==$cask' "$output/verification/receipt.json" >/dev/null || die 'fresh cryptographic candidate verification failed'
make -C "$root" homebrew-destination >"$output/tap.log" 2>&1
# Performance reports retain every sample and compiler/binary identities.
linux_hash=$(jq ${jq_binary_option:+"--binary"} -r '.targets[]|select(.target=="linux/amd64")|.executable_sha256' "$manifest")
for kind in cli tui startup;do
 path=$(jq ${jq_binary_option:+"--binary"} -r --arg kind "$kind" '.reports[$kind].path' "$RELEASE_ACCEPTANCE")
 hash=$(jq ${jq_binary_option:+"--binary"} -r --arg kind "$kind" '.reports[$kind].sha256' "$RELEASE_ACCEPTANCE")
 json_object "$path" && [[ "$(digest "$path")" == "$hash" ]] || die 'retained performance report changed/absent'
 case "$kind" in
 cli) jq ${jq_binary_option:+"--binary"} -e --arg hash "$linux_hash" --arg go "go$(jq ${jq_binary_option:+"--binary"} -r '.compiler' "$manifest")" '
 def names: ["--help/invalid-config=false","--help/invalid-config=true","--version/invalid-config=false","--version/invalid-config=true"] + [[0,100,1000][] as $n|["list","tree","stats","history"][] as $cmd|[false,true][]|"\($n)/\($cmd)/json=\(.)"];
 .passed==true and .manifest.sha256==$hash and .manifest.go==$go and .manifest.selected_acceptance_profile=="reference" and (.cases|length)==84 and
 ([.cases[]|[.name,.run]]|unique|length)==84 and ([.cases[].name]|unique|sort)==(names|sort) and
 all(.cases[]; .samples_ns|sort as $samples | . as $raw | ($raw|length)==100 and all($raw[];type=="number" and .>0 and floor==.)) and
 all(.cases[]; . as $c | ($c.samples_ns|sort) as $s | (if $c.name|startswith("--") then [5000000,7500000,10000000,15000000] else [15000000,20000000,30000000,50000000] end) as $limits |
 $c.passed==true and ($c.error // "")=="" and $c.output_bytes>0 and ($c.warmup_ns|length)==5 and all($c.warmup_ns[];type=="number" and .>0 and floor==.) and ($c.run>=1 and $c.run<=3 and ($c.run|floor)==$c.run) and
 $c.p90_ns==$s[89] and $c.p95_ns==$s[94] and $c.p99_ns==$s[98] and $c.max_ns==$s[99] and $s[89]<$limits[0] and $s[94]<$limits[1] and $s[98]<$limits[2] and $s[99]<$limits[3])
 ' "$path" >/dev/null || die 'reference CLI samples/case matrix failed or have wrong binary/compiler';;
 tui) jq ${jq_binary_option:+"--binary"} -e --arg hash "$linux_hash" --arg sha "$RELEASE_SHA" --arg go "go$(jq ${jq_binary_option:+"--binary"} -r '.compiler' "$manifest")" '
 def names: [[0,100,1000][] as $n|[[80,24],[120,40],[200,60]][] as $s|["prepare","view"][]|"\(.)/\($n)-tasks/\($s[0])x\($s[1])"]+[ ["80x24","120x40","200x60"][]|"calendar/1000-tasks/"+.]+["markdown/32768-bytes","markdown/1048576-bytes","disk-refresh/100-tasks","stress/10000-tasks-1MiB-note-1000-events","stress/read-only-form-1MiB"];
 .passed==true and .manifest.binary_sha256==$hash and .manifest.revision==$sha and .manifest.source_diff=="" and .manifest.go==$go and (.cases|length)==78 and
 ([.cases[]|[.name,.run]]|unique|length)==78 and ([.cases[].name]|unique|sort)==(names|sort) and
 all(.cases[];.passed==true and (.run>=1 and .run<=3 and (.run|floor)==.run) and .consumed_output_size>0 and (.warmup_ns|length)==5 and (.samples_ns|length)==100 and (.allocations|length)==100 and (.allocated_bytes|length)==100 and all(.warmup_ns[],.samples_ns[];type=="number" and .>0 and floor==.) and all(.allocations[],.allocated_bytes[];type=="number" and .>=0 and floor==.) and
 (.preparation_budget==( .name|startswith("prepare/") or startswith("calendar/"))) and ((.preparation_budget|not) or ((.samples_ns|sort) as $s|$s[94]<16700000 and $s[98]<33300000 and $s[99]<50000000)))
 ' "$path" >/dev/null || die 'TUI samples/case matrix failed or have wrong source/binary/compiler';;
 startup) jq ${jq_binary_option:+"--binary"} -e --arg hash "$linux_hash" --arg go "go$(jq ${jq_binary_option:+"--binary"} -r '.compiler' "$manifest")" '.BinarySHA256==$hash and .Go==$go and (.Runs|length)==3 and ([.Runs[].run]|sort)==[1,2,3] and all(.Runs[];(.warmup_ns|length)==5 and (.samples_ns|length)==100 and (.child_max_rss_kib|length)==100 and all(.warmup_ns[],.samples_ns[];type=="number" and .>0 and floor==.) and all(.child_max_rss_kib[];type=="number" and .>=0 and floor==.))' "$path" >/dev/null || die 'packaged startup report incomplete/wrong binary/compiler';;
 esac
done
# Reconcile current remote main as well as the local frozen tools.
# shellcheck source=scripts/gh_deadline.sh
source "$root/scripts/gh_deadline.sh"
gh_deadline_setup "$root" "$output" || die 'could not prepare bounded GitHub driver'
gh api "https://api.github.com/repos/$repo/git/ref/heads/main" >"$output/main.json" || die 'main readback unavailable'
main_sha=$(jq ${jq_binary_option:+"--binary"} -er '.object|select(.type=="commit")|.sha|select(test("^[0-9a-f]{40}$"))' "$output/main.json") || die 'invalid main identity'
gh api "https://api.github.com/repos/$repo/compare/$RELEASE_SHA...$main_sha" >"$output/main-history.json" || die 'main history readback unavailable'
jq ${jq_binary_option:+"--binary"} -e --arg sha "$RELEASE_SHA" '(.status=="identical" or .status=="ahead") and .merge_base_commit.sha==$sha and (.files|type=="array" and length<300) and all(.files[];.filename|test("^(README[.]md|MASTERPLAN[.]md|CONCEPTS[.]md|docs/.*[.]md|docs/verification-evidence/.*)$"))' "$output/main-history.json" >/dev/null || die 'accepted source left main history or main changed build/acceptance inputs'
calls=0
read_release() {
 calls=$((calls+1));local file="$output/releases-$calls.json"
 gh api "https://api.github.com/repos/$repo/releases?per_page=100" --paginate --slurp >"$file" 2>"$file.stderr" || die 'release readback unavailable; no blind write retry'
 jq ${jq_binary_option:+"--binary"} --arg version "$RELEASE_VERSION" '[.[][]|select(.tag_name==$version)]' "$file" >"$output/current-release.json"
 [[ "$(jq ${jq_binary_option:+"--binary"} 'length' "$output/current-release.json")" -le 1 ]] || die 'multiple matching releases; explicit repair required'
}
read_tag() {
 calls=$((calls+1));local file="$output/tag-$calls.http" code=0
 gh api "https://api.github.com/repos/$repo/git/ref/tags/$RELEASE_VERSION" --include >"$file" 2>"$file.stderr" || code=$?
 http=$(head -1 "$file" | tr -d '\r')
 awk 'body{print} /^\r?$/{body=1}' "$file" >"$output/tag.json"
 if [[ "$http" =~ ^HTTP/[0-9.]+\ 404\  && "$code" != 0 ]];then tag_exists=false;return;fi
 [[ "$http" =~ ^HTTP/[0-9.]+\ 200\  && "$code" == 0 ]] || die 'tag state unavailable; no write allowed'
 tag_exists=true;type=$(jq ${jq_binary_option:+"--binary"} -r '.object.type' "$output/tag.json");sha=$(jq ${jq_binary_option:+"--binary"} -r '.object.sha' "$output/tag.json")
 for depth in 1 2 3 4 5;do
  [[ "$sha" =~ ^[0-9a-f]{40}$ ]] || die 'invalid tag object'
  if [[ "$type" == commit ]];then [[ "$sha" == "$RELEASE_SHA" ]] || die 'existing tag conflicts with accepted source';return;fi
  [[ "$type" == tag ]] || die 'unsupported tag object'
  gh api "https://api.github.com/repos/$repo/git/tags/$sha" >"$output/annotated-$depth.json" || die 'annotated tag read failed'
  type=$(jq ${jq_binary_option:+"--binary"} -r '.object.type' "$output/annotated-$depth.json");sha=$(jq ${jq_binary_option:+"--binary"} -r '.object.sha' "$output/annotated-$depth.json")
 done
 die 'tag chain exceeds bound'
}
release_identity() {
 jq ${jq_binary_option:+"--binary"} -e --arg version "$RELEASE_VERSION" --arg sha "$RELEASE_SHA" --rawfile notes "$RELEASE_NOTES" 'length==1 and .[0].tag_name==$version and .[0].target_commitish==$sha and .[0].name==$version and .[0].body==$notes and .[0].prerelease==false and (.[0].id|type=="number" and .>0 and floor==.)' "$output/current-release.json" >/dev/null || die 'existing release differs; no overwrite/repair inferred'
 id=$(jq ${jq_binary_option:+"--binary"} -r '.[0].id' "$output/current-release.json")
}
ensure_tag() {
 read_tag
 if [[ "$tag_exists" == false ]];then
  jq ${jq_binary_option:+"--binary"} -n --arg ref "refs/tags/$RELEASE_VERSION" --arg sha "$RELEASE_SHA" '{ref:$ref,sha:$sha}' >"$output/tag-request.json"
  gh api "https://api.github.com/repos/$repo/git/refs" --method POST --input "$output/tag-request.json" >"$output/tag-create-response.json" 2>"$output/tag-create-response.stderr" || true
  read_tag;[[ "$tag_exists" == true ]] || die 'tag creation uncertain; fresh readback required'
 fi
}
read_assets() {
 calls=$((calls+1));local file="$output/assets-$calls.json"
 gh api "https://api.github.com/repos/$repo/releases/$id/assets?per_page=100" --paginate --slurp >"$file" || die 'asset readback unavailable'
 jq ${jq_binary_option:+"--binary"} '[.[][]]' "$file" >"$output/current-assets.json"
 jq ${jq_binary_option:+"--binary"} -e --slurpfile files "$output/files.json" '([.[].name]|unique|length)==length and all(.[];(.id|type=="number" and .>0 and floor==.) and .state=="uploaded" and (.name as $name|$files[0]|has($name)))' "$output/current-assets.json" >/dev/null || die 'unexpected/duplicate/incomplete asset; explicit draft repair required'
}
check_asset() {
 local name=$1 asset_id file
 asset_id=$(jq ${jq_binary_option:+"--binary"} -r --arg name "$name" '.[]|select(.name==$name)|.id' "$output/current-assets.json")
 [[ "$asset_id" =~ ^[1-9][0-9]*$ ]] || die 'asset missing or ambiguous'
 mkdir -p "$output/readback"
 file="$output/readback/$name"
 gh api "https://api.github.com/repos/$repo/releases/assets/$asset_id" -H 'Accept: application/octet-stream' >"$file" || die 'asset download unavailable; no acceptance'
 [[ "$(digest "$file")" == "$(jq ${jq_binary_option:+"--binary"} -r --arg name "$name" '.[$name]' "$output/files.json")" ]] || die 'remote asset hash differs; no clobber/delete permitted'
}
complete_readback() {
 read_assets
 [[ "$(jq ${jq_binary_option:+"--binary"} 'length' "$output/current-assets.json")" == 9 ]] || die 'complete nine-file release required'
 while IFS= read -r name;do check_asset "$name";done < <(jq ${jq_binary_option:+"--binary"} -r 'keys[]' "$output/files.json")
 bash "$root/scripts/release_check.sh" verify "$output/readback"
}
read_tag;read_release
if [[ "$action" == draft ]];then
 if [[ "$(jq ${jq_binary_option:+"--binary"} 'length' "$output/current-release.json")" == 0 ]];then
  jq ${jq_binary_option:+"--binary"} -n --arg version "$RELEASE_VERSION" --arg sha "$RELEASE_SHA" --rawfile notes "$RELEASE_NOTES" '{tag_name:$version,target_commitish:$sha,name:$version,body:$notes,draft:true,prerelease:false,generate_release_notes:false}' >"$output/draft-request.json"
  status=0;gh api "https://api.github.com/repos/$repo/releases" --method POST --input "$output/draft-request.json" >"$output/draft-response.json" 2>"$output/draft-response.stderr" || status=$?
  printf '%s\n' "$status" >"$output/draft-response.status"
  read_release # Reconcile even a lost response; never repeat POST blindly.
 fi
 release_identity
 [[ "$(jq ${jq_binary_option:+"--binary"} -r '.[0].draft' "$output/current-release.json")" == true ]] || die 'published release is immutable to this draft operation'
 ensure_tag
 read_assets
 while IFS= read -r name;do
  if [[ "$(jq ${jq_binary_option:+"--binary"} --arg name "$name" '[.[]|select(.name==$name)]|length' "$output/current-assets.json")" == 0 ]];then
   read_release;release_identity
   [[ "$(jq ${jq_binary_option:+"--binary"} -r '.[0].draft' "$output/current-release.json")" == true ]] || die 'release became public; stop all asset writes'
   status=0
   gh api "https://uploads.github.com/repos/$repo/releases/$id/assets?name=$name" --method POST --input "$candidate/assets/$name" -H 'Content-Type: application/octet-stream' >"$output/upload-$name.json" 2>"$output/upload-$name.stderr" || status=$?
   printf '%s\n' "$status" >"$output/upload-$name.status"
   read_assets # A lost upload response is resolved by exact-byte readback.
  fi
  check_asset "$name"
 done < <(jq ${jq_binary_option:+"--binary"} -r 'keys[]' "$output/files.json")
 complete_readback
 read_release;release_identity
 [[ "$(jq ${jq_binary_option:+"--binary"} -r '.[0].draft' "$output/current-release.json")" == true ]] || die 'draft changed during verification'
 read_tag;[[ "$tag_exists" == true ]] || die 'complete draft tag absent'
 result=draft-complete
else
 release_identity;complete_readback
 if [[ "$action" == publish && "$(jq ${jq_binary_option:+"--binary"} -r '.[0].draft' "$output/current-release.json")" == true ]];then
  ensure_tag
  read_release;release_identity
  if [[ "$(jq ${jq_binary_option:+"--binary"} -r '.[0].draft' "$output/current-release.json")" == true ]];then
   printf '{"draft":false,"make_latest":"true"}\n' >"$output/publish-request.json"
   gh api "https://api.github.com/repos/$repo/releases/$id" --method PATCH --input "$output/publish-request.json" >"$output/publish-response.json" 2>"$output/publish-response.stderr" || true
  fi
  read_release;release_identity;complete_readback
  [[ "$(jq ${jq_binary_option:+"--binary"} -r '.[0].draft' "$output/current-release.json")" == false ]] || die 'publication not confirmed; no blind PATCH retry'
 fi
 result=$(jq ${jq_binary_option:+"--binary"} -r 'if .[0].draft then "draft-complete" else "published" end' "$output/current-release.json")
 read_tag
 [[ "$tag_exists" == true ]] || die 'accepted release tag absent'
fi
jq ${jq_binary_option:+"--binary"} -n --arg status "$result" --arg sha "$RELEASE_SHA" --arg version "$RELEASE_VERSION" --arg hash "$CANDIDATE_MANIFEST_SHA256" --arg cask "$cask_hash" --arg accepted "$acceptance_hash" --argjson run "$CANDIDATE_RUN_ID" --argjson id "$id" '{schema:1,status:$status,repository:"newbpydev/tusk",source_sha:$sha,version:$version,manifest_sha256:$hash,cask_sha256:$cask,run_id:$run,acceptance_sha256:$accepted,release_id:$id,assets_verified:9,tap_and_anonymous_install_and_metadata:"pending separate readback"}' >"$output/receipt.json"
printf 'Verified existing bytes: %s release %s.\n' "$result" "$id"
