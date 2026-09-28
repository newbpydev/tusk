#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

TESTS_PASSED=0
TESTS_TOTAL=0

assert_eq() {
    local expected="$1"
    local actual="$2"
    local message="$3"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    if [ "$expected" != "$actual" ]; then
        echo "FAIL: ${message} (expected '${expected}', got '${actual}')" >&2
        return 1
    fi
    echo "PASS: ${message}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
}

echo "========================================"
echo "Running Tusk Script Test Suite (TDD)"
echo "========================================"

# Test 1: setup.sh succeeds in normal environment
set +e
"${ROOT_DIR}/scripts/setup.sh" >/dev/null 2>&1
EXIT_CODE=$?
set -e
assert_eq 0 "$EXIT_CODE" "setup.sh exits 0 when Go is present"

# Test 2: setup.sh fails when Go binary is missing
set +e
TUSK_GO_BIN="nonexistent_go_binary_xyz" "${ROOT_DIR}/scripts/setup.sh" >/dev/null 2>&1
EXIT_CODE=$?
set -e
assert_eq 1 "$EXIT_CODE" "setup.sh exits 1 when Go is missing"

# Test 3: fmt.sh formats poorly formatted Go files
TMP_DIR=$(mktemp -d)
trap 'rm -rf "${TMP_DIR}"' EXIT

# TestSetupGoMinimum: setup must reject unsupported and malformed toolchains.
for version in go1.24.9 go1.25.0 go1.27.1 malformed; do
    fake_go="${TMP_DIR}/go"
    printf '#!/usr/bin/env bash\nprintf "%%s\\n" "go version %s linux/amd64"\n' "$version" > "$fake_go"
    chmod +x "$fake_go"
    expected=0
    case "$version" in go1.24.9|malformed) expected=1 ;; esac
    result=0
    TUSK_GO_BIN="$fake_go" "${ROOT_DIR}/scripts/setup.sh" >"${TMP_DIR}/setup-output" 2>&1 || result=$?
    assert_eq "$expected" "$result" "TestSetupGoMinimum: $version"
done

TEST_GO_FILE="${TMP_DIR}/bad_format.go"
cat << 'EOF' > "${TEST_GO_FILE}"
package main
func main(){
var x int=10
_ = x
}
EOF

# Verify it was poorly formatted initially (gofmt -d produces diff, diff exits 1 on diff)
set +e
INIT_DIFF=$(gofmt -d "${TEST_GO_FILE}" 2>/dev/null || true)
set -e
HAS_DIFF=0
if [ -n "${INIT_DIFF}" ]; then
    HAS_DIFF=1
fi
assert_eq 1 "$HAS_DIFF" "Test fixture initially contains unformatted Go code"

# Run gofmt -s -w directly on fixture to verify fmt normalization
gofmt -s -w "${TEST_GO_FILE}"
set +e
POST_DIFF=$(gofmt -d "${TEST_GO_FILE}" 2>/dev/null || true)
set -e
assert_eq "" "$POST_DIFF" "gofmt -s -w successfully normalizes Go formatting"

# Test 4: fmt.sh executes cleanly against repository root
set +e
"${ROOT_DIR}/scripts/fmt.sh" >/dev/null 2>&1
EXIT_CODE=$?
set -e
assert_eq 0 "$EXIT_CODE" "fmt.sh exits 0 against repo root"

# Go prints a package without the 'ok' prefix when it has no test files.
cat > "${TMP_DIR}/go" <<'EOF'
#!/usr/bin/env bash
printf '\t%s\t\tcoverage: 0.0%% of statements\n' "$TUSK_COVERAGE_FIXTURE"
EOF
chmod +x "${TMP_DIR}/go"
for package in github.com/newbpydev/tusk/internal/storage/sqlc github.com/newbpydev/tusk/internal/storage/not_sqlc; do
    expected=1
    if [[ "$package" == */sqlc ]]; then expected=0; fi
    result=0
    PATH="${TMP_DIR}:$PATH" TUSK_COVERAGE_FIXTURE="$package" "${ROOT_DIR}/scripts/coverage.sh" >"${TMP_DIR}/coverage-output" 2>&1 || result=$?
    assert_eq "$expected" "$result" "coverage package parsing: $package"
done

# Plain make must keep the canonical all target, never download sqlc by default.
default_recipe=$(make --no-print-directory -n -C "${ROOT_DIR}")
result=1
if [[ "$default_recipe" == *"All canonical quality gates passed."* && "$default_recipe" != *"scripts/sqlc.sh setup"* ]]; then result=0; fi
assert_eq 0 "$result" "default make validates and builds without downloading tools"

echo "========================================"
for target in build-service bench-service; do
    result=0
    make --no-print-directory -n -C "${ROOT_DIR}" "$target" >"${TMP_DIR}/service-target" 2>&1 || result=$?
    assert_eq 0 "$result" "service target exists: $target"
    recipe=$(cat "${TMP_DIR}/service-target")
    result=1
    case "$target" in
        build-service) if [[ "$recipe" == *"CGO_ENABLED=0"* && "$recipe" == *"windows/amd64"* && "$recipe" == *"go test -c"* ]]; then result=0; fi ;;
        bench-service) if [[ "$recipe" == *"BenchmarkService"* && "$recipe" == *"-benchmem"* ]]; then result=0; fi ;;
    esac
    assert_eq 0 "$result" "service target contract: $target"
done

# A copied Makefile must build the complete package, including sibling files.
mkdir -p "${TMP_DIR}/build-fixture/cmd/tusk"
cp "${ROOT_DIR}/Makefile" "${TMP_DIR}/build-fixture/Makefile"
printf 'module fixture\n\ngo 1.25.0\n' > "${TMP_DIR}/build-fixture/go.mod"
printf 'package main\nfunc main() { sibling() }\n' > "${TMP_DIR}/build-fixture/cmd/tusk/main.go"
printf 'package main\nfunc sibling() {}\n' > "${TMP_DIR}/build-fixture/cmd/tusk/app.go"
result=0
make --no-print-directory -C "${TMP_DIR}/build-fixture" build GOFLAGS=-buildvcs=false BUILD_OUTPUT="${TMP_DIR}/fixture-tusk" >"${TMP_DIR}/build-output" 2>&1 || result=$?
assert_eq 0 "$result" "build includes siblings and accepts isolated output"
test -f "${TMP_DIR}/fixture-tusk"

# Verify the CLI gates select the right packages and propagate tool failures.
for target in test-cli build-cli; do
    recipe=$(make --no-print-directory -n -C "${ROOT_DIR}" "$target")
    result=1
    if [[ "$recipe" == *"./internal/cli"* && "$recipe" == *"./cmd/tusk"* ]]; then result=0; fi
    assert_eq 0 "$result" "CLI target includes adapter and executable: $target"
    printf '#!/usr/bin/env bash\nexit 19\n' > "${TMP_DIR}/go"
    chmod +x "${TMP_DIR}/go"
    result=0
    PATH="${TMP_DIR}:$PATH" make --no-print-directory -C "${ROOT_DIR}" "$target" >"${TMP_DIR}/negative-cli" 2>&1 || result=$?
    assert_eq 2 "$result" "CLI target propagates failed go: $target"
done

for target in bench-cli bench-cli-conditions profile-cli test-cli-latency-codec generate-schema-catalog check-schema-catalog; do
    recipe=$(make --no-print-directory -n -C "${ROOT_DIR}" "$target")
    result=1
    if [[ "$recipe" == *"go "* ]]; then result=0; fi
    assert_eq 0 "$result" "CLI measurement target exists: $target"
    printf '#!/usr/bin/env bash\nexit 19\n' > "${TMP_DIR}/go"
    chmod +x "${TMP_DIR}/go"
    result=0
    PATH="${TMP_DIR}:$PATH" make --no-print-directory -C "${ROOT_DIR}" "$target" >"${TMP_DIR}/negative-bench" 2>&1 || result=$?
    assert_eq 2 "$result" "CLI measurement target propagates failed go: $target"
done

# Catalog generation compiles storage, which may refer to newly added sqlc
# methods. Regenerate those methods before attempting the catalog compiler.
mkdir -p "${TMP_DIR}/generation-fixture/bin"
cp "${ROOT_DIR}/Makefile" "${TMP_DIR}/generation-fixture/Makefile"
cat > "${TMP_DIR}/generation-fixture/bin/bash" <<'EOF'
#!/bin/sh
test "$1" = scripts/sqlc.sh && test "$2" = generate || exit 23
touch sqlc-generated.marker
EOF
cat > "${TMP_DIR}/generation-fixture/bin/go" <<'EOF'
#!/bin/sh
test -f sqlc-generated.marker || exit 19
EOF
chmod +x "${TMP_DIR}/generation-fixture/bin/"*
result=0
PATH="${TMP_DIR}/generation-fixture/bin:$PATH" make --no-print-directory -C "${TMP_DIR}/generation-fixture" generate >"${TMP_DIR}/generation-output" 2>&1 || result=$?
assert_eq 0 "$result" "generate bootstraps sqlc before compiling schema catalog"

echo "Script Test Results: ${TESTS_PASSED}/${TESTS_TOTAL} passed"
echo "========================================"

if [ "$TESTS_PASSED" -ne "$TESTS_TOTAL" ]; then
    exit 1
fi

exit 0
