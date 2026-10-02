#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
scratch=$(mktemp -d)
scratch=$(cd "$scratch" && pwd -P)
trap 'rm -rf "$scratch"' EXIT
failed=0
expect() {
 local want=$1 label=$2 actual=0;shift 2
 "$@" >"$scratch/output" 2>&1 || actual=$?
 if [[ "$actual" == "$want" ]];then printf 'PASS: %s\n' "$label";else printf 'FAIL: %s (%s vs %s)\n' "$label" "$actual" "$want" >&2;cat "$scratch/output" >&2;failed=1;fi
}
# shellcheck disable=SC2329
no_build_prerequisite() {
 make -n -C "$1" bench-cli-release RELEASE_BINARY="$scratch/accepted" CLI_BENCH_OUTPUT="$scratch/cli.json" >"$scratch/dry-run" || return
 ! grep -Eq '(^|[[:space:]])go build|scripts/build[.]sh' "$scratch/dry-run"
}
expect 0 'CLI release measurement has no build prerequisite' no_build_prerequisite "$root"
mkdir -p "$scratch/build-dependent"
cp "$root/Makefile" "$scratch/build-dependent/"
printf '\nbench-cli-release: build\n' >>"$scratch/build-dependent/Makefile"
expect 1 'measurement assertion detects a deliberately added build' no_build_prerequisite "$scratch/build-dependent"
# Keep Make syntax literal until the driver's data validation.
# Literal Make expression is the injection-test input.
# shellcheck disable=SC2016
injection='$(shell touch '"$scratch"'/injected)'
expect 0 'measurement output is literal Make data' make -n -C "$root" bench-cli-release CLI_BENCH_OUTPUT="$injection"
if [[ -e "$scratch/injected" ]];then echo 'FAIL: output executed Make input' >&2;failed=1;fi
if [[ -f "$root/scripts/release_smoke.sh" ]];then
 mkdir -p "$scratch/fixture/scripts" "$scratch/candidate/assets" "$scratch/bin"
 cp "$root/scripts/release_smoke.sh" "$scratch/fixture/scripts/"
 cp "$root/scripts/json_check.sh" "$scratch/fixture/scripts/"
 printf '#!/usr/bin/env bash\n[[ "${FIXTURE_VERIFY_FAIL:-0}" == 0 ]] || exit 61\n' >"$scratch/fixture/scripts/release_check.sh"
 printf '#!/usr/bin/env bash\nif [[ -n ${FIXTURE_BINARY_CALLS:-} ]]; then printf "version executed\\n" >>"$FIXTURE_BINARY_CALLS"; fi\nprintf "tusk version %%s\\n" "${FIXTURE_VERSION:-0.3.0}"\n' >"$scratch/accepted"
 chmod +x "$scratch/accepted"
 cp "$scratch/accepted" "$scratch/original-accepted"
 fixture_os=$(go env GOHOSTOS);fixture_arch=$(go env GOHOSTARCH)
 hash=$(bash "$root/scripts/test/sha256.sh" "$scratch/accepted")
 jq --binary -n --arg hash "$hash" --arg target "$fixture_os/$fixture_arch" '{schema:1,mode:"candidate",version:"0.3.0",source_sha:("a"*40),compiler:"1.27.1",targets:[{target:$target,archive:"payload.tar.gz",executable_sha256:$hash}]}' >"$scratch/candidate/assets/release-manifest.json"
 manifest=$(bash "$root/scripts/test/sha256.sh" "$scratch/candidate/assets/release-manifest.json")
 cat >"$scratch/bin/go" <<'GO'
#!/usr/bin/env bash
if [[ "$1" == env ]];then
 shift
 for key in "$@";do
  case "$key" in GOHOSTOS) printf '%s\n' "$FIXTURE_OS";;GOHOSTARCH) printf '%s\n' "$FIXTURE_ARCH";;*) exit 77;;esac
 done
else
 printf '%s\n' "$*" >>"$FIXTURE_ROOT/calls"
 if [[ -n ${FIXTURE_NATIVE_BINARY:-} ]];then
  [[ $# == 8 && "$1" == run && "$2" == ./scripts/cli-bench && "$3" == --binary && "$4" == "$FIXTURE_NATIVE_BINARY" && "$5" == --output && "$6" == "$FIXTURE_NATIVE_OUTPUT" && "$7" == --acceptance-profile && "$8" == reference ]] || exit 78
 fi
fi
GO
 chmod +x "$scratch/bin/go"
 preflight=(env PATH="$scratch/bin:$PATH" FIXTURE_ROOT="$scratch" FIXTURE_OS="$fixture_os" FIXTURE_ARCH="$fixture_arch" RELEASE_BINARY="$scratch/accepted" RELEASE_MANIFEST="$scratch/candidate/assets/release-manifest.json" CANDIDATE_MANIFEST_SHA256="$manifest" RELEASE_EVIDENCE_SCOPE=local-fixture)
 expect 0 'exact native binary identity fixture' "${preflight[@]}" bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 expect 0 'relative selected binary is executed from its own absolute path' "${preflight[@]}" bash -c 'cd "$FIXTURE_ROOT" && RELEASE_BINARY=accepted bash "$FIXTURE_ROOT/fixture/scripts/release_smoke.sh" preflight'
 expect 61 'inventory failure precedes selected executable version call' "${preflight[@]}" FIXTURE_VERIFY_FAIL=1 FIXTURE_BINARY_CALLS="$scratch/binary-calls" bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 if [[ -e "$scratch/binary-calls" ]];then echo 'FAIL: unverified inventory executed its binary' >&2;failed=1;fi
 expect 1 'manifest version must match the executed binary version' "${preflight[@]}" FIXTURE_VERSION=dev bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 expect 1 'wrong binary digest refused' "${preflight[@]}" RELEASE_BINARY="$scratch/bin/go" bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 cat >"$scratch/bin/uname" <<'UNAME'
#!/usr/bin/env bash
printf 'MINGW64_NT-10.0\n'
UNAME
 cat >"$scratch/bin/cygpath" <<'CYGPATH'
#!/usr/bin/env bash
exit 23
CYGPATH
 chmod +x "$scratch/bin/uname" "$scratch/bin/cygpath"
 cp "$scratch/candidate/assets/release-manifest.json" "$scratch/original-manifest.json"
 jq --binary '.targets[0].target="windows/amd64"' "$scratch/original-manifest.json" >"$scratch/candidate/assets/release-manifest.json"
 windows_manifest=$(bash "$root/scripts/test/sha256.sh" "$scratch/candidate/assets/release-manifest.json")
 expect 1 'failed native path conversion refuses binary acceptance' "${preflight[@]}" FIXTURE_OS=windows FIXTURE_ARCH=amd64 CANDIDATE_MANIFEST_SHA256="$windows_manifest" bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 cat >"$scratch/bin/cygpath" <<'CYGPATH'
#!/usr/bin/env bash
[[ "$1" == -m ]] || exit 24
if [[ "$2" == */accepted ]];then printf 'C:/owned candidate/accepted.exe\n'
elif [[ "$2" == */native-report.json ]];then
 [[ ${FIXTURE_OUTPUT_CONVERSION_FAIL:-0} == 0 ]] || exit 23
 printf 'C:/owned reports/native-report.json\n'
else exit 25;fi
CYGPATH
 expect 0 'Windows CLI measurement receives native binary and report arguments' "${preflight[@]}" FIXTURE_OS=windows FIXTURE_ARCH=amd64 CANDIDATE_MANIFEST_SHA256="$windows_manifest" FIXTURE_NATIVE_BINARY='C:/owned candidate/accepted.exe' FIXTURE_NATIVE_OUTPUT='C:/owned reports/native-report.json' RELEASE_BENCH_OUTPUT="$scratch/native-report.json" bash "$scratch/fixture/scripts/release_smoke.sh" bench-cli
 rm -f "$scratch/native-report.json.identity.json"
 cp "$scratch/calls" "$scratch/calls-before-conversion"
 expect 1 'failed native report conversion refuses measurement' "${preflight[@]}" FIXTURE_OS=windows FIXTURE_ARCH=amd64 CANDIDATE_MANIFEST_SHA256="$windows_manifest" FIXTURE_OUTPUT_CONVERSION_FAIL=1 RELEASE_BENCH_OUTPUT="$scratch/native-report.json" bash "$scratch/fixture/scripts/release_smoke.sh" bench-cli
 expect 0 'failed report conversion never launches measurement' cmp -s "$scratch/calls" "$scratch/calls-before-conversion"
 mv "$scratch/original-manifest.json" "$scratch/candidate/assets/release-manifest.json"
 rm "$scratch/bin/uname" "$scratch/bin/cygpath"
 expect 1 'wrong pinned manifest refused' "${preflight[@]}" CANDIDATE_MANIFEST_SHA256="$(printf 'b%.0s' {1..64})" bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 expect 1 'trusted acceptance needs actual verification receipt' "${preflight[@]}" RELEASE_EVIDENCE_SCOPE=trusted-candidate bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 jq --binary -n --arg hash "$manifest" '{status:"verified",repository:"newbpydev/tusk",workflow:".github/workflows/release.yml",manifest_sha256:$hash,source_sha:("a"*40),version:"v0.3.0"}' >"$scratch/verified.json"
 expect 0 'single trusted verification receipt accepted' "${preflight[@]}" RELEASE_EVIDENCE_SCOPE=trusted-candidate CANDIDATE_VERIFICATION_RECEIPT="$scratch/verified.json" bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 jq --binary '.status="not-verified"' "$scratch/verified.json" >"$scratch/concatenated.json"
 cat "$scratch/verified.json" >>"$scratch/concatenated.json"
 expect 1 'concatenated trusted verification receipt refused' "${preflight[@]}" RELEASE_EVIDENCE_SCOPE=trusted-candidate CANDIDATE_VERIFICATION_RECEIPT="$scratch/concatenated.json" bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 expect 1 'unsupported evidence label refused' "${preflight[@]}" RELEASE_EVIDENCE_SCOPE=anything bash "$scratch/fixture/scripts/release_smoke.sh" preflight
 # A fake measurement runner observes the exact selected path, never rebuilds.
 expect 1 'smoke output cannot modify candidate assets' "${preflight[@]}" RELEASE_SMOKE_OUTPUT="$scratch/candidate/assets/new-smoke" bash "$scratch/fixture/scripts/release_smoke.sh" smoke
 if [[ -e "$scratch/candidate/assets/new-smoke" ]];then echo 'FAIL: smoke modified candidate storage' >&2;failed=1;fi
 expect 1 'measurement output cannot modify candidate assets' "${preflight[@]}" RELEASE_BENCH_OUTPUT="$scratch/candidate/assets/new-bench.json" bash "$scratch/fixture/scripts/release_smoke.sh" bench-cli
 expect 0 'measurement consumes supplied bytes' "${preflight[@]}" RELEASE_BENCH_OUTPUT="$scratch/new-report.json" bash "$scratch/fixture/scripts/release_smoke.sh" bench-cli
 expect 1 'retained measurement identity collision refused' "${preflight[@]}" RELEASE_BENCH_OUTPUT="$scratch/new-report.json" bash "$scratch/fixture/scripts/release_smoke.sh" bench-cli
 expect 0 'measurement preserves selected binary bytes' cmp -s "$scratch/accepted" "$scratch/original-accepted"
 expected_binary=$scratch/accepted
 if [[ "$fixture_os" == windows ]];then expected_binary=$(cygpath -m "$expected_binary");fi
 expect 0 'measurement records selected native binary path' grep -F -- "--binary $expected_binary" "$scratch/calls"
fi
exit "$failed"
