#!/usr/bin/env bash
set -euo pipefail
jq_binary_option=''
case "${OSTYPE:-}" in msys*|cygwin*) jq_binary_option=--binary ;; esac
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
die() { printf 'Release: %s\n' "$*" >&2; exit 1; }
digest() {
    if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | awk '{print $1}'
    else shasum -a 256 "$1" | awk '{print $1}'; fi
}
valid_version() { [[ "$1" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; }
setup_tool() (
    local target binary expected url stage directory install_stage=''
    target="$(go env GOHOSTOS)_$(go env GOHOSTARCH)"
    binary=goreleaser
    [[ "$target" != windows_amd64 ]] || binary+=.exe
    directory=${TUSK_RELEASE_TOOL_DIR:-$root/bin/tools}
    expected=$(jq ${jq_binary_option:+"--binary"} -er --arg target "$target" '.tools.goreleaser.assets[$target].sha256' "$root/scripts/tool-versions.json")
    url=$(jq ${jq_binary_option:+"--binary"} -er --arg target "$target" '.tools.goreleaser.assets[$target].url' "$root/scripts/tool-versions.json")
    stage=$(mktemp -d)
    trap 'rm -rf "$stage"; if [[ -n "$install_stage" ]]; then rm -f "$install_stage"; fi' EXIT
    curl --fail --location --proto '=https' --tlsv1.2 --output "$stage/archive" "$url" || die 'packager download failed'
    [[ "$(digest "$stage/archive")" == "$expected" ]] || die 'packager digest mismatch'
    if [[ "$target" == windows_amd64 ]]; then
        unzip -p "$stage/archive" "$binary" >"$stage/$binary" || die 'packager extraction failed'
    else
        tar -xOzf "$stage/archive" "$binary" >"$stage/$binary" || die 'packager extraction failed'
    fi
    chmod +x "$stage/$binary"
    mkdir -p "$directory"
    install_stage=$(mktemp "$directory/.goreleaser.XXXXXX")
    cp "$stage/$binary" "$install_stage"
    chmod +x "$install_stage"
    mv -f "$install_stage" "$directory/$binary"
    install_stage=$(mktemp "$directory/.goreleaser-archive.XXXXXX")
    cp "$stage/archive" "$install_stage"
    chmod 644 "$install_stage"
    mv -f "$install_stage" "$directory/goreleaser.archive"
    jq ${jq_binary_option:+"--binary"} -n --arg archive "$expected" --arg binary "$(digest "$directory/$binary")" '{archive_sha256:$archive,binary_sha256:$binary}' >"$directory/goreleaser-pin.json"
)
contract() {
    [[ -f "$root/.goreleaser.yaml" ]] || die 'packager configuration missing'
    jq ${jq_binary_option:+"--binary"} -e '
        keys == (["version","project_name","dist","builds","archives","source","checksum","changelog","snapshot","release","homebrew_casks"]|sort) and
        .version == 2 and .project_name == "tusk" and .dist == "dist" and
        .builds == ([{id:"tusk",main:"./cmd/tusk",binary:"tusk",env:["CGO_ENABLED=0"],
          ldflags:["-s -w"],flags:["-trimpath", "-overlay={{ .Env.TUSK_RELEASE_OVERLAY }}"],
          goos:["linux","windows"],goarch:["amd64","arm64"],
          ignore:[{goos:"windows",goarch:"arm64"}],mod_timestamp:"{{ .CommitTimestamp }}"}] | . + [ (.[0] | .id="tusk-macos" | .goos=["darwin"] | del(.ignore)) ]) and
        .archives == ([{id:"tusk",ids:["tusk"],name_template:"tusk_{{ .Version }}_{{ .Os }}_{{ .Arch }}",
          formats:["tar.gz"],format_overrides:[{goos:"windows",formats:["zip"]}],
          builds_info:{owner:"root",group:"root",mtime:"{{ .CommitDate }}"},
          files:([{src:"README.md",dst:"."},{src:"LICENSE",dst:"."},{src:"THIRD_PARTY_NOTICES.md",dst:"."},
            {src:"CONTRIBUTING.md",dst:"."},{src:"SECURITY.md",dst:"."},
            {src:"docs/completions/*",dst:"completions"},{src:"docs/man/*.1",dst:"man"},
            {src:"docs/cli.md",dst:"docs"},{src:"docs/tui.md",dst:"docs"},
            {src:"docs/install.md",dst:"docs"},{src:"docs/service.md",dst:"docs"},
            {src:"docs/assets/*",dst:"docs/assets"}] |
            map(.+{strip_parent:true,info:{owner:"root",group:"root",mode:420,mtime:"{{ .CommitDate }}"}}))}] | . + [(.[0] | .id="tusk-macos" | .ids=["tusk-macos"] | del(.format_overrides))]) and
        .homebrew_casks == [{"name":"tusk","ids":["tusk-macos"],"binaries":["tusk"],"manpages":["man/tusk-add.1","man/tusk-completion.1","man/tusk-delete.1","man/tusk-done.1","man/tusk-edit.1","man/tusk-history.1","man/tusk-list.1","man/tusk-stats.1","man/tusk-tree.1","man/tusk-tui.1","man/tusk-version.1","man/tusk.1"],"completions":{"bash":"completions/tusk.bash","zsh":"completions/tusk.zsh","fish":"completions/tusk.fish"},"homepage":"https://github.com/newbpydev/tusk","description":"Local task management in your terminal","custom_block":"depends_on :macos","skip_upload":true,"repository":{"owner":"newbpydev","name":"homebrew-tap"},"directory":"Casks","url":{"template":"https://github.com/newbpydev/tusk/releases/download/{{ .Tag }}/{{ .ArtifactName }}"}}] and
        .source == {enabled:true,format:"tar.gz",name_template:"tusk_{{ .Version }}_source",prefix_template:"tusk-{{ .Version }}/"} and
        .checksum == {disable:true} and .changelog == {disable:true} and
        .snapshot == {version_template:"{{ .Version }}-dev"} and
        .release == {disable:true,github:{owner:"newbpydev",name:"tusk"}}
    ' "$root/.goreleaser.yaml" >/dev/null || die 'packager violates immutable/local five-target contract'
}
verified_tool() (
    local target binary directory expected stage
    target="$(go env GOHOSTOS)_$(go env GOHOSTARCH)"
    binary=goreleaser
    [[ "$target" != windows_amd64 ]] || binary+=.exe
    directory=${TUSK_RELEASE_TOOL_DIR:-$root/bin/tools}
    expected=$(jq ${jq_binary_option:+"--binary"} -er --arg target "$target" '.tools.goreleaser.assets[$target].sha256' "$root/scripts/tool-versions.json")
    [[ -f "$directory/goreleaser.archive" && -x "$directory/$binary" ]] || die 'run make setup-release first'
    [[ "$(digest "$directory/goreleaser.archive")" == "$expected" ]] || die 'cached packager archive digest mismatch'
    stage=$(mktemp -d)
    trap 'rm -rf "$stage"' EXIT
    if [[ "$target" == windows_amd64 ]]; then unzip -p "$directory/goreleaser.archive" "$binary" >"$stage/binary"
    else tar -xOzf "$directory/goreleaser.archive" "$binary" >"$stage/binary"; fi
    cmp -s "$stage/binary" "$directory/$binary" || die 'installed packager differs from verified archive'
    printf '%s\n' "$directory/$binary"
)
preflight() {
    local tool compiler
    contract
    compiler=$(jq ${jq_binary_option:+"--binary"} -er '.go.release' "$root/scripts/tool-versions.json")
    [[ "$(GOTOOLCHAIN="go$compiler" go env GOVERSION)" == "go$compiler" ]] || die 'pinned release compiler unavailable'
    tool=$(verified_tool)
    env -i PATH="$PATH" "$tool" check --config "$root/.goreleaser.yaml"
    GOTOOLCHAIN="go$compiler" make -C "$root" check-docs check-notices check-generated check-modules
}
package() (
    local mode=$1 sha output parent stage compiler tool tool_directory status=1
    valid_version "${RELEASE_VERSION:-}" || die 'RELEASE_VERSION must be vMAJOR.MINOR.PATCH (no leading zeros or suffix)'
    [[ -n "${RELEASE_OUTPUT:-}" && -n "${RELEASE_SHA:-}" ]] || die 'explicit RELEASE_OUTPUT and RELEASE_SHA required'
    [[ "$RELEASE_SHA" =~ ^[0-9a-f]{40}$ || "$RELEASE_SHA" == HEAD ]] || die 'RELEASE_SHA must be HEAD or a full commit SHA'
    sha=$(git -C "$root" rev-parse --verify "$RELEASE_SHA^{commit}") || die 'source commit not found'
    if git -C "$root" show-ref --verify --quiet "refs/tags/$RELEASE_VERSION"; then
        [[ "$(git -C "$root" rev-parse "$RELEASE_VERSION^{commit}")" == "$sha" ]] || die 'existing tag conflicts with source'
    fi
    parent=$(cd "$(dirname "$RELEASE_OUTPUT")" && pwd -P) || die 'output parent must exist'
    output="$parent/$(basename "$RELEASE_OUTPUT")"
    [[ ! -e "$output" && ! -L "$output" && "$output" != "$root" && "$output" != "$root/.git" && "$output" != "$root/.git/"* ]] || die 'output collision or unsafe location'
    # Output is retained, including failures. Only this invocation-created stage is removed.
    mkdir "$output"
    stage=$(mktemp -d)
    trap 'printf "{\"schema\":1,\"status\":%s}\n" "$status" >"$output/result.json"; rm -rf "$stage"' EXIT
    trap 'exit 130' INT
    trap 'exit 143' TERM
    git clone --quiet --no-hardlinks --no-checkout "$root" "$stage/source" >"$output/build.log" 2>&1
    git -C "$stage/source" fetch --quiet "$root" "$sha" >>"$output/build.log" 2>&1
    git -C "$stage/source" config core.autocrlf false
    git -C "$stage/source" config tar.umask 0022
    git -C "$stage/source" checkout --quiet --detach "$sha" >>"$output/build.log" 2>&1
    [[ -z "$(git -C "$stage/source" status --porcelain)" ]] || die 'source checkout is dirty'
    if ! git -C "$stage/source" show-ref --verify --quiet "refs/tags/$RELEASE_VERSION"; then
        git -C "$stage/source" tag "$RELEASE_VERSION" "$sha"
    fi
    # Only the owned clone has an intended local tag and release repository URL.
    git -C "$stage/source" remote set-url origin https://github.com/newbpydev/tusk.git
    compiler=$(jq ${jq_binary_option:+"--binary"} -er '.go.release' "$stage/source/scripts/tool-versions.json")
    tool_directory=${TUSK_RELEASE_TOOL_DIR:-$root/bin/tools}
    env -u TUSK_SQLC_BIN -u TUSK_SQLC_HOST GOTOOLCHAIN="go$compiler" TUSK_RELEASE_TOOL_DIR="$tool_directory" make -C "$stage/source" setup-sqlc release-check >>"$output/build.log" 2>&1
    tool=$(verified_tool)
    (cd "$stage/source" && GOTOOLCHAIN="go$compiler" bash scripts/release_check.sh overlay "$stage/source" "$stage/overlay" "$RELEASE_VERSION" "$mode") >>"$output/build.log" 2>&1
    cp "$stage/overlay/input.json" "$output/version-input.json"
    # GoReleaser does not template dist. Only this owned absolute field changes.
    jq ${jq_binary_option:+"--binary"} --arg dist "$stage/dist" '.dist=$dist' "$stage/source/.goreleaser.yaml" >"$stage/config.json"
    cp "$stage/config.json" "$output/config-input.json"
    mkdir "$stage/home" "$stage/tmp"
    local args=(release --skip=publish --clean --config "$stage/config.json")
    [[ "$mode" != snapshot ]] || args+=(--snapshot)
    # Credentials, agent settings and user configuration are absent from the packager environment.
    # Each phase retains its own failure boundary.
    # shellcheck disable=SC2129
    (cd "$stage/source" && env -i PATH="$PATH" HOME="$stage/home" TMPDIR="$stage/tmp" \
        GOTOOLCHAIN="go$compiler" GOMODCACHE="$(go env GOMODCACHE)" GOCACHE="$(go env GOCACHE)" \
        TZ=UTC LANG=C.UTF-8 TUSK_RELEASE_OVERLAY="$stage/overlay/overlay.json" \
        "$tool" "${args[@]}") >>"$output/build.log" 2>&1
    (cd "$stage/source" && GOTOOLCHAIN="go$compiler" bash scripts/release_check.sh finalize "$stage/source" "$stage/dist" "$output/assets" "$stage/overlay/input.json" "$sha" "$compiler" "$mode") >>"$output/build.log" 2>&1
    (cd "$stage/source" && GOTOOLCHAIN="go$compiler" bash scripts/release_check.sh verify "$output/assets") >>"$output/build.log" 2>&1
    if [[ "$mode" == candidate ]]; then
        mkdir -p "$output/homebrew/Casks"
        cp "$stage/dist/homebrew/Casks/tusk.rb" "$output/homebrew/Casks/tusk.rb"
        (cd "$stage/source" && GOTOOLCHAIN="go$compiler" bash scripts/release_check.sh check-cask "$output/assets/release-manifest.json" "$output/homebrew/Casks/tusk.rb") >>"$output/build.log" 2>&1
    fi
    [[ -z "$(git -C "$stage/source" status --porcelain)" ]] || die 'packaging mutated source checkout'
    status=0
    printf 'Verified local %s: %s\n' "$mode" "$output/assets"
)
case "${1:-check}" in
    setup) setup_tool ;;
    contract) contract ;;
    check) preflight ;;
    candidate|snapshot) package "$1" ;;
    *) die 'expected setup, contract, check, snapshot or candidate' ;;
esac
