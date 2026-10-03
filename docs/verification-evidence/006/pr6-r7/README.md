# PR #6 seventh complete report batch

All seven reports completed on `5c3f63ff9d818071deb6585b9ea6b5c0b12f9349`:
two macOS jobs, three Linux jobs and Kilo passed, with no open review threads.
Windows failed the unchanged no-LFS checkout before reaching the shell fixtures.
One failed-only same-head retry reproduced Git staging failure near ten seconds;
that outcome invalidated retry alone as a sufficient transient-failure remedy.

The test shared one ten-second context across Git init, staging and checkout.
Each operation now receives its own bounded 30-second resource context, canceled
immediately afterward. Failures include the context error captured before cancel
and elapsed time. This explicitly changes the test resource budget; every Git
argument, required failing LFS filter, pointer scan and product latency/coverage
guard remains unchanged. No persistent artificial delay is added to the tests.

A five-second delay before each real Git command reproduces Red at 10.02 seconds
and passes the same full checkout after repair in 15.15 seconds. An exec-replaced
stalled command still fails at 30.0279 seconds with context deadline exceeded.
Frozen canonical validation with official Go 1.27.1 passes. These controlled Linux
checks do not replace fresh Windows or physical terminal acceptance.

[Acceptance](acceptance.json), [complete reports and initial hypothesis](r7-batch-proof.json),
[delayed-Git recipe](r7-slow-git-recipe.json), [stalled-Git recipe](r7-stalled-git-recipe.json)
and [combined review](combined-review.json) retain the causal evidence and limits.
Both original native failures and controlled logs remain compressed here.
The containing commit records one bounded fixture repair and the compound lesson.
The 51 local closures and physical native/release gates remain unchanged.
