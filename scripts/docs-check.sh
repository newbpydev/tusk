#!/usr/bin/env bash
set -euo pipefail
source_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
root=${1:-$source_root}
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
        if [[ -n "${TUSK_DOCCHECK_BIN:-}" ]]; then
            "$TUSK_DOCCHECK_BIN" "$root/$file" || fail 'README Markdown target policy failed'
        else
            (cd "$source_root" && go run ./scripts/doccheck "$root/$file") || fail 'README Markdown target policy failed'
        fi
        # Read complete URL targets before removing the two verified literals.
        # Quoted HTML, simple Markdown/angle targets and bare references have
        # different terminators. An embedded URL cannot become a second target.
        badge_text=$(LC_ALL=C awk -v ci='https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main' \
            -v license='https://img.shields.io/github/license/newbpydev/tusk' '
            function badge(value) {
                return tolower(value) ~ /img[.]shields[.]io|actions\/workflows\/[^[:space:]]*badge/
            }
            function literal(value, host) {
                # README targets use literal URLs, not renderer-dependent
                # entities, percent escapes, backslashes or split lines.
                if (value == "" || value ~ /[[:space:]]/ || index(value, "\\") ||
                    value ~ /&#|&[[:alpha:]][[:alnum:]]*;|%[[:xdigit:]][[:xdigit:]]/) {
                    printf "README URL target requires a literal value: %s\n", value > "/dev/stderr"
                    exit 42
                }
                if (match(tolower(value), /^(https?:)?\/\//)) {
                    host = substr(value, RLENGTH + 1)
                    sub(/[\/?#].*$/, "", host); sub(/^.*@/, "", host)
                    if (host ~ /[^ -~]/ || index(host, "&")) {
                        printf "README URL host requires literal ASCII: %s\n", host > "/dev/stderr"
                        exit 42
                    }
                }
            }
            function image(value) {
                if ((value ~ /^[[:alpha:]][[:alnum:]+.-]*:/ || value ~ /^\//) && value != ci && value != license) {
                    printf "README advertises an unverified badge: remote image/reference target %s\n", value > "/dev/stderr"
                    exit 42
                }
            }
            function tag_end(text, i, ch, quote) {
                for (i = 1; i <= length(text); i++) {
                    ch = substr(text, i, 1)
                    if (quote != "") { if (ch == quote) quote = "" }
                    else if (ch == "\"" || ch == "\047") quote = ch
                    else if (ch == ">") return i
                }
                return 0
            }
            function target(text, stop) {
                sub(/^[[:blank:]]*/, "", text)
                if (substr(text, 1, 1) == "<") {
                    stop = index(text, ">")
                    if (stop) return substr(text, 2, stop - 2)
                    text = substr(text, 2)
                }
                match(text, /^[^[:space:])]+/)
                return substr(text, 1, RLENGTH)
            }
            {
                tags = $0
                while (match(tolower(tags), /<(img|source|a)([[:space:]>\/]|$)/)) {
                    tags = substr(tags, RSTART)
                    stop = tag_end(tags)
                    if (!stop) exit 42
                    tags = substr(tags, stop + 1)
                }
                # Check complete attribute targets even if their scheme or
                # hostname is obscured, before scanning known badge hosts.
                attrs = $0
                while (match(tolower(attrs), /(^|[[:space:]])(src|href|srcset)[[:space:]]*=[[:space:]]*/)) {
                    attribute = tolower(substr(attrs, RSTART, RLENGTH))
                    attrs = substr(attrs, RSTART + RLENGTH)
                    quote = substr(attrs, 1, 1)
                    if (quote == "\"" || quote == "\047") {
                        attrs = substr(attrs, 2); stop = index(attrs, quote)
                        if (!stop) exit 42
                        value = substr(attrs, 1, stop - 1)
                        attrs = substr(attrs, stop + 1)
                    } else {
                        match(attrs, /^[^[:space:]>]+/)
                        value = substr(attrs, 1, RLENGTH)
                        attrs = substr(attrs, RLENGTH + 1)
                    }
                    literal(value)
                    if (attribute !~ /href/) image(value)
                    if (attribute ~ /srcset/ && index(value, ",")) exit 42
                    if (badge(value) && value != ci && value != license) { print; next }
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
                    if (last == "(") value = target(value)
                    left = before == "" || last ~ /[[:space:]"\047(<]/ || unquoted
                    if (badge(value)) {
                        if (!left || (delimiter != "" && delimiter != ")" && !stop) || (value != ci && value != license)) { print; next }
                        clean = clean substr($0, cursor, start - cursor)
                    } else clean = clean substr($0, cursor, start - cursor) value
                    cursor = start + length(value)
                    rest = substr($0, cursor)
                    unquoted = 0
                }
                print clean rest
            }' "$root/$file") || fail 'README badge scan failed: use complete literal URL targets'
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
