# PR #4 review settlement checkpoint

The implementation commit `486dbcc8804c12696f0043b3cfd1475925595b41` has all
32 observed review threads replied to and resolved: 25 code/test/documentation
outcomes, five declined suggestions and two disproved claims. GitHub reports
MERGEABLE/CLEAN, completed passing checks, no actionable backlog and no parked
human decisions. See the [hosted snapshot](review-pr4-settlement.json).
This records observed review/check settlement, not the final quiet-period
readiness declaration. The documentation-only publication is watched afterward;
merge remains the owner's decision.

Canonical `make validate` passed again before this documentation commit. Every
runtime, test, script, SQL and module source hash still matches the accepted
[performance receipt](review-pr4-r3.json). No new latency run is substituted:
the accepted 84-case/8,400-sample matrix remains the proof for unchanged source.
Native console and hosted release obligations remain Feature 006.

## Final allocation review finding

[4137119488](https://github.com/newbpydev/tusk/pull/4#discussion_r4137119488)
claimed the 110% allocation guard had only 1–2% headroom. A temporary Go source
overlay changed only the test predicate from `n > int64(len(data)*11/10)` to
`n > 0`, forcing the existing failure message to print its measured value.
Implementation and fixture bytes were unchanged. These diagnostic probe
failures are intentional, not failures of the actual guard:

- Go 1.27.1: 434,181 bytes allocated for 434,002 wire bytes.
- Go 1.25.0: 434,180 bytes allocated for 434,002 wire bytes.
- The unmodified canonical focused test passes again.

That is about 100.04%, with roughly 43 KB below the 110% ceiling. The installed
Go runtime uses 8 KiB pages for large objects, as verified in runtime/malloc.go
and internal/runtime/gc/sizeclasses.go. The [runtime rounding code](https://go.dev/src/runtime/msize.go)
uses page alignment here, not the small-object class grid assumed in the review.
The byte guard is retained because it catches oversized buffers that an
allocation-count check or implementation-coupled capacity assertion could miss.
A demonstrated future toolchain change remains a reason to reassess; this
measurement disproves the present claim. The verified reply/resolution is the
durable review record. No runtime or test source changed for this finding.
