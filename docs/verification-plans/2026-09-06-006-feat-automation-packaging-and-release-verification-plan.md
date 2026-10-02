---
feature-id: "006"
plan-source: docs/plans/2026-09-06-006-feat-automation-packaging-and-release-plan.md
surface-profiles: [cli-tui, infrastructure-operations, installation-data-lifecycle, documentation]
status: PR 6 published; second complete report batch locally validated; fresh hosted/native/release gates open
evidence-scope: 51 local scenario closures; two complete PR 6 report batches locally remediated; final canonical pass; fresh hosted/native/release proof pending
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

- [x] 006-V31 **README information order:** A visitor finds purpose, screenshot, available installation and first-task workflow before developer/planning information; source-only status is explicit before a release exists.
- [x] 006-V32 **Source installation:** A fresh checkout keeps replacements, uses documented Go/Make/Bash prerequisites and builds the full command package; no unsupported versioned `go install` path is presented as usable.
- [x] 006-V33 **Unix install draft:** Linux/macOS instructions map architecture to exact asset names, verify hashes before extraction, use user-writable paths and show PATH/version recovery for wrong-architecture/permission/download failures; public URLs remain marked pending until release.
- [x] 006-V34 **Windows install draft:** Native PowerShell hash/extraction/PATH commands distinguish amd64 and use `tusk.exe`; no WSL or Bash-only command is represented as the Windows user route.
- [x] 006-V35 **Quick start replay:** Version, add, list, tree, clean JSON and TUI examples use valid flags and full IDs; no destructive first-run example or hidden prerequisite.
- [x] 006-V36 **Dates/status/JSON examples:** `tomorrow`, supported offsets, priority/status enums and JSON schemas match guides/source; invalid `next week` or an undeclared `jq` dependency fails docs review.
- [x] 006-V37 **Shared configuration:** Explain actual DB/environment/timezone precedence, relative explicit paths, no invalid-config fallback, Windows home fallback, `NO_COLOR` and terminal requirements.
- [x] 006-V38 **Safety/recovery language:** Explain independent force/recursion, no undo, metadata-only history and unknown/committed outcomes; readback precedes retry, and no recovery instruction deletes DB/WAL/SHM.
- [x] 006-V39 **Backup rehearsal:** Stop all fixture owners, preserve closed DB plus remaining sidecars, restore to an isolated location, check integrity and compare tasks/events; prohibit live-DB-only copy and state NORMAL durability limits.
- [x] 006-V40 **Claims and badges:** Every supported platform/install/performance claim has a matching dated receipt; nonexistent release/CI/license badge and universal latency claim fail review.
- [x] 006-V41 **Screenshot/privacy:** Inspect real app image ownership/content, sanitized task data, date/source identity and alt text; social preview is readable and uses an app capture with recorded provenance, not a generated mockup.
- [x] 006-V42 **Metadata/community preview:** Exact proposed description/homepage/topics/social preview and issue/PR/contributor/security guidance agree; optional website/funding/contact fields are not fabricated.
- [x] 006-V43 **Links/rendering:** Relative guide/image/anchor links resolve; headings/code blocks/tables render legibly, badges have valid targets and the drafted public GitHub layout is inspected before activation.
- [x] 006-V44 **U4/license closure:** Owner's selected first-party license/rights confirmation, replacement-aware third-party inventory and verified security contact support all public claims; documentation checks, canonical validation and sequential usability/scope/privacy review pass before the unit commit.

U4 local acceptance is bound to [the receipt](../verification-evidence/006/u4.json).
V33/V34 verify explicitly pending installation drafts, not native execution or
public assets. V35 carries forward the real TUI capture with original provenance;
CLI/JSON and closed-backup rehearsal use current actual processes. V41/V43 cover
inspected image and local Chrome rendering of GitHub Markdown API output; hosted
GitHub metadata/social-preview/license readback remains U8.

### U2 — Versioned packaging and local lifecycle safety

The `scripts/releasecheck/` Go helper is exercised by `make test-release`, full
canonical tests/race/coverage and the minimum compiler; no coverage exemption.
Archive/binary metadata inspection uses standard libraries on any build host.
Unsafe Make parameter fixtures must fail without evaluating shell/Make code.
Pre-commit validation artifacts name their private local validation SHA and do
not close trusted hosted/native acceptance.

Source archive checks use actual Git archives, a narrow initial PAX commit
comment and fixed 0022 tar umask, with executable bits read from Git rather than
OS permissions. Component Red/Green covers config hooks/payload mutations,
substituted/missing manifest inputs, trailing tar data, setuid/setgid/sticky
archive modes and fresh-checkout sqlc.
The cross-built integration fixture contains the real app and payload; the actual
full repository archive is still required for V53. All V45–V62 local results are recorded in [the U2 receipt](../verification-evidence/006/u2.json):
actual complete source archive, byte-identical B/C assets, candidate/snapshot
version smoke, minimum compiler and fresh canonical gates. These private local
validation objects do not close U6 hosted provenance or U5 native release gates.

The canonical config hash excludes temporary-path entropy. A generated runtime
config changes only literal `dist`; retain its actual bytes outside assets and
compare the stable manifest/payload set between owned locations. The first real
packager run exposed unsupported `dist` templating and remains a failed receipt.

- [x] 006-V45 **Immutable version:** All three version forms report the intended validated tag version in actual built binaries; ordinary source and snapshot labels remain clearly unreleased.
- [x] 006-V46 **No mutable linker metadata:** Default `-X` injection or a mutable version global fails the build contract; generated constant/overlay input hash appears in the manifest and the original checkout stays clean.
- [x] 006-V47 **Version/input rejection:** Empty/malformed/non-semver values, shell/code injection text, conflicting tag and unsafe overlay/output paths fail before build/publication; spaces in owned temporary paths remain supported.
- [x] 006-V48 **Complete target matrix:** Exactly the five required OS/architecture payloads exist with matching executable architecture and no invented Windows arm64 target; omission/duplicate/wrong arch fails release-check.
- [x] 006-V49 **Runtime identity:** Build metadata proves CGO disabled, embedded timezone data and selected patched graph; Linux has no dynamic interpreter dependency. Darwin/Windows OS-library use does not become an unsupported fully-static claim.
- [x] 006-V50 **Payload contract:** Unix tar.gz/Windows zip have the right executable, names/modes, README, three completion scripts, all manuals and required notices; extraction exposes the root executable as documented.
- [x] 006-V51 **Licenses:** Missing project license, missing replacement-tree/embedded-asset notice and unclassified dependency obligations fail distribution checks; notices retain upstream grant text.
- [x] 006-V52 **Archive abuse:** Absolute paths, `..` traversal, unsafe links, unexpected database/sidecar/credential files and unexpected executable members fail manifest/member validation.
- [x] 006-V53 **Source archive:** Complete tagged source includes both patched third-party modules and verified gallery bytes; documented checkout build retains the intended graph without an LFS service requirement.
- [x] 006-V54 **Payload consumption:** A clean isolated native home can extract and run version/help without Go, compiler, external service or database initialization; verify archive and executable hashes separately.
- [x] 006-V55 **Manifest/checksums:** Payload corruption, missing hash, wrong target/version/SHA/tool input or a self-hash dependency cycle fails verification; manifest/checksum files identify one complete candidate.
- [x] 006-V56 **No remote side effects:** Both snapshot and local intended-tag candidate modes trap any release API/tag push/tap mutation; no publishing credentials are required or read.
- [x] 006-V57 **Cleanup/data boundaries:** Failure/cancellation cleans only owned temporary build storage, keeps diagnostics and preserves user DB/sidecars/unrelated dirty files; colliding retained outputs are rejected.
- [x] 006-V58 **Binary reproducibility:** Same source/version/compiler/overlay/tool pins built in two owned locations produce matching executable hashes; timestamp/path entropy is detected rather than excused.
- [x] 006-V59 **Archive reproducibility:** Same payloads/source timestamp/tool pins reproduce archive/man/completion bytes; manifest contains stable inputs and separately records non-deterministic workflow identity where appropriate.
- [x] 006-V60 **Failed prerequisites:** Bad tool digest, unavailable pin, failed documentation generation or failed child build produces nonzero release-check/candidate result; preserve existing accepted outputs and no fallback compiler.
- [x] 006-V61 **Notice inventory scope:** Source and executable dependency/asset inventories include local replacement paths and patch provenance; stale inventory fails after a dependency/asset change.
- [x] 006-V62 **U2 closure:** Canonical/generated/minimum checks, manifest/member negative fixtures and current native local smoke pass with synchronized review/receipts before the U2 commit; hosted/native release claims remain open.

### U6 — Hosted candidate and trusted provenance

Local workflow/API/certificate fixtures exercise fail-closed identity, exact main
SHA, all six reusable CI jobs, error-skipping/changed-input rejection, fixed
artifact extraction layout, payload integrity, expiry and lost-response readback.
They are component evidence. All V63–V72 remain pending actual authorized hosted
run/artifact/attestation proof. The verifier requires fresh output outside the
candidate, retained API artifact digest and signed certificate run/attempt/source
identity for all ten subjects; it does not claim a computed transport ZIP hash.
GitHub's minimal run repository is followed by current default-branch readback.

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
| Promotion fixtures | `make test-release-promotion test-repository-metadata lint-release-promotion` | Local U8 boundary/approval/report/API-loss fixtures; no actual consent, signatures or hosted writes |
| Prepare | `make release-prepare` / `make prepare-repository-metadata` with the [documented inputs](../release-acceptance.md) | Implemented U8; no API, unaccepted/unapplied outputs |
| Draft | `make release-draft` with candidate/version/SHA/run/manifest, acceptance, draft approval, notes and new output | Implemented U8; separately authorized tag/private draft/assets; exact full interface in [acceptance records](../release-acceptance.md) |
| Publish/readback | `make release-publish` / `make release-readback` with the same candidate/acceptance/notes and new output | Implemented U8; separate publish approval, complete re-download verification, no rebuild; readback writes no remote state |
| GitHub metadata | `make apply-repository-metadata` with authorization, published receipt and new output | Exact bounded About/topics payload and current documentation SHA; manual social/browser readbacks remain pending |
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

### U6 local engineering checkpoint — 2026-10-01

The manual workflow, reusable same-SHA native/minimum CI, verified Action pins and
fail-closed candidate/API/certificate policy pass canonical, minimum and actionlint
checks. [Receipt](../verification-evidence/006/u6.json) retains observed Red/Green
and sequential security/reliability review. All V63–V72 hosted closure remains open:
no workflow has been pushed/dispatched and fake signatures are component evidence.
The committed local checkpoint enables U5's local driver/selector implementation
under the source-freeze contract; exact hosted artifact acceptance still depends
on an authorized trusted run after all release build inputs are committed.

### U5 local engineering contract — 2026-10-01

Supplied-binary selectors feed existing real CLI/docs/backup and Linux child PTY
tests without rebuilding the candidate. Drivers require pinned manifest/executable
identity and a trusted verification receipt; `local-fixture` is explicitly
preliminary. Fresh evidence lives outside candidate storage and source checkout.
Owned process fixtures are retained on success/failure. Replacement/removal checks
protect task/event identity and data files; newer-schema refusal protects all
pre-existing DB/sidecar bytes and forbids writes into any newly created WAL.
SQLite may create an empty WAL and read-cache SHM on read-only inspection; absence
of these metadata files is not the data-preservation contract. Observed overly
strict presence assertion and failed run are retained, not relabeled as passes.

CLI release measurements use the reference profile. TUI model measurements require
a clean exact-source checkout; packaged startup and binary/measurement compiler
identities remain separate. PowerShell native execution, actual trusted five-target
runtime/terminal proof and candidate timing remain pending. V73–V86 stay open.

### U5 local engineering checkpoint — 2026-10-01

Supplied-artifact selectors, native smoke drivers, retained owned fixtures and
reference measurement entry points pass canonical validation, Go 1.25 checks
and all five application/test cross-builds. The preliminary Linux payload passes
real CLI/docs/backup/replacement/newer-schema and child PTY checks with its hash
unchanged. [Receipt](../verification-evidence/006/u5.json) retains unsafe-output
Red/Green and the corrected SQLite metadata assertion. No production app change,
trusted candidate timing or other-platform native acceptance is claimed. All
V73–V86 and the parent U5 checkbox remain open. The coherent local engineering
commit enables U7 configuration; final source freeze requires a fresh hosted run.

### U7 local cask engineering contract — 2026-10-01

Separate macOS build/archive IDs preserve five targets and nine public assets,
while current `homebrew_casks` selects only the two macOS archives. Upload is
disabled; the cask declares macOS, installs the binary, three static completions
and all 12 manuals, and has no hooks/zap/security bypass. The strict Go audit
compares complete nonblank statements with manifest architecture/hash/member
identity without evaluating downloaded Ruby. Only indentation and blank lines
are ignored; comment/statement boundaries and a pinned no-zap comment are retained.

The candidate bundle now also includes `homebrew/Casks/tusk.rb`, digest-bound in
`candidate-run.json` and separately attested by the same trusted workflow. It
stays outside the nine-file public-asset inventory. Missing/tampered casks refuse
complete candidate acceptance. These workflow/run-schema changes require a fresh
hosted candidate; earlier U6 component receipts remain historical. Failed local
packaging retains the generated cask before auditing. Native Ruby/Homebrew and
Intel/ARM macOS install/removal/security behavior remain unexecuted. The current
read-only tap lookup is HTTP 404; 006-ISS-003 and V87–V94 stay open.

### U7 local engineering checkpoint — 2026-10-01

Pinned GoReleaser creates a macOS-only Intel/ARM cask with the exact archive hashes,
12 manuals and three completions; upload is disabled. Real fresh local packaging
and strict declarative audit pass after retained ordering/comment failures. The
candidate workflow now uploads/attests the separate run-bound cask. Canonical and
minimum-compiler component checks pass; inspector coverage remains above 95%.
[Receipt](../verification-evidence/006/u7.json) retains source/hash/config and review.
Native Ruby/Homebrew/Intel/ARM security/runtime checks are unexecuted; current tap
readback is HTTP404. Parent U7 and V87–V94 remain open. The coherent local commit
enables U8 promotion/readback engineering, with final hosted candidate required.

### U8 local engineering checkpoint — 2026-10-02 UTC

Promotion and repository metadata helpers pass canonical Go 1.27.1 validation,
minimum-Go focused checks and strict Bash lint. Read-only preparation consumes the
real preliminary U7 C manifest/cask and emits nine exact asset digests plus the
approved About/topics/image preview, explicitly unaccepted/unapplied. Fake API
fixtures expose and fix stale main, duplicate timing cases, malformed memory
samples, missing tap, absent draft tag, failed source diff and inherited host/debug
or API-host redirection. Lost create/upload/publish responses reconcile existing
state and downloaded bytes without clobber/delete/retag. The current tap still
returns HTTP404. The custom host compiler notice failure and superseded mixed
source gate remain retained failures; the final pinned gate passes.

[Receipt](../verification-evidence/006/u8.json) records the local boundary scope.
No fixture approval is owner consent or real provenance/native/performance proof.
Parent U8 and V95–V102 remain open; local scenario closures stay 51. Final local
simplification/code review and source freeze follow the coherent engineering
commit. No push, workflow dispatch, remote tag/draft/release, metadata or tap write
has occurred. Hosted/native/publication acceptance requires its separate actual
evidence and concrete owner authorization.

### Final local review-fix and source-freeze checkpoint — 2026-10-02 UTC

All eight local checkpoints are committed. The actual ce-code-review receipt
(`status: complete`, run `20261002-000427-b58e9f8f`, reviewed c027114) retained three
validated findings. Caller-owned Red/Green fixes now require native gate target,
executable and archive digests to match the verified manifest; bound all four
GitHub helpers with a portable stdlib process driver; and require exactly one
candidate/overlay JSON value. The driver preserves streams/arguments/exit codes,
limits every operation to five minutes and joins the owned child/pipes. A shorter
positive `GH_REQUEST_TIMEOUT` is allowed; unbounded/longer values fail. Canonical
Go 1.27.1, Go 1.25 fast/script checks and strict ShellCheck pass; helper coverage
is 97.2%. The real bounded tap readback still returns HTTP404.

[Final local receipt](../verification-evidence/006/review-local/acceptance.json)
retains the completed report, peer admission decisions, all Red/Green/failure
logs and focused fix review. Local lenses/finish roles ran sequentially under the
Task mapping; Claude returned an authentication failure, and the alternate
Composer receipt did not verify its actual model/effort or serving family. No
independent model agreement is claimed. No justified review finding remains open.

The containing coherent review-fix commit is the local source freeze. Next is
separate owner authorization to push this branch and open its reviewable PR;
merge, main candidate dispatch, tags/releases, settings and tap writes remain
separate. Select the final main SHA only after authorized integration, generate
a fresh hosted candidate and rerun affected exact-byte native/cask/reference
gates. Earlier private local candidates are preliminary. Parent U6/U5/U7/U8,
Phase 6/G4 and unexecuted scenarios stay open; local scenario closures remain 51.


### Owner-requested renewed local review loop — 2026-10-02 UTC

The owner deferred publication and requested another ce-simplify-code pass, then
a full ce-code-review and Red/Green remediation loop for all confirmed P0–P2
findings. The prior acceptance at `72c5eeb` remains historical. The active target
is this local review loop; no push, PR, tag, workflow dispatch or release is
authorized. Each completed fix unit must pass canonical validation, synchronize
this triplet and MASTERPLAN, and be committed before advancing. Actual native,
hosted, tap and publication evidence remains pending; the 51 local scenario
closures do not close those parent gates.

### Renewed local review-fix checkpoint — 2026-10-02 UTC

The owner-requested simplification found no worthwhile behavior-preserving change
(0 reuse/quality/efficiency edits; three deliberate structures retained). Full
review `20261002-140940-86eb372b` confirmed two P2 issues; caller focused review
caught a third. Seven Red cases now refuse concatenated approval/acceptance/report
and verification objects; policy fixture hashes support shasum without GNU
sha256sum; a failed Windows native path conversion refuses selected-binary
acceptance. All existing identity, hash, sample and owner-approval guards remain.

[Receipt](../verification-evidence/006/review-r2/acceptance.json) retains the
original completed review, focused addendum, failed/incomplete fixtures, Green
checks and source hashes. Focused checks, Go 1.25 script suite and strict
ShellCheck pass. Fresh final canonical validation/build/generated/docs/notices/workflow checks
pass before this coherent local fix commit; repeat full review on the committed
head.
Local roles ran sequentially, Claude authentication failed and Composer serving
model/effort/independence is unverified. No independent agreement is claimed.

Local scenario closures remain 51; parent U6/U5/U7/U8, Phase 6/G4 and actual
native/hosted/tap/publication gates remain open. No remote mutation or publication
is authorized. The active target remains this review/fix loop.

### Renewed local review loop completion — 2026-10-02 UTC

- [x] Apply ce-simplify-code to the full Feature 006 branch (no worthwhile edits).
- [x] Fix all three confirmed P2 issues with observed Red/Green, focused review,
  minimum-Go/lint/canonical checks and coherent local commit `df10217`.
- [x] Repeat full ce-code-review on that committed head to a clean local pass.

Run `20261002-143551-fd10da2c` is complete with no remaining confirmed P0–P2
finding or unresolved review gate. [Clean repeat receipt](../verification-evidence/006/review-r3/acceptance.json)
retains full coverage, all five rejected peer claims with current guards and
source binding. Local lenses/finish roles ran sequentially; Claude returned
HTTP401 and Composer's actual model/effort/independence is unverified. No
independent agreement is claimed. Both consumed peer jobs are deleted.

Fresh closure canonical validation/build/generated/docs/notices/workflow checks
pass before the governance/evidence commit. The reviewed executable sources remain unchanged. The local review loop
is complete and publication remains deferred by the owner. Actual native/hosted,
tap, cask/performance and publication gates remain open; parent U6/U5/U7/U8,
Phase 6/G4 and the 51 local scenario count do not change. No push/PR, workflow
dispatch, tag, draft/release, settings or tap write occurred.


### Learning and authorized PR publication — 2026-10-02 UTC

The owner invoked ce-compound followed by ce-commit-push-pr, superseding the
previous publication hold for branch push and PR creation. The reviewed executable
sources remain unchanged from the clean repeat review at `df10217`. The
[release-check learning](../solutions/workflow-issues/prove-release-policy-rejection-at-the-intended-boundary.md)
records how to establish causal Red/Green evidence with controlled fixtures.
Full compounding ran sequentially under root AGENTS; frontmatter, links and six
behavior claims pass grounding checks. No glossary or instruction edit was needed.

- [x] Capture the verified learning and synchronize publication authority after a
  fresh canonical gate. [Receipt](../verification-evidence/006/publication/acceptance.json).

The containing documentation unit passed canonical validation before commit
and publication. The complete branch targets GitHub main for hosted review;
publication of a PR does not close release acceptance. Actual native terminals,
trusted-main candidate provenance, exact-byte cask/performance acceptance, tap
availability and release/settings/tap writes remain pending. Parent U6/U5/U7/U8,
Phase 6/G4 and the 51 local scenario count are unchanged. No merge, release tag,
workflow dispatch, repository settings or tap write is authorized by this request.

### PR #6 complete hosted report batch — 2026-10-02 UTC

[PR #6](https://github.com/newbpydev/tusk/pull/6) is open at
`ced4417a648c3dcd21d4e48a215788ca2cce3112`. The owner requested all reports
before one combined remediation pass. All six CI jobs and Kilo's review completed
on that unchanged commit. The 30 review threads were assessed together: 27
change items and three evidence-based replies. Six failing CI jobs reduce to
Windows tool provisioning, macOS fixture assumptions and undeclared ripgrep
in the shell policy fixtures. Additional instances of those portability
assumptions are included in the same bounded unit.

- [x] Apply the valid review/CI changes with causal Red/Green evidence.
- [x] Pass fresh current-state canonical validation and review the combined diff;
  prepare the complete batch as one coherent commit/push unit.
- [ ] Settle the complete post-push hosted report set before release acceptance.

These reports do not close the pending actual native terminal, exact-byte
performance, trusted-main candidate, cask/tap or release-publication gates.
The 51 local scenario closures and open parent units remain as recorded above.

Native preflight now checks the selected executable’s actual `--version` output.
Every native promotion receipt requires `observed_version` equal to the candidate
manifest version. Copied inventories are rechecked before finalize succeeds.
The complete batch is retained in [the R1 receipt](../verification-evidence/006/pr6-r1/acceptance.json).

The first combined canonical gate passed functional/race tests but rejected
release-inspector coverage at 92.6%. Malformed PE tables, ordinal imports and
bounded/corrupt timezone ZIPs now raise focused coverage to 95.5%; final
canonical validation is pending. The failed gate remains retained.

A follow-up inventory regression rejects a standalone notice that differs from
the accepted source even when its bundle hashes/checksums are repaired. The
minimal source-digest binding passed focused Red/Green; final current-state
canonical validation follows the already-passing intermediate gate.

The final current-state canonical gate passed: `make validate build
check-generated check-docs check-notices check-ci check-candidate-workflow
lint-release-promotion`, using official `GOTOOLCHAIN=go1.27.1`. Release-inspector
coverage is 95.2%. The containing commit is the coherent local remediation unit;
publication, visible thread replies/resolution and fresh hosted reports are
verified separately on PR #6. Pending native/release gates remain unchanged.

### PR #6 second complete hosted report batch — 2026-10-02 UTC

All six Native CI jobs and Kilo review completed on unchanged
`79646a31bae6b95980b8bd518b744c15944a4fb7` before this repair unit began.
Linux release/minimum compiler jobs pass. Both macOS runners pass functional
checks but reject release-inspector coverage at 94.9%. Windows now passes tool
setup and exposes build-output quoting, checkout-byte conversion and platform
fixture assumptions. One new import-policy suggestion and four carried summary
claims are assessed against current source and retained evidence together.

- [x] Repair the complete confirmed batch with causal Red/Green.
- [x] Pass fresh canonical validation and review the complete applied diff.
- [ ] Commit/push one coherent repair and settle all fresh hosted reports.

Parent units, the 51 local scenario closures and native/release acceptance
remain unchanged. No merge, release, tag, workflow dispatch, settings or tap
mutation is included.

The second combined repair passes frozen-state `make validate build
check-generated check-docs check-notices check-ci check-candidate-workflow
lint-release-promotion` with official `GOTOOLCHAIN=go1.27.1`. Inspector coverage
is 96.2% on Linux. Five-target verification-test compilation, the actual vendor
checkout inventories under `autocrlf=true`, and all 30 native-binding rejection
checks with zero fake GitHub calls pass. See [the R2 receipt](../verification-evidence/006/pr6-r2/acceptance.json).
The containing commit prepares one coherent repair. Windows ACL execution and
macOS/Windows coverage require the next complete hosted report set.


### PR #6 third complete report-batch repair (2026-10-02)

All seven reports finished on `eaa8791c2f82e1320620354fb2da95f5062cd52c`
before repair edits. Linux/minimum compiler and Kilo pass; both macOS jobs fail
at the release-smoke fixture's physical-path assertion, and Windows passes
functional/race tests but misses native coverage in confirmation and the CLI
measurement process launcher. Previous ACL, checkout-byte and inspector fixes
pass their native checks. This is progressive failure migration.

- [x] Collect the complete third hosted set before starting one combined pass.
- [x] Repair physical/native measurement paths and exercise native confirmation
      and process launching without weakening the 95% package coverage gate.
- [x] Validate the frozen repair and audit the complete applied diff.
- [ ] Commit/push one coherent repair and settle its fresh reports.

The minimum Go 1.25.0 hosted job passes the actual isolated build-output test,
refuting Kilo's new portability suggestion. Physical native terminal, trusted
candidate, exact-byte performance/cask and release/tap acceptance stay pending.

The third combined repair passes frozen-state `make validate build
check-generated check-docs check-notices check-ci check-candidate-workflow
lint-release-promotion` with official `GOTOOLCHAIN=go1.27.1`. Linux coverage is
96.1% for cmd/tusk, 95.6% for CLI measurements and 96.2% for the inspector.
Five-target CLI/verification tests compile. See [the R3 receipt](../verification-evidence/006/pr6-r3/acceptance.json).
The containing commit records one coherent repair. Fresh Windows EOF/cancellation
runtime and package coverage, both macOS smoke fixtures and all other current-head
hosted reports remain required; no native/release acceptance gate closes here.


### PR #6 fourth complete report-batch repair (2026-10-02)

All seven reports finished on `268f0c812c87a4dd894cbc4bec58b1885ac543a1`
before repair edits. Both macOS jobs, all three Linux jobs and Kilo pass. Windows
passes real EOF/answer/invalid-handle/cancellation tests and coverage (cmd/tusk
97.4%, CLI measurements 95.6%), then fails the first shell hash fixture. Git
Bash's default copy-style `ln -s` separates a restricted-PATH executable from
its adjacent runtime. A passing owned resolved-image control and copy-semantics
Red establish the fixture defect; native runtime confirmation remains required.
The related SQLC tool snapshots and archive-member refusal are reviewed together.

- [x] Collect the complete fourth hosted set before repair edits.
- [x] Preserve installed tool runtimes in restricted snapshots and prove the
      SQLC archive-member refusal at its intended boundary.
- [x] Pass frozen-state canonical validation and audit the applied diff.
- [ ] Commit/push one coherent repair and settle its fresh reports.

This remains progressive failure migration: the earlier native repairs pass and
the Windows job reaches a later gate. Coverage, latency and archive security
contracts remain unchanged; physical native and release acceptance stay pending.

The fourth combined repair passes frozen-state `make validate build
check-generated check-docs check-notices check-ci check-candidate-workflow
lint-release-promotion` with official `GOTOOLCHAIN=go1.27.1`. The complete shell
suite also passes the owned link-copy/runtime model, including exact archive
refusal and missing-prerequisite diagnostics. See [the R4 receipt](../verification-evidence/006/pr6-r4/acceptance.json).
The containing commit records one coherent fixture repair. Actual Windows
execution of the repaired snapshots and the full fresh hosted set remain required.
No native terminal or release acceptance gate closes here.
