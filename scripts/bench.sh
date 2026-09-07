#!/usr/bin/env bash
set -euo pipefail

TARGET="${1:-all}"

run_tree() {
    echo "==> Running tree traversal benchmark..."
    if ! output=$(go test -bench=^BenchmarkTreeTraversal$ -run=^$ -benchmem ./internal/core/... 2>&1); then
        echo "$output"
        exit 1
    fi
    echo "$output"

    echo "$output" | awk '
    /^BenchmarkTreeTraversal/ {
        seen++
        ns = $3 + 0.0
        allocs = $7 + 0
        if (ns >= 500.0) {
            printf "Error: BenchmarkTreeTraversal %.1fns/op exceeds budget of 500ns\n", ns
            failed = 1
        }
        if (allocs > 0) {
            printf "Error: BenchmarkTreeTraversal %d allocs/op exceeds budget of 0 allocs\n", allocs
            failed = 1
        }
    }
    END {
        if (!seen) {
            print "Error: no BenchmarkTreeTraversal results parsed"
            exit 1
        }
        if (failed) exit 1
    }'
}

run_build() {
    echo "==> Running tree build benchmark..."
    if ! output=$(go test -bench=^BenchmarkBuildTree$ -run=^$ -benchmem ./internal/core/... 2>&1); then
        echo "$output"
        exit 1
    fi
    echo "$output"

    echo "$output" | awk '
    /^BenchmarkBuildTree/ {
        seen++
        ns = $3 + 0.0
        allocs = $7 + 0
        # 100 tasks per op: budget < 1us/task (< 100,000ns) and < 5 allocs/task (< 500 allocs)
        if (ns >= 100000.0) {
            printf "Error: BenchmarkBuildTree %.1fns/op exceeds budget of 100,000ns (<1us/task)\n", ns
            failed = 1
        }
        if (allocs >= 500) {
            printf "Error: BenchmarkBuildTree %d allocs/op exceeds budget of 500 allocs (<5 allocs/task)\n", allocs
            failed = 1
        }
    }
    END {
        if (!seen) {
            print "Error: no BenchmarkBuildTree results parsed"
            exit 1
        }
        if (failed) exit 1
    }'
}

case "$TARGET" in
    tree)
        run_tree
        ;;
    build)
        run_build
        ;;
    all)
        run_tree
        run_build
        ;;
    *)
        echo "Unknown benchmark target: $TARGET" >&2
        exit 1
        ;;
esac
