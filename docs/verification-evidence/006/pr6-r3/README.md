# PR #6 third complete hosted report batch

All seven reports finished on `eaa8791c2f82e1320620354fb2da95f5062cd52c`
before any repair edits. Both macOS jobs fail the same logical/physical temporary
path comparison; Windows passes functional/race tests but misses native coverage.
Kilo's one minimum-compiler suggestion is refuted by that compiler's actual
passing hosted isolated-output test. [Verdicts](verdicts.json) record the batch.

The physical-path Red is reproduced with an owned symlinked Linux TMPDIR. New
argument fixtures assert native binary/report paths and zero measurement launch
on conversion failure. Portable real Go child tests cover process execution;
Windows confirmation tests own real pipes and files, without changing stdin.
Windows EOF normalization follows the documented ReadFile contract. Local
cross-build Red/Green proves the owned-input seam and compilation, not OS runtime.

[Acceptance](acceptance.json) binds the frozen source and final canonical gate.
The next native run must prove Windows confirmation and coverage, plus both
macOS shell fixtures. The 95% coverage and latency thresholds stay unchanged.
Physical native terminal, trusted candidate, exact-byte performance/cask and
release/tap acceptance remain pending. No merge or release mutation occurred.
