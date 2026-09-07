#!/usr/bin/env bash
set -euo pipefail

output=$(go test -cover -race ./...)
echo "$output"

echo "$output" | awk '
/coverage: [0-9.]+%/ {
    match($0, /coverage: ([0-9.]+)%/, m)
    cov = m[1] + 0.0
    if (cov < 95.0) {
        printf "Error: package coverage %.1f%% is below mandated 95.0%% threshold\n", cov > "/dev/stderr"
        failed = 1
    }
}
END {
    if (failed) exit 1
}
'
