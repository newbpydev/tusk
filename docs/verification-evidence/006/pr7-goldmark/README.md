# PR #7 Goldmark destination correction

The [receipt](acceptance.json) binds the owner-approved parser change to
Red/Green controls and fresh official Go 1.27.1 canonical validation.
Goldmark was already pinned; its version and dependency inventory are unchanged.

The unchanged baseline scanner accepts the two reported bypasses and an unused
continued reference. Passing controls retain both approved badge URLs and
ordinary local links. Additional controls cover unused/duplicate definitions,
nested image labels and conservative code contexts.

The first canonical attempt was interrupted after a self-audit reproduced
two missing list-code tripwires. Its output and inputs are retained separately.
The final complete run passes 707 named script assertions; helper coverage is
97.8%. Pre-command source hashes and the final outcome-only authority update
are recorded separately. Hosted review/CI and release acceptance remain pending
at this local receipt.
