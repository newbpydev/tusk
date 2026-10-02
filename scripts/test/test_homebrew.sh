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
 local want=$1 label=$2 actual=0;shift 2
 "$@" >"$scratch/output" 2>&1 || actual=$?
 if [[ "$actual" == "$want" ]];then printf 'PASS: %s\n' "$label";else printf 'FAIL: %s (%s vs %s)\n' "$label" "$actual" "$want" >&2;cat "$scratch/output" >&2;failed=1;fi
}
expect 0 'disabled-upload macOS cask configuration' bash "$root/scripts/release.sh" contract
mkdir -p "$scratch/fixture/scripts" "$scratch/bin"
cp "$root/scripts/release.sh" "$root/scripts/homebrew.sh" "$scratch/fixture/scripts/"
cp "$root/Makefile" "$scratch/fixture/"
cp "$root/scripts/gh_deadline.sh" "$scratch/fixture/scripts/"
cp -r "$root/scripts/ghdeadline" "$scratch/fixture/scripts/"
for mutation in '.homebrew_casks[0].skip_upload=false' '.homebrew_casks[0].hooks={post:{install:"xattr"}}' '.homebrew_casks[0].zap={trash:["~/.local/share/tusk"]}' '.homebrew_casks[0].ids=["tusk"]' '.homebrew_casks[0].repository.token="secret"' '.homebrew_casks[0].manpages=[]' '.homebrew_casks[0].custom_block="system(\"evil\")"' '.builds[1].goos=["linux","darwin"]';do
 jq ${jq_binary_option:+"--binary"} "$mutation" "$root/.goreleaser.yaml" >"$scratch/fixture/.goreleaser.yaml"
 expect 1 "reject cask mutation: $mutation" bash "$scratch/fixture/scripts/release.sh" contract
done
fixture_gh_shell=$(command -v bash)
fixture_gh_script="$scratch/bin/gh-script"
fixture_gh_binary="$scratch/bin/gh"
if command -v cygpath >/dev/null 2>&1; then
    fixture_gh_shell=$(cygpath -m "$fixture_gh_shell")
    fixture_gh_script=$(cygpath -m "$fixture_gh_script")
    fixture_gh_binary="$scratch/bin/gh.exe"
fi
GH_FIXTURE_OUTPUT="$fixture_gh_binary" make --no-print-directory -C "$root" build-gh-fixture
invoke=(env FIXTURE_GH_BASH="$fixture_gh_shell" FIXTURE_GH_SCRIPT="$fixture_gh_script" PATH="$scratch/bin:$PATH")
printf '#!/usr/bin/env bash\nexit 44\n' >"$scratch/bin/gh-script";chmod +x "$scratch/bin/gh-script"
expect 1 'missing tap remains named prerequisite' "${invoke[@]}" bash "$scratch/fixture/scripts/homebrew.sh" destination
if ! grep -Eq '006-ISS-003' "$scratch/output";then failed=1;fi
cat >"$scratch/bin/gh-script" <<'GH'
#!/usr/bin/env bash
[[ "${GH_HOST:-}" == github.com && -z "${GH_DEBUG:-}" && -z "${DEBUG:-}" ]] || exit 78
[[ "$2" == https://api.github.com/repos/newbpydev/homebrew-tap ]] || exit 79
printf 'true\n'
GH
expect 0 'tap check pins public GitHub host and disables HTTP debug logging' "${invoke[@]}" GH_HOST=wrong.example GH_DEBUG=api DEBUG=1 bash "$scratch/fixture/scripts/homebrew.sh" destination
expect 1 'stalled owned GitHub child has a deadline' "${invoke[@]}" FIXTURE_GH_STALL=1 GH_REQUEST_TIMEOUT=100ms bash "$scratch/fixture/scripts/homebrew.sh" destination
if ! grep -Eq 'GitHub deadline: context deadline exceeded' "$scratch/output";then failed=1;cat "$scratch/output" >&2;fi
expect 0 'cask Make target available' make -n -C "$root" check-homebrew
exit "$failed"
