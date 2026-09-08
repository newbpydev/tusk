.PHONY: all setup fmt vet test test-unit test-compat build-storage race validate coverage bench bench-tree bench-build build clean help
.PHONY: test-scripts

test-scripts:
	@./scripts/test/test_scripts.sh

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

# Select GOTOOLCHAIN=go1.25.0 explicitly for the minimum-compiler proof.
test-compat:
	go version
	@test "$$(go list -m -f '{{.Version}}' modernc.org/sqlite)" = v1.58.0
	@test "$$(go list -m -f '{{.Version}}' modernc.org/libc)" = v1.75.6
	go test -v ./internal/storage -run 'TestSQLiteCompatibility|TestConnectorConstruction|TestConnectionFactory|TestReaderConnection|TestWriterConnection|TestAcquire'
	@./scripts/test/test_scripts.sh

build-storage:
	go version
	CGO_ENABLED=0 go build ./internal/storage
	@set -e; for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		echo "Building storage: $$target (CGO_ENABLED=0)"; \
		CGO_ENABLED=0 GOOS=$${target%/*} GOARCH=$${target#*/} go build ./internal/storage; \
	done
	@set -e; build_tmp=$$(mktemp -d); trap 'rm -rf "$$build_tmp"' EXIT; \
	for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		echo "Compiling storage tests: $$target (execution deferred on other hosts)"; \
		CGO_ENABLED=0 GOOS=$${target%/*} GOARCH=$${target#*/} go test -c -o "$$build_tmp/$${target%/*}-$${target#*/}.test" ./internal/storage; \
	done

race:
	go test -race -v ./...

coverage:
	@./scripts/coverage.sh
bench:
	@./scripts/bench.sh all

bench-tree:
	@./scripts/bench.sh tree

bench-build:
	@./scripts/bench.sh build
validate: fmt vet test race coverage test-scripts
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
	@echo "make test-compat - Prove pinned storage runtime and Go minimum fixtures"
	@echo "make build-storage - Compile CGO-free storage for all five targets"
	@echo "make race      - Run tests with data race detector"
	@echo "make validate    - Run full verification suite (fmt, vet, test, race, coverage)"
	@echo "make coverage    - Run coverage check with race detector"
	@echo "make bench       - Run all core micro-benchmarks"
	@echo "make bench-tree  - Run tree traversal micro-benchmark"
	@echo "make bench-build - Run tree build micro-benchmark"
	@echo "make build     - Compile binary to bin/tusk"
	@echo "make clean     - Clean temporary build artifacts"
