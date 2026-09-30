# U6 JSON formatting correction

Final acceptance: after the owner closed Zed and moved the session to Konsole,
the unchanged formatter candidate passes all 84 CLI case-runs. The earlier
76/84 matrix remains `u6-logs/cli-latency-before-konsole.json`; all samples and
limits are preserved. See [final acceptance](u6-acceptance.json). The chronology
below records the diagnosis and earlier failed measurements as they occurred.

The calendar candidate passed 82/84 reference CLI case-runs. The failing
1,000-task JSON tree p90 values were 15.028957 and 15.174335 ms against the
unchanged 15 ms limit. All p95/p99/max guards passed. This is the observed
red integration gate; its complete report is retained in
`u6-logs/cli-latency-calendar-82-of-84.json`.

The diagnostic profile now covers list and tree independently. It attributed
4.58% flat CPU to ASCII string encoding; timestamps also repeated calendar
conversions for validation and formatting. A constant ASCII safety lookup
replaces repeated character comparisons. Go's UTC `Time.AppendText` performs
RFC3339 validation while formatting. Error classification and all JSON bytes
remain unchanged. SQLite locking, validation, fixtures and thresholds are intact.

Existing `TestJSON_EncodingParity` covers every byte, Unicode, HTML-sensitive
characters and four date fields at the supported year boundaries, comparing
complete task/list/nested-tree output with an independent standard-library
oracle. These tests passed before the implementation and after it. The existing
failing latency matrix supplies red evidence; no duplicate behavioral assertion
was added. A focused diagnostic benchmark was added to the canonical Makefile.

Three isolated JSON measurements before: 915819, 1088429, 1080933 ns/op.
Three after: 810543, 877308, 877990 ns/op. Median improved 18.8%; all six
measurements use one allocation and approximately 480 KiB per 1,000-task tree.
These are diagnostic microbenchmarks, not substitutes for fresh-process limits.

`make validate build check-generated` and Go 1.25 focused JSON/output tests pass.
The full Go 1.25 suite/five-target builds passed immediately before this
formatter-only correction. Source API compatibility is exercised again by the
focused minimum-compiler check. All logs are under `u6-logs/json-formatter-*`.

The completed delta review is in `json-formatter-review/review.json`; six lenses
ran sequentially in the parent context under the repository's tool mapping.
No actionable findings. This is not independent-model evidence. The full CLI
matrix must pass before U6 acceptance. The approved TUI/calendar source is
unchanged. Reuse/quality/efficiency simplification found no further warranted
change: the table is constant, timestamp validation uses the standard library,
and the existing fallback still owns escaping and malformed UTF-8.

## Final results

The full current-binary CLI report passes 76/84, not acceptance. Four first-run
help/version cases miss p90 (worst 11.568 ms); four 1,000-task tree cases miss
query p90 (worst 19.110 ms). Run-2 tree JSON p95 also misses at 20.619 ms.
Run-1 invalid-config help p99/max miss at 15.204/17.680 ms. Later help/version
runs pass. This variability is observed evidence, not proof that every miss is
host noise. Zed remains the session host; no editor/process shutdown, runtime
override, threshold change or repeat-until-green is used.

The current binary hash is
`654c9a0d2275d44e3e4da48f97009239da6956e0fe074e7df944de0e5c952fd0`.
Go 1.25 full tests and five-target builds pass after the formatter change.
Fresh owned Kitty CLI inspection reads back the same task ID and exact due
timestamp. The approved workspace is open again; captures are
`u6-kitty/json-formatter-cli.png` and `json-formatter-app.png`.

The checklist audit found V28/V29 still deferred from U2. Existing tests now
explicitly cover real form draft and selected-task retention through all tiny
and negative dimensions, plus fixed borders during refreshing/saving for
0/1/1000 tasks and four profiles. These are verification-only additions; no
production behavior changed. Focused tests and a new canonical validation pass
in `final-resize-cases.log` and `final-checklist-validate.log`. Inline review
confirmed the assertions observe selection, drafts and frame geometry; V28/V29
are closed. V106 and final V110/U6 acceptance remain open pending the owner's
gate-placement decision. The default remains the existing local gate.
