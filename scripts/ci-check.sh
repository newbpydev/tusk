#!/usr/bin/env bash
set -euo pipefail
jq_binary_option=''
case "${OSTYPE:-}" in msys*|cygwin*) jq_binary_option=--binary ;; esac
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
die() { printf 'CI: %s\n' "$*" >&2; exit 1; }
require() {
    command -v "$1" >/dev/null 2>&1 || die "requires $1 on PATH"
    if [[ "$1" == jq && -n "$jq_binary_option" ]] && ! jq --binary --null-input empty >/dev/null 2>&1; then
        die 'Windows requires jq 1.7+ with binary output support; install a current native jq.exe'
    fi
}
digest() {
    if command -v sha256sum >/dev/null; then sha256sum "$1" | cut -d ' ' -f 1
    else shasum -a 256 "$1" | cut -d ' ' -f 1; fi
}
contract() {
    local checkout="${1:-$root}" lock
    lock="$checkout/scripts/tool-versions.json"
    require jq
    [[ -f "$lock" && -f "$checkout/.github/workflows/ci.yml" ]] || die 'missing workflow or tool catalogue'
    jq ${jq_binary_option:+"--binary"} -e '
      def sha: type == "string" and test("^[0-9a-f]{64}$");
      def version: type == "string" and test("^[0-9]+\\.[0-9]+\\.[0-9]+$");
      .schema == 1 and .go.minimum == "1.25.0" and (.go.release | version) and
      ([.actions[].sha | test("^[0-9a-f]{40}$")] | all) and
      (.actions | keys == ["checkout", "setup_go"]) and
      ([.go.archives[], .tools.actionlint.assets, .tools.goreleaser.assets] |
        all(.[] | .[]; (.sha256 | sha) and (.url | startswith("https://")))) and
      ([.tools[].version] | all(version)) and
      .tools.govulncheck.module == "golang.org/x/vuln" and
      (.tools.govulncheck.sum | test("^h1:[A-Za-z0-9+/]{43}=$")) and
      (.tools.govulncheck.source_sha | test("^[0-9a-f]{40}$")) and
      ([.go.archives[], .tools.actionlint.assets, .tools.goreleaser.assets] |
        all(keys == ["darwin_amd64", "darwin_arm64", "linux_amd64", "linux_arm64", "windows_amd64"]))
    ' "$lock" >/dev/null || die 'invalid immutable tool pin catalogue'
    # JSON is valid YAML; structured contracts avoid a lossy ad-hoc YAML parser.
    jq ${jq_binary_option:+"--binary"} -e --slurpfile pins "$lock" '
      .on == {pull_request:{branches:["main"]},push:{branches:["main"]},workflow_dispatch:{},workflow_call:{inputs:{source_sha:{required:true,type:"string"}}}} and
      .permissions == {contents:"read"} and (.jobs | keys == ["minimum", "native"]) and
      (tostring | test("secrets\\."; "i") | not) and
      .concurrency["cancel-in-progress"] == true and
      .jobs.native.strategy["fail-fast"] == false and
      .jobs.native.strategy.matrix.include == [
        {runner:"ubuntu-24.04",os:"linux",arch:"amd64"},
        {runner:"ubuntu-24.04-arm",os:"linux",arch:"arm64"},
        {runner:"macos-15-intel",os:"darwin",arch:"amd64"},
        {runner:"macos-15",os:"darwin",arch:"arm64"},
        {runner:"windows-2025",os:"windows",arch:"amd64"}] and
      .jobs.minimum["runs-on"] == "ubuntu-24.04" and
      all(.jobs[]; .["timeout-minutes"] == 30 and .defaults.run.shell == "bash" and
        (has("permissions") | not) and (has("secrets") | not)) and
      all(.jobs[].steps[] | select(has("uses"));
        (.uses == ("actions/checkout@" + $pins[0].actions.checkout.sha) and
          .with == {"persist-credentials":false,ref:"${{ inputs.source_sha || github.sha }}"}) or
        (.uses == ("actions/setup-go@" + $pins[0].actions.setup_go.sha) and .with.cache == true)) and
      ([.jobs.native.steps[] | select(.name == "Release compiler").with["go-version"]] == [$pins[0].go.release]) and
      ([.jobs.minimum.steps[] | select(.name == "Minimum compiler").with["go-version"]] == [$pins[0].go.minimum]) and
      ([.jobs.native.steps[] | select(.name == "Canonical gates").run] == ["make setup preflight-ci setup-sqlc validate build check-generated check-ci\nmake check-ci-drift"]) and
      ([.jobs.minimum.steps[] | select(.name == "Canonical gates").run] == ["make setup preflight-ci setup-sqlc validate build check-generated build-tui\nmake check-ci-drift"])
    ' "$checkout/.github/workflows/ci.yml" >/dev/null || die 'workflow violates native/read-only canonical contract'
}
preflight() {
    local go_cmd="${TUSK_GO_BIN:-go}" make_cmd="${TUSK_CI_MAKE:-make}" compiler
    require "$go_cmd"; require "$make_cmd"; require git; require jq; require curl
    local make_banner
    make_banner=$("$make_cmd" --version | sed -n '1p') || die "Make --version failed (status $?)"
    [[ "$make_banner" == *'GNU Make'* ]] || die 'GNU Make is required'
    local identity=() line
    while IFS= read -r line; do identity+=("$line"); done < <("$go_cmd" env GOOS GOARCH GOVERSION CGO_ENABLED GOHOSTOS GOHOSTARCH)
    [[ ${#identity[@]} == 6 ]] || die 'could not read native Go identity'
    [[ "${identity[0]}" == "${identity[4]}" && "${identity[1]}" == "${identity[5]}" ]] || die 'cross compiler cannot replace a native host gate'
    case "${identity[0]}:$(uname -s)" in
        linux:Linux|darwin:Darwin|windows:MINGW*|windows:MSYS*) ;;
        *) die 'native OS/shell mismatch; Windows requires Git Bash, not WSL' ;;
    esac
    [[ "${identity[0]}" == "${TUSK_CI_OS:-${identity[0]}}" && "${identity[1]}" == "${TUSK_CI_ARCH:-${identity[1]}}" ]] || die 'Go OS/architecture differs from native job; cross-build is not acceptance'
    [[ -z "${TUSK_CI_GO:-}" || "${identity[2]}" == "go$TUSK_CI_GO" ]] || die "expected Go $TUSK_CI_GO, found ${identity[2]}"
    [[ "${identity[3]}" == 1 ]] || die 'race/coverage gates require CGO_ENABLED=1'
    compiler="${TUSK_CI_CC:-${CC:-gcc}}"
    if [[ "${identity[0]}" == darwin ]]; then compiler="${TUSK_CI_CC:-${CC:-clang}}"; fi
    require "$compiler"
    local machine
    machine=$("$compiler" -dumpmachine) || die 'native C compiler identity failed'
    case "${identity[0]}/${identity[1]}:$machine" in
        linux/amd64:x86_64*linux*|linux/arm64:aarch64*linux*|darwin/amd64:x86_64*apple*|darwin/arm64:arm64*apple*|darwin/arm64:aarch64*apple*|windows/amd64:x86_64*w64*mingw*) ;;
        *) die "C compiler target $machine does not match native Go ${identity[0]}/${identity[1]}" ;;
    esac
    printf 'Native job: %s/%s %s; C target %s\n' "${identity[0]}" "${identity[1]}" "${identity[2]}" "$machine"
    "$go_cmd" version
    printf '%s\n' "$make_banner"
    "$compiler" --version | sed -n '1p'
}
setup() (
    require jq; require curl
    local target tool_dir binary archive member expected stage
    target="$(go env GOOS)_$(go env GOARCH)"
    tool_dir="${TUSK_CI_TOOL_DIR:-$root/bin/tools}"
    binary=actionlint
    [[ "$target" != windows_amd64 ]] || binary=actionlint.exe
    expected=$(jq ${jq_binary_option:+"--binary"} -er --arg target "$target" '.tools.actionlint.assets[$target].sha256' "$root/scripts/tool-versions.json")
    stage=$(mktemp -d)
    trap 'rm -rf "$stage"' EXIT
    archive="$stage/archive"
    curl --fail --location --proto '=https' --tlsv1.2 --output "$archive" \
        "$(jq ${jq_binary_option:+"--binary"} -er --arg target "$target" '.tools.actionlint.assets[$target].url' "$root/scripts/tool-versions.json")" || die 'actionlint download failed'
    [[ "$(digest "$archive")" == "$expected" ]] || die 'actionlint archive digest mismatch'
    if [[ "$target" == windows_amd64 ]]; then
        require unzip
        unzip -p "$archive" "$binary" >"$stage/$binary" || die 'actionlint extraction failed'
    else
        member=$(tar -tzf "$archive" | grep -x "$binary") || die 'actionlint member missing'
        tar -xOzf "$archive" "$member" >"$stage/$binary" || die 'actionlint extraction failed'
    fi
    chmod +x "$stage/$binary"
    [[ "$("$stage/$binary" -version | head -1)" == "$(jq ${jq_binary_option:+"--binary"} -r '.tools.actionlint.version' "$root/scripts/tool-versions.json")" ]] || die 'incorrect actionlint version'
    mkdir -p "$tool_dir"
    local install_stage
    install_stage=$(mktemp "$tool_dir/.actionlint.XXXXXX")
    trap 'rm -rf "$stage"; rm -f "$install_stage"' EXIT
    cp "$stage/$binary" "$install_stage"
    chmod +x "$install_stage"
    mv -f "$install_stage" "$tool_dir/$binary"
)
drift() {
    local checkout="${1:-$root}"
    git -C "$checkout" diff --exit-code HEAD -- || die 'tracked source drift after canonical gates'
    [[ -z "$(git -C "$checkout" ls-files --others --exclude-standard)" ]] || die 'unexpected untracked source after canonical gates'
}
setup_vulnerabilities() (
    require jq; require go
    local version module expected metadata stage binary tool_dir install_stage=''
    version=$(jq ${jq_binary_option:+"--binary"} -er '.tools.govulncheck.version' "$root/scripts/tool-versions.json")
    module=$(jq ${jq_binary_option:+"--binary"} -er '.tools.govulncheck.module' "$root/scripts/tool-versions.json")
    expected=$(jq ${jq_binary_option:+"--binary"} -er '.tools.govulncheck.sum' "$root/scripts/tool-versions.json")
    metadata=$(go mod download -json "$module@v$version") || die 'analyzer source download failed'
    [[ "$(printf '%s' "$metadata" | jq ${jq_binary_option:+"--binary"} -er '.Sum')" == "$expected" ]] || die 'analyzer source digest mismatch'
    stage=$(mktemp -d)
    trap 'rm -rf "$stage"; if [[ -n "$install_stage" ]]; then rm -f "$install_stage"; fi' EXIT
    GOBIN="$stage" go install "$module/cmd/govulncheck@v$version" || die 'analyzer build failed'
    binary=govulncheck
    [[ "$(go env GOOS)" != windows ]] || binary+=.exe
    go version -m "$stage/$binary" | grep -F "$module" | grep -F "v$version" >/dev/null || die 'incorrect analyzer build identity'
    tool_dir="${TUSK_CI_TOOL_DIR:-$root/bin/tools}"
    mkdir -p "$tool_dir"
    install_stage=$(mktemp "$tool_dir/.govulncheck.XXXXXX")
    cp "$stage/$binary" "$install_stage"
    chmod +x "$install_stage"
    mv -f "$install_stage" "$tool_dir/$binary"
)
case "${1:-check}" in
    contract) contract "${2:-$root}" ;;
    preflight) preflight ;;
    setup) setup ;;
    setup-vulnerabilities) setup_vulnerabilities ;;
    drift) drift "${2:-$root}" ;;
    check)
        contract "$root"
        tool="${TUSK_CI_TOOL_DIR:-$root/bin/tools}/actionlint"
        [[ "$(go env GOOS)" != windows ]] || tool+=.exe
        [[ -x "$tool" ]] || die 'run make setup-ci for pinned actionlint'
        [[ "$("$tool" -version | head -1)" == "$(jq ${jq_binary_option:+"--binary"} -r '.tools.actionlint.version' "$root/scripts/tool-versions.json")" ]] || die 'incorrect actionlint version'
        "$tool" -shellcheck= -pyflakes= "$root/.github/workflows/ci.yml" "$root/.github/workflows/release.yml" "$root/.github/workflows/native-candidate.yml"
        ;;
    *) die 'expected contract, preflight, setup, setup-vulnerabilities, drift or check' ;;
esac
