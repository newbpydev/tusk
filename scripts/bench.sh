#!/usr/bin/env bash
set -euo pipefail

TARGET="${1:-all}"

run_tree() {
    echo "==> Running tree traversal benchmark..."
    if ! output=$(go test -bench=^BenchmarkTreeTraversal$ -run=^$ -benchmem -count=3 ./internal/core/... 2>&1); then
        echo "$output"
        exit 1
    fi
    echo "$output"

    echo "$output" | awk '
    /^BenchmarkTreeTraversal/ {
        seen++
        sum_ns += $3
        if (($7 + 0) > max_allocs) { max_allocs = $7 + 0 }
    }
    END {
        if (!seen) {
            print "Error: no BenchmarkTreeTraversal results parsed"
            exit 1
        }
        mean_ns = sum_ns / seen
        printf "Samples: %d, mean %.1fns/op, max %d allocs/op\n", seen, mean_ns, max_allocs
        if (mean_ns >= 500.0) {
            printf "Error: BenchmarkTreeTraversal mean %.1fns/op over %d samples exceeds budget of 500ns\n", mean_ns, seen
            failed = 1
        }
        if (max_allocs > 0) {
            printf "Error: BenchmarkTreeTraversal %d allocs/op exceeds budget of 0 allocs\n", max_allocs
            failed = 1
        }
        if (failed) exit 1
    }'
}

run_build() {
    echo "==> Running tree build benchmark..."
    if ! output=$(go test -bench=^BenchmarkBuildTree$ -run=^$ -benchmem -count=3 ./internal/core/... 2>&1); then
        echo "$output"
        exit 1
    fi
    echo "$output"

    echo "$output" | awk '
    /^BenchmarkBuildTree/ {
        seen++
        sum_ns += $3
        if (($7 + 0) > max_allocs) { max_allocs = $7 + 0 }
    }
    END {
        if (!seen) {
            print "Error: no BenchmarkBuildTree results parsed"
            exit 1
        }
        mean_ns = sum_ns / seen
        printf "Samples: %d, mean %.1fns/op, max %d allocs/op\n", seen, mean_ns, max_allocs
        # 100 tasks per op: budget < 1us/task (< 100,000ns mean) and < 5 allocs/task (< 500 allocs)
        if (mean_ns >= 100000.0) {
            printf "Error: BenchmarkBuildTree mean %.1fns/op over %d samples exceeds budget of 100,000ns (<1us/task)\n", mean_ns, seen
            failed = 1
        }
        if (max_allocs >= 500) {
            printf "Error: BenchmarkBuildTree %d allocs/op exceeds budget of 500 allocs (<5 allocs/task)\n", max_allocs
            failed = 1
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
