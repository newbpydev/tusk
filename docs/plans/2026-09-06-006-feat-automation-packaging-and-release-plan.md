---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
product_contract_source: 2026-09-06-001-feat-tusk-modern-task-system-plan.md
feature-id: "006"
title: Automation, Packaging, Release and Public Repository Readiness
type: feat
date: 2026-09-06
deepened: 2026-10-01
execution: code
planning_scope: release-engineering-and-public-documentation
---

# Feature Plan 006: Automation, Packaging, Release and Public Repository Readiness

## Goal Capsule

Make Tusk installable and understandable from its GitHub repository, with verified
CLI/TUI binaries, useful documentation and a repeatable release process. Users
should be able to select their platform, verify a download, create their first
task and find the data/recovery instructions without reading planning artifacts.

- **Actors:** Terminal users, script/agent users, contributors and release maintainers.
- **Authority:** [AGENTS.md](../../AGENTS.md), the [product contract](2026-09-06-001-feat-tusk-modern-task-system-plan.md), and [MASTERPLAN.md](../../MASTERPLAN.md). The masterplan controls activation and progress.
- **Surfaces:** CLI/TUI integration, infrastructure/operations, installation/data lifecycle, documentation and GitHub repository settings.
- **Artifact pack:** This plan, its [verification plan](../verification-plans/2026-09-06-006-feat-automation-packaging-and-release-verification-plan.md) and [workorder](../workorders/2026-09-06-006-feat-automation-packaging-and-release-issues-workorder.md), under the existing first-party `docs/` root.
- **Readiness:** Decisions and test contracts are ready for implementation. Execution, native acceptance, hosted checks and publication are unexecuted. U1 can start after an implementation instruction; license and distribution gates apply at their named boundaries.
- **Sequence:** U1 → U3 → U4 → U2 → U6 → U5 → U7 → U8. U1/U2/U3 preserve the original CI/packaging/completion identities.

This request authorizes planning. GitHub metadata changes, README implementation,
workflow dispatch, commits, pushes, tags and release publication are future work.

---

## Product Contract

### Problem frame and current evidence

On 2026-10-01, local and remote `main` both resolve to
`2f2d3ffe8c52d780553f2073fff0d07d89bde4ba`; Feature 005 merged in PR #5 at
`42d7c52`. The checkout was clean before this planning pass. The CLI, TUI,
storage, version command and guides exist. Completion is explicitly disabled in
`internal/cli/root.go`; version is the constant `0.2.0-reboot` in
`cmd/tusk/main.go`. `go.mod` requires Go 1.25.0 and contains local Bubble Tea and
Glamour replacements. Those patched modules and their regular Git gallery bytes
must survive source packaging and fresh checkout.

Live GitHub inspection of [newbpydev/tusk](https://github.com/newbpydev/tusk)
found a public repository with admin access, a description, empty topics/homepage,
no detected project license, no releases/tags and zero Actions workflows. Issues
are enabled; Discussions and Pages are disabled. No first-party `LICENSE`,
`CONTRIBUTING.md`, `SECURITY.md`, `.github/` or `.goreleaser.yaml` exists.
`newbpydev/homebrew-tap` returned 404 to the authenticated inspection; it is an
unverified distribution prerequisite, not an existing destination.

The previous short outline's `implementation-ready` label was unsupported. Its
Go 1.24 assumption and binary Homebrew formula approach are superseded below.
Historical local Feature 002–005 receipts remain evidence for those candidates;
they do not prove this release candidate's native consoles or downloaded assets.

### Requirements

#### Native CI and build governance

- R1. CI runs for pull requests targeting `main`, pushes to `main` and explicit dispatch; PR jobs have read-only repository permissions and no publishing secrets.
- R2. Execute canonical native quality/build gates on all five release OS/architecture pairs using the pinned release compiler; additionally prove the Go 1.25.0 source minimum on Linux and compile all five targets with that minimum.
- R3. Windows gates use Git Bash, GNU Make and a matching native C compiler for Go race builds; no WSL result or cross-build closes native Windows acceptance.
- R4. Formatting, generated SQL/schema and module checks fail on source drift; canonical validation preserves the existing ≥95% nonexempt package coverage policy.
- R5. Pin Actions by full upstream commit SHA and downloaded tools by version/digest; preserve the accepted runtime dependencies and local patches.

#### Packaging, identity and data-preserving installation

- R6. Release version is an immutable generated constant matching the tag; snapshots and ordinary source builds cannot claim a published version.
- R7. Package CGO-disabled Linux amd64/arm64, Darwin amd64/arm64 and Windows amd64 binaries, with embedded timezone data and no runtime Go/compiler/database-service requirement.
- R8. Distribute deterministic target archives, completions, man pages, project license, third-party notices, a complete source archive and SHA256 checksums.
- R9. Bind source SHA, version, compiler, tool pins, generated version input, target, payload hashes and workflow identity in a release manifest and verified provenance.
- R10. Local checks/snapshots/candidates do not publish, create remote tags or update taps; failures retain diagnostics and leave unrelated checkout files and user data intact.
- R11. Installing, replacing, upgrading or uninstalling a binary preserves tasks/events and the database/WAL/SHM; newer-schema refusal never attempts an automatic downgrade.
- R12. Document and replay a consistent offline backup/restore procedure, including sidecars and SQLite NORMAL durability limitations, without claiming a new backup command.

#### Storage-free completion and manual generation

- R13. Completion and manual generation work with broken database/date configuration, redirected streams and no terminal or storage initialization.
- R14. `completion bash|zsh|fish` emits only the requested script to stdout; invalid shell, arity or flags exit 2, and output failures exit 1 with a safe stderr diagnostic.
- R15. Shell completion suggests static commands, flags and enums only; it does not query task IDs, notes, tags or any database.
- R16. Generate manuals/completion files from the actual fresh Cobra command tree with stable dates/order and fail on checked-in drift.

#### Public documentation and repository presentation

- R17. README leads with purpose, a real screenshot, verified installation choices and a short first-task workflow; deeper CLI/TUI/install/recovery/contributor guides remain linked.
- R18. Every runnable documentation example matches real command/date/JSON/exit behavior; clearly identify IDs/placeholders and optional tools such as `jq`.
- R19. Support and performance claims cite dated candidate evidence with platform, terminal, fixtures and limits; screenshots have sanitized data, alt text and provenance.
- R20. Fill repository description, homepage, topics and social preview with reviewed values; verify GitHub's rendered README, links, badges, license detection and release visibility after publication.
- R21. Owner selects the project license and confirms authority before redistribution; archive all required notices, including dependencies reached through local replacements.
- R22. Add concise contributor and security/support guidance with working GitHub contact routes, issue/PR templates and a version policy; do not invent an email, response SLA, donation link or maintenance team.

#### Native runtime, terminal and performance acceptance

- R23. Execute exact packaged bytes natively on every target for first-use storage, path/Unicode/space handling, named timezones, WAL concurrency, JSON/pipes and graceful/forced interruption.
- R24. Inspect packaged TUI input, resize, plain mode, save/conflict/delete/recovery and terminal restoration on native terminals; synthetic or headless checks remain separate.
- R25. Re-run candidate CLI/TUI performance evidence on the declared measurement host; keep the reference criteria and all failed/raw samples visible.

#### Trusted distribution and unit closure

- R26. A trusted manual candidate workflow builds and attests downloadable artifacts without creating a public release; native tests consume those exact archived bytes.
- R27. Privileged promotion verifies the trusted workflow/ref/SHA, artifact identity and accepted manifest digest; PR artifacts, branch names and release text cannot become executable shell input.
- R28. Create a complete draft, verify downloaded assets, then publish only an explicitly authorized tag/version/candidate; retries cannot replace published bytes or move a release tag.
- R29. Provide a tested macOS Homebrew cask in the owner-controlled tap, using the verified release URLs/checksums and preserving data on uninstall.
- R30. Each implementation unit retains red/green, review and applicable evidence, synchronizes its triplet/masterplan and passes `make validate` before its coherent local commit; publication and merge need their own authorization.

### Acceptance examples

- AE1. A Linux user downloads the matching archive, verifies SHA256, puts `tusk` on PATH, runs version/help without a database, then adds a task and opens it in the TUI.
- AE2. A Windows user follows native PowerShell extraction/hash/PATH instructions, creates data under a path containing spaces/Unicode and exits the TUI with the console usable afterward.
- AE3. A script user reads one clean JSON value from `list --all --json`; completion generation with unusable `TUSK_DB_PATH` still creates no data directory.
- AE4. A maintainer's incomplete/tampered candidate or mismatched SHA cannot publish. Replacing/removing a verified binary preserves fixture tasks/history.
- AE5. A visitor understands available installation routes, the tested platform scope and limitations from GitHub's About panel and README; badges never suggest an absent release or unexecuted gate.

### Scope boundaries

In scope: CI portability repairs justified by red tests, release version/build
metadata, five-platform archives, static completions/manuals, native acceptance,
license/notices, install/backup docs, README, community files, repository metadata,
macOS Homebrew distribution and authorized release settlement.

Deferred to follow-up: apt/rpm/winget/Scoop registries, container images, a curl-to-shell
installer, a self-updater, Apple signing/notarization and Windows code signing.
Unsigned download behavior is documented and inspected; failure to install through
normal platform controls blocks that route rather than adding a security bypass.
A separately hosted GitHub Pages website, custom domain, cloud sync, telemetry,
new schema, task semantics and application feature work are outside this scope.

---

## Planning Contract

### Key technical decisions

- KTD1. **Build tooling is separate from runtime.** Keep Go 1.25.0 as the source compatibility floor. Pin Go 1.27.1 for production artifacts and native CI, GoReleaser OSS v2.18.2 and actionlint v1.7.12, observed in official release feeds on 2026-10-01. U1 records verified archive hashes and Action SHAs in `scripts/tool-versions.json`; sqlc retains v1.31.1 and its existing digest catalogue. Recheck availability/security before execution; a changed pin is a synchronized amendment, never a floating `latest` download. Vulnerability tooling must support the selected compiler; unavailable analysis stays a release blocker.
- KTD2. **Native matrix, separate minimum job.** Use `ubuntu-24.04` amd64, `ubuntu-24.04-arm` arm64, `macos-15-intel` amd64, `macos-15` arm64 and `windows-2025` amd64. Assert actual OS/architecture/tool versions. These are tested baselines, not claims about every older OS. The minimum-compiler job runs full canonical gates and existing `build-tui` cross-builds on Linux. Windows release users are additionally inspected in Windows Terminal on Windows 11. If a runner is unavailable, the gate stays pending; do not silently drop an architecture.
- KTD3. **Portability without a second gate system.** All automation invokes Make targets. Windows uses `shell: bash`, Git for Windows and verified GNU Make/GCC on PATH; Go remains the native Windows toolchain. Native macOS has clang and Bash/Make; Linux has GCC. Race/coverage use CGO support, while release builds use `CGO_ENABLED=0`. Ensure LF checkout, bounded jobs (30 minutes), read-only CI credentials, fail-fast disabled and post-gate source-diff failure. Cache downloaded modules, never promote a cached executable as accepted evidence.
- KTD4. **One fresh command-tree builder.** Refactor only the construction seam in `internal/cli/root.go` so runtime, static completion and `scripts/docgen` share an invocation-owned tree. No exported mutable Cobra singleton. Generator options refuse service/TUI/confirmation access. Disable autogenerated help footer/today timestamps in manuals. No public `man` command is added.
- KTD5. **Immutable version overlay.** Move the existing `Version` constant to `cmd/tusk/version.go` with a development label. A release helper generates a validated, immutable Go constant file and JSON `go build -overlay` mapping in owned temporary storage. GoReleaser receives that build flag; explicitly override its default `-X` linker injections. Record the generated input hash, use `-trimpath`, and avoid wall-clock build timestamps. Clean checkout and local replacements remain unchanged. Package-level mutable version variables are prohibited. Test actual built output, not only the manifest text.
- KTD6. **Stable, complete payloads.** Names are `tusk_<version>_<linux|darwin|windows>_<amd64|arm64>.<tar.gz|zip>`; version omits the leading `v`. Unix archives contain `tusk`, Windows `tusk.exe`. Include `README.md`, `LICENSE`, `THIRD_PARTY_NOTICES.md`, `completions/tusk.{bash,zsh,fish}` and `man/*.1`; archive members have no absolute/traversal paths or user database. Root binary remains directly extractable. Include a tagged source tarball with patched `third_party/` modules. `checksums.txt` hashes payloads, source and notices; `release-manifest.json` inventories those hashes and the checksum-file digest without a self-hash cycle. Attest payloads and manifest separately.
- KTD7. **CGO-free is a precise claim.** Assert binary build metadata has CGO disabled and the expected dependencies/tzdata. Linux ELF inspection proves no interpreter or shared-library requirement. Darwin/Windows may use normal OS libraries; do not describe them as fully static ELF binaries. Target runtime tests, rather than a filename, prove architecture compatibility.
- KTD8. **Build once, accept and promote.** U2 adds local snapshot and tagged-candidate modes. Candidate mode uses an owned temporary checkout at an exact SHA and a local-only intended version tag; `goreleaser release --skip=publish --clean` packages it without remote writes. U6 creates the trusted hosted candidate workflow and attests its output. U5 downloads and tests that candidate. U8 uploads and publishes those accepted files through GitHub release APIs, without rebuilding. GoReleaser OSS remains the packager; promotion is a small manifest-driven script, not a presumed OSS publish-existing subcommand.
- KTD9. **Release lifecycle and authorization.** Proposed first tag is `v0.3.0`, continuing beyond the current `0.2.0-reboot` development label; owner confirms the actual tag/SHA in U8. Semantic versions reject malformed input, existing conflicting tags and a tag outside authorized `main` history. No automatic public release on ordinary push/PR. Draft creation/promotion are separate Make targets and explicit maintainer actions. Before publication all assets, native receipts, notices and hashes are complete. Enable immutable releases only as an owner-reviewed setting; if unavailable, retain a no-overwrite/no-retag policy. Bad published versions are deprecated in notes and superseded by a new version, not overwritten.
- KTD10. **Least privilege across the promotion boundary.** Candidate generation uses no tap or release-publishing secrets; only its provenance job gets `id-token: write` and `attestations: write`. Promotion uses repository contents write for the authorized draft/release and validates artifact run, trusted workflow, main SHA and manifest digest. Action checkout credentials are not persisted. No `pull_request_target` execution of PR code, untrusted privileged cache/artifact reuse, raw environment dumps or shell interpolation of externally supplied text. Use structured arguments/body files for release notes and metadata writes.
- KTD11. **Homebrew cask, staged after acceptance.** Use current `homebrew_casks` configuration with upload disabled. Proposed tap is `newbpydev/homebrew-tap`, `Casks/tusk.rb`; owner must establish destination/access before U7. macOS Intel/ARM archives, SHA256, binary, completions and manpages are configured and tested; no uninstall/zap hook touches data and no quarantine-removal hook is added. Validate a local candidate cask before publication; publish the reviewed cask only after the release URL exists. Tap update is initially an explicit maintainer commit/push, avoiding a permanent cross-repository secret; a failure leaves verified direct downloads usable but Phase 6 open. Linuxbrew is not advertised by this macOS route.
- KTD12. **Honest installation routes.** Before the first release, document source checkout plus `make setup build` and the current native support limits. After verified publication, lead with binary downloads for all five targets and macOS cask installation; Windows examples use PowerShell, Unix examples use POSIX shell. Versioned `go install github.com/newbpydev/tusk/cmd/tusk@...` is unsupported while local replacements exist. Show hash verification before extraction/install, PATH checks, upgrade/uninstall and recovery from wrong architecture, permission errors, download failure or checksum mismatch. No user must install Go, Make, Kitty or `jq` to run a prebuilt binary.
- KTD13. **README and About are the repository landing page.** Interpret the requested GitHub page as the existing repository home. Homepage points to `https://github.com/newbpydev/tusk#readme`. Proposed description: `Local task management in your terminal: a keyboard-driven TUI and scriptable CLI, backed by SQLite.` Proposed topics: `go`, `golang`, `cli`, `tui`, `task-manager`, `terminal`, `sqlite`, `bubbletea`, `productivity`, `offline`, `command-line`. Social preview uses a sanitized real TUI capture with readable title; manually upload only if the API cannot support that field. Leave absent funding/website links empty and preserve existing settings unless a reviewed change has a concrete purpose.
- KTD14. **Evidence serves users.** README order: purpose/status, real screenshot with alt text, install options/platform table, quick start, core features, CLI/JSON and TUI pointers, storage/config/backup, verification/support limits, contributing/security/license. Show only a few working badges: CI, released version and selected license after each exists. Reuse inspected Feature 005 screenshot evidence only with its original date/source label; fresh packaged-app capture is preferred for release. Link raw benchmark/acceptance receipts and contextualize finite-sample measurements. Do not turn internal workorders or implementation IDs into the user's installation workflow.
- KTD15. **License is an owner decision.** Recommend MIT for first-party code, pending owner selection and rights confirmation. Existing third-party licenses remain intact. Generate a dependency/license inventory that includes local replaced trees and embedded images; unclassified obligations block distribution. `SECURITY.md` uses GitHub private vulnerability reporting only after verified enabled; otherwise name the repository owner's GitHub contact route without claiming confidentiality for public issues. Issues/PR templates request OS/arch/version/reproduction, redact personal task data and distinguish security reports.
- KTD16. **Keep terminal and measurement evidence distinct.** Bash runs all Make/tests/latency; owned Kitty windows on Linux/macOS operate only the actual app with isolated data. Windows uses an owned native Windows Terminal/PowerShell session for separate native-console evidence. Retain identity/dimensions/observations, sanitized captures and cleanup. Existing benchmark targets rebuild their binary; U5 adds `bench-cli-release` and `bench-tui-release` to consume `RELEASE_BINARY` unchanged, with hash-preservation failure fixtures. Source-model TUI measurements remain separate from packaged startup. The accepted Feature 005 Ryzen profile was scoped to PR #5; do not silently extend it to release acceptance. U5 applies the reference CLI criteria, retains any separate calibrated result and seeks an explicit release decision if the unchanged reference gate fails.

### High-level technical design

```mermaid
flowchart LR
  U1[U1 Native CI] --> U3[U3 Completions and manuals]
  U3 --> U4[U4 README and license readiness]
  U4 --> U2[U2 Local packaging]
  U2 --> U6[U6 Hosted candidate and provenance]
  U6 --> U5[U5 Exact-asset native acceptance]
  U5 --> U7[U7 Homebrew candidate]
  U7 --> U8[U8 Authorized publication and metadata]
```

```mermaid
stateDiagram-v2
  [*] --> LocalSnapshot
  LocalSnapshot --> HostedCandidate: trusted SHA and version
  HostedCandidate --> AcceptedCandidate: exact hashes and all gates
  AcceptedCandidate --> CompleteDraft: authorized tag and upload
  CompleteDraft --> Published: downloaded verification and authorization
  Published --> TapAvailable: reviewed cask publication
  TapAvailable --> Settled: public install and landing-page readback
  HostedCandidate --> Blocked: build or provenance failure
  AcceptedCandidate --> Blocked: missing native receipt
  CompleteDraft --> Blocked: asset mismatch or partial upload
  Published --> Deprecated: defect discovered
  Deprecated --> HostedCandidate: new version, never replace old bytes
```

### Output structure and public-content specification

Planned additions include `.github/workflows/{ci,release}.yml`, issue/PR templates,
`.goreleaser.yaml`, `scripts/tool-versions.json`, `scripts/docgen/`, release/check
helpers and their tests, `docs/install.md`, `docs/releasing.md`, `docs/man/`,
`docs/completions/`, `docs/assets/`, `CONTRIBUTING.md`, `SECURITY.md`, `LICENSE`
and `THIRD_PARTY_NOTICES.md`. These paths do not yet exist. Build outputs stay in
an ignored, owner-selected `dist/`; durable receipts belong under
`docs/verification-evidence/006/` only during execution.

README quick start uses `tusk --version`, `tusk add 'Plan the release' --priority
high --due tomorrow`, `tusk list --all`, `tusk tree`, `tusk list --all --json` and
`tusk tui`. An ID-dependent example explicitly says to copy the complete ID from
the add/list result. Explain `?`, `a`, `e`, `q`, the 80×24 minimum and the linked
TUI guide. Optional `jq` examples state that prerequisite. Deletion is documented
in the guide with independent recursion/force consent and no undo; never use a
destructive command as a first-run demonstration.

The install guide has separate Unix and PowerShell examples, actual release
asset names, archive checksum/provenance verification, user-writable install
directories, PATH verification, completions/man installation and precise platform
support. Source instructions keep the local patched directories and distinguish
the minimum compiler from the release compiler. Explain nonempty `TUSK_DB_PATH`,
absolute `XDG_DATA_HOME`, then the home fallback on every OS; Windows does not
silently switch to AppData. Backup guidance stops all owners, preserves the
closed DB and any remaining WAL/SHM together, restores to an isolated directory,
checks integrity and compares tasks/events before use. Never copy only a live DB.

---

## Implementation Units

Every unit starts with an observed failing test or negative fixture before
implementation. Documentation and metadata units use missing/broken-content or
readback checks, rather than manufacturing application unit tests. Each unit
owns its named files, adds only the needed Make targets, synchronizes the triplet
and masterplan and passes canonical validation before its local commit. Planned
target names below are defined in the verification plan; none is runnable yet.

### U1. Native CI matrix and portable canonical tooling

**Goal / requirements:** R1–R5, R30; product U21. Make canonical checks executable on native targets.

**Dependencies:** Features 004/005 accepted; this planning pack and a new implementation instruction.

**Files / ownership:** `.github/workflows/ci.yml`, `scripts/tool-versions.json`, `scripts/ci-check.sh`, `scripts/test/test_ci.sh`, necessary existing setup/fmt/coverage/sqlc/script-test changes, `.gitattributes`, `Makefile`.

**Approach:** Apply KTD1–KTD3/KTD10. Add workflow lint/contract and prerequisite checks; preserve generator/runtime pins. Fix only reproduced BSD/GNU/MSYS/path incompatibilities. CI reports all native jobs and a minimum-Go job; generated checks run through existing targets and the tracked diff is checked afterward.

**Red-first tests:** Missing matrix/race gate, floating Action ref, failed tool download/digest, absent Make/compiler and stale generated output each fail the intended fixture. A spaced/Unicode checkout and LF/CRLF fixture exposes unsafe path parsing before modifying formatting scripts.

**Verification:** V01–V16; planned `make test-ci check-ci`, existing `make validate build check-generated`; hosted U1 closure requires exact-SHA job URLs. If publication is not authorized, record local U1 completion and keep hosted acceptance open.

**Failure / recovery:** Bootstrap failure explains the missing dependency; no continuing green job or global PATH/config modification. Preserve logs and repair bounded portability defects before retry.

**Reviews:** Architecture, portability, correctness, supply-chain security, testing, simplicity.

### U3. Storage-free completions and deterministic manuals

**Goal / requirements:** R13–R16, R30; product U23 completion boundary.

**Dependencies:** U1; generation must exist before release payload assembly.

**Files / ownership:** `internal/cli/root.go`, `internal/cli/completion.go`, `internal/cli/completion_test.go`, `internal/cli/root_test.go`, `scripts/docgen/main.go`, `scripts/docgen/main_test.go`, `scripts/test/test_completions.sh`, `docs/completions/`, `docs/man/`, `Makefile`, relevant `go.mod/go.sum` changes only if Cobra doc introduces a required import.

**Approach:** KTD4 and static R15 suggestions. Preserve existing invocation/error/output handling and share construction with the generator. Register exactly three supported shells; no ID/tag service completion or application startup hook. Generate all command manuals and parse/load real scripts with disposable shell configuration.

**Red-first tests:** Missing completion command, malformed/extra shell arguments, terminal/service/config callback invocation, short/failed output writer and wall-clock/manual drift. Bash/Zsh/Fish load fixtures assert representative subcommand/flag/enum suggestions without creating storage.

**Verification:** V17–V30; planned `make test-completions generate-docs check-docs`, existing focused CLI/full/canonical gates. U3 real-shell inspection uses owned Kitty on the applicable host.

**Failure / recovery:** Invalid usage exits 2 with empty stdout; filesystem/output failure exits 1 without partial accepted generated output. Missing shell is a pending acceptance result, not a skipped pass.

**Reviews:** CLI contract, lazy startup/performance, documentation generation, shell portability, testing.

### U4. README, installation guides, community files and license readiness

**Goal / requirements:** R12, R17–R22, R30; public-content draft and redistribution prerequisites.

**Dependencies:** U3; owner license decision gates license addition and U2 distributable payload closure.

**Files / ownership:** `README.md`, `docs/install.md`, `docs/releasing.md`, `docs/cli.md`, `docs/tui.md`, `docs/assets/`, `CONTRIBUTING.md`, `SECURITY.md`, `LICENSE`, `THIRD_PARTY_NOTICES.md`, `.github/ISSUE_TEMPLATE/`, `.github/pull_request_template.md`, `scripts/test/test_docs.sh`, documentation/notice check helper and `Makefile`.

**Approach:** KTD12–KTD15. Produce a source-install-first README until releases exist. Draft the precise release-install sections and metadata values in the release guide for U8 activation. Inventory third-party code/assets; confirm owner license before writing the selected first-party grant. Reuse safe existing app evidence with accurate provenance or capture it during applicable later app checks. Test links/examples, not marketing adjectives.

**Red-first checks:** Missing install platform/source prerequisite, nonexistent release/badge, unsupported versioned go-install recommendation, broken relative link, missing notice and invalid date grammar fail the documentation fixtures. Replay backup instructions on a closed temporary data fixture.

**Verification:** V31–V44; planned `make test-docs check-notices`, existing canonical validation and human documentation-usability review. Hosted metadata is previewed here and applied/read back in U8.

**Failure / recovery:** Unknown license/contact/destination remains an owned gate; do not invent values. Broken install route stays labeled unavailable. Revert only owned documentation/content changes if readback contradicts the proposed behavior.

**Reviews:** Product/scope, documentation usability, licensing/privacy boundaries, data safety, simplicity.

### U2. Reproducible CGO-free payloads and immutable version metadata

**Goal / requirements:** R5–R11, R21, R30; product U22 packaging boundary.

**Dependencies:** U4 accepted license/notices and U3 generated docs; U1 tool pins.

**Files / ownership:** `.goreleaser.yaml`, `.gitignore`, `cmd/tusk/main.go`, `cmd/tusk/version.go`, `cmd/tusk/main_test.go`, `scripts/release.sh`, `scripts/release_check.sh`, `scripts/test/test_release.sh`, `Makefile`, `docs/releasing.md`.

**Approach:** KTD5–KTD9. Add separate check/snapshot/candidate modes and an owned temporary checkout/overlay. Package all five targets from the full main package and include local replacements/source. Check payload membership, architecture/build metadata, version/hash/notice consistency and deterministic binary rebuilds; archive determinism uses the same version, source time and tool pins. No network publishing credentials are needed locally.

**Red-first tests:** Constant/tag mismatch, malformed version injection, default `-X` metadata, missing target/tzdata/patched source/notice, wrong hash, unsafe archive member and failed generation/download. Snapshot and candidate fixtures trap attempted remote publication or changes to the original checkout.

**Verification:** V45–V62; planned `make test-release release-check release-snapshot release-candidate`; existing canonical/generated/minimum-Go gates. Local native smoke uses the extracted local payload, while hosted exact-byte acceptance is U5.

**Failure / recovery:** Refuse output collisions unless the path is an explicitly owned throwaway run directory; retain failed receipts outside cleanup. No cleanup glob can remove user databases or unrelated outputs. Changed bytes invalidate the prior manifest and acceptance.

**Reviews:** Build/API contract, reproducibility, portability, data lifecycle, supply chain, reliability.

### U6. Trusted hosted candidate workflow and artifact provenance

**Goal / requirements:** R9, R10, R26, R27, R30; generate the bytes native acceptance will inspect.

**Dependencies:** U2; hosted operation requires authorized publication of the workflow/code and dispatch.

**Files / ownership:** `.github/workflows/release.yml`, candidate/promotion contracts in `scripts/release.sh`, `scripts/test/test_release.sh`, `Makefile`, `docs/releasing.md`.

**Approach:** Manual default-branch candidate dispatch takes an exact trusted SHA/version, calls canonical packaging, records manifest/workflow/run identity, attests payloads/manifest and uploads bounded artifacts with 90-day retention. Publication jobs are separate and disabled by default; U8 consumes the accepted run. Reusable CI gates run on the candidate SHA before packaging. Build and provenance inputs include the immutable version overlay and tool lock.

**Red-first tests:** PR/fork artifact promotion, different SHA/version/run, tampered checksum, missing attestation and release/tap secret access fail before remote writes. Fake GitHub API boundaries expose canceled upload and lost responses without public publication.

**Verification:** V63–V72; planned `make test-release check-ci verify-candidate`; hosted workflow URL, artifact ID/digest and verified provenance are required and unexecuted here.

**Failure / recovery:** An incomplete/canceled candidate is never accepted; rerun as a new run ID with new hashes. Artifact expiry blocks promotion until a new candidate receives acceptance. No trusting an artifact based only on its filename.

**Reviews:** Security/authz, deployment/reliability, artifact contract, testing/evidence, scope.

### U5. Exact-artifact native lifecycle, terminal and performance acceptance

**Goal / requirements:** R7, R11, R12, R18, R19, R23–R25, R30; inherited native handoffs.

**Dependencies:** U6 hosted candidate and exact artifact download/verification.

**Files / ownership:** `scripts/release_smoke.sh`, `scripts/release_smoke.ps1`, `scripts/test/test_release.sh`, `scripts/cli-bench/main.go` and tests only if needed for supplied-binary verification, candidate startup selectors/tests in `cmd/tusk/`, `Makefile`, `docs/install.md`, `docs/releasing.md`, execution receipts under `docs/verification-evidence/006/`. Any production fix requires a new candidate and affected prior gates.

**Approach:** Execute packaged bytes on all KTD2 targets; Linux/macOS owned Kitty and Windows native console checks satisfy their separate tiers. Test untouched/space/Unicode homes, explicit paths, embedded zones, first data creation, JSON/pipe/signals, WAL ownership and TUI lifecycle. Compare normalized tasks/events around binary replacement and removal; newer-schema refusal uses a fixture and file hashes. Replay consistent backup/restore. Run measured CLI/TUI gates sequentially on the declared reference host with actual candidate binaries.

**Red-first tests:** Existing-data recreation, wrong architecture, destructive uninstall, newer-schema mutation and lost terminal restoration are failing lifecycle/smoke fixtures before repair. A benchmark fixture first fails when the existing build prerequisite overwrites its supplied binary. Native runtime failures become bounded review-fix units; historical cross-build passes cannot close them.

**Verification:** V73–V86; planned `make test-release-smoke release-smoke`, existing canonical and benchmark gates with unique retained outputs; native receipts include exact archive/executable SHA256 and terminal/OS versions.

**Failure / recovery:** Preserve DB/WAL/SHM on failure; read back after uncertain mutations. Missing native terminal access or reference latency failure keeps the respective release gate pending. Do not substitute synthetic PTY or hosted console logs for visual app acceptance.

**Reviews:** Data integrity, terminal/concurrency lifecycle, portability, performance, documentation usability.

### U7. Homebrew cask candidate and destination readiness

**Goal / requirements:** R11, R17, R21, R29, R30; tested alternate macOS installation.

**Dependencies:** U5; owner-controlled tap location/access verified before hosted tap changes.

**Files / ownership:** `.goreleaser.yaml` cask section, release helpers/tests, `docs/install.md`, `docs/releasing.md`, `Makefile`; eventual `Casks/tusk.rb` in the separately owned tap, never assumed part of this checkout's commit.

**Approach:** KTD11. Generate the cask without upload, audit syntax/checksums/architecture/manual/completion directives, test local candidate installation and removal on native Intel/ARM macOS. Temporary local asset URLs used for prepublication tests are labeled fixture evidence; U8 proves final public URLs and live tap install. Do not create a cross-repository token by default.

**Red-first tests:** Wrong arch/hash, absent asset, invalid cask syntax, accidental tap upload and database-deleting uninstall fail before the cask is enabled. Missing tap destination produces a named prerequisite failure.

**Verification:** V87–V94; planned `make test-homebrew check-homebrew`, native Homebrew lint/install/version/completion/man/removal and data-preservation records.

**Failure / recovery:** Tap unavailability cannot masquerade as a successful brew route. Preserve working direct downloads; retry only tap publication when the accepted release/cask bytes remain unchanged.

**Reviews:** Distribution/operations, data lifecycle, supply chain, macOS portability, documentation.

### U8. Authorized release, public metadata and final settlement

**Goal / requirements:** R17–R22, R27–R30; product U22/U23 release closure and requested GitHub landing page.

**Dependencies:** U7 and every native/hosted/performance gate; explicit release/tag/SHA and hosted-setting authorization.

**Files / ownership:** Release-promotion helper/tests, `README.md`, `docs/install.md`, `docs/releasing.md`, approved public assets, GitHub repository metadata/social preview, separate tap publication and synchronized product/feature governance artifacts.

**Approach:** Compare current SHA and accepted manifest with terminal/hosted receipts. Create a complete draft with existing accepted artifacts, re-download and verify every asset/version/provenance, then publish the authorized tag. Publish the reviewed cask and prove anonymous direct-download/live-tap installs. Activate README release links/badges/platform claims; apply reviewed About/topics/homepage/social preview and verify rendered GitHub content at the published documentation SHA. Documentation-only changes after the binary SHA retain both identities and must not claim identical commits.

**Final source freeze:** Finish and locally validate/commit all release helper,
test-driver, benchmark-selector and cask configuration changes before selecting
the publishable source SHA. If later units changed those inputs after the first
U6 candidate, regenerate through the existing candidate workflow and rerun affected
U5/U7 gates against its exact bytes. These are acceptance reruns, not new units.
Only documentation/receipt changes may carry a later separate SHA without a new
binary candidate; any changed build input invalidates acceptance. Hosted gate
closure may follow a validated local engineering checkpoint, but a pending
external result cannot become a completed release unit or Phase 6 checkbox.

**Red-first checks:** Missing receipt, tampered asset, conflicting existing tag, stale source acceptance, expired candidate artifact and unsafe retry refuse publication. Repository readback fixtures expose incomplete metadata, bad links or nonexistent badges. A failed release/tap/metadata response is reconciled through readback before retry.

**Verification:** V95–V102; planned `make test-release test-docs verify-candidate release-draft release-publish`, existing fresh `make validate`; final browser/manual/API evidence and anonymous installs required. Draft/publish commands are intentionally separate from snapshot/check targets.

**Failure / recovery:** Before publication, preserve incomplete draft state for targeted repair; no deletion/replacement of published assets or retagging. After publication, deprecate defects and prepare a new version. Partial tap/metadata success is recorded and retried independently; Phase 6 remains open until required readbacks pass.

**Reviews:** Adversarial release audit, security, deployment/data integrity, documentation/product, evidence/coherence.

---

## Verification Contract and Definition of Done

The [verification plan](../verification-plans/2026-09-06-006-feat-automation-packaging-and-release-verification-plan.md)
maps all 30 requirements to 102 unexecuted scenarios, exact Make interfaces,
fixtures and evidence tiers. The [workorder](../workorders/2026-09-06-006-feat-automation-packaging-and-release-issues-workorder.md)
owns planning corrections and external decisions. Code/config/documentation
tests, generated-diff checks, canonical ≥95% coverage/race validation and focused
review precede each unit commit. Native tests run release payloads, not a locally
rebuilt look-alike. A receipt records UTC time, source and documentation SHAs,
binary/archive/manifest hashes, host/tool/terminal identity, exact command or
observed actions, results, failed samples and links to retained evidence.

Local engineering is complete when U1/U3/U4/U2 and all locally executable portions
of U6/U5/U7/U8 pass and are separately committed. Hosted/native/release acceptance
stays open for missing external evidence. Phase 6 is complete only when all 102
scenarios have applicable execution evidence, every required platform/terminal
and installed artifact is accepted, license/destination decisions are closed,
the authorized release and cask are publicly installable, metadata/README readback
passes and the synchronized product G4/master checklist are actually closed.
No planning checkbox closes a release scenario.

Publication authorization is a final action against a concrete candidate, not
permission inferred from planning or a local commit. GitHub settings changes use
the reviewed payload and current-state comparison; setting availability or an
essential unknown becomes a workorder gate, not an invented field value.

---

## Open Decisions, Dependencies and Risks

- **License/rights:** Owner chooses license and confirms redistribution authority before U4 license closure/U2 distribution. MIT is a proposal, not an applied grant; 006-ISS-001.
- **Native access:** Release maintainer arranges all five native targets and required terminal sessions before U5; blocked/absent results remain pending; 006-ISS-002.
- **Tap ownership:** Owner establishes or supplies the accessible tap before U7. Proposed destination was not found during inspection; 006-ISS-003.
- **Release identity/authority:** Owner confirms actual version, SHA and hosted actions against U8's completed candidate. Proposed v0.3.0 is not a pushed tag; 006-ISS-004.
- **Security contact/settings:** Maintainer verifies private vulnerability reporting and the supported contact route before U4 public claims; 006-ISS-005.
- **Runtime/platform control:** OS unsigned-download restrictions, changing runner availability and analyzer/compiler support are execution risks with failing/pending gates. No broad OS support, authenticated draft install or cross-build substitutes for the named evidence.
- **Artifact/source drift:** A source/tool/version/payload change invalidates affected native/performance/provenance receipts. A documentation-only publication records its separate SHA and preserves executable identity.

---

## References and Planning Review

Repository sources: `go.mod`, `cmd/tusk/main.go`, `internal/cli/root.go`,
`Makefile`, `scripts/sqlc.sh`, `README.md`, `docs/cli.md`, `docs/tui.md`,
the Feature 002–005 triplets and
[owned-Kitty learning](../solutions/workflow-issues/verify-owned-kitty-window-without-focus-or-environment-leaks.md).
No runtime gate was executed in this planning pass.

External guidance, checked 2026-10-01:

- [Go module install restrictions](https://go.dev/ref/mod#go-install) support the source-checkout route while local replacements remain.
- [Go build overlays](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies) and [GoReleaser Go build configuration](https://goreleaser.com/customization/builds/builders/go/) support immutable version generation without linker-injected globals.
- [Go downloads](https://go.dev/dl/), [GoReleaser v2.18.2](https://github.com/goreleaser/goreleaser/releases/tag/v2.18.2) and [actionlint v1.7.12](https://github.com/rhysd/actionlint/releases/tag/v1.7.12) supply the observed tool versions, not proof of this project's compatibility.
- [GitHub runner matrix](https://docs.github.com/en/actions/reference/runners/github-hosted-runners) supports the selected native runner labels.
- [GitHub secure workflow use](https://docs.github.com/en/actions/reference/security/secure-use) supports SHA pins, least privilege and trusted artifact promotion.
- [GoReleaser deprecations](https://goreleaser.com/resources/deprecations/) and [Homebrew casks](https://goreleaser.com/customization/publish/homebrew_casks/) supersede binary `brews` formula guidance; [Cask Cookbook](https://docs.brew.sh/Cask-Cookbook) governs the installed files/lifecycle.
- [GitHub artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations) and [immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases) support verified provenance and complete-draft promotion.
- [GitHub README guidance](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-readmes) and [repository topics](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/classifying-your-repository-with-topics) support the public presentation decisions.

Planning review is sequential in the main thread under the repository tool map.
Architecture/dependency, product/scope, coherence/feasibility, correctness/recovery,
test/evidence, security/supply-chain, data lifecycle, performance, portability,
documentation usability and simplicity lenses inspect the triplet together.
Their final dispositions live in the workorder; no independent reviewer or
executed application acceptance is claimed.

### 1. GitHub Actions CI Pipeline (`.github/workflows/ci.yml`)
- Triggers on push and pull requests to `main`.
- Matrix: Ubuntu Latest, macOS Latest, Windows Latest.
- Steps:
  - Checkout repository.
  - Setup Go 1.24+.
  - Run `make validate` (`fmt`, `vet`, `test`, `race`).
  - Run binary build test.

### 2. Multi-Platform Packaging with GoReleaser (`.goreleaser.yaml`)
- Builds:
  - Linux (`amd64`, `arm64`)
  - Darwin (`amd64`, `arm64`)
  - Windows (`amd64`)
- CGO-free configuration guarantees static binaries without dynamic C library dependencies.
- Generates SHA256 checksums and Homebrew formula.

### 3. Shell Completions & Documentation
- Cobra completion commands: `tusk completion bash`, `tusk completion zsh`, `tusk completion fish`.
- Automated man page generation via `cobra/doc`.

---

## Verification Scenarios

1. Local dry-run of GoReleaser (`goreleaser check` and `goreleaser build --snapshot --clean`).
2. Verification that generated shell completion scripts load without syntax errors in Bash and Zsh.
3. CI workflow linter check validating GitHub Actions YAML syntax.
