#!/usr/bin/env bash
set -euo pipefail
jq_binary_option=''
case "${OSTYPE:-}" in msys*|cygwin*) jq_binary_option=--binary ;; esac
jq() { command jq ${jq_binary_option:+"--binary"} "$@"; }
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
failed=0
expect() {
 local want=$1 label=$2 actual=0;shift 2
 "$@" >"$scratch/result" 2>&1 || actual=$?
 if [[ "$actual" == "$want" ]];then printf 'PASS: %s\n' "$label"
 else printf 'FAIL: %s (%s vs %s)\n' "$label" "$actual" "$want" >&2;cat "$scratch/result" >&2;failed=1;fi
}
expect 0 'native candidate workflow contract' bash "$root/scripts/native_candidate.sh" contract
expect 0 'native verifier explicitly grants read-only attestation access' jq -e '.permissions.attestations=="read"' "$root/.github/workflows/native-candidate.yml"
# The verifier must never come from the candidate-controlled source input.
# shellcheck disable=SC2016
expect 0 'native verifier checkout is bound to the trusted main workflow' jq -e '.jobs.native.steps[0].with.ref=="${{ github.workflow_sha }}" and (.jobs.native.if|contains("github.ref == '\''refs/heads/main'\''")) and (.jobs.native.if|contains("github.workflow_ref == '\''newbpydev/tusk/.github/workflows/native-candidate.yml@refs/heads/main'\''")) and (.jobs.native.if|contains("github.workflow_sha == github.sha"))' "$root/.github/workflows/native-candidate.yml"
mkdir -p "$scratch/fixture/scripts" "$scratch/fixture/.github/workflows"
if [[ ! -f "$root/scripts/native_candidate.sh" ]];then exit 1;fi
cp "$root/scripts/native_candidate.sh" "$scratch/fixture/scripts/"
cp "$root/scripts/tool-versions.json" "$scratch/fixture/scripts/"
for mutation in '.permissions.attestations="write"' '.permissions.contents="write"' '.on.push={}' '.jobs.native.strategy["fail-fast"]=true' '.jobs.native.strategy.matrix.include|=.[0:4]' '.jobs.native.steps[0].with["persist-credentials"]=true' '.jobs.native.steps[3].with["run-id"]="999"' '.jobs.native.steps[4].env.CANDIDATE_MANIFEST_SHA256="unselected"' '.jobs.native.steps[4]["continue-on-error"]=true' '.jobs.native.steps[5]["if"]="success()"';do
 jq "$mutation" "$root/.github/workflows/native-candidate.yml" >"$scratch/fixture/.github/workflows/native-candidate.yml"
 expect 1 "reject unsafe hosted contract: $mutation" bash "$scratch/fixture/scripts/native_candidate.sh" contract
done
# The candidate expression must stay literal in the mutated workflow.
# shellcheck disable=SC2016
for mutation in '.jobs.native.if="true"' 'del(.jobs.native.if)' '.jobs.native.steps[0].with.ref="${{ inputs.source_sha }}"'; do
 jq "$mutation" "$root/.github/workflows/native-candidate.yml" >"$scratch/fixture/.github/workflows/native-candidate.yml"
 expect 1 "reject untrusted verifier dispatch: $mutation" bash "$scratch/fixture/scripts/native_candidate.sh" contract
done
mkdir -p "$scratch/candidate/assets" "$scratch/bin" "$scratch/payload"
printf '#!/usr/bin/env bash\nprintf "tusk version 0.3.0\\n"\n' >"$scratch/payload/tusk"
chmod +x "$scratch/payload/tusk"
tar -czf "$scratch/candidate/assets/tusk.tar.gz" -C "$scratch/payload" tusk
hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/payload/tusk")
archive_hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/candidate/assets/tusk.tar.gz")
jq -n --arg hash "$hash" --arg archive_hash "$archive_hash" '{files_sha256:{"tusk.tar.gz":$archive_hash},version:"0.3.0",source_sha:("a"*40),targets:[{target:"linux/amd64",archive:"tusk.tar.gz",executable_sha256:$hash}]}' >"$scratch/candidate/assets/release-manifest.json"
manifest=$(bash "$root/scripts/test/sha256.sh" "$scratch/candidate/assets/release-manifest.json")
cat >"$scratch/bin/make" <<'MAKE'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >>"$FIXTURE_ROOT/calls"
case "${3:-}" in
 verify-candidate)
  [[ ${FIXTURE_VERIFY_FAIL:-0} == 0 ]] || exit 61
  mkdir "$CANDIDATE_VERIFICATION_DIR"
  printf '{"status":"verified"}\n' >"$CANDIDATE_VERIFICATION_DIR/receipt.json" ;;
 release-smoke)
  [[ -f "$CANDIDATE_VERIFICATION_RECEIPT" && -f "$RELEASE_BINARY" ]] || exit 62
  [[ ${FIXTURE_SMOKE_FAIL:-0} == 0 ]] || exit 63
  mkdir "$RELEASE_SMOKE_OUTPUT"
  printf '{"process_status":0,"observed_version":"0.3.0"}\n' >"$RELEASE_SMOKE_OUTPUT/receipt.json" ;;
 *) exit 77 ;;
esac
MAKE
cat >"$scratch/bin/go" <<'GO'
#!/usr/bin/env bash
if [[ "$1" == env && "$2" == GOOS ]];then printf 'linux\n'
elif [[ "$1" == env && "$2" == GOHOSTOS && "$3" == GOHOSTARCH ]];then printf 'linux\namd64\n'
else exit 77;fi
GO
chmod +x "$scratch/bin/make" "$scratch/bin/go"
driver=(env PATH="$scratch/bin:$PATH" FIXTURE_ROOT="$scratch" CANDIDATE_DIR="$scratch/candidate" RELEASE_SHA=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa RELEASE_VERSION=v0.3.0 CANDIDATE_RUN_ID=123 CANDIDATE_MANIFEST_SHA256="$manifest")
expect 0 'verified archive smoke preserves exact executable' "${driver[@]}" RELEASE_NATIVE_OUTPUT="$scratch/accepted" bash "$root/scripts/native_candidate.sh" smoke
expect 0 'selected executable is unchanged' cmp -s "$scratch/payload/tusk" "$scratch/accepted/payload/tusk"
expect 0 'automated receipt leaves manual terminal pending' jq -e '.status=="automated-passed" and .manual_terminal=="pending" and .target=="linux/amd64" and .candidate_run_id==123 and .observed_version=="0.3.0"' "$scratch/accepted/receipt.json"
# jq expands its own argument inside the literal expression.
# shellcheck disable=SC2016
expect 0 'native receipt binds the selected archive digest' jq -e --arg hash "$archive_hash" '.archive=="tusk.tar.gz" and .archive_sha256==$hash' "$scratch/accepted/receipt.json"
cp "$scratch/candidate/assets/release-manifest.json" "$scratch/manifest-saved.json"
jq '.files_sha256["tusk.tar.gz"]=("0"*64)' "$scratch/manifest-saved.json" >"$scratch/candidate/assets/release-manifest.json"
wrong_manifest=$(bash "$root/scripts/test/sha256.sh" "$scratch/candidate/assets/release-manifest.json")
: >"$scratch/calls"
expect 1 'archive mismatch refuses extraction and execution' "${driver[@]}" CANDIDATE_MANIFEST_SHA256="$wrong_manifest" RELEASE_NATIVE_OUTPUT="$scratch/wrong-archive" bash "$root/scripts/native_candidate.sh" smoke
expect 1 'archive mismatch never reaches smoke' grep -q 'release-smoke' "$scratch/calls"
cp "$scratch/manifest-saved.json" "$scratch/candidate/assets/release-manifest.json"
expect 1 'existing output is never replaced' "${driver[@]}" RELEASE_NATIVE_OUTPUT="$scratch/accepted" bash "$root/scripts/native_candidate.sh" smoke
expect 1 'output cannot mutate the source checkout' "${driver[@]}" RELEASE_NATIVE_OUTPUT="$scratch/fixture/inside" bash "$scratch/fixture/scripts/native_candidate.sh" smoke
expect 1 'output cannot mutate candidate storage' "${driver[@]}" RELEASE_NATIVE_OUTPUT="$scratch/candidate/inside" bash "$root/scripts/native_candidate.sh" smoke
: >"$scratch/calls"
expect 61 'failed cryptographic verification refuses execution' "${driver[@]}" FIXTURE_VERIFY_FAIL=1 RELEASE_NATIVE_OUTPUT="$scratch/unverified" bash "$root/scripts/native_candidate.sh" smoke
expect 1 'failed verification never reaches smoke' grep -q 'release-smoke' "$scratch/calls"
expect 1 'failed verification leaves no passing receipt' test -f "$scratch/unverified/receipt.json"
expect 63 'failed smoke remains a failed invocation' "${driver[@]}" FIXTURE_SMOKE_FAIL=1 RELEASE_NATIVE_OUTPUT="$scratch/smoke-failed" bash "$root/scripts/native_candidate.sh" smoke
expect 1 'failed smoke leaves no passing receipt' test -f "$scratch/smoke-failed/receipt.json"
printf tampered >"$scratch/payload/tusk"
tar -czf "$scratch/candidate/assets/tusk.tar.gz" -C "$scratch/payload" tusk
tampered_archive=$(bash "$root/scripts/test/sha256.sh" "$scratch/candidate/assets/tusk.tar.gz")
jq --arg hash "$tampered_archive" '.files_sha256["tusk.tar.gz"]=$hash' "$scratch/manifest-saved.json" >"$scratch/candidate/assets/release-manifest.json"
expect 1 'extracted digest mismatch refuses execution' "${driver[@]}" RELEASE_NATIVE_OUTPUT="$scratch/tampered" bash "$root/scripts/native_candidate.sh" smoke
# Exercise the actual pinned-lint boundary, using a tool fixture that only records
# its arguments. Omitting the new workflow must fail this assertion.
mkdir -p "$scratch/fixture/bin/tools"
cp "$root/Makefile" "$scratch/fixture/"
cp "$root/scripts/ci-check.sh" "$scratch/fixture/scripts/"
cp "$root/.github/workflows/ci.yml" "$root/.github/workflows/release.yml" "$root/.github/workflows/native-candidate.yml" "$scratch/fixture/.github/workflows/"
cat >"$scratch/fixture/bin/tools/actionlint" <<'LINT'
#!/usr/bin/env bash
if [[ "$1" == -version ]];then printf '1.7.12\n'
else printf '%s\n' "$@" >"$FIXTURE_ROOT/lint-arguments";fi
LINT
chmod +x "$scratch/fixture/bin/tools/actionlint"
expect 0 'pinned lint gate executes on complete workflow fixture' env PATH="$scratch/bin:$PATH" FIXTURE_ROOT="$scratch" bash "$scratch/fixture/scripts/ci-check.sh" check
expect 0 'pinned lint includes the native packaged workflow' grep -F -q "$scratch/fixture/.github/workflows/native-candidate.yml" "$scratch/lint-arguments"
exit "$failed"
