#!/usr/bin/env bash
set -euo pipefail

# Packages with no executable statements to cover (pure interfaces, generated
# code) are exempt from the no-test-files failure. Space-separated substring
# patterns, one per package path. Everything else without tests still fails.
COVERAGE_EXEMPT="${COVERAGE_EXEMPT:-internal/ports generated sqlc}"

if ! output=$(go test -cover -race ./... 2>&1); then
    echo "$output"
    exit 1
fi
echo "$output"

export COVERAGE_EXEMPT
echo "$output" | awk '
BEGIN {
    n = split(ENVIRON["COVERAGE_EXEMPT"], exempt, " ")
}
/\[no test files\]/ {
    exempted = 0
    for (i = 1; i <= n; i++) {
        if (exempt[i] != "" && index($0, exempt[i]) > 0) { exempted = 1 }
    }
    if (!exempted) {
        print "Error: package has no test files: " $0
        failed = 1
    }
}
/coverage: [0-9.]+%/ {
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
