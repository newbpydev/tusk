#!/usr/bin/env bash
set -euo pipefail

# Packages with no executable statements of their own to cover (pure
# interfaces, generated code) are exempt from both the no-test-files failure
# and the coverage threshold. Space-separated path patterns matched against
# the package path field only (exact match or trailing path segments), so
# similarly-named testable packages are never swallowed. Everything else
# without tests still fails. Override via COVERAGE_EXEMPT="pat1 pat2".
COVERAGE_EXEMPT="${COVERAGE_EXEMPT:-internal/ports generated sqlc}"

if ! output=$(go test -cover -race ./... 2>&1); then
    echo "$output"
    exit 1
fi
echo "$output"

export COVERAGE_EXEMPT
echo "$output" | awk '
function is_exempt(pkg) {
    n = split(ENVIRON["COVERAGE_EXEMPT"], exempt, " ")
    for (i = 1; i <= n; i++) {
        e = exempt[i]
        if (e == "") { continue }
        if (pkg == e || match(pkg, "(^|/)" e "$")) { return 1 }
    }
    return 0
}
/\[no test files\]/ {
    if (!is_exempt($2)) {
        print "Error: package has no test files: " $0
        failed = 1
    }
}
/coverage: [0-9.]+%/ {
    if (is_exempt($2)) { next }
    for (i = 1; i <= NF; i++) {
        if ($i ~ /^[0-9.]+%$/) {
            cov = substr($i, 1, length($i)-1) + 0.0
            if (cov < 95.0) {
                print "Error: package coverage " cov "% is below mandated 95.0% threshold"
                failed = 1
            }
        }
    }
}
END {
    if (failed) exit 1
}
'
