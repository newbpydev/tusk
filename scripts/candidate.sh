#!/usr/bin/env bash
set -euo pipefail
jq_binary_option=''
case "${OSTYPE:-}" in msys*|cygwin*) jq_binary_option=--binary ;; esac
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
repo=newbpydev/tusk
export GH_HOST=github.com
unset GH_DEBUG DEBUG
workflow=.github/workflows/release.yml
die() { printf 'Candidate: %s\n' "$*" >&2; exit 1; }
digest() {
    if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | awk '{print $1}'
    else shasum -a 256 "$1" | awk '{print $1}'; fi
}
inputs() {
    [[ "${RELEASE_VERSION:-}" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] || die 'invalid intended version'
    [[ "${RELEASE_SHA:-}" =~ ^[0-9a-f]{40}$ ]] || die 'full source SHA required'
    [[ "${CANDIDATE_RUN_ID:-${GITHUB_RUN_ID:-}}" =~ ^[1-9][0-9]*$ ]] || die 'positive run ID required'
}
identity() {
    inputs
    [[ "${GITHUB_REPOSITORY:-}" == "$repo" && "${GITHUB_EVENT_NAME:-}" == workflow_dispatch && "${GITHUB_REF:-}" == refs/heads/main ]] || die 'only manual dispatch from trusted main is allowed'
    [[ "${GITHUB_SHA:-}" == "$RELEASE_SHA" && "${GITHUB_WORKFLOW_SHA:-}" == "$RELEASE_SHA" && "${GITHUB_WORKFLOW_REF:-}" == "$repo/$workflow@refs/heads/main" ]] || die 'dispatch, builder and requested source must be the same main SHA'
    [[ "${GITHUB_RUN_ATTEMPT:-}" =~ ^[1-9][0-9]*$ ]] || die 'positive attempt required'
}
contract() {
    jq ${jq_binary_option:+"--binary"} -e --slurpfile pins "$root/scripts/tool-versions.json" '
      keys == (["name","on","permissions","concurrency","jobs"]|sort) and
      (.on|keys) == ["workflow_dispatch"] and
      .on.workflow_dispatch.inputs.source_sha == {description:"Exact main dispatch SHA (40 lowercase hex)",required:true,type:"string"} and
      .on.workflow_dispatch.inputs.version == {description:"Intended vMAJOR.MINOR.PATCH",required:true,type:"string"} and
      (.on.workflow_dispatch.inputs|keys) == ["source_sha","version"] and
      .permissions == {contents:"read"} and
      .concurrency == {group:"candidate-${{ github.run_id }}", "cancel-in-progress":false} and
      (.jobs|keys) == ["build","ci","identity","provenance"] and
      .jobs.ci == {needs:["identity"],uses:"./.github/workflows/ci.yml",with:{source_sha:"${{ inputs.source_sha }}"}} and
      ($pins[0].release_actions|keys)==["attest_build_provenance","download_artifact","upload_artifact"] and
      all($pins[0].release_actions[];.sha|test("^[0-9a-f]{40}$")) and
      (.jobs.identity|keys)==(["runs-on","timeout-minutes","defaults","steps"]|sort) and
      (.jobs.build|keys)==(["runs-on","timeout-minutes","defaults","steps","needs","outputs"]|sort) and
      (.jobs.provenance|keys)==(["runs-on","timeout-minutes","defaults","steps","needs","permissions"]|sort) and
      all(.jobs[].steps[]?; (keys- ["name","run","env","uses","with","id"] | length)==0) and
      .jobs.build.outputs=={artifact_id:"${{ steps.upload.outputs.artifact-id }}",artifact_digest:"${{ steps.upload.outputs.artifact-digest }}"} and
      .jobs.identity.steps[1].env=={RELEASE_VERSION:"${{ inputs.version }}",RELEASE_SHA:"${{ inputs.source_sha }}"} and
      .jobs.build.steps[3].env==.jobs.identity.steps[1].env and .jobs.build.steps[4].env==.jobs.identity.steps[1].env and
      .jobs.provenance.steps[3].env=={GOTOOLCHAIN:"local"} and
      .jobs.build.needs == ["identity","ci"] and .jobs.provenance.needs == ["build"] and
      .jobs.provenance.permissions == {contents:"read",actions:"read","id-token":"write",attestations:"write"} and
      all(.jobs | to_entries[] | select(.key != "ci");
        .value["runs-on"] == "ubuntu-24.04" and .value["timeout-minutes"] == 30 and
        .value.defaults.run.shell == "bash" and (.value|has("if")|not) and
        (.key == "provenance" or (.value|has("permissions")|not))) and
      (.jobs.identity.steps|length) == 2 and (.jobs.build.steps|length) == 6 and (.jobs.provenance.steps|length) == 5 and
      all(.jobs[] | tostring; test("secrets\\.|--clobber|release create|git push";"i")|not) and
      all(.jobs[].steps[]? | select(has("uses"));
        .uses == ("actions/checkout@"+$pins[0].actions.checkout.sha) or
        .uses == ("actions/setup-go@"+$pins[0].actions.setup_go.sha) or
        .uses == ("actions/upload-artifact@"+$pins[0].release_actions.upload_artifact.sha) or
        .uses == ("actions/download-artifact@"+$pins[0].release_actions.download_artifact.sha) or
        .uses == ("actions/attest-build-provenance@"+$pins[0].release_actions.attest_build_provenance.sha)) and
      all(.jobs[].steps[]? | select((.uses? // "") | startswith("actions/checkout@")); .with == {ref:"${{ github.sha }}","persist-credentials":false,"fetch-depth":0}) and
      all(.jobs[].steps[]? | select(has("run")); (.run|contains("${{")|not) and (has("if")|not)) and
      .jobs.identity.steps[1].run == "make candidate-identity" and
      .jobs.build.steps[1].with == {"go-version":$pins[0].go.release,cache:true} and
      .jobs.build.steps[2].run == "make setup-sqlc setup-release release-check" and
      .jobs.build.steps[3].run == "make release-candidate RELEASE_OUTPUT=\"$RUNNER_TEMP/candidate\"" and
      .jobs.build.steps[4].run == "make record-candidate CANDIDATE_DIR=\"$RUNNER_TEMP/candidate\"" and
      .jobs.build.steps[5].id == "upload" and
      .jobs.build.steps[5].with == {name:"tusk-candidate-${{ github.run_id }}-${{ github.run_attempt }}",path:"${{ runner.temp }}/candidate/assets/*\n${{ runner.temp }}/candidate/candidate-run.json\n${{ runner.temp }}/candidate/homebrew/Casks/tusk.rb","if-no-files-found":"error","retention-days":90,"compression-level":0,overwrite:false} and
      .jobs.provenance.steps[1].with == {"go-version":$pins[0].go.release,cache:true} and
      .jobs.provenance.steps[2].with == {"artifact-ids":"${{ needs.build.outputs.artifact_id }}",path:"${{ runner.temp }}/candidate","merge-multiple":true} and
      .jobs.provenance.steps[3].run == "make verify-release RELEASE_ASSETS=\"$RUNNER_TEMP/candidate/assets\"\nmake check-homebrew RELEASE_MANIFEST=\"$RUNNER_TEMP/candidate/assets/release-manifest.json\" HOMEBREW_CASK=\"$RUNNER_TEMP/candidate/homebrew/Casks/tusk.rb\"" and
      .jobs.provenance.steps[4].with == {"subject-path":"${{ runner.temp }}/candidate/assets/*\n${{ runner.temp }}/candidate/candidate-run.json\n${{ runner.temp }}/candidate/homebrew/Casks/tusk.rb","push-to-registry":false} and
      (.jobs.build|has("env")|not) and (.jobs.identity|has("env")|not) and (.jobs.provenance|has("env")|not) and
      all(.jobs.build.steps[] | .env? // {}; has("GH_TOKEN")|not)
    ' "$root/.github/workflows/release.yml" >/dev/null || die 'workflow violates isolated trusted-candidate contract'
}
record() {
    identity
    local directory=${CANDIDATE_DIR:?} manifest hash
    manifest="$directory/assets/release-manifest.json"
    [[ -f "$manifest" && ! -L "$manifest" && ! -e "$directory/candidate-run.json" && ! -L "$directory/candidate-run.json" ]] || die 'manifest absent or run receipt collision'
    jq ${jq_binary_option:+"--binary"} -e --arg sha "$RELEASE_SHA" --arg version "${RELEASE_VERSION#v}" '.mode=="candidate" and .source_sha==$sha and .version==$version' "$manifest" >/dev/null || die 'manifest does not match dispatch'
    bash "$root/scripts/release_check.sh" check-cask "$manifest" "$directory/homebrew/Casks/tusk.rb"
    local cask_hash
    cask_hash=$(digest "$directory/homebrew/Casks/tusk.rb")
    hash=$(digest "$manifest")
    jq ${jq_binary_option:+"--binary"} -n --arg repo "$repo" --arg workflow "$workflow" --arg sha "$RELEASE_SHA" --arg version "$RELEASE_VERSION" --arg hash "$hash" --arg cask "$cask_hash" --argjson run "$GITHUB_RUN_ID" --argjson attempt "$GITHUB_RUN_ATTEMPT" '{schema:1,repository:$repo,workflow:$workflow,ref:"refs/heads/main",source_sha:$sha,version:$version,run_id:$run,run_attempt:$attempt,manifest_sha256:$hash,cask_sha256:$cask}' >"$directory/candidate-run.json"
    printf 'Candidate manifest SHA256: %s\n' "$hash"
}
verify() (
    inputs
    # shellcheck source=scripts/json_check.sh
    source "$root/scripts/json_check.sh"
    local directory=${CANDIDATE_DIR:?} output=${CANDIDATE_VERIFICATION_DIR:?} manifest="$CANDIDATE_DIR/assets/release-manifest.json" attempt artifact_id artifact_digest file hash
    [[ "${CANDIDATE_MANIFEST_SHA256:-}" =~ ^[0-9a-f]{64}$ ]] || die 'accepted manifest digest required'
    [[ -f "$manifest" && ! -L "$manifest" && "$(digest "$manifest")" == "$CANDIDATE_MANIFEST_SHA256" ]] || die 'manifest digest mismatch'
    json_object "$directory/candidate-run.json" || die 'run receipt missing or unsafe'
    jq ${jq_binary_option:+"--binary"} -e --arg sha "$RELEASE_SHA" --arg version "$RELEASE_VERSION" --arg hash "$CANDIDATE_MANIFEST_SHA256" --argjson run "$CANDIDATE_RUN_ID" 'keys==(["schema","repository","workflow","ref","source_sha","version","run_id","run_attempt","manifest_sha256","cask_sha256"]|sort) and .schema==1 and .repository=="newbpydev/tusk" and .workflow==".github/workflows/release.yml" and .ref=="refs/heads/main" and .source_sha==$sha and .version==$version and .run_id==$run and .manifest_sha256==$hash and (.run_attempt|type=="number" and .>0 and floor==.)' "$directory/candidate-run.json" >/dev/null || die 'untrusted run receipt'
    jq ${jq_binary_option:+"--binary"} -e --arg sha "$RELEASE_SHA" --arg version "${RELEASE_VERSION#v}" '.mode=="candidate" and .source_sha==$sha and .version==$version' "$manifest" >/dev/null || die 'manifest identity mismatch'
    [[ -f "$directory/homebrew/Casks/tusk.rb" && ! -L "$directory/homebrew/Casks/tusk.rb" ]] || die 'reviewed cask missing or unsafe'
    [[ "$(digest "$directory/homebrew/Casks/tusk.rb")" == "$(jq ${jq_binary_option:+"--binary"} -r '.cask_sha256' "$directory/candidate-run.json")" ]] || die 'cask differs from run receipt'
    attempt=$(jq ${jq_binary_option:+"--binary"} -r '.run_attempt' "$directory/candidate-run.json")
    directory=$(cd "$directory" && pwd -P) || die 'candidate directory missing'
    output="$(cd "$(dirname "$output")" && pwd -P)/$(basename "$output")" || die 'verification output parent missing'
    [[ "$output" != "$directory" && "$output" != "$directory/"* && "$output" != "$root" && "$output" != "$root/"* ]] || die 'verification must not mutate candidate or source storage'
    # Receipt directories are new, never replaced; failed boundaries stay inspectable.
    mkdir "$output" || die 'verification output collision/parent missing'
    # shellcheck source=scripts/gh_deadline.sh
    source "$root/scripts/gh_deadline.sh"
    gh_deadline_setup "$root" "$output" || die 'could not prepare bounded GitHub driver'
    gh --version >"$output/verifier-version.txt"
    gh api "https://api.github.com/repos/$repo/actions/runs/$CANDIDATE_RUN_ID/attempts/$attempt" >"$output/run.json"
    jq ${jq_binary_option:+"--binary"} -e --arg sha "$RELEASE_SHA" --argjson run "$CANDIDATE_RUN_ID" --argjson attempt "$attempt" '.id==$run and .run_attempt==$attempt and .path==".github/workflows/release.yml" and .event=="workflow_dispatch" and .head_branch=="main" and .head_sha==$sha and .status=="completed" and .conclusion=="success" and .repository.full_name=="newbpydev/tusk" and .head_repository.full_name=="newbpydev/tusk" and .head_repository.id==.repository.id' "$output/run.json" >/dev/null || die 'run is incomplete or untrusted'
    # Actions returns a minimal repository; default_branch needs its own readback.
    gh api "https://api.github.com/repos/$repo" >"$output/repository.json"
    jq ${jq_binary_option:+"--binary"} -e --slurpfile run "$output/run.json" '.full_name=="newbpydev/tusk" and .default_branch=="main" and .id==$run[0].repository.id' "$output/repository.json" >/dev/null || die 'repository/default branch changed'
    gh api "https://api.github.com/repos/$repo/actions/runs/$CANDIDATE_RUN_ID/attempts/$attempt/jobs?per_page=100" >"$output/jobs.json"
    jq ${jq_binary_option:+"--binary"} -e '.total_count==9 and (.jobs|length)==9 and all(.jobs[];.status=="completed" and .conclusion=="success") and (["identity","build","provenance","linux/amd64 — release compiler","linux/arm64 — release compiler","darwin/amd64 — release compiler","darwin/arm64 — release compiler","windows/amd64 — release compiler","Linux — minimum compiler and five-target cross-builds"] | all(.[]; . as $name | $jobs[0].jobs | map(select(.name==$name or (.name|endswith(" / "+$name)))) | length==1))' --slurpfile jobs "$output/jobs.json" "$output/jobs.json" >/dev/null || die 'same-run native/minimum/build/provenance gates incomplete'
    gh api "https://api.github.com/repos/$repo/actions/runs/$CANDIDATE_RUN_ID/artifacts?per_page=100" >"$output/artifacts.json"
    jq ${jq_binary_option:+"--binary"} -e --arg sha "$RELEASE_SHA" --arg name "tusk-candidate-$CANDIDATE_RUN_ID-$attempt" --argjson run "$CANDIDATE_RUN_ID" --slurpfile runmeta "$output/run.json" '[.artifacts[]|select(.name==$name)] as $a | .total_count==(.artifacts|length) and ($a|length)==1 and ($a[0].id|type=="number" and .>0 and floor==.) and $a[0].expired==false and ($a[0].expires_at|fromdateiso8601)>now and ($a[0].digest|test("^sha256:[0-9a-f]{64}$")) and $a[0].workflow_run.id==$run and $a[0].workflow_run.head_sha==$sha and $a[0].workflow_run.head_branch=="main" and $a[0].workflow_run.repository_id==$runmeta[0].repository.id and $a[0].workflow_run.head_repository_id==$runmeta[0].repository.id' "$output/artifacts.json" >/dev/null || die 'artifact missing, expired, duplicated or from a different source'
    artifact_id=$(jq ${jq_binary_option:+"--binary"} -r --arg name "tusk-candidate-$CANDIDATE_RUN_ID-$attempt" '.artifacts[]|select(.name==$name)|.id' "$output/artifacts.json")
    artifact_digest=$(jq ${jq_binary_option:+"--binary"} -r --arg name "tusk-candidate-$CANDIDATE_RUN_ID-$attempt" '.artifacts[]|select(.name==$name)|.digest' "$output/artifacts.json")
    # Inspect names/types before using an inventory name as a path.
    bash "$root/scripts/release_check.sh" verify "$directory/assets" >"$output/inventory.log" 2>&1
    bash "$root/scripts/release_check.sh" check-cask "$manifest" "$directory/homebrew/Casks/tusk.rb" >"$output/cask.log" 2>&1
    local subjects=("$directory/homebrew/Casks/tusk.rb" "$directory/candidate-run.json" "$manifest" "$directory/assets/checksums.txt")
    while IFS= read -r file; do subjects+=("$directory/assets/$file"); done < <(jq ${jq_binary_option:+"--binary"} -r '.files_sha256|keys[]' "$manifest")
    local index=0
    for file in "${subjects[@]}"; do
        hash=$(digest "$file")
        gh attestation verify "$file" --repo "$repo" --signer-workflow "$repo/$workflow" --signer-digest "$RELEASE_SHA" --source-ref refs/heads/main --source-digest "$RELEASE_SHA" --deny-self-hosted-runners --predicate-type https://slsa.dev/provenance/v1 --format json >"$output/attestation-$index.json"
        jq ${jq_binary_option:+"--binary"} -e --arg sha "$RELEASE_SHA" --arg hash "$hash" --arg uri "https://github.com/$repo/actions/runs/$CANDIDATE_RUN_ID/attempts/$attempt" 'any(.[]; .verificationResult as $v | $v.signature.certificate as $c | $c.issuer=="https://token.actions.githubusercontent.com" and $c.buildSignerURI=="https://github.com/newbpydev/tusk/.github/workflows/release.yml@refs/heads/main" and $c.buildSignerDigest==$sha and $c.sourceRepositoryURI=="https://github.com/newbpydev/tusk" and $c.sourceRepositoryDigest==$sha and $c.sourceRepositoryRef=="refs/heads/main" and $c.runnerEnvironment=="github-hosted" and $c.buildTrigger=="workflow_dispatch" and $c.runInvocationURI==$uri and any($v.statement.subject[];.digest.sha256==$hash))' "$output/attestation-$index.json" >/dev/null || die 'verified certificate/subject belongs to a different run/source'
        index=$((index+1))
    done
    jq ${jq_binary_option:+"--binary"} -n --arg sha "$RELEASE_SHA" --arg version "$RELEASE_VERSION" --arg hash "$CANDIDATE_MANIFEST_SHA256" --arg cask "$(digest "$directory/homebrew/Casks/tusk.rb")" --arg digest "$artifact_digest" --argjson run "$CANDIDATE_RUN_ID" --argjson attempt "$attempt" --argjson id "$artifact_id" '{schema:1,status:"verified",repository:"newbpydev/tusk",workflow:".github/workflows/release.yml",source_sha:$sha,version:$version,run_id:$run,run_attempt:$attempt,artifact_id:$id,artifact_api_digest:$digest,manifest_sha256:$hash,cask_sha256:$cask,transport:"API artifact digest retained; every payload and run receipt verified cryptographically, no claim of independently hashing the transport ZIP"}' >"$output/receipt.json"
    printf 'Verified trusted candidate: run %s attempt %s artifact %s\n' "$CANDIDATE_RUN_ID" "$attempt" "$artifact_id"
)
case "${1:-contract}" in
    identity) identity ;;
    contract) contract ;;
    record) record ;;
    verify) verify ;;
    *) die 'expected identity, contract, record or verify' ;;
esac
