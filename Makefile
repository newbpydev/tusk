.PHONY: all setup fmt vet test test-unit race validate coverage bench bench-tree bench-build build clean help

all: validate build

setup:
	@./scripts/setup.sh

fmt:
	@./scripts/fmt.sh

vet:
	go vet ./...

test:
	go test -v ./...

test-unit:
	go test -v -short ./...

race:
	go test -race -v ./...

coverage:
	@./scripts/coverage.sh
bench:
	go test -bench=. -run=^$$ -benchmem ./internal/core/...

bench-tree:
	go test -bench=BenchmarkTreeTraversal -run=^$$ -benchmem ./internal/core/...

bench-build:
	go test -bench=BenchmarkBuildTree -run=^$$ -benchmem ./internal/core/...

validate: fmt vet test race coverage
	@echo "All canonical quality gates passed."

build:
	mkdir -p bin
	go build -o bin/tusk cmd/tusk/main.go

clean:
	rm -rf bin/ coverage.out .tusk-test*.db

help:
	@echo "Tusk Canonical Build & Quality Gates"
	@echo "===================================="
	@echo "make setup     - Verify environment and Go toolchain"
	@echo "make fmt       - Format all Go source files"
	@echo "make vet       - Run go vet static analysis"
	@echo "make test      - Run all tests"
	@echo "make test-unit - Run short unit tests"
	@echo "make race      - Run tests with data race detector"
	@echo "make validate  - Run full verification suite (fmt, vet, test, race)"
	@echo "make coverage    - Run coverage check with race detector"
	@echo "make bench       - Run all core micro-benchmarks"
	@echo "make bench-tree  - Run tree traversal micro-benchmark"
	@echo "make bench-build - Run tree build micro-benchmark"
	@echo "make build     - Compile binary to bin/tusk"
	@echo "make clean     - Clean temporary build artifacts"
