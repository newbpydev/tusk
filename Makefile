.DEFAULT_GOAL := all
.PHONY: all setup fmt vet test test-unit test-compat build-storage race validate coverage bench bench-tree bench-build build clean help
.PHONY: setup-sqlc generate check-generated test-scripts bench-storage
.PHONY: test-ci
test-ci:
	bash scripts/test/test_ci.sh

.PHONY: test-docgen
DOCGEN_TEST_FLAGS ?=
test-docgen:
	go test $(DOCGEN_TEST_FLAGS) -v ./scripts/docgen

.PHONY: test-docs test-notices check-notices generate-notices
test-docs:
	bash scripts/test/test_docs.sh

test-notices:
	bash scripts/test/test_notices.sh

check-notices:
	bash scripts/notices.sh check

generate-notices:
	bash scripts/notices.sh generate

.PHONY: test-release setup-release release-check release-snapshot release-candidate verify-release build-release-fixture
RELEASE_VERSION ?=
RELEASE_SHA ?= HEAD
RELEASE_OUTPUT ?=
# Freeze raw values and pass them as environment data, never recipe shell code.
override RELEASE_VERSION := $(value RELEASE_VERSION)
override RELEASE_SHA := $(value RELEASE_SHA)
override RELEASE_OUTPUT := $(value RELEASE_OUTPUT)
export RELEASE_VERSION RELEASE_SHA RELEASE_OUTPUT
override RELEASE_ASSETS := $(value RELEASE_ASSETS)
export RELEASE_ASSETS
test-release:
	go test -v ./cmd/tusk -run '^TestVersion_'
	go test -v ./scripts/releasecheck $(TUSK_RELEASE_TEST_ARGS)
	bash scripts/test/test_release.sh
	bash scripts/test/test_candidate.sh

setup-release:
	bash scripts/release.sh setup

release-check:
	bash scripts/release.sh check

release-snapshot:
	bash scripts/release.sh snapshot

release-candidate:
	bash scripts/release.sh candidate

verify-release:
	bash scripts/release_check.sh verify "$$RELEASE_ASSETS"

.PHONY: coverage-release
override TUSK_RELEASE_COVERAGE_FILE := $(value TUSK_RELEASE_COVERAGE_FILE)
export TUSK_RELEASE_COVERAGE_FILE
coverage-release:
	go test -coverprofile="$$TUSK_RELEASE_COVERAGE_FILE" ./scripts/releasecheck

# Test-owned checkout and output; called by the inspector integration tests.
build-release-fixture:
	CGO_ENABLED=0 GOTOOLCHAIN=go1.27.1 go build -trimpath -overlay="$$TUSK_RELEASE_OVERLAY" -ldflags='-s -w' -o "$$TUSK_RELEASE_FIXTURE_BINARY" ./cmd/tusk

.PHONY: build-gh-deadline
build-gh-deadline:
	@test -n "$$GH_DEADLINE_OUTPUT"
	go build -trimpath -buildvcs=false -o "$$GH_DEADLINE_OUTPUT" scripts/ghdeadline/main.go

.PHONY: generate-docs check-docs tidy-modules test-completions
test-completions: build check-docs
	bash scripts/test/test_completions.sh

generate-docs:
	go run ./scripts/docgen --output docs

check-docs:
	go run ./scripts/docgen --output docs --check

tidy-modules:
	go mod tidy

.PHONY: setup-ci preflight-ci check-ci check-ci-drift
setup-ci:
	bash scripts/ci-check.sh setup

preflight-ci:
	bash scripts/ci-check.sh preflight

check-ci:
	bash scripts/ci-check.sh check

check-ci-drift:
	bash scripts/ci-check.sh drift
.PHONY: build-service bench-service

build-service:
	go version
	@set -e; build_tmp=$$(mktemp -d); trap 'rm -rf "$$build_tmp"' EXIT; \
	for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		echo "Building service and tests: $$target (CGO_ENABLED=0; native execution separate)"; \
		CGO_ENABLED=0 GOOS=$${target%/*} GOARCH=$${target#*/} go build ./internal/service/...; \
		CGO_ENABLED=0 GOOS=$${target%/*} GOARCH=$${target#*/} go test -c -o "$$build_tmp/service-$${target%/*}-$${target#*/}.test" ./internal/service; \
		CGO_ENABLED=0 GOOS=$${target%/*} GOARCH=$${target#*/} go test -c -o "$$build_tmp/dateparse-$${target%/*}-$${target#*/}.test" ./internal/service/dateparse; \
	done

bench-service:
	go version
	go test ./internal/service -run '^$$' -bench '^BenchmarkService$$' -benchmem -benchtime=1x

setup-sqlc:
	bash scripts/sqlc.sh setup

generate:
	bash scripts/sqlc.sh generate
	$(MAKE) generate-schema-catalog

check-generated: check-schema-catalog
	bash scripts/sqlc.sh check

.PHONY: generate-schema-catalog check-schema-catalog
generate-schema-catalog:
	TUSK_UPDATE_SCHEMA_CATALOG=1 go test ./internal/storage -run '^TestEmbeddedSchemaCatalog$$' -count=1

check-schema-catalog:
	go test ./internal/storage -run '^TestEmbeddedSchemaCatalog$$' -count=1

test-scripts:
	@./scripts/test/test_scripts.sh
	@bash scripts/test/test_sqlc.sh
	@bash scripts/test/test_ci.sh
	@bash scripts/test/test_docs.sh
	@bash scripts/test/test_notices.sh
	@bash scripts/test/test_release.sh
	@bash scripts/test/test_candidate.sh
	@bash scripts/test/test_release_smoke.sh
	@bash scripts/test/test_homebrew.sh
	@bash scripts/test/test_promotion.sh
	@bash scripts/test/test_metadata.sh

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
bench-storage:
	go version
	go test ./internal/storage -run '^$$' -bench '^BenchmarkStorage$$' -benchmem -benchtime=200ms

bench:
	@./scripts/bench.sh all

bench-tree:
	@./scripts/bench.sh tree

bench-build:
	@./scripts/bench.sh build
validate: fmt vet test race coverage test-scripts check-modules
	@echo "All canonical quality gates passed."

BUILD_OUTPUT ?= bin/tusk
CLI_TEST_RUN ?= .
TUI_TEST_RUN ?= .
TUI_TEST_FLAGS ?=
.PHONY: bench-tui
TUI_BENCH_OUTPUT ?= docs/verification-evidence/005/tui-latency.json
bench-tui: build
	TUSK_TUI_BENCH_OUTPUT="$(abspath $(TUI_BENCH_OUTPUT))" go test ./internal/tui -run '^TestTUIMeasurements$$' -count=1 -timeout=20m -v
	TUSK_TUI_STARTUP_OUTPUT="$(abspath $(TUI_BENCH_OUTPUT)).startup.json" go test ./cmd/tusk -run '^TestTUIStartupMeasurements$$' -count=1 -timeout=5m -v

.PHONY: test-tui build-tui build-tui-fixture

# Manual fault-injected app, never test-result verification in Kitty.
build-tui-fixture:
	mkdir -p bin
	CGO_ENABLED=0 go build -tags=tusk_fixture -o bin/tusk-tui-fixture ./scripts/fixtures/tui

test-tui:
	go test $(TUI_TEST_FLAGS) -v ./internal/tui ./internal/terminaltext ./internal/cli ./cmd/tusk -run '$(TUI_TEST_RUN)'

build-tui: build-cli
	@set -e; build_tmp=$$(mktemp -d); trap 'rm -rf "$$build_tmp"' EXIT; \
	for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		echo "Compiling TUI tests: $$target (CGO_ENABLED=0; native execution separate)"; \
		CGO_ENABLED=0 GOOS=$${target%/*} GOARCH=$${target#*/} go test -c -o "$$build_tmp/tui-$${target%/*}-$${target#*/}.test" ./internal/tui; \
		CGO_ENABLED=0 GOOS=$${target%/*} GOARCH=$${target#*/} go test -c -o "$$build_tmp/text-$${target%/*}-$${target#*/}.test" ./internal/terminaltext; \
	done

.PHONY: test-cli build-cli

test-cli:
	go test -v ./internal/cli ./cmd/tusk -run '$(CLI_TEST_RUN)'

build-cli:
	go version
	@test "$$(go list -m -f '{{.Version}}' modernc.org/sqlite)" = v1.58.0
	@test "$$(go list -m -f '{{.Version}}' modernc.org/libc)" = v1.75.6
	@set -e; build_tmp=$$(mktemp -d); trap 'rm -rf "$$build_tmp"' EXIT; \
	for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		echo "Building CLI and tests: $$target (CGO_ENABLED=0; native execution separate)"; \
		CGO_ENABLED=0 GOOS=$${target%/*} GOARCH=$${target#*/} go build -o "$$build_tmp/tusk-$${target%/*}-$${target#*/}" ./cmd/tusk; \
		CGO_ENABLED=0 GOOS=$${target%/*} GOARCH=$${target#*/} go test -c -o "$$build_tmp/cli-$${target%/*}-$${target#*/}.test" ./internal/cli; \
		CGO_ENABLED=0 GOOS=$${target%/*} GOARCH=$${target#*/} go test -c -o "$$build_tmp/main-$${target%/*}-$${target#*/}.test" ./cmd/tusk; \
	done

build:
	mkdir -p "$(dir $(BUILD_OUTPUT))"
	CGO_ENABLED=0 go build -o "$(BUILD_OUTPUT)" ./cmd/tusk

clean:
	rm -rf bin/ coverage.out .tusk-test*.db

help:
	@echo "Tusk Canonical Build & Quality Gates"
	@echo "===================================="
	@echo "make setup     - Verify environment and Go toolchain"
	@echo "make setup-ci preflight-ci check-ci - Pinned workflow tooling and native CI contract"
	@echo "make test-ci check-ci-drift - Negative CI fixtures and post-gate source drift"
	@echo "make generate-docs check-docs test-completions - Static completion scripts and deterministic manuals"
	@echo "make test-docs check-notices generate-notices - Public guides and reviewed redistribution inventory"
	@echo "make fmt       - Format all Go source files"
	@echo "make vet       - Run go vet static analysis"
	@echo "make test      - Run all tests"
	@echo "make test-unit - Run short unit tests"
	@echo "make test-compat - Prove pinned storage runtime and Go minimum fixtures"
	@echo "make build-storage - Compile CGO-free storage for all five targets"
	@echo "make build-service - Compile CGO-free service and tests for all five targets"
	@echo "make bench-service - Measure fixed single-sample service workloads"
	@echo "make race      - Run tests with data race detector"
	@echo "make validate    - Run fmt, vet, test, race, coverage, script and module checks"
	@echo "make check-modules - Check go.mod/go.sum with go mod tidy -diff"
	@echo "make coverage    - Run coverage check with race detector"
	@echo "make bench-storage - Measure storage queries and open costs"
	@echo "make bench       - Run all core micro-benchmarks"
	@echo "make bench-tree  - Run tree traversal micro-benchmark"
	@echo "make bench-build - Run tree build micro-benchmark"
	@echo "make build     - Compile binary to bin/tusk"
	@echo "make clean     - Clean temporary build artifacts"

.PHONY: bench-cli
CLI_BENCH_PROFILE ?= reference
CLI_BENCH_OUTPUT ?= $(if $(filter reference,$(CLI_BENCH_PROFILE)),docs/verification-evidence/004/latency.json,)

bench-cli: build
	go run ./scripts/cli-bench --binary "$(BUILD_OUTPUT)" --output "$(CLI_BENCH_OUTPUT)" --acceptance-profile "$(CLI_BENCH_PROFILE)"

.PHONY: profile-cli
CLI_PROFILE_OUTPUT ?= /tmp/tusk-cli.cpu
profile-cli:
	go test ./scripts/cli-bench -run '^$$' -bench '^BenchmarkCLIProfile$$' -benchtime=3s -cpuprofile "$(CLI_PROFILE_OUTPUT)" -o "$(CLI_PROFILE_OUTPUT).test"

.PHONY: bench-cli-json
bench-cli-json:
	go test ./internal/cli -run '^$$' -bench '^BenchmarkJSON(Escaped)?$$' -benchmem -count=3

.PHONY: check-modules
check-modules:
	go mod tidy -diff

.PHONY: test-cli-latency-codec
test-cli-latency-codec:
	go test -v ./internal/storage -run '^TestCodec_'

.PHONY: bench-cli-conditions
CLI_CONDITIONS_OUTPUT ?= /tmp/tusk-cli-conditions.json
bench-cli-conditions: build
	TUSK_CLI_CONDITIONS=1 TUSK_CLI_BINARY="$(abspath $(BUILD_OUTPUT))" TUSK_CLI_CONDITIONS_OUTPUT="$(abspath $(CLI_CONDITIONS_OUTPUT))" go test -v ./scripts/cli-bench -run '^TestCLIConditions$$' -count=1

# The pinned analyzer must support the actual release compiler.
.PHONY: setup-vulnerabilities check-vulnerabilities
setup-vulnerabilities:
	bash scripts/ci-check.sh setup-vulnerabilities

VULNCHECK_BIN ?= bin/tools/govulncheck
check-vulnerabilities:
	"$(VULNCHECK_BIN)" ./cmd/tusk

.PHONY: candidate-identity record-candidate verify-candidate check-candidate-workflow test-candidate
# Candidate identity crosses recipes only as raw environment values.
$(foreach parameter,CANDIDATE_DIR CANDIDATE_MANIFEST_SHA256 CANDIDATE_RUN_ID CANDIDATE_VERIFICATION_DIR,$(eval override $(parameter) := $$(value $(parameter))))
export CANDIDATE_DIR CANDIDATE_MANIFEST_SHA256 CANDIDATE_RUN_ID CANDIDATE_VERIFICATION_DIR
candidate-identity:
	bash scripts/candidate.sh identity
record-candidate:
	bash scripts/candidate.sh record
verify-candidate:
	bash scripts/candidate.sh verify
check-candidate-workflow:
	bash scripts/candidate.sh contract
test-candidate:
	bash scripts/test/test_candidate.sh

.PHONY: test-release-smoke
test-release-smoke:
	go test -v ./internal/cli ./cmd/tusk -run '^TestReleaseSelection_'
	bash scripts/test/test_release_smoke.sh

.PHONY: release-smoke release-smoke-processes bench-cli-release bench-tui-release
$(foreach parameter,RELEASE_BINARY RELEASE_MANIFEST RELEASE_EVIDENCE_SCOPE CANDIDATE_VERIFICATION_RECEIPT RELEASE_SMOKE_OUTPUT,$(eval override $(parameter) := $$(value $(parameter))))
export RELEASE_BINARY RELEASE_MANIFEST RELEASE_EVIDENCE_SCOPE CANDIDATE_VERIFICATION_RECEIPT RELEASE_SMOKE_OUTPUT
override RELEASE_CLI_BENCH_OUTPUT := $(value CLI_BENCH_OUTPUT)
override RELEASE_TUI_BENCH_OUTPUT := $(value TUI_BENCH_OUTPUT)
export RELEASE_CLI_BENCH_OUTPUT RELEASE_TUI_BENCH_OUTPUT
release-smoke:
	bash scripts/release_smoke.sh smoke
release-smoke-processes:
	go test -v ./internal/cli -run '^(TestProcess_Workflow|TestDocsExamples_QuickStartAndClosedBackup|TestReleaseLifecycle_)' -count=1
	go test -v ./cmd/tusk -run '^TestTUIProcess_TerminalLifecycle$$' -count=1
bench-cli-release:
	RELEASE_BENCH_OUTPUT="$$RELEASE_CLI_BENCH_OUTPUT" bash scripts/release_smoke.sh bench-cli
bench-tui-release:
	RELEASE_BENCH_OUTPUT="$$RELEASE_TUI_BENCH_OUTPUT" bash scripts/release_smoke.sh bench-tui

.PHONY: test-homebrew check-homebrew homebrew-candidate homebrew-destination
$(foreach parameter,HOMEBREW_CASK,$(eval override $(parameter) := $$(value $(parameter))))
export HOMEBREW_CASK
test-homebrew:
	go test -v ./scripts/releasecheck -run '^TestCask_'
	bash scripts/test/test_homebrew.sh
check-homebrew:
	bash scripts/homebrew.sh check
homebrew-candidate:
	bash scripts/homebrew.sh render
homebrew-destination:
	bash scripts/homebrew.sh destination

.PHONY: test-release-promotion release-prepare release-draft release-publish release-readback
$(foreach parameter,RELEASE_PROMOTION_OUTPUT RELEASE_ACCEPTANCE RELEASE_AUTHORIZATION RELEASE_NOTES,$(eval override $(parameter) := $$(value $(parameter))))
export RELEASE_PROMOTION_OUTPUT RELEASE_ACCEPTANCE RELEASE_AUTHORIZATION RELEASE_NOTES
test-release-promotion:
	bash scripts/test/test_promotion.sh
release-prepare:
	bash scripts/promote.sh prepare
release-draft:
	bash scripts/promote.sh draft
release-publish:
	bash scripts/promote.sh publish
release-readback:
	bash scripts/promote.sh readback

.PHONY: lint-release-promotion
lint-release-promotion:
	shellcheck scripts/promote.sh scripts/test/test_promotion.sh scripts/repository_metadata.sh scripts/test/test_metadata.sh scripts/candidate.sh scripts/test/test_candidate.sh scripts/homebrew.sh scripts/test/test_homebrew.sh

.PHONY: test-repository-metadata prepare-repository-metadata apply-repository-metadata
$(foreach parameter,METADATA_OUTPUT METADATA_AUTHORIZATION METADATA_RELEASE_RECEIPT,$(eval override $(parameter) := $$(value $(parameter))))
export METADATA_OUTPUT METADATA_AUTHORIZATION METADATA_RELEASE_RECEIPT
test-repository-metadata:
	bash scripts/test/test_metadata.sh
prepare-repository-metadata:
	bash scripts/repository_metadata.sh prepare
apply-repository-metadata:
	bash scripts/repository_metadata.sh apply
