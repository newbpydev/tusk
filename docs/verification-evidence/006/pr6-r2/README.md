# PR #6 second complete hosted report batch

All seven checks finished on `79646a31bae6b95980b8bd518b744c15944a4fb7`
before this repair began. Three Linux/minimum compiler checks pass; Windows
and both macOS jobs fail. Kilo adds one import-policy thread and carries four
summary claims. [Verdicts](verdicts.json) record the combined assessment.

Causal Red/Green covers literal build paths, canonical checkout bytes, vendor
attribute precedence, verification test cross-compilation and import policy
diagnostics. The native job excerpts retain the observed failures for platform
fixtures. New archive/cask/metadata cases raise focused inspector coverage to
96.2% on Linux without changing the 95% gate. A fresh owned checkout with
`autocrlf=true` passes the actual dependency inventories and notices. All-target
native receipt rejection tests also assert that no GitHub calls occur.

[Acceptance](acceptance.json) identifies current verification, input hashes,
invalid probes and evidence limits. Final frozen-state canonical validation passed, including functional/race
checks and 96.2% release-inspector coverage. Publication and the complete next
hosted report set are verified separately.

Windows permission fixtures keep strict read and directory-creation refusal
assertions. They limit privileges only on an owned impersonation thread; their
next native execution remains required. This does not close physical native
terminal, trusted-main candidate, exact-byte performance/cask or release/tap
acceptance. No merge, release, tag, dispatch, settings or tap write occurred.
