# PR #5 review remediation

### PR #5 hosted review R1 (2026-10-01)

Three stale-write UX findings are fixed with observed failing regressions:
paused form saves show an in-modal Ctrl+R hint; a paused delete supports r and
Ctrl+R, revokes old consent and loads a fresh preview; rejected toggles pause
writes until successful readback. Canonical `make validate build` passes.
Owned Kitty inspection at 100x32 confirms visible stale-form feedback and
retained raw draft. Delete-read abandonment and renewed consent are automated
regressions, not additional Kitty evidence. Receipts: `docs/verification-evidence/005/review-pr5/`.
Remaining hosted threads and Feature 006 native/release gates stay open.

### PR #5 hosted review R2 (2026-10-01)

Seven navigation/input/presentation findings have regression coverage and fixes:
paging counts task rows and group gaps with visible overlap; accepted search is
trimmed; collapse/projection share the end-inclusive predicate; search and filter
paste is atomically bounded and control-checked; picker windows start on titles
and retain the selected ID; details no longer render the form-template editor;
zero due dates display Not set in summary and recovery. Picker offset rounds up
to the next title, rather than down, so both selected lines remain visible at odd
heights. `make validate build` passes. Owned Kitty at 100x32 confirms whitespace
search acceptance, filter presentation and clean quit against an isolated DB.
The focused regression and Kitty receipts are under `review-pr5/`. Lifecycle and
portability review remains open; native/hosted release gates remain Feature 006.
