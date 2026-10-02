#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
die() { printf 'Release smoke: %s\n' "$*" >&2; exit 1; }
digest() {
 if command -v sha256sum >/dev/null 2>&1;then sha256sum "$1" | awk '{print $1}';else shasum -a 256 "$1" | awk '{print $1}';fi
}
native_path() {
 if [[ "$(go env GOHOSTOS)" == windows ]];then cygpath -m "$1";else printf '%s\n' "$1";fi
}
preflight() {
 [[ -f "${RELEASE_BINARY:-}" && ! -L "$RELEASE_BINARY" && -f "${RELEASE_MANIFEST:-}" && ! -L "$RELEASE_MANIFEST" ]] || die 'regular extracted binary and manifest required'
 [[ "${CANDIDATE_MANIFEST_SHA256:-}" =~ ^[0-9a-f]{64}$ && "$(digest "$RELEASE_MANIFEST")" == "$CANDIDATE_MANIFEST_SHA256" ]] || die 'manifest differs from selected digest'
 local host=() line target
 while IFS= read -r line;do host+=("$line");done < <(go env GOHOSTOS GOHOSTARCH)
 [[ ${#host[@]} == 2 ]] || die 'native host identity unavailable'
 target="${host[0]}/${host[1]}"
 case "$target:$(uname -s)" in linux/amd64:Linux|linux/arm64:Linux|darwin/amd64:Darwin|darwin/arm64:Darwin|windows/amd64:MINGW*|windows/amd64:MSYS*) ;; *) die 'required native target/shell; Windows uses native Git Bash or PowerShell, not WSL';; esac
 binary_hash=$(digest "$RELEASE_BINARY")
 jq -e --arg target "$target" --arg hash "$binary_hash" '.mode=="candidate" and ([.targets[]|select(.target==$target and .executable_sha256==$hash)]|length)==1' "$RELEASE_MANIFEST" >/dev/null || die 'binary digest/architecture not in candidate'
 bash "$root/scripts/release_check.sh" verify "$(dirname "$RELEASE_MANIFEST")"
 case "${RELEASE_EVIDENCE_SCOPE:-trusted-candidate}" in
  local-fixture) printf 'Preliminary local fixture; cannot close native release acceptance.\n' ;;
  trusted-candidate)
   [[ -f "${CANDIDATE_VERIFICATION_RECEIPT:-}" && ! -L "$CANDIDATE_VERIFICATION_RECEIPT" ]] || die 'actual trusted-candidate verification receipt required'
   jq -e --arg hash "$CANDIDATE_MANIFEST_SHA256" --slurpfile m "$RELEASE_MANIFEST" '.status=="verified" and .repository=="newbpydev/tusk" and .workflow==".github/workflows/release.yml" and .manifest_sha256==$hash and .source_sha==$m[0].source_sha and .version==("v"+$m[0].version)' "$CANDIDATE_VERIFICATION_RECEIPT" >/dev/null || die 'verification receipt does not match candidate' ;;
  *) die 'invalid evidence scope';;
 esac
 RELEASE_BINARY="$(cd "$(dirname "$RELEASE_BINARY")" && pwd -P)/$(basename "$RELEASE_BINARY")"
 export RELEASE_BINARY TUSK_RELEASE_BINARY="$(native_path "$RELEASE_BINARY")"
}
owned_output() {
 local requested=$1 parent candidate output
 [[ -n "$requested" && -d "$(dirname "$requested")" ]] || die 'explicit new output with existing parent required'
 parent=$(cd "$(dirname "$requested")" && pwd -P)
 candidate=$(cd "$(dirname "$RELEASE_MANIFEST")/.." && pwd -P)
 output="$parent/$(basename "$requested")"
 case "$output" in "$candidate"|"$candidate"/*|"$root"|"$root"/*) die 'output must be outside candidate storage and source checkout';; esac
 [[ ! -e "$output" && ! -L "$output" ]] || die 'output already exists'
 printf '%s\n' "$output"
}
new_output() {
 [[ -n "${RELEASE_BENCH_OUTPUT:-}" && ! -e "$RELEASE_BENCH_OUTPUT" && ! -L "$RELEASE_BENCH_OUTPUT" && -d "$(dirname "$RELEASE_BENCH_OUTPUT")" ]] || die 'explicit new report with existing parent required'
 [[ ! -e "$RELEASE_BENCH_OUTPUT.identity.json" && ! -L "$RELEASE_BENCH_OUTPUT.identity.json" && ! -e "$RELEASE_BENCH_OUTPUT.startup.json" && ! -L "$RELEASE_BENCH_OUTPUT.startup.json" ]] || die 'identity report collision'
 RELEASE_BENCH_OUTPUT=$(owned_output "$RELEASE_BENCH_OUTPUT")
 jq -n --arg hash "$binary_hash" --arg manifest "$CANDIDATE_MANIFEST_SHA256" --arg scope "${RELEASE_EVIDENCE_SCOPE:-trusted-candidate}" --slurpfile m "$RELEASE_MANIFEST" '{schema:1,binary_sha256:$hash,manifest_sha256:$manifest,scope:$scope,source_sha:$m[0].source_sha,binary_compiler:$m[0].compiler,measurement_compiler:"reported separately by Go measurement harness"}' >"$RELEASE_BENCH_OUTPUT.identity.json"
}
case "${1:-preflight}" in
 preflight) preflight ;;
 smoke)
  preflight
  [[ -n "${RELEASE_SMOKE_OUTPUT:-}" && ! -e "$RELEASE_SMOKE_OUTPUT" && ! -L "$RELEASE_SMOKE_OUTPUT" ]] || die 'new retained RELEASE_SMOKE_OUTPUT directory required'
  RELEASE_SMOKE_OUTPUT=$(owned_output "$RELEASE_SMOKE_OUTPUT")
  mkdir "$RELEASE_SMOKE_OUTPUT" || die 'smoke output parent missing/collision'
  TUSK_RELEASE_FIXTURE_DIR="$(cd "$RELEASE_SMOKE_OUTPUT" && pwd -P)/fixtures"
  mkdir "$TUSK_RELEASE_FIXTURE_DIR"
  TUSK_RELEASE_FIXTURE_DIR=$(native_path "$TUSK_RELEASE_FIXTURE_DIR")
  export TUSK_RELEASE_FIXTURE_DIR
  status=0
  make -C "$root" release-smoke-processes >"$RELEASE_SMOKE_OUTPUT/processes.log" 2>&1 || status=$?
  jq -n --arg hash "$binary_hash" --arg manifest "$CANDIDATE_MANIFEST_SHA256" --arg scope "${RELEASE_EVIDENCE_SCOPE:-trusted-candidate}" --arg host "$(uname -s)" --argjson status "$status" '{schema:1,binary_sha256:$hash,manifest_sha256:$manifest,scope:$scope,host:$host,process_status:$status,manual_terminal:"pending; automated child PTY is separate",fixtures:"owned fixtures retained including DB/WAL/SHM on failure"}' >"$RELEASE_SMOKE_OUTPUT/receipt.json"
  [[ "$(digest "$RELEASE_BINARY")" == "$binary_hash" ]] || die 'supplied executable changed during smoke'
  exit "$status" ;;
 bench-cli)
  preflight;new_output
  status=0
  (cd "$root" && go run ./scripts/cli-bench --binary "$RELEASE_BINARY" --output "$RELEASE_BENCH_OUTPUT" --acceptance-profile reference) || status=$?
  [[ "$(digest "$RELEASE_BINARY")" == "$binary_hash" ]] || die 'supplied executable changed during measurements'
  exit "$status" ;;
 bench-tui)
  preflight;new_output
  [[ "$(go env GOHOSTOS)/$(go env GOHOSTARCH)" == linux/amd64 ]] || die 'TUI retained reference measurements need native Linux amd64'
  # Model timings use the accepted source; packaged startup uses the selected bytes.
  [[ "$(git -C "$root" rev-parse HEAD)" == "$(jq -r '.source_sha' "$RELEASE_MANIFEST")" ]] || die 'TUI model measurement checkout must equal candidate source SHA'
  [[ -z "$(git -C "$root" status --porcelain)" ]] || die 'TUI model measurement source must be clean'
  status=0
  (cd "$root" && TUSK_TUI_BENCH_OUTPUT="$RELEASE_BENCH_OUTPUT" go test ./internal/tui -run '^TestTUIMeasurements$' -count=1 -timeout=20m -v) || status=$?
  (cd "$root" && TUSK_TUI_STARTUP_OUTPUT="$RELEASE_BENCH_OUTPUT.startup.json" go test ./cmd/tusk -run '^TestTUIStartupMeasurements$' -count=1 -timeout=5m -v) || status=$?
  [[ "$(digest "$RELEASE_BINARY")" == "$binary_hash" ]] || die 'supplied executable changed during measurements'
  exit "$status" ;;
 *) die 'expected preflight, smoke, bench-cli or bench-tui';;
esac
