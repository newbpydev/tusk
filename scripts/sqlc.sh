#!/usr/bin/env bash
set -euo pipefail

sqlc_sha256() {
    if command -v sha256sum >/dev/null; then sha256sum "$1" | cut -d ' ' -f 1
    else shasum -a 256 "$1" | cut -d ' ' -f 1; fi
}

verify_sqlc_archive() {
    local archive="$1" digest="$2" output="$3" actual members details
    actual=$(sqlc_sha256 "$archive") || return 1
    [[ "$actual" == "$digest" ]] || { echo 'sqlc archive digest mismatch' >&2; return 1; }
    members=$(tar -tzf "$archive") || return 1
    [[ "$members" == sqlc || "$members" == sqlc.exe ]] || { echo 'Unexpected sqlc archive members' >&2; return 1; }
    details=$(tar -tvzf "$archive") || return 1
    [[ "${details:0:1}" == '-' ]] || { echo 'sqlc archive member is not a regular file' >&2; return 1; }
    tar -xOzf "$archive" "$members" > "$output" || return 1
    [[ -s "$output" ]] || return 1
}

sqlc_main() (
    set -euo pipefail
    action="${1:?expected setup, generate or check}"
    case "$action" in
        setup) required=(curl) ;;
        generate) required=(gofmt) ;;
        check) required=(gofmt diff) ;;
        *) echo "Unknown sqlc action: $action" >&2; exit 2 ;;
    esac
    for dependency in "${required[@]}"; do
        command -v "$dependency" >/dev/null || { echo "sqlc $action requires $dependency on PATH" >&2; exit 1; }
    done
    root="${2:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
    root=$(cd "$root" && pwd)
    scratch=$(mktemp -d)
    stage=''
    backup=''
    install_stage=''
    cleanup() {
        rm -rf "$scratch"
        if [[ -n "$stage" ]]; then rm -rf "$stage"; fi
        if [[ -n "$install_stage" ]]; then rm -f "$install_stage"; fi
        if [[ -n "$backup" && -e "$backup" && ! -e "$root/internal/storage/sqlc" ]]; then
            mv "$backup" "$root/internal/storage/sqlc"
        fi
    }
    trap cleanup EXIT
    tool="${TUSK_SQLC_BIN:-$root/bin/tools/sqlc}"
    case "$(uname -s)" in MINGW*|MSYS*) tool="${TUSK_SQLC_BIN:-$root/bin/tools/sqlc.exe}" ;; esac
    if [[ "$action" == setup ]]; then
        host="${TUSK_SQLC_HOST:-$(uname -s)/$(uname -m)}"
        case "$host" in
            Linux/x86_64) target=linux_amd64; digest=497ae4fcdfa64c5b0c311ffe4c2bd991e43991e82e5367792ed78bc2dca27354 ;;
            Linux/aarch64|Linux/arm64) target=linux_arm64; digest=b7cae247740d0c51a1e657479e5b2d21e6fef428f596682a01bc55bf4ab8a23d ;;
            Darwin/x86_64) target=darwin_amd64; digest=c5af76772e3785d21663a62697056b383f07629979b1bd25b93872e73dbd519b ;;
            Darwin/arm64) target=darwin_arm64; digest=21602158c99eb1f2bae197a66abfb1941d1e9e50b23125bb193349c6b1acc71e ;;
            MINGW*/x86_64|MSYS*/x86_64) target=windows_amd64; digest=40d138ec18b1cc80d2be7305917fd4deceda4e0c32d78ba5d8faa4bfa3bc0fc0; tool="${TUSK_SQLC_BIN:-$root/bin/tools/sqlc.exe}" ;;
            *) echo "Unsupported sqlc host: $host" >&2; exit 1 ;;
        esac
        curl --fail --location --proto '=https' --tlsv1.2 --output "$scratch/sqlc.tar.gz" "https://github.com/sqlc-dev/sqlc/releases/download/v1.31.1/sqlc_1.31.1_${target}.tar.gz"
        verify_sqlc_archive "$scratch/sqlc.tar.gz" "$digest" "$scratch/sqlc"
        chmod +x "$scratch/sqlc"
        [[ "$("$scratch/sqlc" version)" == v1.31.1 ]] || { echo 'Incorrect sqlc version' >&2; exit 1; }
        mkdir -p "$(dirname "$tool")"
        install_stage=$(mktemp "$(dirname "$tool")/.sqlc.XXXXXX")
        cp "$scratch/sqlc" "$install_stage"
        chmod +x "$install_stage"
        mv -f "$install_stage" "$tool"
        exit
    fi
    [[ -x "$tool" ]] || { echo 'Run make setup-sqlc to install sqlc v1.31.1' >&2; exit 1; }
    [[ "$("$tool" version)" == v1.31.1 ]] || { echo 'Incorrect sqlc version; expected v1.31.1' >&2; exit 1; }
    mkdir -p "$scratch/db"
    cp "$root/sqlc.yaml" "$scratch/sqlc.yaml"
    cp "$root/db/queries.sql" "$scratch/db/queries.sql"
    cp -R "$root/db/migrations" "$scratch/db/migrations"
    (cd "$scratch" && "$tool" generate)
    [[ -d "$scratch/internal/storage/sqlc" ]] || { echo 'Generator produced no output' >&2; exit 1; }
    gofmt -s -w "$scratch/internal/storage/sqlc"/*.go
    if [[ "$action" == check ]]; then
        diff -ru "$scratch/internal/storage/sqlc" "$root/internal/storage/sqlc"
        exit
    fi
    [[ ! -L "$root/internal/storage/sqlc" ]] || { echo 'Generated directory must not be a symlink' >&2; exit 1; }
    mkdir -p "$root/internal/storage"
    stage=$(mktemp -d "$root/internal/storage/.sqlc.XXXXXX")
    cp -R "$scratch/internal/storage/sqlc/." "$stage/"
    if [[ -e "$root/internal/storage/sqlc" ]]; then
        backup="$stage.previous"
        mv "$root/internal/storage/sqlc" "$backup"
    fi
    mv "$stage" "$root/internal/storage/sqlc"
    stage=''
    if [[ -n "$backup" ]]; then rm -rf "$backup"; backup=''; fi
)

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then sqlc_main "$@"; fi
