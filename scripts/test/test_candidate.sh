#!/usr/bin/env bash
set -euo pipefail
jq_binary_option=''
case "${OSTYPE:-}" in msys*|cygwin*) jq_binary_option=--binary ;; esac
export jq_binary_option
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
failed=0
expect() {
    local want=$1 label=$2 actual=0
    shift 2
    "$@" >"$scratch/output" 2>&1 || actual=$?
    if [[ "$actual" == "$want" ]]; then printf 'PASS: %s\n' "$label"
    else printf 'FAIL: %s (expected %s, got %s)\n' "$label" "$want" "$actual" >&2; cat "$scratch/output" >&2; failed=1; fi
}
sha=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
identity=(env GITHUB_REPOSITORY=newbpydev/tusk GITHUB_EVENT_NAME=workflow_dispatch GITHUB_REF=refs/heads/main GITHUB_SHA="$sha" GITHUB_WORKFLOW_SHA="$sha" GITHUB_WORKFLOW_REF=newbpydev/tusk/.github/workflows/release.yml@refs/heads/main GITHUB_RUN_ID=123 GITHUB_RUN_ATTEMPT=1 RELEASE_SHA="$sha" RELEASE_VERSION=v0.3.0)
expect 0 'manual main candidate identity' "${identity[@]}" bash "$root/scripts/candidate.sh" identity
for override in GITHUB_REPOSITORY=evil/tusk GITHUB_EVENT_NAME=pull_request GITHUB_REF=refs/heads/topic GITHUB_SHA=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb GITHUB_WORKFLOW_SHA=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb GITHUB_WORKFLOW_REF=evil/tusk/.github/workflows/release.yml@refs/heads/main GITHUB_RUN_ID=0 RELEASE_VERSION='v0.3.0;exit'; do
    expect 1 "reject identity $override" "${identity[@]}" "$override" bash "$root/scripts/candidate.sh" identity
done
expect 0 'ID-selected artifact extracts at the reviewed root' jq ${jq_binary_option:+"--binary"} -e '.jobs.provenance.steps[2].with["merge-multiple"] == true' "$root/.github/workflows/release.yml"
expect 0 'candidate workflow contract' bash "$root/scripts/candidate.sh" contract
# The dollar references are jq variables.
# shellcheck disable=SC2016
expect 0 'provenance owns its pinned Go compiler' jq ${jq_binary_option:+"--binary"} -e --slurpfile pins "$root/scripts/tool-versions.json" '[.jobs.provenance.steps[] | select((.uses // "") | startswith("actions/setup-go@"))] | length==1 and .[0].uses==("actions/setup-go@"+$pins[0].actions.setup_go.sha) and .[0].with["go-version"]==$pins[0].go.release' "$root/.github/workflows/release.yml"
expect 0 'reviewed cask uploaded and attested' jq ${jq_binary_option:+"--binary"} -e '.jobs.build.steps[5].with.path|endswith("/candidate/homebrew/Casks/tusk.rb")' "$root/.github/workflows/release.yml"
mkdir -p "$scratch/workflow/scripts" "$scratch/workflow/.github/workflows"
if [[ -f "$root/scripts/candidate.sh" && -f "$root/.github/workflows/release.yml" ]]; then
    cp "$root/scripts/candidate.sh" "$scratch/workflow/scripts/"
    cp "$root/scripts/tool-versions.json" "$scratch/workflow/scripts/"
    for mutation in '.jobs.provenance.steps[4]["continue-on-error"]=true' '.jobs.provenance.steps[4]["if"]="always()"' '.jobs.build.steps[3].env.RELEASE_SHA=("b"*40)' '.on.push={}' '.permissions.contents="write"' '.jobs.build.permissions={"id-token":"write"}' '.jobs.build.needs=["identity"]' '.jobs.ci.secrets="inherit"' '.jobs.provenance.steps[0].with["persist-credentials"]=true' '.jobs.build.steps[4].env.GH_TOKEN="secret"' '.jobs.build.steps[5].with["retention-days"]=1' '.jobs.provenance.steps[2].uses="actions/download-artifact@main"'; do
        jq ${jq_binary_option:+"--binary"} "$mutation" "$root/.github/workflows/release.yml" >"$scratch/workflow/.github/workflows/release.yml"
        expect 1 "reject workflow $mutation" bash "$scratch/workflow/scripts/candidate.sh" contract
    done
fi
# API/cryptographic verifier boundary fakes; no native or hosted proof implied.
mkdir -p "$scratch/fixture/scripts" "$scratch/candidate/assets" "$scratch/bin"
if [[ ! -f "$root/scripts/candidate.sh" ]]; then exit 1; fi
cp "$root/scripts/candidate.sh" "$scratch/fixture/scripts/"
 cp "$root/scripts/json_check.sh" "$scratch/fixture/scripts/"
cp "$root/Makefile" "$scratch/fixture/"
cp "$root/scripts/gh_deadline.sh" "$scratch/fixture/scripts/"
cp -r "$root/scripts/ghdeadline" "$scratch/fixture/scripts/"
cat >"$scratch/fixture/scripts/release_check.sh" <<'CHECK'
#!/usr/bin/env bash
printf '%s\n' "$1" >>"$FIXTURE_ROOT/verifier-calls"
case "$1:${FIXTURE_VERIFIER_FAIL:-none}" in verify:inventory|check-cask:cask) exit 61;; esac
printf 'inventory inspected\n'
CHECK
cp "$root/scripts/test/sha256.sh" "$scratch/sha256.sh"
printf payload >"$scratch/candidate/assets/payload.tar.gz"
file_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/candidate/assets/payload.tar.gz")
jq ${jq_binary_option:+"--binary"} -n --arg sha "$sha" --arg hash "$file_hash" '{schema:1,mode:"candidate",version:"0.3.0",source_sha:$sha,files_sha256:{"payload.tar.gz":$hash},checksums_sha256:$hash}' >"$scratch/candidate/assets/release-manifest.json"
printf payload >"$scratch/candidate/assets/checksums.txt"
manifest_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/candidate/assets/release-manifest.json")
jq ${jq_binary_option:+"--binary"} -n --arg sha "$sha" --arg hash "$manifest_hash" '{schema:1,repository:"newbpydev/tusk",workflow:".github/workflows/release.yml",ref:"refs/heads/main",source_sha:$sha,version:"v0.3.0",run_id:123,run_attempt:1,manifest_sha256:$hash}' >"$scratch/candidate/candidate-run.json"
jq ${jq_binary_option:+"--binary"} -n --arg sha "$sha" '{id:123,run_attempt:1,path:".github/workflows/release.yml",event:"workflow_dispatch",head_branch:"main",head_sha:$sha,status:"completed",conclusion:"success",repository:{id:8,full_name:"newbpydev/tusk"},head_repository:{id:8,full_name:"newbpydev/tusk"}}' >"$scratch/run.json"
jq ${jq_binary_option:+"--binary"} -n '{id:8,full_name:"newbpydev/tusk",default_branch:"main"}' >"$scratch/repo.json"
jq ${jq_binary_option:+"--binary"} -n --arg sha "$sha" '{total_count:1,artifacts:[{id:456,name:"tusk-candidate-123-1",expired:false,expires_at:"2099-01-01T00:00:00Z",digest:("sha256:"+("c"*64)),workflow_run:{id:123,repository_id:8,head_repository_id:8,head_branch:"main",head_sha:$sha}}]}' >"$scratch/artifact.json"
jq ${jq_binary_option:+"--binary"} -n '{total_count:9,jobs:(["identity","ci / linux/amd64 — release compiler","ci / linux/arm64 — release compiler","ci / darwin/amd64 — release compiler","ci / darwin/arm64 — release compiler","ci / windows/amd64 — release compiler","ci / Linux — minimum compiler and five-target cross-builds","build","provenance"]|map({name:.,status:"completed",conclusion:"success"}))}' >"$scratch/jobs.json"
cat >"$scratch/bin/gh-script" <<'GH'
#!/usr/bin/env bash
set -euo pipefail
[[ "${GH_HOST:-}" == github.com && -z "${GH_DEBUG:-}" && -z "${DEBUG:-}" ]] || exit 78
printf '%s\n' "$*" >>"$FIXTURE_ROOT/gh-calls"
case "$1" in
    --version) echo 'gh version 2.102.0 (fixture)' ;;
    api)
        [[ "$2" == https://api.github.com/repos/* ]] || exit 79
        endpoint=${2#https://api.github.com/}
        [[ "${FIXTURE_API_FAIL:-0}" != 1 ]] || exit 47
        [[ "$*" != *'--method POST'* && "$*" != *'--method PATCH'* && "$*" != *'--method DELETE'* ]] || exit 77
        case "$endpoint" in
            repos/newbpydev/tusk) cat "$FIXTURE_ROOT/repo.json" ;;
            */artifacts*) cat "$FIXTURE_ROOT/artifact.json" ;;
            */jobs*) cat "$FIXTURE_ROOT/jobs.json" ;;
            */attempts/1) cat "$FIXTURE_ROOT/run.json" ;;
            *) exit 77 ;;
        esac ;;
    attestation)
        [[ "$2" == verify && "$*" == *'--source-ref refs/heads/main'* && "$*" == *'--source-digest aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'* && "$*" == *'--signer-digest aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'* && "$*" == *'--deny-self-hosted-runners'* ]] || exit 77
        [[ "${FIXTURE_ATTEST:-ok}" != missing ]] || exit 1
        hash=$(bash "$FIXTURE_ROOT/sha256.sh" "$3")
        if [[ "$3" == */payload.tar.gz && -n "${FIXTURE_PAYLOAD_HASH:-}" ]]; then hash=$FIXTURE_PAYLOAD_HASH; fi
        jq ${jq_binary_option:+"--binary"} -n --arg hash "$hash" --arg run "${FIXTURE_CERT_RUN:-123}" '[{verificationResult:{signature:{certificate:{issuer:"https://token.actions.githubusercontent.com",buildSignerURI:"https://github.com/newbpydev/tusk/.github/workflows/release.yml@refs/heads/main",buildSignerDigest:("a"*40),sourceRepositoryURI:"https://github.com/newbpydev/tusk",sourceRepositoryDigest:("a"*40),sourceRepositoryRef:"refs/heads/main",runnerEnvironment:"github-hosted",buildTrigger:"workflow_dispatch",runInvocationURI:("https://github.com/newbpydev/tusk/actions/runs/"+$run+"/attempts/1")}},statement:{subject:[{digest:{sha256:$hash}}]}}}]' ;;
    *) exit 77 ;;
esac
GH
chmod +x "$scratch/bin/gh-script"
fixture_gh_shell=$(command -v bash)
fixture_gh_script="$scratch/bin/gh-script"
fixture_gh_binary="$scratch/bin/gh"
if command -v cygpath >/dev/null 2>&1; then
    fixture_gh_shell=$(cygpath -m "$fixture_gh_shell")
    fixture_gh_script=$(cygpath -m "$fixture_gh_script")
    fixture_gh_binary="$scratch/bin/gh.exe"
fi
GH_FIXTURE_OUTPUT="$fixture_gh_binary" make --no-print-directory -C "$root" build-gh-fixture
verify=(env FIXTURE_GH_BASH="$fixture_gh_shell" FIXTURE_GH_SCRIPT="$fixture_gh_script" PATH="$scratch/bin:$PATH" GH_HOST=wrong.example GH_DEBUG=api DEBUG=1 FIXTURE_ROOT="$scratch" CANDIDATE_DIR="$scratch/candidate" CANDIDATE_MANIFEST_SHA256="$manifest_hash" CANDIDATE_RUN_ID=123 RELEASE_SHA="$sha" RELEASE_VERSION=v0.3.0)
expect 1 'missing cask must refuse complete candidate acceptance' "${verify[@]}" CANDIDATE_VERIFICATION_DIR="$scratch/missing-cask" bash "$scratch/fixture/scripts/candidate.sh" verify
mkdir -p "$scratch/candidate/homebrew/Casks"
printf 'fixture cask' >"$scratch/candidate/homebrew/Casks/tusk.rb"
cask_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/candidate/homebrew/Casks/tusk.rb")
jq ${jq_binary_option:+"--binary"} --arg hash "$cask_hash" '.cask_sha256=$hash' "$scratch/candidate/candidate-run.json" >"$scratch/new-receipt"
mv "$scratch/new-receipt" "$scratch/candidate/candidate-run.json"
expect 0 'trusted API and certificate identity accepted' "${verify[@]}" CANDIDATE_VERIFICATION_DIR="$scratch/accepted" bash "$scratch/fixture/scripts/candidate.sh" verify
for boundary in inventory cask; do
 : >"$scratch/gh-calls"
 expect 61 "candidate refuses failed $boundary verification" "${verify[@]}" FIXTURE_VERIFIER_FAIL="$boundary" CANDIDATE_VERIFICATION_DIR="$scratch/failed-$boundary" bash "$scratch/fixture/scripts/candidate.sh" verify
 [[ ! -e "$scratch/failed-$boundary/receipt.json" ]] || failed=1
 if grep -q '^attestation ' "$scratch/gh-calls"; then echo 'FAIL: failed verifier reached attestation' >&2; failed=1; fi
done
expect 1 'verification output cannot mutate source checkout' "${verify[@]}" CANDIDATE_VERIFICATION_DIR="$scratch/fixture/unwanted-verification" bash "$scratch/fixture/scripts/candidate.sh" verify
[[ ! -e "$scratch/fixture/unwanted-verification" ]] || failed=1
expect 47 'lost read response retains failure without acceptance' "${verify[@]}" FIXTURE_API_FAIL=1 CANDIDATE_VERIFICATION_DIR="$scratch/lost" bash "$scratch/fixture/scripts/candidate.sh" verify
if [[ -e "$scratch/lost/receipt.json" ]]; then echo 'FAIL: lost response accepted candidate' >&2; failed=1; fi
expect 0 'fresh readback retry succeeds' "${verify[@]}" CANDIDATE_VERIFICATION_DIR="$scratch/retry" bash "$scratch/fixture/scripts/candidate.sh" verify
printf tampered >"$scratch/candidate/assets/payload.tar.gz"
expect 1 'changed payload cannot use original attestation' "${verify[@]}" FIXTURE_PAYLOAD_HASH="$file_hash" CANDIDATE_VERIFICATION_DIR="$scratch/tampered" bash "$scratch/fixture/scripts/candidate.sh" verify
printf payload >"$scratch/candidate/assets/payload.tar.gz"
printf tampered >"$scratch/candidate/homebrew/Casks/tusk.rb"
expect 1 'changed cask cannot use original run receipt' "${verify[@]}" CANDIDATE_VERIFICATION_DIR="$scratch/tampered-cask" bash "$scratch/fixture/scripts/candidate.sh" verify
printf 'fixture cask' >"$scratch/candidate/homebrew/Casks/tusk.rb"
expect 1 'wrong accepted manifest digest refused' "${verify[@]}" CANDIDATE_MANIFEST_SHA256="$(printf 'd%.0s' {1..64})" CANDIDATE_VERIFICATION_DIR="$scratch/wrong-manifest" bash "$scratch/fixture/scripts/candidate.sh" verify
expect 1 'verification cannot write inside payloads' "${verify[@]}" CANDIDATE_VERIFICATION_DIR="$scratch/candidate/assets/inside" bash "$scratch/fixture/scripts/candidate.sh" verify
if [[ -e "$scratch/candidate/assets/inside" ]]; then echo 'FAIL: verifier modified candidate storage' >&2; failed=1; fi
expect 1 'verification receipt collision preserved' "${verify[@]}" CANDIDATE_VERIFICATION_DIR="$scratch/accepted" bash "$scratch/fixture/scripts/candidate.sh" verify
expect 1 'missing attestation rejected' "${verify[@]}" FIXTURE_ATTEST=missing CANDIDATE_VERIFICATION_DIR="$scratch/missing" bash "$scratch/fixture/scripts/candidate.sh" verify
expect 1 'different signed run rejected' "${verify[@]}" FIXTURE_CERT_RUN=999 CANDIDATE_VERIFICATION_DIR="$scratch/wrong-cert" bash "$scratch/fixture/scripts/candidate.sh" verify
original=$(cat "$scratch/run.json")
for mutation in '.event="pull_request"' '.head_repository.full_name="evil/tusk"' '.head_repository.id=9' '.head_sha=("b"*40)' '.run_attempt=2' '.conclusion="cancelled"' '.path=".github/workflows/evil.yml"' '.head_branch="topic"'; do
    printf '%s' "$original" | jq ${jq_binary_option:+"--binary"} "$mutation" >"$scratch/run.json"
    expect 1 "reject run $mutation" "${verify[@]}" CANDIDATE_VERIFICATION_DIR="$scratch/negative-$RANDOM" bash "$scratch/fixture/scripts/candidate.sh" verify
done
printf '%s' "$original" >"$scratch/run.json"
original=$(cat "$scratch/artifact.json")
for mutation in '.artifacts[0].expired=true' '.artifacts[0].expires_at="2000-01-01T00:00:00Z"' '.artifacts[0].workflow_run.id=999' '.artifacts += [.artifacts[0]]' '.artifacts[0].digest="bad"'; do
    printf '%s' "$original" | jq ${jq_binary_option:+"--binary"} "$mutation" >"$scratch/artifact.json"
    expect 1 "reject artifact $mutation" "${verify[@]}" CANDIDATE_VERIFICATION_DIR="$scratch/negative-$RANDOM" bash "$scratch/fixture/scripts/candidate.sh" verify
done
printf '%s' "$original" >"$scratch/artifact.json"
jq ${jq_binary_option:+"--binary"} '.jobs[2].conclusion="skipped"' "$scratch/jobs.json" >"$scratch/new-jobs"; mv "$scratch/new-jobs" "$scratch/jobs.json"
expect 1 'skipped native gate rejected' "${verify[@]}" CANDIDATE_VERIFICATION_DIR="$scratch/skipped" bash "$scratch/fixture/scripts/candidate.sh" verify
# Literal Make input must remain data before any script or API invocation.
# shellcheck disable=SC2016
make -C "$root" verify-candidate CANDIDATE_DIR='$(shell touch '"$scratch"'/injected)' >"$scratch/make-output" 2>&1 || true
if [[ -e "$scratch/injected" ]]; then echo 'FAIL: candidate Make value executed code' >&2; failed=1; fi
[[ ! -e "$scratch/accepted/receipt.json" ]] && failed=1
exit "$failed"
