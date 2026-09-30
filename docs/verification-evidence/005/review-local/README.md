# Feature 005 owner-requested local code review

The full branch from 4f4873c78c484da12f766e700bb69b92f46335a9 to
review-dispatch HEAD a39af7af8cf9a4253fdbd8876e64003644a9ab90 was reviewed,
including current working changes. Three sequential main-context rounds fixed
two P2 defects and one P3 glossary inconsistency, then found no further issues.
See [acceptance.json](acceptance.json), [review.md](review.md),
[applied findings](applied-findings.json) and the [review-only patch](review-fixes.patch).

| Finding | Red evidence | Green / final evidence |
| --- | --- | --- |
| #1, added patched-dependency source bypass | dependency-inventory-red.log: both actual integrity guards passed an extra source file | dependency-inventory-green.log, final-validate.log, final-min-go.log |
| #2, action-time date reference | date-reference-red-corrected.log: today/tomorrow edits returned No changes; Apply used yesterday's DST bounds | date-reference-green.log, final-validate.log, final-min-go.log |
| #3, misleading TUI glossary | CONCEPTS descriptions contradicted actual Update/formKey/helpLines | review-fixes.patch; source inspection |

`date-reference-red.log` is a discarded exploratory probe: its UTC service
fixture and fake UI clock/location were inconsistent. It does not establish
the intended defect. Only the corrected, discriminating reproduction is used
for finding #2. The source guard regression copies modules into temporary
fixtures and executes the real test under a timeout; no dependency source in
the working checkout was modified.

Automated commands ran in Bash: canonical validation/build/generated checks,
Go 1.25 focused race tests and five CGO-free application/test cross-builds.
Non-Linux binaries were compiled, not executed. Final CLI and TUI reports retain
all samples and manifest conditions, with every case in every run passing.
The earlier candidate and unchanged control CLI reports both passed 82/84;
these remain failed reports, not accepted substitutes. Their runtime CLI,
service and storage sources were identical. One subsequent complete measurement
passed 84/84 without changing source, limits, host settings or trimming outliers.
`tui-latency-date-fix.json` predates the inventory-test correction and is retained
as historical evidence. `tui-latency.json` matches all 326 final source hashes;
its startup companion contains 300 observations and 15 warmups, not another gate.
The full measurement diff hash is an at-measurement snapshot; later changes are
review documentation only. The accepted binary and source hashes remain exact.

Real CLI/TUI checks ran in an owned 80x24 Kitty window with a temporary database.
The calendar, tomorrow Save/filter, Darkmatter surfaces and terminal restoration
were inspected. The valid calendar/filter screenshots were captured through KDE
Spectacle after verifying the exact owned X11 ID and unchanged active-window
identity. Earlier XGetImage captures returned stale GPU backing pixels and were
discarded. The valid receipt names this limitation; no stale image is retained
here as acceptance evidence. The owned window and control worktree were closed;
other user sessions were untouched. Midnight/DST triggers are deterministic test
evidence; the live app used the normal clock.

Review/validator roles ran sequentially in the main context as AGENTS.md directs,
not as independent agents. The requested Claude Opus peer failed authentication
(HTTP 401) and returned no verified review; actual model/effort and independence
remain unverified. An inline adversarial fallback completed. Stable numbering
comes from the skill's mechanical merger; synthesized metadata removes its
inferred independence labels because the source returns explicitly attest
main-context execution. No confidence or severity was promoted.

The three feature artifacts and MASTERPLAN are synchronized. Prior AGENTS.md
and untracked solution edits are preserved. ce-code-review's dirty-tree rule
leaves these verified fixes uncommitted. No new implementation unit, push, PR,
merge or release is claimed. Native/hosted V111-V112 remain Feature 006 obligations.
