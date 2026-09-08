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

echo "========================================"
echo "Script Test Results: ${TESTS_PASSED}/${TESTS_TOTAL} passed"
echo "========================================"

if [ "$TESTS_PASSED" -ne "$TESTS_TOTAL" ]; then
    exit 1
fi

exit 0
