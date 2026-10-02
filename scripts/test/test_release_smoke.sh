#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
failed=0
expect() {
 local want=$1 label=$2 actual=0;shift 2
 "$@" >"$scratch/output" 2>&1 || actual=$?
 if [[ "$actual" == "$want" ]];then printf 'PASS: %s\n' "$label";else printf 'FAIL: %s (%s vs %s)\n' "$label" "$actual" "$want" >&2;cat "$scratch/output" >&2;failed=1;fi
}
expect 0 'CLI release measurement has no build prerequisite' make -n -C "$root" bench-cli-release RELEASE_BINARY="$scratch/accepted" CLI_BENCH_OUTPUT="$scratch/cli.json"
# Keep Make syntax literal until the driver's data validation.
injection='$(shell touch '"$scratch"'/injected)'
expect 0 'measurement output is literal Make data' make -n -C "$root" bench-cli-release CLI_BENCH_OUTPUT="$injection"
if [[ -e "$scratch/injected" ]];then echo 'FAIL: output executed Make input' >&2;failed=1;fi
if [[ -f "$root/scripts/release_smoke.sh" ]];then
 mkdir -p "$scratch/fixture/scripts" "$scratch/candidate/assets" "$scratch/bin"
 cp "$root/scripts/release_smoke.sh" "$scratch/fixture/scripts/"
 printf '#!/usr/bin/env bash\nexit 0\n' >"$scratch/fixture/scripts/release_check.sh"
 printf 'accepted bytes' >"$scratch/accepted"
 fixture_os=$(go env GOHOSTOS);fixture_arch=$(go env GOHOSTARCH)
 hash=$(sha256sum "$scratch/accepted" | cut -d' ' -f1)
 jq -n --arg hash "$hash" --arg target "$fixture_os/$fixture_arch" '{schema:1,mode:"candidate",version:"0.3.0",source_sha:("a"*40),compiler:"1.27.1",targets:[{target:$target,archive:"payload.tar.gz",executable_sha256:$hash}]}' >"$scratch/candidate/assets/release-manifest.json"
 manifest=$(sha256sum "$scratch/candidate/assets/release-manifest.json" | cut -d' ' -f1)
 cat >"$scratch/bin/go" <<'GO'
#!/usr/bin/env bash
if [[ "$1" == env ]];then printf '%s\n' "$FIXTURE_OS" "$FIXTURE_ARCH";else printf '%s\n' "$*" >>"$FIXTURE_ROOT/calls";fi
GO
 chmod +x "$scratch/bin/go"
 preflight=(env PATH="$scratch/bin:$PATH" FIXTURE_ROOT="$scratch" FIXTURE_OS="$fixture_os" FIXTURE_ARCH="$fixture_arch" RELEASE_BINARY="$scratch/accepted" RELEASE_MANIFEST="$scratch/candidate/assets/release-manifest.json" CANDIDATE_MANIFEST_SHA256="$manifest" RELEASE_EVIDENCE_SCOPE=local-fixture)
 expect 0 'exact native binary identity fixture' "${preflight[@]}" bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 expect 1 'wrong binary digest refused' "${preflight[@]}" RELEASE_BINARY="$scratch/bin/go" bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 expect 1 'wrong pinned manifest refused' "${preflight[@]}" CANDIDATE_MANIFEST_SHA256="$(printf 'b%.0s' {1..64})" bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 expect 1 'trusted acceptance needs actual verification receipt' "${preflight[@]}" RELEASE_EVIDENCE_SCOPE=trusted-candidate bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 expect 1 'unsupported evidence label refused' "${preflight[@]}" RELEASE_EVIDENCE_SCOPE=anything bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 # A fake measurement runner observes the exact selected path, never rebuilds.
 expect 1 'smoke output cannot modify candidate assets' "${preflight[@]}" RELEASE_SMOKE_OUTPUT="$scratch/candidate/assets/new-smoke" bash "$scratch/fixture/scripts/release_smoke.sh" smoke
 if [[ -e "$scratch/candidate/assets/new-smoke" ]];then echo 'FAIL: smoke modified candidate storage' >&2;failed=1;fi
 expect 1 'measurement output cannot modify candidate assets' "${preflight[@]}" RELEASE_BENCH_OUTPUT="$scratch/candidate/assets/new-bench.json" bash "$scratch/fixture/scripts/release_smoke.sh" bench-cli
 expect 0 'measurement consumes supplied bytes' "${preflight[@]}" RELEASE_BENCH_OUTPUT="$scratch/new-report.json" bash "$scratch/fixture/scripts/release_smoke.sh" bench-cli
 expect 1 'retained measurement identity collision refused' "${preflight[@]}" RELEASE_BENCH_OUTPUT="$scratch/new-report.json" bash "$scratch/fixture/scripts/release_smoke.sh" bench-cli
 [[ "$(cat "$scratch/accepted")" == 'accepted bytes' ]] || failed=1
 grep -F -- "--binary $scratch/accepted" "$scratch/calls" >/dev/null || failed=1
fi
exit "$failed"
