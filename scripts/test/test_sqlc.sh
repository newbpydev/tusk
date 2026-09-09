#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
source "$ROOT_DIR/scripts/sqlc.sh"
test_dir=$(mktemp -d)
trap 'rm -rf "$test_dir"' EXIT
mkdir -p "$test_dir/project/db/migrations" "$test_dir/project/internal/storage/sqlc" "$test_dir/bin"
printf 'version: "2"\n' > "$test_dir/project/sqlc.yaml"
printf 'package sqlc\n' > "$test_dir/project/db/queries.sql"
printf 'schema\n' > "$test_dir/project/db/migrations/001.sql"
cat > "$test_dir/bin/sqlc" <<'EOF'
#!/usr/bin/env bash
set -eu
if [[ "$1" == version ]]; then echo "${FAKE_SQLC_VERSION:-v1.31.1}"; exit; fi
mkdir -p internal/storage/sqlc
cp db/queries.sql internal/storage/sqlc/queries.go
if [[ "${FAKE_SQLC_FAIL:-0}" == 1 ]]; then exit 1; fi
EOF
chmod +x "$test_dir/bin/sqlc"
export TUSK_SQLC_BIN="$test_dir/bin/sqlc"
expect_failure() { if "$@" >"$test_dir/failure.log" 2>&1; then echo "Expected failure: $*" >&2; exit 1; fi; }
[[ -x "$ROOT_DIR/scripts/sqlc.sh" && -x "$ROOT_DIR/scripts/test/test_sqlc.sh" ]] || { echo 'sqlc entrypoints must be executable' >&2; exit 1; }
"$ROOT_DIR/scripts/sqlc.sh" generate "$test_dir/project"
bash "$ROOT_DIR/scripts/sqlc.sh" check "$test_dir/project"
printf 'package stale\n' > "$test_dir/project/internal/storage/sqlc/queries.go"
expect_failure bash "$ROOT_DIR/scripts/sqlc.sh" check "$test_dir/project"
expect_failure env FAKE_SQLC_FAIL=1 bash "$ROOT_DIR/scripts/sqlc.sh" generate "$test_dir/project"
[[ "$(cat "$test_dir/project/internal/storage/sqlc/queries.go")" == 'package stale' ]]
bash "$ROOT_DIR/scripts/sqlc.sh" generate "$test_dir/project"
touch "$test_dir/project/internal/storage/sqlc/extra.go"
expect_failure bash "$ROOT_DIR/scripts/sqlc.sh" check "$test_dir/project"
rm "$test_dir/project/internal/storage/sqlc/extra.go" "$test_dir/project/internal/storage/sqlc/queries.go"
expect_failure bash "$ROOT_DIR/scripts/sqlc.sh" check "$test_dir/project"
expect_failure env FAKE_SQLC_VERSION=v0.0.0 bash "$ROOT_DIR/scripts/sqlc.sh" generate "$test_dir/project"
expect_failure env TUSK_SQLC_HOST=unsupported bash "$ROOT_DIR/scripts/sqlc.sh" setup "$test_dir/project"
printf 'not an archive' > "$test_dir/bad.tar.gz"
expect_failure verify_sqlc_archive "$test_dir/bad.tar.gz" 0000000000000000000000000000000000000000000000000000000000000000 "$test_dir/extracted"
expect_failure verify_sqlc_archive "$test_dir/bad.tar.gz" "$(sqlc_sha256 "$test_dir/bad.tar.gz")" "$test_dir/extracted"
mkdir -p "$test_dir/archive/nested"
printf 'binary fixture' > "$test_dir/archive/sqlc"
tar -czf "$test_dir/good.tar.gz" -C "$test_dir/archive" sqlc
verify_sqlc_archive "$test_dir/good.tar.gz" "$(sqlc_sha256 "$test_dir/good.tar.gz")" "$test_dir/extracted"
cmp "$test_dir/archive/sqlc" "$test_dir/extracted"
tar -czPf "$test_dir/traversal.tar.gz" -C "$test_dir/archive" nested/../sqlc
expect_failure verify_sqlc_archive "$test_dir/traversal.tar.gz" "$(sqlc_sha256 "$test_dir/traversal.tar.gz")" "$test_dir/extracted"
touch "$test_dir/archive/extra"
tar -czf "$test_dir/extra.tar.gz" -C "$test_dir/archive" sqlc extra
expect_failure verify_sqlc_archive "$test_dir/extra.tar.gz" "$(sqlc_sha256 "$test_dir/extra.tar.gz")" "$test_dir/extracted"
rm "$test_dir/archive/sqlc"
ln -s extra "$test_dir/archive/sqlc"
tar -czf "$test_dir/symlink.tar.gz" -C "$test_dir/archive" sqlc
expect_failure verify_sqlc_archive "$test_dir/symlink.tar.gz" "$(sqlc_sha256 "$test_dir/symlink.tar.gz")" "$test_dir/extracted"
before=$(sqlc_sha256 "$TUSK_SQLC_BIN")
printf '#!/usr/bin/env bash\nexit 1\n' > "$test_dir/bin/curl"
chmod +x "$test_dir/bin/curl"
expect_failure env PATH="$test_dir/bin:$PATH" TUSK_SQLC_HOST=Linux/x86_64 bash "$ROOT_DIR/scripts/sqlc.sh" setup "$test_dir/project"
[[ "$(sqlc_sha256 "$TUSK_SQLC_BIN")" == "$before" ]]
echo 'sqlc script fixtures passed (source archive, drift, partial failure, version, host, digest, traversal, symlink, download failure)'

for scenario in setup:curl generate:gofmt check:diff; do
    action=${scenario%:*}
    missing=${scenario#*:}
    restricted="$test_dir/without-$missing"
    mkdir -p "$restricted"
    for dependency in bash dirname mktemp uname rm curl gofmt diff; do
        if [[ "$dependency" != "$missing" ]]; then
            ln -s "$(command -v "$dependency")" "$restricted/$dependency"
        fi
    done
    expect_failure env PATH="$restricted" TUSK_SQLC_HOST=Linux/x86_64 bash "$ROOT_DIR/scripts/sqlc.sh" "$action" "$test_dir/project"
    if ! grep -F "sqlc $action requires $missing on PATH" "$test_dir/failure.log" >/dev/null; then
        cat "$test_dir/failure.log" >&2
        echo "Missing prerequisite was not explained: $missing" >&2
        exit 1
    fi
done
echo 'sqlc direct execution and missing prerequisite fixtures passed'
