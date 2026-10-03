# PR #6 complete report batch R1

All six native CI jobs and Kilo's 30-thread review completed on `ced4417` before
any remediation change, as requested by the owner. The batch has 27 change
items and three evidence-based replies. [Verdicts](verdicts.json) and the
[receipt](acceptance.json) distinguish local changes from hosted acceptance.

The fixes cover native tool provisioning and fixture portability, causal policy
checks, provenance compiler pinning, completion flags, executable selection,
interrupted receipts, atomic staging, copied inventory, executable/library and
populated timezone inspection, and shipped documentation links. Native receipts
now bind observed `--version` output; promotion requires the same version for all
five native targets.

The copied-inventory regression replaces an inspected dist archive at the later
source-inventory read boundary. It fails with the old successful return and
passes when finalize rechecks the actual copied inventory. No production test
hook or timing race is needed. Other Red/Green logs retain the intended failure
boundaries. The receipt labels invalid probes and partial runs explicitly.

Current-state canonical validation passes, including race, coverage (95.2%),
generated/docs/notices/workflow checks and ShellCheck. The containing commit
publishes this single combined remediation unit; thread replies and resolution
are verified separately online. This local unit leaves the 51
scenario closures and U6/U5/U7/U8, Phase 6/G4 unchanged. Physical native terminals,
trusted-main candidate provenance, exact-byte cask/performance, public tap and
release publication still require their planned evidence and authority.
