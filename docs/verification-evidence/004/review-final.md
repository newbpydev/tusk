# Feature 004 final local code review

Reviewed `main` (`6a8aff0`) through `c336461` on
`feat/cli-interface-and-scripting`. User authorized fixing all P0–P2 findings and repeating
review to a clean pass. No PR existed at review time.

## Applied findings

| Finding | Severity | Original location | Resolution |
|---|---|---|---|
| #2 | P2 | docs/cli.md:8 | Correct runtime scheduling contract; mark the superseded single-processor trial as historical in all three plan documents. |
| #1 | P2 | docs/cli.md:148 | Correct verification surface: automated Makefile checks and measurements in Codex Bash; required visible terminal scenarios in an owned Kitty window. |

The stale Goal Capsule latency status was reconciled with the existing U6
acceptance. Historical failed measurements retain their original outcomes.
Round 2 reviewed the entire applied documentation diff: **zero open P0–P2**.
No executable-code defect was confirmed. Runtime and test source are unchanged.

## Verification

Fresh `make validate build check-generated` passed in Codex Bash, including full
and race tests, coverage, 34/34 script fixtures and generated schema/sqlc checks.
[Compressed gate log](review-final-gate.log.gz).
All **158** source hashes match [U6 acceptance](u6-acceptance.json).
The accepted latency report and Kitty artifacts remain historical evidence for
that unchanged implementation; this review did not claim new measurements or
new terminal inspection. Documentation corrections required no new behavior test.

## Coverage

Local correctness, project standards, maintainability, security, reliability,
API contract, performance, data migration, testing, agent parity and past-learning
passes ran sequentially in the main context, following the repository tool map.
They are distinct review lenses, not independent model reviews.

Cross-model adversarial coverage is degraded: Claude (requested claude-opus-5-5, high) returned a provider authentication error; the one replacement, Grok via Cursor (requested grok-4.7-xhigh, model-implied effort), timed out at the shared deadline without a usable artifact. Neither attempt counts as a completed independent review. Requested serving identities were not verified. The replacement was reaped and both job directories were removed.

Both documentation findings were confirmed by re-reading the guide, the actual
entrypoint, `TestRuntimePolicy`, AGENTS.md and the matching triplet paragraphs.
Validator counts: 2 confirmed; 0 rejected, unresolved, malformed, failed or
shortcut-skipped. Zero confidence suppressions, missing-quote demotions, evidence
backfills or mode-aware demotions. No confidence promotion from local reviewer agreement.
No retained residual risks or testing-gap findings. Native Windows/macOS execution
and hosted release gates remain intentionally assigned to Feature 006.

Runtime review covered parsing/admission, all eight durable commands, JSON and
human output, close/commit/error semantics, terminal consent/cancellation,
query/tree/sort ownership, strict storage codecs, generated schema validation,
real-process recovery and the benchmark harness. Large historical logs and
measurements were inspected selectively. No migration DDL change was introduced.

## Requirements completeness

Every Feature 004 requirement has corresponding implementation/evidence.

| Requirement | Status | Evidence |
|---|---|---|
| R1 | met | internal/cli/root.go; root_test.go; cmd/tusk/process_linux_test.go |
| R2 | met | internal/cli/root.go; flags.go; root_test.go |
| R3 | met | internal/cli/config.go; cmd/tusk/app.go; app_test.go |
| R4 | met | internal/cli/root.go; cmd/tusk/app.go; root_test.go |
| R5 | met | go.mod; cmd/tusk/app.go; internal/cli/compatibility_test.go; U6 acceptance cross-build receipt |
| R6 | met | internal/cli/add.go; mutations_test.go |
| R7 | met | internal/cli/edit.go; mutations_test.go |
| R8 | met | internal/cli/done.go; edit.go; mutations_test.go |
| R9 | met | internal/cli/delete.go; delete_test.go; cmd/tusk/confirm_test.go |
| R10 | met | internal/cli/list.go; queries_test.go; internal/service/queries.go |
| R11 | met | internal/cli/tree.go; queries_test.go; internal/core/tree.go |
| R12 | met | internal/cli/stats.go; history.go; queries_test.go |
| R13 | met | internal/cli/json.go; json_tasks.go; json_test.go |
| R14 | met | internal/cli/output.go; json.go; output_test.go |
| R15 | met | internal/cli/root.go; errors.go; output.go; output_test.go; process_test.go |
| R16 | met | internal/cli/root.go; root_test.go; output_test.go |
| R17 | met | internal/cli/format.go; format_test.go; U6 accepted Kitty evidence |
| R18 | met | internal/cli/sanitize.go; errors.go; format_test.go; json_test.go |
| R19 | met | internal/cli/terminal.go; cmd/tusk/terminal.go; format_test.go; process_linux_test.go |
| R20 | met | cmd/tusk/confirm*.go; internal/cli/delete.go; delete_test.go; process_linux_test.go |
| R21 | met | internal/cli/process_test.go; cmd/tusk/process_linux_test.go |
| R22 | met | scripts/cli-bench/; Makefile; docs/verification-evidence/004/u6-latency-accepted.json (retained acceptance, not freshly measured) |
| R23 | met | internal/cli/compatibility_test.go; U6 acceptance cross-build receipt; native/hosted gates intentionally deferred |
| R24 | met | internal/cli imports; docs/cli.md; shared ports.TaskService |
| R25 | met | unit acceptance receipts; branch unit commits; synchronized triplet and MASTERPLAN.md |

### Implementation units

- [x] U1: Feature 004 U1 acceptance entry and unit commit; u1-checkpoint.json is historical
- [x] U5: docs/verification-evidence/004/u5-accepted.json
- [x] U4: docs/verification-evidence/004/u4-accepted.json
- [x] U2: docs/verification-evidence/004/u2-accepted.json
- [x] U7: docs/verification-evidence/004/u7-accepted.json
- [x] U3: docs/verification-evidence/004/u3-accepted.json
- [x] U6: docs/verification-evidence/004/u6-acceptance.json

Reverse scope check found no unrequested behavior rule. Native/hosted release
proof is an intentional deferred obligation, not an omitted local unit.

## Known pattern

[Preserve transaction outcomes through error redaction](../../solutions/database-issues/preserve-transaction-outcomes-through-error-redaction.md):
CLI diagnostics preserve unknown outcomes ahead of safe causes; tests cover
value/pointer errors, no mutation replay and fresh-owner readback.
No durable-action gap was found in the scriptable CLI surface.

## Verdict

Clean local P0–P2 pass after two documentation fixes. This is local review
acceptance; no publication, hosted review, native release validation or merge
was performed. Feature 005 remains planning-only.

Actionable findings: none.

[Structured receipt](review-final.json) and [detailed review records](review-final-details.tar.gz)
contain scope, dispositions, validation and follow-up checks.
