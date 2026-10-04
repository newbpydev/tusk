# PR #7 Windows hierarchy-rollup fixture correction

The [receipt](acceptance.json) binds the source-CI correction to the retained
native Windows failure, controlled Red/Green and fresh official Go 1.27.1
canonical validation. All 710 named script assertions pass.

The original test passed 30 Linux repetitions but failed natively on Windows.
A fixed clock and reversed UUID entropy reproduce its unsupported row-position
assumption deterministically. Selecting the intended parent by ID passes 20
repetitions and retains every original count/progress assertion and transition.
Production Go source is unchanged. These model checks provide automated evidence;
they do not satisfy physical-console acceptance.

Pre-command source hashes and final outcome-only authority updates are recorded
separately. Fresh Windows source CI and final release acceptance remain pending
at this local checkpoint.
