#!/usr/bin/env bash
set -euo pipefail
jq_binary_option=''
case "${OSTYPE:-}" in msys*|cygwin*) jq_binary_option=--binary ;; esac
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
# shellcheck source=scripts/json_check.sh
source "$root/scripts/json_check.sh"
repo=newbpydev/tusk
export GH_HOST=github.com
unset GH_DEBUG DEBUG
mode=${1:-prepare}
die() { printf 'Repository metadata: %s\n' "$*" >&2;exit 1; }
digest() { if command -v sha256sum >/dev/null 2>&1;then sha256sum "$1"|awk '{print $1}';else shasum -a 256 "$1"|awk '{print $1}';fi; }
regular() { [[ -f "$1" && ! -L "$1" ]]; }
[[ "$mode" == prepare || "$mode" == apply ]] || die 'expected prepare or apply'
payload="$root/.github/repository-metadata.json"
json_object "$payload" || die 'single metadata payload object required'
# Reject unexpected writable fields, endpoints, private settings or script content.
jq ${jq_binary_option:+"--binary"} -e 'keys==(["repository","about","topics","social_preview","manual_readback"]|sort) and .repository=="newbpydev/tusk" and (.about|keys)==["description","homepage"] and .about.description=="Local task management in your terminal: a keyboard-driven TUI and scriptable CLI, backed by SQLite." and .about.homepage=="https://github.com/newbpydev/tusk#readme" and .topics=={names:["go","golang","cli","tui","task-manager","terminal","sqlite","bubbletea","productivity","offline","command-line"]} and .social_preview=="docs/assets/tusk-tui.png"' "$payload" >/dev/null || die 'metadata payload outside reviewed field contract'
[[ -n "${METADATA_OUTPUT:-}" ]] || die 'new retained METADATA_OUTPUT required'
parent=$(cd "$(dirname "$METADATA_OUTPUT")" && pwd -P) || die 'output parent missing'
output="$parent/$(basename "$METADATA_OUTPUT")"
case "$output" in "$root"|"$root"/*) die 'output must be outside source checkout';; esac
mkdir "$output" || die 'output exists; use fresh readback directory'
trap 'printf "{\"mode\":\"%s\",\"exit_status\":%s}\n" "$mode" "$?" >"$output/result.json"' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
jq ${jq_binary_option:+"--binary"} '.about' "$payload" >"$output/about-request.json"
jq ${jq_binary_option:+"--binary"} '.topics' "$payload" >"$output/topics-request.json"
payload_hash=$(digest "$payload")
if [[ "$mode" == prepare ]];then
 regular "$root/docs/assets/tusk-tui.png" || die 'reviewed preview missing'
 jq ${jq_binary_option:+"--binary"} -n --arg hash "$payload_hash" --arg image "$(digest "$root/docs/assets/tusk-tui.png")" '{schema:1,status:"prepared-unapplied",repository:"newbpydev/tusk",payload_sha256:$hash,preview_sha256:$image,social_preview:"manual owner upload/readback",private_security_settings:"not changed",public_release_install_and_readme:"pending actual evidence"}' >"$output/plan.json"
 printf 'Prepared exact About/topics requests and preview identity; no remote writes.\n';exit 0
fi
if ! json_object "${METADATA_AUTHORIZATION:-}" || ! json_object "${METADATA_RELEASE_RECEIPT:-}";then die 'explicit owner authorization and public release readback required';fi
documentation_sha=$(git -C "$root" rev-parse HEAD)
[[ "$documentation_sha" =~ ^[0-9a-f]{40}$ && -z "$(git -C "$root" status --porcelain)" ]] || die 'clean approved documentation SHA required'
jq ${jq_binary_option:+"--binary"} -e --arg sha "$documentation_sha" --arg payload "$payload_hash" --arg release "$(digest "$METADATA_RELEASE_RECEIPT")" '
 keys==(["schema","status","action","repository","documentation_sha","payload_sha256","release_receipt_sha256","approval_reference"]|sort) and .schema==1 and .status=="owner-approved" and .action=="metadata" and .repository=="newbpydev/tusk" and .documentation_sha==$sha and .payload_sha256==$payload and .release_receipt_sha256==$release and (.approval_reference|type=="string" and length>0)
 ' "$METADATA_AUTHORIZATION" >/dev/null || die 'approval does not match exact payload/documentation/release readback'
jq ${jq_binary_option:+"--binary"} -e '.schema==1 and .status=="published" and .repository=="newbpydev/tusk" and .assets_verified==9 and (.release_id|type=="number" and .>0 and floor==.) and (.source_sha|test("^[0-9a-f]{40}$")) and (.version|test("^v(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)$"))' "$METADATA_RELEASE_RECEIPT" >/dev/null || die 'verified public release required'
id=$(jq ${jq_binary_option:+"--binary"} -r '.release_id' "$METADATA_RELEASE_RECEIPT")
# shellcheck source=scripts/gh_deadline.sh
source "$root/scripts/gh_deadline.sh"
gh_deadline_setup "$root" "$output" || die 'could not prepare bounded GitHub driver'
gh api "https://api.github.com/repos/$repo/releases/$id" >"$output/release.json"
jq ${jq_binary_option:+"--binary"} -e --slurpfile r "$METADATA_RELEASE_RECEIPT" '.id==$r[0].release_id and .draft==false and .tag_name==$r[0].version and .target_commitish==$r[0].source_sha' "$output/release.json" >/dev/null || die 'public release identity changed'
gh api "https://api.github.com/repos/$repo/git/ref/heads/main" >"$output/main.json"
jq ${jq_binary_option:+"--binary"} -e --arg sha "$documentation_sha" '.object.type=="commit" and .object.sha==$sha' "$output/main.json" >/dev/null || die 'approved documentation is not current main'
gh api "https://api.github.com/repos/$repo" >"$output/repository-before.json"
gh api "https://api.github.com/repos/$repo/topics" >"$output/topics-before.json"
if ! jq ${jq_binary_option:+"--binary"} -e --slurpfile p "$payload" '.full_name=="newbpydev/tusk" and .description==$p[0].about.description and .homepage==$p[0].about.homepage' "$output/repository-before.json" >/dev/null;then
 gh api "https://api.github.com/repos/$repo" --method PATCH --input "$output/about-request.json" >"$output/about-response.json" 2>"$output/about-response.stderr" || true
fi
# Lost mutation responses are reconciled before the next independent operation.
gh api "https://api.github.com/repos/$repo" >"$output/repository-after.json"
jq ${jq_binary_option:+"--binary"} -e --slurpfile p "$payload" '.full_name=="newbpydev/tusk" and .description==$p[0].about.description and .homepage==$p[0].about.homepage' "$output/repository-after.json" >/dev/null || die 'About update unconfirmed; read back in a fresh invocation'
if ! jq ${jq_binary_option:+"--binary"} -e --slurpfile p "$payload" '(.names|sort)==($p[0].topics.names|sort)' "$output/topics-before.json" >/dev/null;then
 gh api "https://api.github.com/repos/$repo/topics" --method PUT --input "$output/topics-request.json" >"$output/topics-response.json" 2>"$output/topics-response.stderr" || true
fi
gh api "https://api.github.com/repos/$repo/topics" >"$output/topics-after.json"
jq ${jq_binary_option:+"--binary"} -e --slurpfile p "$payload" '(.names|sort)==($p[0].topics.names|sort)' "$output/topics-after.json" >/dev/null || die 'topic update unconfirmed; retry only missing state after readback'
jq ${jq_binary_option:+"--binary"} -n --arg sha "$documentation_sha" --arg hash "$payload_hash" '{schema:1,status:"about-topics-verified",repository:"newbpydev/tusk",documentation_sha:$sha,payload_sha256:$hash,manual_social_preview_readme_license_security_install:"pending separate observed evidence"}' >"$output/receipt.json"
printf 'About/topics match approved values; manual landing-page gates remain pending.\n'
