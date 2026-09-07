#!/usr/bin/env bash
set -euo pipefail

if ! output=$(go test -cover -race ./... 2>&1); then
    echo "$output"
    exit 1
fi
echo "$output"

echo "$output" | awk '
/\[no test files\]/ {
    print "Error: package has no test files: " $0
    failed = 1
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
