#!/usr/bin/env bash
set -euo pipefail
jq_binary_option=''
case "${OSTYPE:-}" in msys*|cygwin*) jq_binary_option=--binary ;; esac
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
die() { printf 'Native candidate: %s\n' "$*" >&2;exit 1; }
digest() {
 if command -v sha256sum >/dev/null 2>&1;then sha256sum "$1" | awk '{print $1}';else shasum -a 256 "$1" | awk '{print $1}';fi
}
contract() {
 local guard="github.event_name == 'workflow_dispatch' && github.repository == 'newbpydev/tusk' && github.ref == 'refs/heads/main' && github.workflow_ref == 'newbpydev/tusk/.github/workflows/native-candidate.yml@refs/heads/main' && github.workflow_sha == github.sha"
 jq ${jq_binary_option:+"--binary"} -e --slurpfile pins "$root/scripts/tool-versions.json" '
  .on|keys==["workflow_dispatch"]
 ' "$root/.github/workflows/native-candidate.yml" >/dev/null || die 'manual dispatch required'
 jq ${jq_binary_option:+"--binary"} -e --arg guard "$guard" --slurpfile pins "$root/scripts/tool-versions.json" '
  .permissions=={contents:"read",actions:"read",attestations:"read"} and
  .jobs.native.if==$guard and
  (.on.workflow_dispatch.inputs|keys)==["candidate_run_attempt","candidate_run_id","manifest_sha256","source_sha","version"] and
  all(.on.workflow_dispatch.inputs[];.required==true and .type=="string") and
  (.jobs|keys)==["native"] and .jobs.native["timeout-minutes"]==30 and
  .jobs.native.defaults.run.shell=="bash" and
  .jobs.native["runs-on"]=="${{ matrix.runner }}" and
  .jobs.native.strategy["fail-fast"]==false and
  .jobs.native.strategy.matrix.include==[
   {runner:"ubuntu-24.04",target:"linux/amd64"},
   {runner:"ubuntu-24.04-arm",target:"linux/arm64"},
   {runner:"macos-15-intel",target:"darwin/amd64"},
   {runner:"macos-15",target:"darwin/arm64"},
   {runner:"windows-2025",target:"windows/amd64"}] and
  (.jobs.native|has("permissions")|not) and (.jobs.native|has("secrets")|not) and
  (.jobs.native.steps|length)==6 and
  all(.jobs.native.steps[]; (has("continue-on-error")|not)) and
  all(.jobs.native.steps[]|select(has("run"));(.run|contains("${{")|not)) and
  .jobs.native.steps[0].uses==("actions/checkout@"+$pins[0].actions.checkout.sha) and
  .jobs.native.steps[0].with=={ref:"${{ github.workflow_sha }}","persist-credentials":false,"fetch-depth":0} and
  .jobs.native.steps[1].uses==("actions/setup-go@"+$pins[0].actions.setup_go.sha) and
  .jobs.native.steps[1].with=={"go-version":$pins[0].go.release,cache:true} and
  .jobs.native.steps[2].if=="runner.os == '\''Windows'\''" and
  .jobs.native.steps[2].run=="test -x /c/mingw64/bin/make.exe\nprintf '\''%s\\n'\'' '\''C:/mingw64/bin'\'' >> \"$GITHUB_PATH\"" and
  .jobs.native.steps[3].uses==("actions/download-artifact@"+$pins[0].release_actions.download_artifact.sha) and
  .jobs.native.steps[3].with=={name:"tusk-candidate-${{ inputs.candidate_run_id }}-${{ inputs.candidate_run_attempt }}",path:"${{ runner.temp }}/candidate","run-id":"${{ inputs.candidate_run_id }}","github-token":"${{ github.token }}",repository:"newbpydev/tusk"} and
  .jobs.native.steps[4].run=="make native-candidate-smoke RELEASE_NATIVE_OUTPUT=\"$RUNNER_TEMP/native\"" and
  .jobs.native.steps[4].env=={
   GH_TOKEN:"${{ github.token }}",GOTOOLCHAIN:"local",CGO_ENABLED:"0",
   CANDIDATE_DIR:"${{ runner.temp }}/candidate",CANDIDATE_RUN_ID:"${{ inputs.candidate_run_id }}",
   CANDIDATE_MANIFEST_SHA256:"${{ inputs.manifest_sha256 }}",
   RELEASE_SHA:"${{ inputs.source_sha }}",RELEASE_VERSION:"${{ inputs.version }}",
   RELEASE_NATIVE_TARGET:"${{ matrix.target }}"} and
  .jobs.native.steps[5].if=="always()" and
  .jobs.native.steps[5].uses==("actions/upload-artifact@"+$pins[0].release_actions.upload_artifact.sha) and
  .jobs.native.steps[5].with=={name:"native-${{ matrix.runner }}-${{ github.run_id }}-${{ github.run_attempt }}",path:"${{ runner.temp }}/native","if-no-files-found":"warn","retention-days":90,overwrite:false}
 ' "$root/.github/workflows/native-candidate.yml" >/dev/null || die 'unsafe native candidate workflow'
}
smoke() {
 local candidate output parent manifest target archive archive_hash archive_expected expected binary hash status=0
 candidate=$(cd "${CANDIDATE_DIR:?}" && pwd -P)
 [[ -n "${RELEASE_NATIVE_OUTPUT:-}" ]] || die 'new retained output required'
 parent=$(cd "$(dirname "$RELEASE_NATIVE_OUTPUT")" && pwd -P) || die 'output parent missing'
 output="$parent/$(basename "$RELEASE_NATIVE_OUTPUT")"
 case "$output" in "$root"|"$root"/*|"$candidate"|"$candidate"/*) die 'output must be outside candidate/source storage';; esac
 mkdir "$output" || die 'output collision'
 export CANDIDATE_DIR="$candidate" CANDIDATE_VERIFICATION_DIR="$output/verification"
 # Verification inspects every archive and authenticates source/run/subject
 # bindings before extracting or running any downloaded executable.
 make -C "$root" verify-candidate >"$output/verification.log" 2>&1
 manifest="$candidate/assets/release-manifest.json"
 local host=() line
 while IFS= read -r line;do host+=("$line");done < <(go env GOHOSTOS GOHOSTARCH)
 [[ ${#host[@]} == 2 ]] || die 'native host identity unavailable'
 target="${host[0]}/${host[1]}"
 [[ "$target" == "${RELEASE_NATIVE_TARGET:-$target}" ]] || die 'runner target differs from native host'
 archive=$(jq ${jq_binary_option:+"--binary"} -er --arg target "$target" '.targets[]|select(.target==$target)|.archive' "$manifest")
 expected=$(jq ${jq_binary_option:+"--binary"} -er --arg target "$target" '.targets[]|select(.target==$target)|.executable_sha256' "$manifest")
 archive_expected=$(jq ${jq_binary_option:+"--binary"} -er --arg archive "$archive" '.files_sha256[$archive]' "$manifest")
 archive_hash=$(digest "$candidate/assets/$archive")
 [[ "$archive_hash" == "$archive_expected" ]] || die 'selected archive differs from verified manifest'
 mkdir "$output/payload"
 case "$target" in
  windows/amd64) unzip -q "$candidate/assets/$archive" -d "$output/payload";binary="$output/payload/tusk.exe" ;;
  linux/amd64|linux/arm64|darwin/amd64|darwin/arm64) tar -xzf "$candidate/assets/$archive" -C "$output/payload";binary="$output/payload/tusk" ;;
  *) die 'unsupported native target';;
 esac
 hash=$(digest "$binary")
 [[ "$hash" == "$expected" ]] || die 'extracted executable differs from verified target'
 export RELEASE_BINARY="$binary" RELEASE_MANIFEST="$manifest" CANDIDATE_VERIFICATION_RECEIPT="$output/verification/receipt.json" RELEASE_EVIDENCE_SCOPE=trusted-candidate RELEASE_SMOKE_OUTPUT="$output/smoke"
 make -C "$root" release-smoke >"$output/smoke.log" 2>&1 || status=$?
 [[ "$(digest "$binary")" == "$hash" ]] || die 'executable changed during native checks'
 [[ "$status" == 0 ]] || return "$status"
 jq ${jq_binary_option:+"--binary"} -n --arg target "$target" --arg hash "$hash" --arg archive "$archive" --arg archive_hash "$archive_hash" --arg manifest "$CANDIDATE_MANIFEST_SHA256" --arg sha "$RELEASE_SHA" --arg version "$RELEASE_VERSION" --arg host "$(uname -s)" --argjson run "$CANDIDATE_RUN_ID" --slurpfile smoke "$output/smoke/receipt.json" '{schema:1,status:"automated-passed",scope:"trusted-candidate",target:$target,executable_sha256:$hash,archive:$archive,archive_sha256:$archive_hash,manifest_sha256:$manifest,source_sha:$sha,version:$version,candidate_run_id:$run,host:$host,observed_version:$smoke[0].observed_version,manual_terminal:"pending",terminal_automation:"Linux child PTY only; macOS/Windows visual console checks remain separate"}' >"$output/receipt.json"
}
case "${1:-contract}" in
 contract) contract ;;
 smoke) smoke ;;
 *) die 'expected contract or smoke';;
esac
