---
feature-id: "006"
plan-source: docs/plans/2026-09-06-006-feat-automation-packaging-and-release-plan.md
verification-plan: docs/verification-plans/2026-09-06-006-feat-automation-packaging-and-release-verification-plan.md
status: Implementation active - U1/U3/U4 locally accepted; U2 next; release gates open
evidence-scope: U1 Red/Green and canonical receipts; hosted/native proof pending
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
| 006-ISS-001 | R21 / repository owner, licensing | P1 | Resolved locally; payload checks pending | MIT grant and complete replacement-aware notices accepted; U2 verifies payload inclusion | Owner decision and U4 receipt; V44 passes locally; V51/V61 remain pending |
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
| 006-ISS-013 | Existing benchmark build prerequisites / performance | P1 | Fixed in plan | Add supplied-binary mode before release measurements; do not overwrite accepted artifacts | Hash-preservation Red/Green and retained candidate CLI/TUI matrices; V85,V86 |
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

---

## Open gate details

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
- **Blocking effect / revisit:** Owner selection is resolved. Grant and reviewed inventory are now accepted in U4. Payload inclusion/native redistribution still gate U2/U5 distribution. No native/publication exception is authorized.

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
