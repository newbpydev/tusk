# PR #6 ninth complete report batch

All six CI reports and Kilo finished on fbde2f8 before any edit or reply. Five CI
jobs pass; Windows fails a later notices positive control. Kilo's one new finding
identifies lost OS-operand negative coverage. This unit repairs both together.

The OS case uses the same passing owned windows/amd64 identity and requests
linux/amd64. Exit 1 and its exact job-identity diagnostic are required. A dropped
OS predicate escapes the original full shell suite and is detected after repair.

A permanent notices fixture models jq's Windows LF-to-CRLF output translation
and reproduces the exact incomplete-inventory refusal. Explicit binary output
across maintainer/fixture callers and the Make compiler lookup makes the complete
controlled shell suite pass. jq 1.7+ is a documented build-only prerequisite.
No JSON predicate, source grant bytes, product limit or Go runtime source changes.
Actual native lists were not logged, so CRLF attribution remains a supported
hypothesis until fresh native execution. The broader Linux model's bootstrap
failure is retained separately from the actual Windows notices failure.

[Acceptance](acceptance.json), [complete reports](r9-batch-proof.json),
[OS mutation control](r9-os-guard-model.sh), [JSON-output control](r9-jq-crlf-model.sh)
and [combined review](combined-review.json) preserve evidence and limits.
The 51 local closures and physical native/trusted-main/exact-byte release gates
remain open. Fresh hosted proof is required after the containing repair commit.
