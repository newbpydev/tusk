#!/usr/bin/env bash
set -euo pipefail
root=${1:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}
fail() { printf 'Documentation: %s\n' "$*" >&2; exit 1; }
required() { grep -Eiq -- "$2" "$root/$1" || fail "$1 lacks $3"; }
files=(README.md docs/install.md docs/releasing.md docs/cli.md docs/tui.md CONTRIBUTING.md SECURITY.md docs/assets/README.md)
for file in "${files[@]}" LICENSE THIRD_PARTY_NOTICES.md docs/assets/tusk-tui.png; do
    [[ -s "$root/$file" ]] || fail "missing $file"
done
required README.md 'source installation' 'available installation route'
required README.md 'no public release' 'unreleased status'
required docs/install.md 'Go 1\.25' 'minimum Go'
required docs/install.md '^Install Git,.*Bash' 'source Bash prerequisite'
required docs/install.md 'GNU Make' 'source Make prerequisite'
required docs/install.md 'third_party' 'local replacement explanation'
required docs/install.md 'PowerShell' 'native Windows route'
required docs/install.md 'Get-FileHash' 'Windows checksum verification'
required docs/install.md 'sha256sum|shasum' 'Unix checksum verification'
required docs/install.md 'CGO_ENABLED=0' 'self-contained source build'
required docs/install.md 'sidecars|WAL/SHM' 'backup safety'
required docs/install.md 'NORMAL' 'durability limits'
required SECURITY.md 'https://github.com/newbpydev' 'verified owner route'
required SECURITY.md '^GitHub private vulnerability reporting ((is|was) enabled and verified on [0-9]{4}-[0-9]{2}-[0-9]{2}\.|is (currently )?(disabled|not enabled)\.)[[:blank:]]*$' 'actual private-reporting state'
[[ $(grep -Eic '^[[:blank:]]*GitHub private vulnerability reporting[[:blank:]]' "$root/SECURITY.md") == 1 ]] ||
    fail 'SECURITY.md requires one private-reporting state declaration'
required LICENSE 'Permission is hereby granted' 'selected MIT grant'
required THIRD_PARTY_NOTICES.md 'third_party/bubbletea' 'replacement notice'
required THIRD_PARTY_NOTICES.md 'third_party/glamour' 'replacement notice'
unsupported_claims="go install github.com/newbpydev/tusk[^[:space:]]*@|--due[ =]+[\"']*next week|Guaranteed sub-15ms|guaranteed.*latency"
for file in "${files[@]}"; do
    if grep -Eq -- "$unsupported_claims" "$root/$file"; then
        fail "unsupported install/date/performance claim in $file"
    fi
    # Release links still require publication. Source CI/license badges were
    # activated separately after live workflow and license readback.
    if [[ "$file" == README.md ]] && grep -Eq '/releases/download/' "$root/$file"; then
        fail 'README advertises an unverified badge or release'
    fi
    if [[ "$file" == README.md ]]; then
        # Read complete URL targets before removing the two verified literals.
        # Quoted HTML, simple Markdown/angle targets and bare references have
        # different terminators. An embedded URL cannot become a second target.
        badge_text=$(awk -v ci='https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main' \
            -v license='https://img.shields.io/github/license/newbpydev/tusk' '
            function badge(value) {
                return tolower(value) ~ /img[.]shields[.]io|actions\/workflows\/[^[:space:]]*badge/
            }
            {
                # A quoted src value is one target, including any spaces.
                attrs = $0
                while (match(tolower(attrs), /(^|[[:space:]])src[[:space:]]*=[[:space:]]*["\047]/)) {
                    quote = substr(attrs, RSTART + RLENGTH - 1, 1)
                    attrs = substr(attrs, RSTART + RLENGTH)
                    stop = index(attrs, quote)
                    value = stop ? substr(attrs, 1, stop - 1) : attrs
                    if (badge(value) && (!stop || (value != ci && value != license))) { print; next }
                    attrs = stop ? substr(attrs, stop + 1) : ""
                }
                cursor = 1; rest = $0; clean = ""; unquoted = 0
                while (match(tolower(rest), /(https?:)?\/\//)) {
                    start = cursor + RSTART - 1
                    before = substr($0, 1, start - 1)
                    tail = substr($0, start)
                    last = substr(before, length(before), 1)
                    delimiter = ""
                    if (last == "\"" || last == "\047") delimiter = last
                    else if (last == "(") delimiter = ")"
                    else if (last == "<") delimiter = ">"
                    stop = delimiter == "" ? 0 : index(tail, delimiter)
                    if (delimiter != "") value = stop ? substr(tail, 1, stop - 1) : tail
                    else {
                        unquoted = tolower(before) ~ /(^|[[:space:]])(src|href)=$/
                        if (unquoted) match(tail, /^[^[:space:]>"\047]+/)
                        else match(tail, /^[^[:space:]"\047]+/)
                        value = substr(tail, 1, RLENGTH)
                    }
                    left = before == "" || last ~ /[[:space:]"\047(<]/ || unquoted
                    if (badge(value)) {
                        if (!left || (delimiter != "" && !stop) || (value != ci && value != license)) { print; next }
                        clean = clean substr($0, cursor, start - cursor)
                    } else clean = clean substr($0, cursor, start - cursor) value
                    cursor = start + length(value)
                    rest = substr($0, cursor)
                    unquoted = 0
                }
                print clean rest
            }' "$root/$file") || fail 'README badge scan failed'
        if grep -iE 'img\.shields\.io|actions/workflows/[^[:space:]]*badge' <<<"$badge_text" >/dev/null; then
            fail 'README advertises an unverified badge'
        fi
    fi
    # Links in these guides use simple Markdown targets, without titles/spaces.
    while IFS= read -r target; do
        if [[ "$target" == '<'*'>' ]]; then target=${target:1:${#target}-2}; fi
        case "$target" in https://*|http://*|mailto:*) continue ;; esac
        base=${target%%#*}
        if [[ -z "$base" ]]; then linked="$root/$file"; else linked="$root/$(dirname "$file")/$base"; fi
        [[ -e "$linked" ]] || fail "$file has broken link: $target"
        if [[ "$target" == *'#'* ]]; then
            fragment=${target#*#}
            anchors=$(awk '/^#+ / { sub(/^#+ /, ""); print }' "$linked" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9 _-]//g; s/ /-/g')
            grep -Fxq -- "$fragment" <<<"$anchors" || fail "$file has missing anchor: $target"
        fi
    done < <(awk '{ line=$0; while (match(line, /\]\([^ )]+\)/)) { value=substr(line,RSTART+2,RLENGTH-3); print value; line=substr(line,RSTART+RLENGTH) } }' "$root/$file")
done
printf 'Public guide contracts and local links pass; hosted rendering remains separate.\n'
