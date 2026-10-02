---
feature-id: "006"
plan-source: docs/plans/2026-09-06-006-feat-automation-packaging-and-release-plan.md
verification-plan: docs/verification-plans/2026-09-06-006-feat-automation-packaging-and-release-verification-plan.md
status: PR 6 published; second complete report batch locally validated; fresh hosted/native/release gates open
evidence-scope: 51 local scenario closures; two complete PR 6 report batches locally remediated; final canonical pass; fresh hosted/native/release proof pending
deepened: 2026-10-01
---

# Feature 006 Issue Workorder

This workorder accompanies the [plan](../plans/2026-09-06-006-feat-automation-packaging-and-release-plan.md)
and [verification plan](../verification-plans/2026-09-06-006-feat-automation-packaging-and-release-verification-plan.md).
[MASTERPLAN.md](../../MASTERPLAN.md) controls execution. Planning records do not
replace application, hosted, terminal, license-owner or publication evidence.

**Fixed in plan** means the decision and retest contract are now explicit; the
implementation has not been fixed or tested. **Open decision** requires the
named owner at the named boundary. **Open verification** requires future actual
execution. Owners are responsible roles, not a claim a human has accepted an
assignment. No release exception or performance waiver is assumed.

---

## Issue register

| ID | Source / owner-lens | Severity | Status | Impact / next action | Closure evidence |
| --- | --- | --- | --- | --- | --- |
| 006-ISS-001 | R21 / repository owner, licensing | P1 | Resolved locally; hosted redistribution pending | MIT grant and complete replacement-aware notices accepted; U2 verifies payload inclusion | Owner decision and U4 receipt; V44 passes locally; V51/V61 locally pass; hosted distribution pending |
| 006-ISS-002 | R2,R23,R24 / release maintainer, portability | P1 | Open verification | Arrange all native target/terminal access before U5; absent hosts block release | Exact-asset native runtime and owned terminal records; V73–V84 |
| 006-ISS-003 | R29 / repository owner, distribution | P1 | Open decision | Establish accessible tap/destination before U7; proposed tap was not found | Owner-controlled repository/path and Intel/ARM local/live cask proof; V87–V98 |
| 006-ISS-004 | R28 / repository owner, release | P1 | Open decision | Confirm version/SHA and explicit hosted/release authority at concrete U8 candidate | Authorized identity, draft/asset readback and actual public result; V95–V102 |
| 006-ISS-005 | R22 / maintainer, security/support | P2 | Verified locally; U8 activation pending | Private reporting disabled; verified owner profile/public request route avoids confidentiality claim | U4 API readback and SECURITY.md; V42/V44 local pass; V100 pending |
| 006-ISS-006 | Original outline / coherence | P1 | Fixed in plan | Replace unsupported readiness with executable unit and evidence contracts | Eight unit fields, 30 requirements and 102 pending scenarios; full execution still required |
| 006-ISS-007 | Original Go 1.24 assumption / compatibility | P1 | Fixed in plan | Keep Go 1.25.0 source floor and separate pinned production compiler | Native/minimum jobs and tool lock; V03,V04,V12,V49 |
| 006-ISS-008 | CI/race/format / portability, testing | P1 | Fixed in plan | Distinguish native Windows Bash/Make/compiler, CGO race and CGO-free release; fail source drift | Negative prerequisites, native jobs and drift fixtures; V05–V11,V16 |
| 006-ISS-009 | Local replacements / install usability | P1 | Fixed in plan | Replace unsupported versioned go-install advice with source-checkout and binary routes | Fresh source and install-example replay; V14,V32–V34,V53,V84 |
| 006-ISS-010 | Constant version / architecture | P1 | Fixed in plan | Use generated constant overlay, explicit linker settings and actual binary version assertions | Injection/determinism/checkout-isolation fixtures; V45–V47,V58,V59 |
| 006-ISS-011 | Deprecated binary formula / distribution | P2 | Fixed in plan | Generate current macOS cask with upload disabled; separate tap publication | Cask config/native/local/live install and removal; V87–V94,V98 |
| 006-ISS-012 | Candidate/rebuild/provenance / security, evidence | P1 | Fixed in plan | Build once, attest, test exact downloaded bytes and promote existing accepted files | Trusted-run/hash/provenance/tamper/readback matrix; V63–V72,V95–V98 |
| 006-ISS-013 | Existing benchmark build prerequisites / performance | P1 | Resolved locally; candidate matrices pending | Add supplied-binary mode before release measurements; do not overwrite accepted artifacts | Hash-preservation Red/Green and retained candidate CLI/TUI matrices; V85,V86 |
| 006-ISS-014 | Inherited native/release handoffs / coherence | P1 | Fixed in plan | Map product V66–V73 and Feature 002–005 release obligations without closing them in planning | Whole inherited contracts mapped to actual receipts; verification handoff table |
| 006-ISS-015 | Install/backup/uninstall / data integrity | P1 | Fixed in plan | Preserve DB/WAL/SHM and compare domain data around replacement/restore | Native current/newer-schema/backup/remove fixtures; V39,V79–V81,V91 |
| 006-ISS-016 | Missing repository metadata / documentation | P2 | Fixed in plan | Use concrete About/homepage/topics/social preview and inspect actual rendered GitHub | Preview plus before/after API/browser evidence; V42,V43,V100,V101 |
| 006-ISS-017 | Public support/performance/image claims / product/privacy | P2 | Fixed in plan | Cite actual dated evidence, sanitized captures and verified badges/install routes | Documentation negative checks and released README replay; V40,V41,V99,V102 |
| 006-ISS-018 | Tool/analyzer/runner availability / operations | P1 | Open verification | Freeze verified pins in U1 and prove analyzer compatibility before release; a missing native job stays pending | Verified tool/action digests, exact compiler and successful analysis/native URLs; V03,V12,V13,V60,V65 |
| 006-ISS-019 | Partial publication/retries / reliability | P1 | Fixed in plan | Reconcile lost responses; preserve published bytes and use a new version for repairs | API failure fixtures and complete draft/public readback; V70,V95–V98,V102 |
| 006-ISS-020 | Review host mapping / evidence quality | P2 | Recorded constraint | Execute applicable lenses sequentially in main thread; do not claim independent agreement | Review-lens coverage below and explicit independence limitation |
| 006-ISS-021 | U1 formatting / portability | P1 | Fixed locally | Space/Unicode filenames were split by xargs; replaced with literal find -exec arguments | Observed lstat failures, CRLF/path Green and both compiler canonical gates; U1 receipt |
| 006-ISS-022 | U1 native identity / correctness | P1 | Fixed locally | Reject target/host mismatch and Windows Go invoked through a Linux/WSL shell | Both impersonation Red fixtures, native preflight Green and fresh canonical gates; U1 receipt |

| 006-ISS-023 | U3 completion protocol / CLI contract | P1 | Fixed locally | Cobra debug stderr was sharing the buffered stdout writer; discard library-only debug and preserve safe invocation diagnostics | Stream/protocol Red/Green, all three widgets and canonical gate; U3 receipt |
| 006-ISS-024 | U3 stale generated member / correctness | P2 | Fixed locally | Reject extra directories, symlinks and unequal flat output members | Extra-directory Red/Green and atomic repair/recovery fixtures; U3 receipt |
| 006-ISS-025 | U3 manual syntax / usability | P2 | Fixed locally | Escape angle-bracket arguments in detached documentation tree so md2man retains syntax | Synopsis Red/Green and inspected before/after manual captures; U3 receipt |
| 006-ISS-026 | U3 ordinary startup / efficiency | P2 | Fixed locally | Limit Cobra global callback registration to hidden shell completion requests | Observed ordinary-tree registration Red, lazy Green and canonical gate; U3 receipt |

| 006-ISS-027 | U4 license inventory / supply chain | P1 | Fixed locally | Block omitted assets, nested SQLite grants and Go/timezone notices in addition to unknown/changed/graph-mismatched entries | Omission Red fixtures, complete-set Green and current canonical gate; U4 receipt |
| 006-ISS-028 | U4 generated notices / portability | P2 | Fixed locally | Set readable 0644 permissions on staged notice before promotion | Permission Red/Green and minimum/compiler script gates; U4 receipt |

| 006-ISS-029 | U2 packager contract / supply chain | P1 | Resolved locally | Reject hooks, publishers, extra builds/files and checksum overrides through an exact structured config contract | Observed Red/Green, real packager and final canonical/minimum receipts in u2.json; foreign/hosted acceptance separate |
| 006-ISS-030 | U2 manifest inputs / integrity | P1 | Resolved locally | Bind exact required input keys and hashes to source members and the compiler lock | Observed Red/Green, real packager and final canonical/minimum receipts in u2.json; foreign/hosted acceptance separate |
| 006-ISS-031 | U2 trailing archive data / integrity | P1 | Resolved locally | Finish gzip CRC verification with bounded zero padding; reject hidden trailing streams and unsafe global metadata | Observed Red/Green, real packager and final canonical/minimum receipts in u2.json; foreign/hosted acceptance separate |
| 006-ISS-032 | U2 source prerequisites/modes / portability | P2 | Resolved locally | Install pinned sqlc in owned checkout, fix source tar umask/LF and derive executable modes from Git | Observed Red/Green, real packager and final canonical/minimum receipts in u2.json; foreign/hosted acceptance separate |
| 006-ISS-035 | U6 ID download layout / correctness | P1 | Resolved locally; hosted proof pending | Explicit merge-multiple extracts the unique selected artifact at its reviewed root | Official pinned Action source and observed layout contract Red/Green; hosted execution pending |
| 006-ISS-036 | U6 verifier output / data integrity | P1 | Resolved locally; hosted proof pending | Require new retained output outside candidate and Git storage | Observed verifier accepted and modified candidate storage; boundary regression Green |
| 006-ISS-037 | U6 workflow skip/input controls / security | P1 | Resolved locally; hosted proof pending | Reject error-skipping/conditional steps, changed source env and unexpected job/step fields | Three accepted unsafe workflow mutations observed Red; exact structured contract Green |
| 006-ISS-040 | U7 cask audit and provenance / supply chain | P2 | Resolved locally; hosted/native pending | Bind/upload/attest generated cask separately; audit pinned statement order/comment and preserve failures before cleanup | Real packager A/B Red and C Green, missing/tampered cask fixtures, canonical validation; u7.json |
| 006-ISS-041 | U8 current-main/source authority / correctness | P1 | Resolved locally | Refuse changed remote main build inputs and failed local Git diff; allow only later governance prose | Observed Red/Green and final canonical gate; u8.json |
| 006-ISS-042 | U8 retained measurement completeness / evidence | P1 | Resolved locally | Unique full three-run case matrices, recomputed reference tail guards and numeric nonnegative memory samples | Duplicate-case and malformed sample Red/Green; u8.json |
| 006-ISS-043 | U8 tag/tap preflight / release integrity | P1 | Resolved locally; actual tap open | Require live tap readiness and exact accepted source tag in a completed authorized draft | Missing-tap/tag fixtures; actual tap404 remains 006-ISS-003; u8.json |
| 006-ISS-044 | U8 GitHub host/log boundary / security | P1 | Resolved locally | Pin public host, literal API/upload endpoints and remove inherited HTTP debug flags; no filename-label or clobber parser | API-host/host/debug Red/Green, structured binary upload and final canonical gate; u8.json |
| 006-ISS-045 | Final release-tool reliability / bounded processes | P2 | Resolved locally | Shared portable driver bounds every maintainer gh child; streams/exit codes and write readback remain intact | Actual review #2, owned child tests, canonical/minimum/lint; review-local/acceptance.json |
| 006-ISS-046 | Final native receipt / exact executable and archive identity | P2 | Resolved locally; actual native acceptance pending | Require native target/executable/archive fields to match the verified manifest | Actual review #1, four malformed-binding Red/Green, current promotion fixtures; review-local/acceptance.json |
| 006-ISS-047 | Candidate strict JSON / parser agreement | P2 | Resolved locally | Require EOF after one decoded candidate/overlay value, retaining whitespace and unknown-field checks | Actual verify-release appended-object reproduction, four trailing-data Red/Green, canonical race/coverage; review-local/acceptance.json |
| 006-ISS-048 | Renewed review / ambiguous JSON approval and acceptance | P2 | Resolved locally | Require exactly one object before every boundary predicate; reject failed record followed by passing record | Seven parser/metadata/smoke Red/Green cases; review-r2/acceptance.json |
| 006-ISS-049 | Renewed review / native policy test hash prerequisites | P2 | Resolved locally; physical native proof pending | Portable test-only hashing with GNU or shasum; no new mandatory Perl/coreutils prerequisite | Shasum-only PATH Red/Green and minimum-Go script suite; review-r2/acceptance.json |
| 006-ISS-050 | Focused fix review / supplied native binary path failure | P2 | Resolved locally; physical native proof pending | Propagate cygpath conversion failure before exporting selected executable | Corrected Windows fixture Red 0 vs 1 and Green refusal; strict lint; review-r2/acceptance.json |
| 006-ISS-039 | U5 evidence output ownership / integrity | P2 | Resolved locally | Reject smoke/measurement outputs inside candidate/source storage; normalize physical absolute output paths | Observed boundary Red/Green, canonical validation; u5.json |
| 006-ISS-038 | U6 minimal API repository / contract | P2 | Resolved locally; hosted proof pending | Fetch current repository/default branch separately from minimal Actions run repository | Official REST schema and minimal-run fixture Red/Green; real trusted candidate pending |
| 006-ISS-034 | U2 special archive modes / security | P2 | Resolved locally | Reject setuid/setgid/sticky bits for every file/directory before payload acceptance | Observed Red/Green, real packager and final canonical/minimum receipts in u2.json; foreign/hosted acceptance separate |
| 006-ISS-033 | U2 literal distribution path / correctness | P1 | Resolved locally | Generate external runtime config changing only literal dist to owned storage; retain canonical hash and actual override separately | Observed Red/Green, real packager and final canonical/minimum receipts in u2.json; foreign/hosted acceptance separate |

---

## Open gate details

U2 owns the build-time `scripts/releasecheck/` Go inspector behind its shell
wrapper. Standard-library archive/binary parsing avoids platform-specific tools;
the package retains canonical coverage. Raw Make release parameters require
injection rejection before build. Private local validation commits/artifacts
are preliminary engineering evidence, not trusted hosted release inputs.

U1's [local receipt](../verification-evidence/006/u1.json) binds source/tool hashes,
Red/Green, sequential review, Go 1.27.1 and Go 1.25.0 canonical gates and five
minimum-Go cross-builds. Tool/archive/Action pins and release-compiler analyzer
support are locally verified for 006-ISS-018; five native hosted job URLs and
candidate-specific acceptance remain pending. U1 is committed as 080426e.

U3's [local receipt](../verification-evidence/006/u3.json) closes V17–V30 on Linux
with current-source canonical/minimum checks, deterministic generated output and
owned Kitty inspection. All three sequential simplification lenses were applied;
ordinary startup callback registration was removed. The owned window was closed.
Native Windows ACL and hosted runtime proof remain pending. U3 is committed as e543502.

U4's [local receipt](../verification-evidence/006/u4.json) closes local V31–V44
with the MIT grant, complete notices, docs negative fixtures, actual-process
backup/quick-start rehearsal, fresh source install and Chrome preview. Public
release/metadata/native-install acceptance remains pending. U2 follows the U4 commit.

### 006-ISS-001: Project license and redistribution rights

- **Phase found:** Planning, 2026-10-01; owner/lens: repository owner/licensing; P1.
- **Affected:** R21, U4/U2/U7/U8; V44,V51,V61,V93,V95.
- **Evidence:** No first-party LICENSE exists and live GitHub `licenseInfo` is null. Vendored patches/assets retain separate upstream grants.
- **Expected:** A selected owner-authorized first-party license plus complete dependency/asset notices accompanies redistributed files.
- **Decision:** The owner answered “MIT; rights confirmed” on 2026-10-01. U4 may add the MIT grant; retain third-party texts and patch provenance. An incompatible/unclassified obligation requires investigation before packaging distribution.
- **Closure:** Owner decision recorded without credentials, license detection after publication, complete archive/source notices and replacement-aware inventory checks.
- **Blocking effect / revisit:** Owner selection is resolved. Grant and reviewed inventory are now accepted in U4. U2 payload inclusion is locally verified; actual hosted/native redistribution remains U5/U8. No native/publication exception is authorized.

### 006-ISS-002: Native platforms and real terminal access

- **Phase found:** Planning; owner/lens: release maintainer/portability; P1.
- **Affected:** R2,R23,R24, U1/U5; V03,V73–V84 and inherited Feature 002–005 obligations.
- **Evidence:** Prior local cross-builds exist. Feature 004 V90–V91, Feature 005 V111–V112 and remaining product TUSK-V66 platform coverage are still unchecked.
- **Decision plan:** Use five native hosted runners plus owned Linux/macOS Kitty and Windows 11 Windows Terminal app sessions. Record OS/arch/window/terminal identity and exact downloaded byte hashes. A console-less hosted runner supplies runtime proof, not visual acceptance.
- **Closure:** Required native tests and inspected terminal observations/captures at supported dimensions, with full artifact identity.
- **Blocking effect / revisit:** Blocks release acceptance/U8 until all required results exist. Revisit when scheduling U5; unavailable desktop/host remains pending without waiver.

### 006-ISS-003: Homebrew destination and ownership

- **Phase found:** Planning; owner/lens: repository owner/distribution; P1.
- **Affected:** R29, U7/U8; V87–V98.
- **Evidence:** Authenticated API lookup of proposed `newbpydev/homebrew-tap` returned 404; existence/access is unverified.
- **Decision plan:** Owner establishes or supplies the controlled tap/path. Generate cask without upload; verify native local installs, then explicitly commit/push the reviewed cask after the public release URLs exist. Avoid a new permanent cross-repository token.
- **Closure:** Accessible destination/branch, reviewed exact cask, native Intel/ARM install/remove data preservation and anonymous live-tap result.
- **Blocking effect / revisit:** Blocks U7 hosted distribution readiness and final settlement; local package/docs preparation can proceed. Revisit at U7 intake. Direct-download success alone does not waive planned Homebrew delivery.

### 006-ISS-004: Version, candidate identity and publication authority

- **Phase found:** Planning; owner/lens: repository owner/release; P1.
- **Affected:** R27,R28,R30, U6/U8; V63–V72,V95–V102.
- **Evidence:** GitHub has no releases/tags/workflows; current code reports `0.2.0-reboot`. This invocation is planning authority.
- **Decision plan:** Proposed first release is v0.3.0. Confirm actual version/full SHA and desired hosted actions against the concrete accepted manifest. Keep pushes/workflow dispatch, tags, draft/public release and metadata changes within that authorization; planning/local commits imply none of them.
- **Closure:** Accepted CI/native/performance/license receipts, explicit owner action scope, exact draft bytes/readback, published immutable identity, public install/tap and GitHub content readback.
- **Blocking effect / revisit:** Blocks unauthorized hosted actions and U8 publication. Local red/green/packaging/preview work is not blocked. Revisit before any hosted action and at U8 final candidate audit.

### 006-ISS-005: Security-reporting contact and settings

- **Phase found:** Planning; owner/lens: maintainer/security-support; P2.
- **Affected:** R22, U4/U8; V42,V44,V100.
- **Evidence:** No SECURITY.md exists; private vulnerability reporting availability/activation was not verified. An admin role does not establish a contact SLA or security team.
- **Decision plan:** Verify and document GitHub private reporting if enabled; otherwise give an accurate owner GitHub contact route and state that public issue content is public. Do not fabricate email/private support promises.
- **Closure:** Tested real contact links, approved concise policy and hosted feature readback.
- **Current verification:** GitHub API returned `enabled:false` on 2026-10-01; the owner profile returned the expected login/URL. SECURITY.md states disabled private reporting and provides a public request route without confidentiality. V42/V44 pass locally; V100 activation/readback remains pending.
- **Blocking effect / revisit:** Recheck the setting and documented route before U8 activation.

### 006-ISS-018: Verified release tools and native runner availability

- **Phase found:** Planning; owner/lens: CI/release maintainer/operations; P1.
- **Affected:** R2,R5,R9,R26, U1/U2/U6; V03,V12,V13,V60,V65.
- **Evidence:** Official release feeds currently expose Go 1.27.1, GoReleaser v2.18.2 and actionlint v1.7.12. Current runner documentation lists the five selected labels. Availability does not prove this graph's compatibility, analyzer support or exact executable digests.
- **Decision plan:** U1 verifies/downloads pins and records upstream Action SHAs/archive digests. Prove vulnerability analyzer support for the release compiler. Pin changes require synchronized review; preserve the minimum-Go and dependency graph independently.
- **Closure:** Exact versions/digests and supported analyzer results; five native canonical/build jobs and minimum-Go proof for the candidate SHA. No floating fallback or silently skipped unsupported job.
- **Blocking effect / revisit:** Local tool execution must pass before its consuming unit; missing current analysis/native proof blocks release. Revisit U1 bootstrap and final candidate preflight.

---

## Corrected planning findings and retest expectations

The original readiness label was not a sufficient plan. It lacked unit ownership,
failure contracts, platform prerequisites, data-preserving installation and public
onboarding. The plan now assigns stable U1/U2/U3 identities and adds U4–U8 in
dependency order. Findings 006-ISS-006 through 017 and 019 have concrete corrections
and scenario owners, while their runtime closure evidence remains unexecuted.

The main-thread review identified an additional concrete evidence failure:
`bench-cli` and `bench-tui` depend on `build`, which would overwrite a supplied
candidate binary. U5 must add a tested consume-existing mode and preserve binary
hashes before/after timing. Add failing hash-preservation fixtures first; then
retain the full reference matrices and candidate startup measurements. The
Feature 005 PR #5 Ryzen calibration remains scoped to its original acceptance
unless the owner explicitly authorizes release use.

The complete-draft/promotion contract avoids a second rebuild after native
acceptance and avoids relying on a nonexistent presumed GoReleaser OSS
publish-existing-artifacts command. The current cask path supersedes deprecated
binary-formula advice while the source/ZIP/tar routes remain available. The
metadata payload is concrete and bounded; a separate Pages website is deferred.

The sequencing review also requires final source freeze after release helpers,
measurement selectors and cask/test drivers are locally validated and committed.
A changed input after the initial U6 candidate requires a new trusted candidate
and affected U5/U7 acceptance reruns. This correction prevents later test/config
units from leaving publication tied to stale candidate evidence. Receipt/README
updates may have a later separately recorded documentation SHA.

Any issue discovered during implementation receives a new stable issue ID with
reproduction, expected/actual behavior, affected requirement/unit/scenario,
red/green/canonical evidence, owner and next action. Do not relabel these planning
corrections as executed fixes or suppress fresh candidate failures.

---

## Review-lens sign-offs

These are document-review dispositions, not implementation acceptance. Repository
tool mapping requires sequential main-thread execution. No independent peer,
sub-agent or cross-model agreement is claimed.

| Lens | Planning result | Basis / corrections | Execution retest |
| --- | --- | --- | --- |
| Architecture/dependency sequencing | Reviewed inline | Eight units preserve original IDs; completion/license precede packages, hosted candidate precedes exact-byte native proof | Unit-owned changes and candidate flow |
| Product/scope and simplicity | Reviewed inline | Public installation/README/metadata fulfill user request; no new website, service, updater or registries; manual tap write avoids credential infrastructure | README/install usability and actual scope |
| Coherence/feasibility | Reviewed inline | Canonical Make interfaces, tool minimum/release distinction, native target matrix and owned external gates align | Links/count/coverage audit plus all native prerequisites |
| Correctness/reliability | Reviewed inline | Strict version/manifest checks, no rebuild promotion and partial-write readback/idempotency are explicit | Negative fixtures/API loss/cancellation |
| Test strategy/evidence quality | Reviewed inline | 102 unchecked scenarios; local/hosted/native/publication tiers; supplied-binary benchmark defect corrected in contract | Observed Red/Green and exact candidate receipts |
| Security/supply chain | Reviewed inline | Read-only PRs, full-SHA pins, trusted artifact provenance, no implicit secret/tap access or shell interpolation | Pin verification, tamper/fork rejection and credential boundaries |
| Data integrity/lifecycle | Reviewed inline | No schema work; preserve DB/WAL/SHM, offline restore and newer-schema refusal | Native semantic/byte comparisons and unknown readback |
| Performance | Reviewed inline | Existing release measurement budgets retained; no historical/calibrated result silently expanded; prevent candidate rebuild | Retained CLI/TUI reports and binary hashes |
| Portability/terminal lifecycle | Reviewed inline | Five native runners, Windows native compiler/console and real Kitty surface distinction | Exact native binaries, shells and inspected terminal captures |
| Documentation/accessibility/privacy | Reviewed inline | Platform install routes, readable screenshot/alt text, no fabricated contacts/badges and actual GitHub render readback | Replayed examples, sanitized captures and public readback |

---

## Planning audit and implementation release gate

Planning audit checks the readable triplet, 30 requirements, eight stable unit
contracts, ordered dependencies, 102 unique unchecked scenarios, coverage and
links, planned Make command owners and explicit release/decision boundaries.
The product triplet, registry and masterplan are synchronized with these
contracts. `git diff --check` and document-only validation are planning checks;
they do not establish application or release acceptance.

Observed planning result on 2026-10-01: the temporary `planning-audit` Make target
passed 30 requirement IDs/coverage rows, eight ordered unit contracts, 102 unique
unchecked scenarios, 20 issue-register IDs and 185 local links/anchors. It also
confirmed all 73 historical product scenario check states were preserved, all
eight Phase 6 implementation units remained unchecked and the eight changed files
were planning/governance documents only. `git diff --check` passed. No application
tests, builds, benchmarks or hosted mutations ran; temporary audit files were
removed after the check.

Leave these items unchecked until execution:

- [ ] U1 → U3 → U4 → U2 → U6 → U5 → U7 → U8 implemented with coherent per-unit commits.
- [ ] Observed Red/Green and focused tests retained for all feature-bearing code/config changes.
- [ ] Fresh canonical validation, coverage/race, generated and module gates pass before every commit.
- [ ] Minimum compiler and five native candidate-SHA quality/build jobs pass.
- [ ] Bash/Zsh/Fish and deterministic manual generation pass with no storage/terminal initialization.
- [ ] Complete payload/source/license/notice/checksum/version/provenance evidence passes.
- [ ] Owner license/rights, native access, security contact and tap destination gates closed.
- [ ] Exact candidate bytes pass native CLI/TUI/storage/install/backup/upgrade/remove checks.
- [ ] Owned Kitty/native Windows Terminal observations and all retained CLI/TUI release measurements accepted.
- [ ] Authorized draft/public release, reviewed cask and anonymous install verification complete.
- [ ] GitHub metadata/social preview, rendered README/badges/links and license readback accepted.
- [ ] Final receipt matches binary and documentation identities; inherited product/feature gates synchronized with actual evidence.
- [ ] Remaining unaccepted release issues: 0.

No implementation/release checkbox is closed by this planning pass. Open owner
decisions are addressed at their explicit unit boundaries; they are not hidden
assumptions or release waivers.

### U2 local engineering checkpoint — 2026-10-01

V45–V62 pass locally. [Receipt](../verification-evidence/006/u2.json) retains all
failed and successful packaging/component measurements, current canonical/minimum
gates and owned Kitty observations. Inspector coverage is 95.9% without exemption.
The source archive builds with minimum Go; all nine B/C assets reproduce exactly.
Preliminary private validation SHAs are not final source acceptance. U6 hosted
provenance, foreign native execution/performance and publication remain open.

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
