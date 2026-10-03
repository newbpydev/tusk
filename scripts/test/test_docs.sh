#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
failed=0
expect() {
    local want=$1 name=$2 actual=0
    shift 2
    "$@" >"$scratch/result" 2>&1 || actual=$?
    if [[ "$actual" == "$want" ]]; then
        printf 'PASS: %s\n' "$name"
    else
        printf 'FAIL: %s (expected %s, got %s)\n' "$name" "$want" "$actual" >&2
        cat "$scratch/result" >&2
        failed=1
    fi
}
expect 0 'public documentation contract' bash "$root/scripts/docs-check.sh" "$root"
fixture="$scratch/checkout ü"
mkdir -p "$fixture"
cp -R "$root/docs" "$fixture/docs"
for file in README.md CHANGELOG.md CONTRIBUTING.md SECURITY.md LICENSE THIRD_PARTY_NOTICES.md AGENTS.md MASTERPLAN.md; do
    if [[ -f "$root/$file" ]]; then cp "$root/$file" "$fixture/"; fi
done
expect 0 'complete documentation fixture is a passing control' bash "$root/scripts/docs-check.sh" "$fixture"
cp "$root/README.md" "$fixture/README.md"
printf '\n[![Native CI](https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/newbpydev/tusk/actions/workflows/ci.yml)\n[![MIT](https://img.shields.io/github/license/newbpydev/tusk)](LICENSE)\n' >>"$fixture/README.md"
expect 0 'verified native CI and detected MIT badges are allowed' bash "$root/scripts/docs-check.sh" "$fixture"
cp "$root/README.md" "$fixture/README.md"
cp "$fixture/SECURITY.md" "$scratch/security-saved"
printf '# Security reporting\nGitHub private vulnerability reporting is enabled and verified on 2026-10-03.\nUse https://github.com/newbpydev/tusk/security/advisories/new to report privately.\n' >"$fixture/SECURITY.md"
expect 0 'verified enabled private reporting is allowed' bash "$root/scripts/docs-check.sh" "$fixture"
for state in 'is disabled' 'is currently disabled' 'is not enabled' 'is currently not enabled'; do
    printf '# Security reporting\nGitHub private vulnerability reporting %s.\nUse https://github.com/newbpydev/tusk to contact the maintainer.\n' "$state" >"$fixture/SECURITY.md"
    expect 0 "allow explicit disabled private reporting state: $state" bash "$root/scripts/docs-check.sh" "$fixture"
done
for state in 'was never enabled' 'is no longer enabled' 'is not enabled and verified'; do
    printf '# Security reporting\nGitHub private vulnerability reporting %s\nUse https://github.com/newbpydev/tusk/security/advisories/new.\n' "$state" >"$fixture/SECURITY.md"
    expect 1 "reject ambiguous private reporting state: $state" bash "$root/scripts/docs-check.sh" "$fixture"
done
mv "$scratch/security-saved" "$fixture/SECURITY.md"
cp "$fixture/SECURITY.md" "$scratch/security-saved"
for state in 'is enabled and verified on 2026-10-03' 'was enabled and verified on 2026-10-03' 'is disabled' 'is currently disabled' 'is not enabled' 'is currently not enabled'; do
    printf '# Security reporting\nGitHub private vulnerability reporting %s.  \nUse https://github.com/newbpydev/tusk.\n' "$state" >"$fixture/SECURITY.md"
    expect 0 "allow complete reporting-state line with trailing blanks: $state" bash "$root/scripts/docs-check.sh" "$fixture"
    printf '# Security reporting\nGitHub private vulnerability reporting %s. It is now disabled.\nUse https://github.com/newbpydev/tusk.\n' "$state" >"$fixture/SECURITY.md"
    expect 1 "reject trailing reporting-state text: $state" bash "$root/scripts/docs-check.sh" "$fixture"
done
printf '# Security reporting\nGitHub private vulnerability reporting is enabled and verified on 2026-10-03.\nGitHub private vulnerability reporting is disabled.\nUse https://github.com/newbpydev/tusk.\n' >"$fixture/SECURITY.md"
expect 1 'reject conflicting reporting-state declarations' bash "$root/scripts/docs-check.sh" "$fixture"
printf '# Security reporting\nGitHub private vulnerability reporting is disabled.\nGitHub private vulnerability reporting is disabled.\nUse https://github.com/newbpydev/tusk.\n' >"$fixture/SECURITY.md"
expect 1 'require one reporting-state declaration' bash "$root/scripts/docs-check.sh" "$fixture"
mv "$scratch/security-saved" "$fixture/SECURITY.md"
for missing in docs/install.md LICENSE docs/assets/tusk-tui.png; do
    if [[ -f "$fixture/$missing" ]]; then
        mv "$fixture/$missing" "$scratch/saved"
        expect 1 "reject missing $missing" bash "$root/scripts/docs-check.sh" "$fixture"
        mv "$scratch/saved" "$fixture/$missing"
    fi
done
for bad in '[Broken](docs/absent.md)' '[Bad anchor](docs/cli.md#absent-heading)' 'go install github.com/newbpydev/tusk/cmd/tusk@v0.3.0' '![Release](https://img.shields.io/github/v/release/newbpydev/tusk)' '![Wrong repo](https://github.com/evil/tusk/actions/workflows/ci.yml/badge.svg)' '![Static pass](https://img.shields.io/badge/CI-passing-green)' 'tusk add sample --due "next week"' 'Guaranteed sub-15ms on every platform'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$bad" >>"$fixture/README.md"
    expect 1 "reject unsupported public content: $bad" bash "$root/scripts/docs-check.sh" "$fixture"
done
cp "$root/README.md" "$fixture/README.md"
for badge in '<img src="https://img.shields.io/github/v/release/newbpydev/tusk">' '[badge]: https://img.shields.io/badge/CI-passing-green' '<img src="https://github.com/evil/tusk/actions/workflows/ci.yml/badge.svg">' '[badge]: https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main-evil'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$badge" >>"$fixture/README.md"
    expect 1 "reject unverified raw badge: $badge" bash "$root/scripts/docs-check.sh" "$fixture"
done
cp "$root/README.md" "$fixture/README.md"
printf '\n<img src="https://img.shields.io/github/license/newbpydev/tusk">\n[ci]: https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main\n' >>"$fixture/README.md"
expect 0 'verified raw HTML and reference badges are allowed' bash "$root/scripts/docs-check.sh" "$fixture"
cp "$root/README.md" "$fixture/README.md"
printf '\n<img src="//img.shields.io/badge/CI-passing-green">\n' >>"$fixture/README.md"
expect 1 'reject protocol-relative unverified badge' bash "$root/scripts/docs-check.sh" "$fixture"
for badge in '<img src="https://IMG.SHIELDS.IO/badge/CI-passing-green">' '<img src="//Img.Shields.Io/badge/CI-passing-green">'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$badge" >>"$fixture/README.md"
    expect 1 "reject mixed-case unverified badge host: $badge" bash "$root/scripts/docs-check.sh" "$fixture"
done
for url in 'https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main' 'https://img.shields.io/github/license/newbpydev/tusk'; do
    for badge in "![Verified]($url)" "<img src=\"$url\">" "<img src='$url'>" "<img src=$url>" "[verified]: $url" "[verified]: <$url>" "![Verified](<$url>)" "($url)"; do
        cp "$root/README.md" "$fixture/README.md"
        printf '\n%s\n' "$badge" >>"$fixture/README.md"
        expect 0 "allow complete approved badge target: $badge" bash "$root/scripts/docs-check.sh" "$fixture"
    done
    for target in "https://evil.example/proxy/$url" "https://evil.example/?url=$url" "https://evil.example/proxy/($url)" "https://evil.example/<$url>" "prefix-$url" "$url-extra" "$url)extra"; do
        for badge in "![Unverified]($target)" "<img src=\"$target\">" "[unverified]: $target"; do
            # A closing parenthesis terminates a simple Markdown target.
            [[ "$target" == "$url)extra" && "$badge" == '!['* ]] && continue
            cp "$root/README.md" "$fixture/README.md"
            printf '\n%s\n' "$badge" >>"$fixture/README.md"
            expect 1 "reject incomplete or embedded approved badge target: $badge" bash "$root/scripts/docs-check.sh" "$fixture"
            if ! grep -Fq 'README advertises an unverified badge' "$scratch/result"; then
                printf 'FAIL: badge refusal did not reach its boundary: %s\n' "$badge" >&2
                failed=1
            fi
        done
    done
    cp "$root/README.md" "$fixture/README.md"
    printf '\n[Other](https://example.com)![Verified](%s)\n' "$url" >>"$fixture/README.md"
    expect 0 'allow adjacent independent Markdown targets' bash "$root/scripts/docs-check.sh" "$fixture"
    for target in "prefix $url" "https://evil.example/proxy/ $url" "$url suffix"; do
        cp "$root/README.md" "$fixture/README.md"
        printf '\n<img src="%s">\n' "$target" >>"$fixture/README.md"
        expect 1 'quoted HTML source is one complete target, including spaces' bash "$root/scripts/docs-check.sh" "$fixture"
    done
done
cp "$root/README.md" "$fixture/README.md"
mkdir -p "$scratch/failing-awk"
printf '#!/usr/bin/env bash\nexit 73\n' >"$scratch/failing-awk/awk"
chmod +x "$scratch/failing-awk/awk"
expect 1 'refuse documentation certification when badge scanner fails' env PATH="$scratch/failing-awk:$PATH" bash "$root/scripts/docs-check.sh" "$fixture"
for url in 'https://img&#46;shields&#46;io/badge/CI-passing-green' 'https://img.shie&#118;lds.io/badge/CI-passing-green' 'https://img&#x2e;shields&#x2e;io/badge/CI-passing-green' 'https://img&period;shields&period;io/badge/CI-passing-green' '&#104;ttps://img&#46;shields&#46;io/badge/CI-passing-green' 'https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badg&#101;.svg?branch=evil' 'https://img%2eshields%2eio/badge/CI-passing-green'; do
    for target in "<img src=\"$url\">" "<img src='$url'>" "<img src=$url>" "<a href=\"$url\">badge</a>" "![Unverified]($url)" "[unverified]: $url" "[unverified]: <$url>"; do
        cp "$root/README.md" "$fixture/README.md"
        printf '\n%s\n' "$target" >>"$fixture/README.md"
        expect 1 "reject encoded URL target: $target" bash "$root/scripts/docs-check.sh" "$fixture"
    done
done
for target in '![Unverified](https://img\.shields\.io/badge/CI-passing-green)' $'<img src="https://img.\nshields.io/badge/CI-passing-green">' $'<img src="https://img.\tshields.io/badge/CI-passing-green">' $'![Unverified](https://img.\nshields.io/badge/CI-passing-green)' '<img src="https://img&#46shields&#46io/badge/CI-passing-green">'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$target" >>"$fixture/README.md"
    expect 1 'reject escaped or split URL targets' bash "$root/scripts/docs-check.sh" "$fixture"
done
cp "$root/README.md" "$fixture/README.md"
printf '\nPlain prose &amp; alt text remain independent of literal targets.\n<img src="docs/assets/tusk-tui.png" alt="Tasks &amp; notes">\n[Other](https://example.com/?one=1&two=2)\n' >>"$fixture/README.md"
expect 0 'allow local images, ordinary query separators and entities outside URL targets' bash "$root/scripts/docs-check.sh" "$fixture"
for dot in '。' '．' '｡'; do
    for target in "<img src=\"https://img${dot}shields${dot}io/badge/CI-passing-green\">" "![Unverified](https://img${dot}shields${dot}io/badge/CI-passing-green)" "[unverified]: https://img${dot}shields${dot}io/badge/CI-passing-green"; do
        cp "$root/README.md" "$fixture/README.md"
        printf '\n%s\n' "$target" >>"$fixture/README.md"
        expect 1 'reject Unicode-equivalent remote hosts' bash "$root/scripts/docs-check.sh" "$fixture"
    done
done
encoded='&#104;&#116;&#116;&#112;&#115;&#58;&#47;&#47;&#105;&#109;&#103;&#46;&#115;&#104;&#105;&#101;&#108;&#100;&#115;&#46;&#105;&#111;&#47;&#98;&#97;&#100;&#103;&#101;&#47;&#67;&#73;&#45;&#112;&#97;&#115;&#115;&#105;&#110;&#103;&#45;&#103;&#114;&#101;&#101;&#110;'
for target in "<picture><source srcset=\"$encoded\"><img src=\"docs/assets/tusk-tui.png\"></picture>" "<img srcset='$encoded' src=\"docs/assets/tusk-tui.png\">" "<source srcset=$encoded>"; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$target" >>"$fixture/README.md"
    expect 1 'reject fully encoded responsive image targets' bash "$root/scripts/docs-check.sh" "$fixture"
done
for target in '<source srcset="docs/assets/tusk-tui.png 1x, docs/assets/tusk-tui.png 2x">' '<source srcset="docs/assets/tusk-tui.png,docs/assets/tusk-tui.png">'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$target" >>"$fixture/README.md"
    expect 1 'reject unsupported compound responsive targets' bash "$root/scripts/docs-check.sh" "$fixture"
done
for target in '[docs](https://example.com/a "Title")' '[docs](https://example.com/a '\''Title'\'')' 'Example https://example.com/search?q=a%20b' '`[docs](https://example.com/a%20b)`' '``Here `literal` [docs](https://example.com/a%20b)``' 'Text `[status](see CI)`' $'```text\n[docs](https://example.com/a%20b)\n```' $'~~~~text\n[docs](https://example.com/a%20b)\n~~~\n[docs](https://example.com/a%20b)\n~~~~'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$target" >>"$fixture/README.md"
    expect 0 'allow titled links and URL examples outside target constraints' bash "$root/scripts/docs-check.sh" "$fixture"
done
for url in 'https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main' 'https://img.shields.io/github/license/newbpydev/tusk'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n![Verified](%s "Title ü")\n<source srcset="%s">\n' "$url" "$url" >>"$fixture/README.md"
    expect 0 'allow verified titled badges and single responsive targets' bash "$root/scripts/docs-check.sh" "$fixture"
done
cp "$root/README.md" "$fixture/README.md"
cp "$fixture/docs/assets/tusk-tui.png" "$fixture/docs/assets/图.png"
printf '\n![Local](docs/assets/图.png)\n<source srcset="docs/assets/图.png">\n[Unicode path](https://example.com/路径)\n' >>"$fixture/README.md"
expect 0 'allow Unicode local filenames and remote paths with ASCII hosts' bash "$root/scripts/docs-check.sh" "$fixture"
for text in 'Text [status](see CI)' 'Text [status](<see CI)' '[status]: <see CI' '[^note]: https://example.com/search?q=a%20b'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$text" >>"$fixture/README.md"
    expect 0 'leave non-target text and footnotes outside URL constraints' bash "$root/scripts/docs-check.sh" "$fixture"
done
for text in '<img src="https://img.shields&shy.io/badge/CI-passing-green">' "![Unverified]("$'\n'"$encoded"$'\n)' "![Unverified][continued]"$'\n\n'"[continued]:"$'\n  '"$encoded"; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$text" >>"$fixture/README.md"
    expect 1 'reject entity authorities and continued encoded targets' bash "$root/scripts/docs-check.sh" "$fixture"
done
for url in 'https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main' 'https://img.shields.io/github/license/newbpydev/tusk'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n![Verified](\n%s\n)\n\n![Verified](%s\n)\n\n![Verified][continued]\n\n[continued]:\n  %s\n' "$url" "$url" "$url" >>"$fixture/README.md"
    expect 0 'allow verified URI values continued after their markers' bash "$root/scripts/docs-check.sh" "$fixture"
done
for text in "> ![Unverified][badge]"$'\n>\n> '"[badge]: <$encoded>" "> > ![Unverified][badge]"$'\n> >\n> > '"[badge]: $encoded"; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$text" >>"$fixture/README.md"
    expect 1 'reject encoded reference targets inside block quotes' bash "$root/scripts/docs-check.sh" "$fixture"
done
for text in $'```invalid`info\n![Unverified](https://img。shields.io/badge/CI-passing-green)' $'    ```text\n![Unverified](https://img。shields.io/badge/CI-passing-green)' $'\t```text\n![Unverified](https://img。shields.io/badge/CI-passing-green)' $'```text\nexample\n```\r\n![Unverified](https://img。shields.io/badge/CI-passing-green)' $'> ```text\n> example\n![Unverified](https://img。shields.io/badge/CI-passing-green)'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$text" >>"$fixture/README.md"
    expect 1 'fence boundaries cannot suppress following live targets' bash "$root/scripts/docs-check.sh" "$fixture"
done
for text in $'    ![Example](https://example.com/a%20b)' $'\t![Example](https://example.com/a%20b)' $'> ```text\n> ![Example](https://example.com/a%20b)\n> ```' $'   ```text\n![Example](https://example.com/a%20b)\n   ```'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$text" >>"$fixture/README.md"
    expect 0 'allow supported indented and quoted fenced code excerpts' bash "$root/scripts/docs-check.sh" "$fixture"
done
cp "$root/README.md" "$fixture/README.md"
printf '\nSee `[x](https://example.com/a%%20b\ntail)` for details.\n' >>"$fixture/README.md"
expect 1 'wrapped inline code spans retain the documented target tripwire' bash "$root/scripts/docs-check.sh" "$fixture"
for text in $'> ```text\n> example\n>\t```\n> ![Unverified](https://img。shields.io/badge/CI-passing-green)' $'- item\nlazy paragraph continuation\n\n    ![Unverified](https://img。shields.io/badge/CI-passing-green)' $'- item\n  - child\n\n    ![Unverified](https://img。shields.io/badge/CI-passing-green)' $'- item\n\n    ![Unverified](https://img。shields.io/badge/CI-passing-green)' $'1. item\n\n    ![Unverified](https://img。shields.io/badge/CI-passing-green)' $'> - item\n>\n>     ![Unverified](https://img。shields.io/badge/CI-passing-green)' $'- item\n  ```text\n![Unverified](https://img。shields.io/badge/CI-passing-green)' $'1. item\n   ~~~text\n![Unverified](https://img。shields.io/badge/CI-passing-green)' "- [badge]: <$encoded>"$'\n\n![Unverified][badge]' "> - [badge]: $encoded"$'\n>\n> ![Unverified][badge]'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$text" >>"$fixture/README.md"
    expect 1 'list containers retain literal target and reference checks' bash "$root/scripts/docs-check.sh" "$fixture"
done
cp "$root/README.md" "$fixture/README.md"
printf '\n- item\n\n        ![Example](https://example.com/a%%20b)\n' >>"$fixture/README.md"
expect 1 'list-contained code retains the documented target tripwire' bash "$root/scripts/docs-check.sh" "$fixture"
cp "$root/README.md" "$fixture/README.md"
printf '\n- item\n\nNormal paragraph.\n\n    ![Example](https://example.com/a%%20b)\n' >>"$fixture/README.md"
expect 0 'code exclusion resumes after an explicit list boundary' bash "$root/scripts/docs-check.sh" "$fixture"
for url in 'https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main' 'https://img.shields.io/github/license/newbpydev/tusk'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n- ![Verified](%s "Title")\n\n- [badge]: <%s>\n\n![Verified][badge]\n' "$url" "$url" >>"$fixture/README.md"
    expect 0 'verified list links and reference destinations remain allowed' bash "$root/scripts/docs-check.sh" "$fixture"
done
for text in '<img alt=">" src'$'\n '"=\"$encoded\">" "<img src"$'\n '"=\"$encoded\">" "<source srcset"$'\n '"=\"$encoded\">" "<a href"$'\n '"=\"$encoded\">badge</a>" '![Build: passing](https://evil.example/status.svg)' '<img src="https://evil.example/status.svg">' '<source srcset="https://evil.example/status.svg">' '[badge]: https://evil.example/status.svg'; do
    cp "$root/README.md" "$fixture/README.md"
    printf '\n%s\n' "$text" >>"$fixture/README.md"
    expect 1 'remote image sources and reference targets require approval' bash "$root/scripts/docs-check.sh" "$fixture"
done
cp "$root/README.md" "$fixture/README.md"
printf "\ntusk add sample --due='next week'\n" >>"$fixture/README.md"
expect 1 'reject single-quoted unsupported due-date claim' bash "$root/scripts/docs-check.sh" "$fixture"
cp "$root/README.md" "$fixture/README.md"
if [[ -f "$fixture/docs/install.md" ]]; then
    sed '/^Install Git,/s/\*\*Bash\*\*/a POSIX shell/' "$root/docs/install.md" >"$fixture/docs/install.md"
    expect 1 'source route declares Bash prerequisite' bash "$root/scripts/docs-check.sh" "$fixture"
    if ! grep -Fq 'lacks source Bash prerequisite' "$scratch/result"; then
        echo 'FAIL: missing Bash did not reach its prerequisite boundary' >&2
        cat "$scratch/result" >&2
        failed=1
    fi
fi
exit "$failed"
