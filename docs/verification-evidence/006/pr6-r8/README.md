# PR #6 eighth complete report batch

The complete current-head set contains six CI reports and Kilo: five CI jobs
and Kilo pass; Windows fails one ambient-identity negative test. All previous
Windows repairs execute and pass, including both no-LFS checkouts, the 57 script
assertions, profile compiler arguments and restricted SQLC fixtures.

The old negative requests windows/amd64 from the actual host, which is a valid
identity on Windows. Use the existing owned windows/amd64 identity and request
windows/arm64 instead. Exit 1 and its exact native-architecture diagnostic are
required. An owned Windows-identity model reproduces the false failure before
repair and exercises the same full shell suite afterward. It models identity
selection in the CI test only; it does not claim native Windows execution.

[Acceptance](acceptance.json), [complete reports](r8-batch-proof.json),
[control recipe](r8-windows-identity-model.sh) and [combined review](combined-review.json)
retain causal proof. Frozen official-Go canonical validation passes; fresh hosted results remain
pending. The containing commit includes one fixture repair and its
compound lesson. Physical native/release gates and the 51 local closures stay open.
