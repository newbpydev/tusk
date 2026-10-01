---
feature-id: "006"
plan-source: docs/plans/2026-09-06-006-feat-automation-packaging-and-release-plan.md
surface-profiles: [cli-tui, infrastructure-operations, installation-data-lifecycle, documentation]
status: Implementation active - U1/U3 locally accepted; U4 next
evidence-scope: Five local U1 scenario closures; hosted/native release gates pending
deepened: 2026-10-01
---

# Feature 006 Verification Plan

## Verification contract

Verify a complete release and public repository landing page, from canonical
native CI through package generation, exact-byte native acceptance, authorized
publication and a user's installation/first-task workflow. The
[plan](../plans/2026-09-06-006-feat-automation-packaging-and-release-plan.md) decides
behavior; the [workorder](../workorders/2026-09-06-006-feat-automation-packaging-and-release-issues-workorder.md)
records defects and external gates. [MASTERPLAN.md](../../MASTERPLAN.md) alone
activates units and records completion.

At implementation intake all scenarios below are unexecuted. Planning inspected source, local/remote SHA,
GitHub settings/workflows/releases and official tool documentation. No build,
application test, canonical validation, benchmark, native/Kitty check or hosted
mutation occurred. Historical Feature 002–005 receipts remain labeled with their
original candidate identities.

Evidence tiers are distinct:

- **Focused:** Red/green command, generator, script/configuration and negative-fixture checks.
- **Aggregate:** Canonical `make validate`, build, generated/module drift and minimum-compiler proof.
- **Local packaging:** Complete archives, checksums, immutable version input, reproducibility and snapshot/candidate isolation.
- **Hosted:** Trusted exact-SHA native matrix, candidate run/artifact identity and provenance verification.
- **Native/manual:** Actual downloaded binaries, platform storage/consoles, owned Kitty/native Windows Terminal app inspection and shell/Homebrew use.
- **Publication:** Authorized draft/tag/release/tap/metadata, anonymous downloads, GitHub rendered-content readback and final evidence settlement.

Passing one tier does not close another. Missing hosts, shells, credentials or
desktop access leave their scenario pending. Routine code fixes require new
red/green evidence and a new candidate; documentation-only changes record a
separate documentation SHA without relabeling executable identity.

---

## Requirement coverage

Every unprefixed R/U/V refers to Feature 006. Stable scenario IDs are `006-V01`
through `006-V102`. Requirements R1–R30 come from the plan; all scenarios have an
owning unit in the sections below.

| Requirement | Owning units | Scenarios | Evidence tier |
| --- | --- | --- | --- |
| R1 CI triggers/permissions | U1 | V01, V02, V12, V15 | Focused/hosted |
| R2 Native/minimum compiler | U1, U5 | V03, V04, V16, V73, V74 | Aggregate/hosted/native |
| R3 Windows prerequisites | U1, U5 | V05, V06, V75, V77, V82 | Focused/hosted/native |
| R4 Canonical drift/coverage | U1 | V07–V11, V16 | Focused/aggregate/hosted |
| R5 Tool/runtime pins | U1, U2 | V12–V14, V48, V49, V60 | Focused/packaging |
| R6 Immutable version | U2 | V45–V47, V58, V59 | Focused/packaged binary |
| R7 Five CGO-free targets | U2, U5 | V48, V49, V54, V73–V76 | Packaging/native |
| R8 Complete payloads | U2 | V50–V54, V61 | Packaging/negative fixtures |
| R9 Manifest/provenance | U2, U6 | V55, V58, V59, V65–V68, V72 | Packaging/hosted |
| R10 Nonpublishing local modes | U2, U6 | V56, V57, V62, V63, V69–V71 | Focused/hosted |
| R11 Data-preserving lifecycle | U2, U5, U7 | V57, V79–V81, V91, V92 | Focused/native |
| R12 Backup/recovery docs | U4, U5 | V39, V81 | Documentation/native disk |
| R13 Storage-free generators | U3 | V17, V18, V24, V29 | Focused/process |
| R14 Completion stream/exit contract | U3 | V17–V20, V23 | Focused/shell |
| R15 Static completion | U3 | V21, V22 | Focused/shell |
| R16 Deterministic generated docs | U3 | V24–V30 | Generator/drift/manual |
| R17 README/install usability | U4, U7, U8 | V31–V34, V40, V44, V94, V98, V99 | Documentation/native/publication |
| R18 Runnable examples | U4, U5 | V35–V39, V75, V76, V84 | Focused/native |
| R19 Evidence-backed claims/assets | U4, U5, U8 | V40, V41, V82–V86, V99, V102 | Documentation/manual/performance |
| R20 GitHub metadata/rendering | U4, U8 | V42, V43, V100, V101 | Preview/API/browser |
| R21 License/redistribution | U4, U2, U7 | V44, V51, V61, V93 | Owner/license/payload review |
| R22 Contributor/security guidance | U4, U8 | V42–V44, V100 | Documentation/hosted readback |
| R23 Native storage/process behavior | U5 | V73–V81 | Native runtime/disk/process |
| R24 Native TUI/terminal restoration | U5 | V82, V83 | Native/manual |
| R25 Candidate performance | U5 | V85, V86 | Declared-host retained measurements |
| R26 Trusted hosted candidate | U6 | V63–V66, V69–V72 | Focused/hosted |
| R27 Privileged artifact boundary | U6, U8 | V64–V70, V95, V97 | Security/hosted/API fixtures |
| R28 Authorized immutable promotion | U8 | V95–V98, V102 | Owner/hosted/publication |
| R29 Homebrew route | U7, U8 | V87–V94, V98 | Native Homebrew/publication |
| R30 Per-unit closure/governance | All | V16, V30, V44, V62, V72, V86, V94, V102 | Aggregate/review/governance |

### Inherited product and feature handoffs

| Existing obligation | Feature 006 scenarios | Closure boundary |
| --- | --- | --- |
| Product TUSK-V67 native canonical gates | V01–V16 | Candidate-SHA hosted native jobs, not local lint |
| Product TUSK-V68 target artifact execution | V63–V76 | Exact packaged hashes executed on all five targets |
| Product TUSK-V69 complete release manifest | V45–V62 | Payload/version/checksum/notice proofs |
| Product TUSK-V70 upgrade/downgrade/removal | V79–V81, V91, V92 | Real disk and platform installation data preservation |
| Product TUSK-V71 distribution/readiness | V87–V102 | Owner decisions, native install, authorized release/tap |
| Product TUSK-V72 completion/manuals | V17–V30 | Storage-free generation and three real shells |
| Product TUSK-V73 documentation/handoff | V31–V44, V84, V99–V102 | Replayed examples, rendered readback and final synchronized receipts |
| Remaining platform coverage of TUSK-V66 | V82, V83 | Exact packaged TUI and native terminal inspection |
| Feature 002 V33 native Windows storage | V75, V78–V81 | Native path/permissions/reopen/WAL tests, no WSL substitute |
| Feature 003 remaining native/date/consumer proof | V74–V78, V82–V84 | Embedded-zone, consumer/runtime and real-app behavior |
| Feature 004 V90–V91 native/hosted release | V01–V16, V45–V102 | Exact CLI/native/packaging/publication receipts |
| Feature 005 V111–V112 native/hosted release | V63–V102 | Native TUI, licenses/install/publication; earlier local acceptance retained |

Handoff links preserve the older scenario IDs and unchecked release gates.
Executing a Feature 006 scenario closes an inherited item only when the whole
inherited contract has matching evidence; partial platform evidence is recorded
without checking the umbrella item.

---

## Scenarios

### U1 — Native CI and portable canonical tooling

- [ ] 006-V01 **Triggers:** PR to main, main push and explicit dispatch select CI; unrelated tag/branch events do not publish anything.
- [ ] 006-V02 **Fork/permissions:** Read-only fork PR executes quality checks with no release/tap secrets or write permission; no privileged PR-code checkout.
- [ ] 006-V03 **Five native jobs:** Each pinned runner reports the expected OS/arch and release compiler, runs setup/validate/generated check/build, and emits current-SHA job URLs; a missing job fails acceptance.
- [x] 006-V04 **Minimum compiler:** Go 1.25.0 full canonical checks run on Linux and existing `build-tui` compiles application/tests for all five targets; release builds use the separate pinned compiler.
- [ ] 006-V05 **Windows prerequisite normal/negative:** Git Bash, GNU Make and native GCC match the Windows Go environment; missing compiler/Make or wrong native architecture fails with actionable diagnostics before a false green race job.
- [ ] 006-V06 **Checkout portability:** Space/Unicode paths and LF/CRLF fixtures preserve Go sources and script behavior on GNU/BSD/MSYS tooling; reproduce any incompatibility before changing scripts.
- [x] 006-V07 **Formatting drift:** A deliberately unformatted tracked fixture fails the post-format source-diff gate; CI cannot quietly repair and pass.
- [ ] 006-V08 **Generator drift:** Wrong sqlc pin, changed queries/schema, missing/extra generated file and fake generator failure fail their appropriate canonical checks without replacing accepted generated code.
- [ ] 006-V09 **Module drift:** Dependency metadata changes fail `check-modules`; local replace trees are present and third-party source remains unchanged after checks.
- [ ] 006-V10 **Coverage/race:** Missing tests, nonexempt coverage below 95% and a race-fixture failure propagate; no new coverage exemption masks a release helper/package.
- [x] 006-V11 **Gate propagation:** Failed fmt/vet/test/race/coverage/script child commands cannot yield a successful aggregate workflow result; logs name the failed gate.
- [x] 006-V12 **Immutable action/tool pins:** Missing/full-SHA mismatch/floating action ref and missing version/digest fail configuration checks; pins are verified against the upstream repository/release asset.
- [x] 006-V13 **Tool setup recovery:** Network failure, truncated archive, wrong hash or missing tool cannot execute unverified bytes or overwrite the last accepted tool; retry uses an owned cache path.
- [ ] 006-V14 **Fresh source:** Fresh checkout and tagged-source fixtures retain Bubble Tea/Glamour replacement trees and gallery bytes without requiring unavailable LFS hydration; build input inventory detects absent patch files.
- [ ] 006-V15 **Cancellation/cache/log hygiene:** Superseded PR runs can cancel; the matrix still reports independent failures. No executable promotion cache, secret/environment dump, stale job or skipped platform is treated as current evidence.
- [ ] 006-V16 **U1 closure:** Fresh canonical/generator/native/minimum checks and sequential portability/security review bind the unit SHA and tool lock; synchronize governance before the U1 commit while missing hosted execution remains pending.

U1 [local receipt](../verification-evidence/006/u1.json) closes the five focused/aggregate scenarios checked above. Other U1 scenarios retain their local partial evidence there while required hosted/native portions remain pending. No native Windows/macOS execution or workflow dispatch is claimed.

### U3 — Static shell completions and manuals

- [x] 006-V17 **Three scripts:** Bash/Zsh/Fish commands each write only the intended script to stdout and return 0 with redirected streams and absent terminal callbacks.
- [x] 006-V18 **No initialization:** Broken path, invalid timezone/parent environment, missing home and rejecting service/TUI/confirmation factories do not affect generation or create DB/data directories.
- [x] 006-V19 **Usage:** Missing/unknown shell, extra argument and unsupported flag return 2, leave stdout empty and write one safe stderr diagnostic.
- [x] 006-V20 **Output failure:** Short writer/broken pipe return 1; no recursive diagnostic, database open or accepted truncated file.
- [x] 006-V21 **Static suggestions:** Representative subcommands, flags and status/priority enums are offered; IDs, user tags, titles and notes are never fetched. User filesystem filename completion, if Cobra emits it, does not open task storage.
- [x] 006-V22 **Fresh tree/repetition:** Repeated/concurrent construction for different invocations does not share mutable commands/options; generation leaves existing root/help/version/JSON behavior unchanged.
- [x] 006-V23 **Real shell load/use:** `bash -n`, `zsh -n` and `fish -n` pass; disposable real sessions source the matching script and obtain representative completions without shell-startup-file changes.
- [x] 006-V24 **Manual generation purity:** Generator obtains the real CLI tree with storage/terminal factories unavailable and emits all registered command manuals, including completion and TUI help.
- [x] 006-V25 **Determinism:** Two generations with different wall-clock times and checkout locations produce the same normalized bytes; no current-date/autogenerated footer leak.
- [x] 006-V26 **Drift detection:** Changed command/flag help, missing output or extra stale manual/completion fails `check-docs`; rerunning generation repairs only owned outputs.
- [x] 006-V27 **Generation failure recovery:** Read-only destination or failing formatter leaves previous accepted outputs intact; no partial generation is represented as green.
- [x] 006-V28 **Generated module graph:** Any Cobra doc dependency is justified, pinned and covered by module checks; no database/TUI ownership or global-state regression enters the builder seam.
- [x] 006-V29 **Lazy startup non-regression:** Existing help/version/syntax isolation plus completion/manual paths run with inaccessible storage; new registrations do not make root construction perform terminal discovery or network/storage work.
- [x] 006-V30 **U3 closure:** Canonical tests, real shell records and owned Kitty inspection of completion/man usage match generated files and synchronized docs; no build/test results are displayed as Kitty app evidence.

U3 local acceptance is bound to [the receipt](../verification-evidence/006/u3.json),
including owned Kitty screenshots and three actual shell widgets. Windows ACL
fixtures are implemented but not natively executed here; foreign native/hosted
release obligations remain pending. The unit does not claim independent review.

### U4 — Public documentation, license and community readiness

- [ ] 006-V31 **README information order:** A visitor finds purpose, screenshot, available installation and first-task workflow before developer/planning information; source-only status is explicit before a release exists.
- [ ] 006-V32 **Source installation:** A fresh checkout keeps replacements, uses documented Go/Make/Bash prerequisites and builds the full command package; no unsupported versioned `go install` path is presented as usable.
- [ ] 006-V33 **Unix install draft:** Linux/macOS instructions map architecture to exact asset names, verify hashes before extraction, use user-writable paths and show PATH/version recovery for wrong-architecture/permission/download failures; public URLs remain marked pending until release.
- [ ] 006-V34 **Windows install draft:** Native PowerShell hash/extraction/PATH commands distinguish amd64 and use `tusk.exe`; no WSL or Bash-only command is represented as the Windows user route.
- [ ] 006-V35 **Quick start replay:** Version, add, list, tree, clean JSON and TUI examples use valid flags and full IDs; no destructive first-run example or hidden prerequisite.
- [ ] 006-V36 **Dates/status/JSON examples:** `tomorrow`, supported offsets, priority/status enums and JSON schemas match guides/source; invalid `next week` or an undeclared `jq` dependency fails docs review.
- [ ] 006-V37 **Shared configuration:** Explain actual DB/environment/timezone precedence, relative explicit paths, no invalid-config fallback, Windows home fallback, `NO_COLOR` and terminal requirements.
- [ ] 006-V38 **Safety/recovery language:** Explain independent force/recursion, no undo, metadata-only history and unknown/committed outcomes; readback precedes retry, and no recovery instruction deletes DB/WAL/SHM.
- [ ] 006-V39 **Backup rehearsal:** Stop all fixture owners, preserve closed DB plus remaining sidecars, restore to an isolated location, check integrity and compare tasks/events; prohibit live-DB-only copy and state NORMAL durability limits.
- [ ] 006-V40 **Claims and badges:** Every supported platform/install/performance claim has a matching dated receipt; nonexistent release/CI/license badge and universal latency claim fail review.
- [ ] 006-V41 **Screenshot/privacy:** Inspect real app image ownership/content, sanitized task data, date/source identity and alt text; social preview is readable and uses an app capture with recorded provenance, not a generated mockup.
- [ ] 006-V42 **Metadata/community preview:** Exact proposed description/homepage/topics/social preview and issue/PR/contributor/security guidance agree; optional website/funding/contact fields are not fabricated.
- [ ] 006-V43 **Links/rendering:** Relative guide/image/anchor links resolve; headings/code blocks/tables render legibly, badges have valid targets and the drafted public GitHub layout is inspected before activation.
- [ ] 006-V44 **U4/license closure:** Owner's selected first-party license/rights confirmation, replacement-aware third-party inventory and verified security contact support all public claims; documentation checks, canonical validation and sequential usability/scope/privacy review pass before the unit commit.

### U2 — Versioned packaging and local lifecycle safety

- [ ] 006-V45 **Immutable version:** All three version forms report the intended validated tag version in actual built binaries; ordinary source and snapshot labels remain clearly unreleased.
- [ ] 006-V46 **No mutable linker metadata:** Default `-X` injection or a mutable version global fails the build contract; generated constant/overlay input hash appears in the manifest and the original checkout stays clean.
- [ ] 006-V47 **Version/input rejection:** Empty/malformed/non-semver values, shell/code injection text, conflicting tag and unsafe overlay/output paths fail before build/publication; spaces in owned temporary paths remain supported.
- [ ] 006-V48 **Complete target matrix:** Exactly the five required OS/architecture payloads exist with matching executable architecture and no invented Windows arm64 target; omission/duplicate/wrong arch fails release-check.
- [ ] 006-V49 **Runtime identity:** Build metadata proves CGO disabled, embedded timezone data and selected patched graph; Linux has no dynamic interpreter dependency. Darwin/Windows OS-library use does not become an unsupported fully-static claim.
- [ ] 006-V50 **Payload contract:** Unix tar.gz/Windows zip have the right executable, names/modes, README, three completion scripts, all manuals and required notices; extraction exposes the root executable as documented.
- [ ] 006-V51 **Licenses:** Missing project license, missing replacement-tree/embedded-asset notice and unclassified dependency obligations fail distribution checks; notices retain upstream grant text.
- [ ] 006-V52 **Archive abuse:** Absolute paths, `..` traversal, unsafe links, unexpected database/sidecar/credential files and unexpected executable members fail manifest/member validation.
- [ ] 006-V53 **Source archive:** Complete tagged source includes both patched third-party modules and verified gallery bytes; documented checkout build retains the intended graph without an LFS service requirement.
- [ ] 006-V54 **Payload consumption:** A clean isolated native home can extract and run version/help without Go, compiler, external service or database initialization; verify archive and executable hashes separately.
- [ ] 006-V55 **Manifest/checksums:** Payload corruption, missing hash, wrong target/version/SHA/tool input or a self-hash dependency cycle fails verification; manifest/checksum files identify one complete candidate.
- [ ] 006-V56 **No remote side effects:** Both snapshot and local intended-tag candidate modes trap any release API/tag push/tap mutation; no publishing credentials are required or read.
- [ ] 006-V57 **Cleanup/data boundaries:** Failure/cancellation cleans only owned temporary build storage, keeps diagnostics and preserves user DB/sidecars/unrelated dirty files; colliding retained outputs are rejected.
- [ ] 006-V58 **Binary reproducibility:** Same source/version/compiler/overlay/tool pins built in two owned locations produce matching executable hashes; timestamp/path entropy is detected rather than excused.
- [ ] 006-V59 **Archive reproducibility:** Same payloads/source timestamp/tool pins reproduce archive/man/completion bytes; manifest contains stable inputs and separately records non-deterministic workflow identity where appropriate.
- [ ] 006-V60 **Failed prerequisites:** Bad tool digest, unavailable pin, failed documentation generation or failed child build produces nonzero release-check/candidate result; preserve existing accepted outputs and no fallback compiler.
- [ ] 006-V61 **Notice inventory scope:** Source and executable dependency/asset inventories include local replacement paths and patch provenance; stale inventory fails after a dependency/asset change.
- [ ] 006-V62 **U2 closure:** Canonical/generated/minimum checks, manifest/member negative fixtures and current native local smoke pass with synchronized review/receipts before the U2 commit; hosted/native release claims remain open.

### U6 — Hosted candidate and trusted provenance

- [ ] 006-V63 **Manual candidate isolation:** Authorized default-branch dispatch builds the requested exact main SHA/version and uploads a candidate without a public release/tag/tap write.
- [ ] 006-V64 **Trusted source/run:** Reject fork/PR source, untrusted workflow/artifact, nonexistent SHA, unauthorized branch/tag and a run identity that does not match the manifest.
- [ ] 006-V65 **Required native CI:** Candidate pipeline references successful canonical native and minimum jobs for the same source; a pending/skipped/failing/wrong-SHA check blocks accepted candidate production.
- [ ] 006-V66 **Build once:** Hosted packaging creates one versioned complete payload set; acceptance downloads that set. Local rebuilt bytes cannot be substituted under the same accepted identity.
- [ ] 006-V67 **Provenance verification:** Verify attestations for archives/source and the manifest against the trusted repository/workflow/ref/SHA; archive hash and overlay/tool inputs agree with the manifest.
- [ ] 006-V68 **Tamper/mismatch:** Altered artifact, checksum, manifest/version or wrong workflow provenance is rejected before native install/promotion; no trusting filenames or self-reported metadata alone.
- [ ] 006-V69 **Credential separation:** Candidate jobs lack release/tap secrets and content-write permission; only provenance needs OIDC/attestation write, with no credential persistence or private task/environment leaks.
- [ ] 006-V70 **Partial failure/readback:** Canceled/timed-out artifact/attestation upload cannot leave a successful candidate; fake lost-response API tests reconcile identity/state before retry.
- [ ] 006-V71 **Retention/rerun:** Expired/incomplete artifacts block promotion; a new run has new identity and receives new native acceptance even if source is unchanged.
- [ ] 006-V72 **U6 closure:** Hosted job/artifact URLs, artifact ID/digest, verified attestations and local canonical/security review bind exact source/version/hash evidence; unexecuted hosted work remains pending until actually dispatched.

### U5 — Native packaged-app, lifecycle and performance acceptance

- [ ] 006-V73 **Native bytes on five targets:** Download, verify, extract and execute each exact packaged artifact on its native OS/arch; record archive/executable SHA256, OS and tool identity. Cross-build existence is insufficient.
- [ ] 006-V74 **Zones/minimum platform:** Named zones work without external tzdata; run leap/month/DST/service fixtures on the native release compiler/OS; support table lists only inspected baselines.
- [ ] 006-V75 **Paths/Windows storage handoff:** First-use/explicit/relative/space/Unicode paths, home/XDG/env precedence, nonregular/read-only/corrupt/newer-schema paths and reopening are exercised natively; failures preserve originals and sidecars.
- [ ] 006-V76 **JSON/streams:** Help/version are storage-free; mutation/query examples preserve one clean JSON value and expected exit codes; closed/slow stdout and safe diagnostics cause no duplicate mutation or mixed human JSON.
- [ ] 006-V77 **Signals/native consoles:** Graceful cancellation and supported forced termination/Windows console events restore or report outcome correctly; no Unix-only signal test is labeled Windows proof.
- [ ] 006-V78 **WAL/concurrency:** Independent native owners read/write real disk fixtures, preserve coherent hierarchy/history and produce bounded lock/conflict behavior without lost rollup.
- [ ] 006-V79 **Replace/upgrade:** Existing accepted source-version data reopens with candidate binary, tasks/events/config paths remain stable, and no schema/data recreation occurs; unsupported legacy PostgreSQL import is excluded.
- [ ] 006-V80 **Newer schema/downgrade/remove:** Fixture newer schema is refused without changing DB/sidecar bytes; binary replacement/removal leaves data intact and restoration of the supported binary can reopen it.
- [ ] 006-V81 **Backup/recovery:** Replay closed-owner backup/isolated restore/integrity and old/new unknown-outcome readback; compare normalized tasks/events and preserve sidecars until successful verification.
- [ ] 006-V82 **Native TUI workflow:** Exact packaged binary opens in owned Kitty on Linux/macOS and owned Windows Terminal on Windows 11; inspect create/edit/move/toggle/filter/search/history/delete/recovery and external CLI readback with isolated data.
- [ ] 006-V83 **Terminal lifecycle:** Inspect 80×24/120×40/200×60, shrink/restore, Unicode, NO_COLOR, paste/focus, q/Ctrl+C/startup failure and relaunch; preserve drafts and restore shell modes/cursor/alternate screen. Attach sanitized app captures and dated observations.
- [ ] 006-V84 **Install/example replay:** Run documented native Unix/PowerShell commands and first-task examples against extracted candidate bytes, identifying every optional tool; unsigned-download restrictions are observed without disabling platform controls.
- [ ] 006-V85 **CLI release timing:** Run all 28 cases × three complete runs against the exact candidate Linux amd64 executable on the declared host; retain five warmups and 100 consecutive samples/case/run, tails/max/miss counts/output correctness, compiler/binary hash and host conditions. Reference query p90/p95/p99/max <15/20/30/50 ms and help/version <5/7.5/10/15 ms; prior PR #5 calibration alone cannot close this release gate.
- [ ] 006-V86 **TUI performance/U5 closure:** Measure candidate projection/frame/View/Markdown/startup with Feature 005's retained fixtures and budgets; no concurrent build/test workload during timing. Retain all samples/failures, native terminal/runtime receipts, canonical gate and sequential data/portability/performance review before unit closure.

### U7 — Homebrew cask and distribution destination

- [ ] 006-V87 **Destination ownership:** Verify the authorized accessible tap/repository/branch and exact cask path; nonexistent destination/insufficient permission fails explicitly before any hosted mutation.
- [ ] 006-V88 **Cask generation:** Generate with upload disabled, current nondeprecated cask configuration, correct macOS Intel/ARM URLs/digests, binary/completion/man mappings and no unsupported Linuxbrew claim.
- [ ] 006-V89 **Prepublication native install:** Audit and install a reviewed local candidate cask on native Intel/ARM macOS; fixture local asset URLs are labeled prepublication evidence and never presented as verified public downloads.
- [ ] 006-V90 **Cask negative paths:** Wrong architecture/hash, missing asset or invalid directives fail safely, preserve existing binary/data and cannot cause a partial successful brew route.
- [ ] 006-V91 **Upgrade/remove:** Candidate cask replacement and uninstall preserve tasks/history/DB/WAL/SHM, with no zap/data-removal or quarantine-bypass hooks; `tusk` PATH and completion/man behavior match docs.
- [ ] 006-V92 **Native app after install:** Installed cask binary matches the accepted payload hash; launch the actual TUI/CLI, quit/relaunch and read fixture data afterward on both macOS architectures.
- [ ] 006-V93 **No automatic external write:** Snapshot/candidate/draft modes do not publish to tap; missing license/accepted release metadata blocks cask release readiness and no permanent cross-repository token is required.
- [ ] 006-V94 **U7 closure:** Native Homebrew lint/install/remove/version/manual/completion records and destination evidence match the cask; docs/metadata preview and canonical/review gates pass, while final live public URL/tap proof remains U8.

### U8 — Authorized publication and GitHub landing-page settlement

- [ ] 006-V95 **Exact candidate preflight:** Match authorized SHA/version and manifest digest against current accepted CI/provenance/native/performance/license/tap receipts; missing, expired, mismatched or unaccepted gate refuses promotion.
- [ ] 006-V96 **Complete draft/tag:** Authorized new semver tag points to the accepted main SHA; draft has every exact accepted asset/notice/checksum/provenance and reviewed notes. Conflicting existing tag/release refuses overwrite or retag.
- [ ] 006-V97 **Re-download/reconcile/publish:** Verify every draft asset hash/version/provenance after upload, reconcile timeout/lost response, then publish only within explicit authorization. Retry does not duplicate/replace a published asset; immutable setting or no-overwrite policy is verified.
- [ ] 006-V98 **Public install/tap:** Anonymous direct downloads and native live macOS tap installs verify final public URLs, checksums and versions; tap/asset failure keeps final release settlement open despite any successful publication step.
- [ ] 006-V99 **README activation:** Release installation links, license/CI/release badges, screenshots, support table and performance statements match actual evidence; README/guide examples are replayed against the released bytes.
- [ ] 006-V100 **GitHub metadata readback:** About description/homepage/topics/social preview, license detection and security/community contact features match approved values after current-state comparison; unavailable settings are named gates, never invented successes.
- [ ] 006-V101 **Rendered/anonymous landing page:** Inspect actual GitHub README at documentation SHA, social preview and release pages; links/images/alt text/anchors/tables/code blocks render, and visitors can find installation and use without workflow artifacts.
- [ ] 006-V102 **Final governance/recovery:** Final receipt binds executable-source SHA, documentation SHA, manifest/assets, native/hosted/browser evidence, per-unit commits and public release/tap URLs. All open issues have resolved evidence; fresh canonical validation and current review pass before synchronization/commit and inherited G4/scenario closure. Publication defects use a new version, preserving old bytes/receipts.

---

## Commands and environments

All software checks/builds use Make in Codex Bash or the native hosted Bash
environment. Standalone tool commands below describe the methods a planned Make
target wraps; they do not authorize bypassing canonical targets. Kitty runs only
the real installed app and shell completion/manual use.

| Tier | Canonical command/interface | Availability / owner / evidence |
| --- | --- | --- |
| Environment | `make setup` | Exists; U1 adds CI Make/compiler/tool preflight without runtime requirements |
| Aggregate | `make validate build check-generated` | Exists; every unit commit, native jobs and final candidate source |
| Minimum | `GOTOOLCHAIN=go1.25.0 make validate build-tui` | Exists; Linux runtime/minimum and five-target application/test compile proof |
| CLI focused | `make test-cli CLI_TEST_RUN='<pattern>'` | Exists; U3 completion/root tests |
| TUI focused | `make test-tui TUI_TEST_RUN='<pattern>'` | Exists; bounded regression fixes only |
| CI configuration | `make test-ci check-ci` | Planned U1; negative fixtures plus pinned actionlint/config contract checks |
| Shells | `make test-completions` | Planned U3; binary generation plus Bash/Zsh/Fish syntax/load/use; missing shell fails/pends |
| Generated docs | `make generate-docs check-docs` | Planned U3; deterministic real-tree manuals/completions and diff checks |
| Public docs | `make test-docs check-notices` | Planned U4; link/example/claim fixtures and replacement-aware license inventory |
| Release unit fixtures | `make test-release` | Planned U2; script/version/manifest/API fakes; later U6/U8 extend the same contract |
| Config/local snapshot | `make release-check release-snapshot` | Planned U2; wraps pinned GoReleaser check/full snapshot, archives and local native smoke; never publishes |
| Intended-tag candidate | `make release-candidate RELEASE_VERSION=<semver> RELEASE_SHA=<full-sha> RELEASE_OUTPUT=<new-owned-dir>` | Planned U2; separate owned checkout/local-only tag, no publishing, deterministic manifest |
| Exact artifact verification | `make verify-candidate CANDIDATE_DIR=<downloaded-dir> CANDIDATE_MANIFEST_SHA256=<digest>` | Planned U6; verifies target set/hashes and trusted hosted provenance |
| Native lifecycle | `make test-release-smoke release-smoke RELEASE_BINARY=<extracted-binary> RELEASE_MANIFEST=<path>` | Planned U5; shell/PowerShell native drivers, owned temp fixture paths and semantic snapshots |
| CLI release measurement | `make bench-cli-release RELEASE_BINARY=<extracted-binary> CLI_BENCH_OUTPUT=<new-report>` | Planned U5; invokes the existing runner without rebuilding the supplied binary; hash-preservation fixture required |
| TUI release measurement | `make bench-tui-release RELEASE_BINARY=<extracted-binary> TUI_BENCH_OUTPUT=<new-report>` | Planned U5; source-model measurements plus separate packaged startup/identity selection, retaining Feature 005 workloads/budgets |
| Homebrew | `make test-homebrew check-homebrew CANDIDATE_DIR=<dir>` | Planned U7; generated cask syntax/audit/native installation; hosted tap mutation remains separate |
| Draft | `make release-draft CANDIDATE_RUN_ID=<trusted-id> CANDIDATE_MANIFEST_SHA256=<accepted-digest> RELEASE_VERSION=<semver> RELEASE_SHA=<sha>` | Planned U8; explicit authorized hosted write, accepted bytes only |
| Publish | `make release-publish RELEASE_TAG=<authorized-tag> CANDIDATE_MANIFEST_SHA256=<accepted-digest>` | Planned U8; verifies complete draft/readback/authorization, never rebuilds |
| GitHub metadata | Structured GitHub API or documented CLI plus browser readback | U8; exact approved payload, before/after state and documentation SHA; social preview manual upload if necessary |
| Terminal | Owned Kitty Linux/macOS; owned Windows Terminal/PowerShell Windows 11 | U3/U5/U7; actual apps, isolated data, target window identity and inspected captures |

Exact new helper/function names are implementation details. The target inputs
above are decided public Make interfaces and must be reflected in help/docs and
negative fixtures. All output paths are quoted; choose fresh evidence directories
instead of overwriting earlier reports. No `make clean` runs against user data or
as an evidence cleanup shortcut.

---

## Fixtures, failure injection and execution receipts

Use owned temporary roots for checkout/tool archives, shell config, candidate
output and home/data. In application fixtures, set task-specific path variables
and pass `TUSK_DB_PATH`/`XDG_DATA_HOME` only to the child; do not repurpose the
coordinator's HOME. Windows drivers use native path syntax and named isolated
directories. All child processes have deadlines, cleanup and join/readback rules.

CI/release script fixtures use fake tool/API executables with bounded scripted
responses for download, digest, missing prerequisite, build failure, upload
partial failure and lost response. Use structured text files for notes/JSON;
malicious version/ref/release text must remain data. Archives cover member types,
paths, missing payloads and corruption. Filesystem failures are deterministic on
privileged runners rather than assuming chmod proves an access error.

Data fixtures include empty/current schema, populated nested tasks/history,
Unicode/spaced paths, held writer, corrupt/newer schema and unknown commit
outcomes. Compare normalized domain tasks/events around upgrade/restore, and
byte hashes around refusal. No new schema or legacy PostgreSQL migration is
introduced. Deletion/unknown-outcome behavior reuses accepted service contracts.

The exact candidate source/overlay/tools/archive/executable/manifest hashes are
recorded together. Model-level benchmarks and packaged startup evidence remain
different measurements. Current benchmark recipes rebuild `BUILD_OUTPUT`; U5
must observe a failing hash-preservation fixture and add the supplied-binary seam
before claiming to measure downloaded bytes. No copying a candidate onto a path
that an existing build prerequisite silently overwrites.

TUI measurement retains the [Feature 005 performance protocol](2026-09-06-005-feat-interactive-tui-application-verification-plan.md#performance-protocol):
three complete runs, five warmups and 100 samples per case; frame preparation at
80×24/120×40/200×60 with 0/100/1,000 tasks, 32 KiB selected notes and 100 displayed
history events has p95 <16.7 ms, p99 <33.3 ms and maximum <50 ms. Pure View timing,
cold Markdown, refresh and packaged startup remain separate observations without
an invented startup SLA. Retain 10,000-task/1 MiB/resize-storm stress observations
and resource quiescence. The measurement driver and model code must match the
accepted source SHA; compiled source-model timing is labeled separately from
the downloaded executable's startup/runtime measurements.

Freeze and commit test drivers/benchmark selectors/release/cask configuration
before final candidate generation. If U5/U7/U8 changed build inputs after the
initial U6 run, repeat the trusted candidate and affected native/performance/cask
checks; never keep the earlier source SHA as proof of a changed build. Receipt
and README-only changes record their own documentation identity. All pending
external gates remain unchecked after a local engineering checkpoint.

Execution later creates `docs/verification-evidence/006/` with per-unit receipts,
native job links, terminal capture provenance, complete retained benchmark
reports, candidate manifest/provenance verification, owner decisions and final
release/metadata readbacks. Each receipt records UTC time, unit/scenario IDs,
source/documentation SHA, exact command or observed action, expected/actual
result, host/tool/terminal versions, all relevant hashes and evidence paths.
Keep failure reports and the prior candidate's status. Red/green examples above
are planned failure contracts, not observations of executed tests.

---

## Planning audit record

The 2026-10-01 planning audit checks requirement/unit/scenario coverage, cross-links,
planned command ownership, external gate owners, unchanged historical receipts
and unchecked execution scenarios. Its document-only results are recorded in the
workorder. Application/native/hosted/release execution results must be recorded
later; none is inferred from a well-formed document or live read-only GitHub query.
