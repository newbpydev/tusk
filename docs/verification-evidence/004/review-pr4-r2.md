# PR #4 review remediation, round 2

Baseline: `665082f22c82bc54e6f5126905cc79f2c2e56cb9`. All eight new suggestions
were investigated sequentially under the repository instruction. Four are fixed
and four are addressed differently. No Feature 005 implementation was started.
The existing four evidence-based no-change decisions remain in the round 1
receipt; the review summary repeats them without new counter-evidence.

## Signal approach decision (ce-pov)

Choose the current Stop-before-cancel policy with a precise contract. The
alternatives were re-raising a buffered rapid signal and delaying escalation
until a grace timer or stall detector fires. The incumbent restores the prior
disposition before publishing cancellation in `cmd/tusk/signal.go`; the subprocess
test observes that boundary and verifies forced exit when reader cancellation
permanently fails. `threadConfirmation` still joins its reader on graceful exit.
KTD5 requires that join; KTD7 already permits partial output on output failure.

The receive-to-Stop window can absorb a rapid second signal. This is a real limit,
not a guarantee the test proves. [Go signal documentation](https://pkg.go.dev/os/signal)
confirms nonblocking channel delivery and restoration on Stop. [Linux signal(7)](https://man7.org/linux/man-pages/man7/signal.7.html)
confirms that standard signals do not queue. Draining a buffered signal does not
guarantee exactly-two-signal escalation across OS coalescing and would require
platform-specific re-raising. A grace timer also cannot guarantee intact output
if it expires during a flush. Neither extra mechanism is justified by a contract
requiring exact delivery counts or a minimum cleanup grace period; no such
contract exists here. The user-facing docs now state force-exit, partial-output,
inherited-disposition and rapid-signal limits. Revisit if a supervisor protocol
requires a timed grace period or evidence shows the documented escalation fails.

The decision was made by the main agent. No independent cross-model peers ran,
because AGENTS.md maps agent dispatch to sequential main-thread work. That limits
independent corroboration; the implementation, regression and primary sources
are the evidence, not an asserted panel consensus.

## Dispositions

| Thread | Verdict | Evidence and decision |
|---|---|---|
| [4136736392](https://github.com/newbpydev/tusk/pull/4#discussion_r4136736392) | fixed-differently | Clarified the contract in signal.go, docs/cli.md and KTD7: escalation applies to a signal delivered after cancellation is published, with the normal inherited signal disposition. Rapid signals can coalesce; exactly two back-to-back deliveries were never a reliable OS-level guarantee. The dequeue-to-Stop window is real, but draining/re-raising does not establish that stronger guarantee and adds platform-specific termination behavior. Keeping Stop-before-cancel and documenting the boundary; the real-process regression deliberately observes that boundary. See the approach decision in docs/verification-evidence/004/review-pr4-r2.md and the Linux signal(7) queueing rules: https://man7.org/linux/man-pages/man7/signal.7.html. |
| [4136736404](https://github.com/newbpydev/tusk/pull/4#discussion_r4136736404) | fixed-differently | Made the tradeoff explicit in docs/cli.md and KTD7: a later interrupt may force termination during progressing cleanup and may truncate output; users must reopen and read state before retrying mutations. Keeping immediate escalation instead of introducing a grace timer/stall detector. KTD7 already permits partial stdout on failed writes; no intact-output guarantee applies to force termination. Graceful cancellation still joins the reader. The approach decision considered both alternatives; it was evaluated sequentially in the main thread under AGENTS.md, so no independent cross-model panel ran. |
| [4136736409](https://github.com/newbpydev/tusk/pull/4#discussion_r4136736409) | fixed | Separated the test budgets: startup and barriers have a 30-second outer safety limit, while a five-second exit timer begins only after the second signal. Added a four-second startup-delay case, which failed at the ready barrier under the previous three-second budget and now passes. Timeout cleanup kills and joins the child before returning; both SIGINT and SIGTERM remain covered. |
| [4136736420](https://github.com/newbpydev/tusk/pull/4#discussion_r4136736420) | fixed | The trusted timezone diagnostic now states that --timezone overrides TUSK_TIMEZONE and gives UTC as a valid example. The regression asserts that guidance for an explicitly empty flag; rejected values remain private and invalid configuration still opens no service. |
| [4136736432](https://github.com/newbpydev/tusk/pull/4#discussion_r4136736432) | fixed | Removed the unreachable ErrChildrenPresent entry from the trusted-constants list. The dedicated --recursive remediation branch remains authoritative and its existing regression passes; error precedence is unchanged. |
| [4136736450](https://github.com/newbpydev/tusk/pull/4#discussion_r4136736450) | fixed-differently | Added an explicit exemption beside staticReader, as suggested: corruption tests supply one-use rows, while the allocation fixture reuses unmodified unfiltered rows to isolate service allocations from repository cloning. The fake is explicitly unsuitable for filtered or mutating repeated reads. Production TaskReader ownership remains unchanged and real-repository detachment tests continue to pass. |
| [4136736459](https://github.com/newbpydev/tusk/pull/4#discussion_r4136736459) | fixed-differently | Made timingParentSetup fail explicitly for unknown modes, with a misspelled pre-allocated-pipe regression that was red before the change. All nine valid modes retain an independent expected-result table. Thus a typo in the diagnostic loop cannot silently become baseline; no package-level mutable mode list is introduced, respecting AGENTS.md. |
| [4136736467](https://github.com/newbpydev/tusk/pull/4#discussion_r4136736467) | fixed | Added check-modules to validate and documented it in make help. Script fixtures verify both contracts; the validate-recipe assertion failed before the change. The canonical gate now executes go mod tidy -diff and passes without modifying module metadata. |

## Verification

`make validate build check-generated` passes, including full/race tests,
coverage, 37 script fixtures, module metadata and generated checks.
`GOTOOLCHAIN=go1.25.0 make test-unit check-modules` passes. Owned Kitty inspection
verified the explicit-empty-flag diagnostic and recovery with UTC against an
isolated database; the window was closed. See [screen](review-pr4-r2-kitty.png),
[text](review-pr4-r2-kitty.txt) and [source hashes](review-pr4-r2.json).
All automated tests and timing ran in default Codex Bash.

Red/green logs reproduce the delayed-startup failure, misleading timezone hint, silent
unknown diagnostic mode, and missing module gate. The dead diagnostic removal
and fixture exemption are refactoring/documentation under existing passing tests.


### Current-head latency verification remains open

The first complete matrix failed only 1,000-task JSON tree run 2 (p90 16.405353
ms). A bounded complete repeat after closing the Kitty window failed the same
case in run 1 (15.008861 ms). An unchanged control at `665082f` in an isolated
checkout failed that case in run 2 (16.275519 ms). Each report retains all 84
case-runs and 8,400 samples; all p95/p99/maximum guards pass. The control result
shows the failure is not unique to the new review edits. It does not establish
which host/runtime factor caused it or make any failed report a pass.

No query, storage or tree implementation changed in this round. No performance
fix or acceptance-policy change is justified by this evidence alone. The prior
round's passing matrix remains historical; the current binary has not passed
fresh latency acceptance. No retries until green or sample trimming are used.
The second matrix and the control are diagnostic comparisons, not replacements
for the first failed verification. This is a local acceptance blocker separate
from the functional gate and hosted review checks.

- [First matrix](review-pr4-r2-latency.json.gz)
- [Bounded repeat](review-pr4-r2-latency-repeat.json.gz)
- [Unchanged control](review-pr4-r2-latency-control.json.gz)
