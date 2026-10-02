#!/usr/bin/env bash
set -euo pipefail
mode=${1:-check}
root=${2:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}
case "$mode" in check|generate) ;; *) echo 'Expected check or generate' >&2; exit 2 ;; esac
cd "$root"
scratch=$(mktemp -d)
stage=''
trap 'rm -rf "$scratch"; if [[ -n "$stage" ]]; then rm -f "$stage"; fi' EXIT
fail() { printf 'Notices: %s\n' "$*" >&2; exit 1; }
digest() {
    if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | awk '{print $1}'
    else shasum -a 256 "$1" | awk '{print $1}'; fi
}
inventory=scripts/notices.json
[[ -s LICENSE && -s "$inventory" ]] || fail 'first-party license or reviewed inventory missing'
jq -e '.contract == "tusk-notices/v1" and (.modules|length)>0 and
    all(.modules[]; (.licenses|length)>0) and
    all((.modules[].licenses[], .extras[].licenses[], .assets[]);
        (.classification | IN("MIT", "MIT-declaration", "BSD-2-Clause", "BSD-3-Clause", "MPL-2.0", "Apache-2.0", "attribution", "public-domain", "BSD-3-Clause AND MIT AND BSD-2-Clause AND public-domain")) and
        (.sha256|test("^[0-9a-f]{64}$")) and (.file|test("(^/|(^|/)\\.\\.(/|$))")|not)) and
    ([.extras[]|{name,root,files:[.licenses[].file]}] == [
        {name:"Go standard library",root:"go",files:["LICENSE"]},
        {name:"Embedded IANA timezone database",root:"go",files:["lib/time/README"]}
    ])' "$inventory" >/dev/null || fail 'unclassified, incomplete or unsafe inventory'
# Resolve full-graph license sources without adding build/test-only sums to the
# accepted module files. Go verifies downloads against the checksum database.
cp go.mod "$scratch/go.mod"
cp go.sum "$scratch/go.sum"
go mod download -modfile="$scratch/go.mod" all
go list -modfile="$scratch/go.mod" -m -json all >"$scratch/modules"
jq -s '[.[]|select(.Main != true)|{path:.Path,version:.Version,replacement:(.Replace.Path // "")}]|sort_by(.path)' "$scratch/modules" >"$scratch/current"
jq '[.modules[]|{path,version,replacement}]|sort_by(.path)' "$inventory" >"$scratch/accepted"
cmp -s "$scratch/current" "$scratch/accepted" || fail 'module graph differs from reviewed inventory'
go_root=$(go env GOROOT)
go_root=${go_root//\\//}
find docs/assets third_party/glamour/styles/gallery -type f \( -name '*.png' -o -name '*.jpg' -o -name '*.svg' \) | LC_ALL=C sort >"$scratch/assets-current"
jq -r '.assets[].file' "$inventory" | LC_ALL=C sort >"$scratch/assets-accepted"
cmp -s "$scratch/assets-current" "$scratch/assets-accepted" || fail 'checked-in asset inventory is incomplete'
{
    cat <<'EOF'
# Third-party notices

Generated from the reviewed `scripts/notices.json` by `make generate-notices`.
This includes the entire selected Go module graph (runtime, build and test
dependencies), Go standard-library terms, embedded IANA timezone data and
checked-in assets. Extra upstream terms are retained verbatim. Testdata/logos
from the module cache are not redistributed by Tusk's source/payload archives.

Local replacements `third_party/bubbletea` and `third_party/glamour` retain
upstream grants. Their `TUSK-PATCH.json` files record patch/source provenance;
Glamour gallery images retain their upstream MIT terms. The real sample capture
in docs/assets is first-party MIT with separate recorded image provenance.

go-localereader v0.0.1 declares MIT and names Yasuhiro Matsumoto in README.md,
rather than supplying a standalone license file. Its declaration is retained.
HashiCorp golang-lru's MPL-2.0 applies to its covered source; no changes to that
module are made. Source is available at the recorded upstream version. Including
its notice here does not change Tusk's first-party MIT license.
Covered golang-lru v2.0.7 source archive:
https://proxy.golang.org/github.com/hashicorp/golang-lru/v2/@v/v2.0.7.zip

EOF
    while IFS=$'\t' read -r module version replacement; do
        module_root=$(jq -rs --arg path "$module" '.[]|select(.Path==$path)|.Replace.Dir // .Dir // empty' "$scratch/modules")
        [[ -n "$module_root" ]] || fail "download sources first: go mod download $module@$version"
        module_root=${module_root//\\//}
        find "$module_root" -type f \( -iname 'license*' -o -iname 'licence*' -o -iname 'copying*' -o -iname 'copyright*' -o -iname 'notice*' \) \
            ! -path '*/testdata/*' ! -name '*.go' ! -name 'LICENSE-LOGO' | while IFS= read -r file; do
                printf '%s\n' "${file#"$module_root"/}"
            done | LC_ALL=C sort >"$scratch/grants-current"
        if [[ "$module" == github.com/mattn/go-localereader && ! -s "$scratch/grants-current" ]]; then
            printf 'README.md\n' >"$scratch/grants-current"
        fi
        jq -r --arg path "$module" '.modules[]|select(.path==$path)|.licenses[].file' "$inventory" | LC_ALL=C sort >"$scratch/grants-accepted"
        cmp -s "$scratch/grants-current" "$scratch/grants-accepted" || fail "incomplete grant inventory: $module"
        printf '\n## %s %s\n\n' "$module" "$version"
        # Literal Markdown backticks, never shell command substitutions.
        # shellcheck disable=SC2016
        if [[ -n "$replacement" ]]; then printf 'Replacement: `%s`\n\n' "$replacement"; fi
        printf 'Module source/version: https://pkg.go.dev/%s@%s\n\n' "$module" "$version"
        while IFS=$'\t' read -r file classification hash; do
            [[ -f "$module_root/$file" && "$(digest "$module_root/$file")" == "$hash" ]] || fail "missing/changed grant: $module/$file"
            printf '### %s (%s)\n\n```text\n' "$file" "$classification"
            cat "$module_root/$file"
            printf '\n```\n'
        done < <(jq -r --arg path "$module" '.modules[]|select(.path==$path)|.licenses[]|[.file,.classification,.sha256]|@tsv' "$inventory")
    done < <(jq -r '.modules[]|[.path,.version,.replacement]|@tsv' "$inventory")
    while IFS=$'\t' read -r name file classification hash; do
        [[ -f "$go_root/$file" && "$(digest "$go_root/$file")" == "$hash" ]] || fail "missing/changed $name notice"
        printf '\n## %s (%s)\n\n```text\n' "$name" "$classification"
        cat "$go_root/$file"
        printf '\n```\n'
    done < <(jq -r '.extras[]|.name as $name|.licenses[]|[$name,.file,.classification,.sha256]|@tsv' "$inventory")
    printf '\n## Checked-in assets\n\n'
    while IFS=$'\t' read -r file classification hash source; do
        [[ -f "$file" && "$(digest "$file")" == "$hash" ]] || fail "missing/changed asset: $file"
        # shellcheck disable=SC2016
        printf -- '- `%s`: %s; %s. SHA256 `%s`.\n' "$file" "$classification" "$source" "$hash"
    done < <(jq -r '.assets[]|[.file,.classification,.sha256,.source]|@tsv' "$inventory")
} >"$scratch/notices"
if [[ "$mode" == check ]]; then
    cmp -s "$scratch/notices" THIRD_PARTY_NOTICES.md || fail 'missing/stale generated notices; review inventory then make generate-notices'
else
    # Preserve existing notices if validation/generation fails; rename only complete bytes.
    stage=$(mktemp "$root/.notices.XXXXXX")
    cp "$scratch/notices" "$stage"
    chmod 644 "$stage"
    mv "$stage" THIRD_PARTY_NOTICES.md
fi
printf 'Reviewed dependency, replacement and asset notices verified.\n'
