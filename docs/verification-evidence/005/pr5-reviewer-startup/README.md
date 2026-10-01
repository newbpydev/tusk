# PR #5 Kilo workspace startup repair

PR: https://github.com/newbpydev/tusk/pull/5

Original head: `a002e2eefb36828364a5e8e2cd76cd912d6cad3e`.
Original failing check: `110156456257`, recorded in `kilo-failure.json`.

Kilo failed before review with workspace checkout exit 128. Git LFS tried to
smudge `third_party/glamour/styles/gallery/ascii.png`, whose object was absent
from Tusk's LFS store. The Glamour v0.9.1 source archive had contributed eight
pointer files and the upstream rules that activate LFS. This host has no Git LFS,
so earlier local source hashing and app tests accepted the pointer bytes without
exercising the remote object lookup.

All eight PNGs are now ordinary Git blobs. `assets.json` records their upstream
v0.9.1 URLs, byte sizes, object SHA-256s and original pointer SHA-256s. Every image
matches both the size and object hash in its original pointer: 4,090,310 bytes in
total. The source manifest retains the original upstream pointer hashes and
explicitly declares the materialized files and `.gitattributes` correction.
Existing golden-file attributes and both runtime renderer patches are preserved.

`make test-tui TUI_TEST_RUN='^TestTUI_DependencyCheckoutNeedsNoLFS$'
TUI_TEST_FLAGS='-count=1'` fails before the fix with exit 128 and `smudge filter lfs
failed` (`red.log`). The new regression creates temporary indexed repositories
from each local dependency, then checks out into fresh directories with a
required failing LFS filter. It also rejects unresolved pointer payloads, even
when attributes do not activate LFS. The broader dependency checks pass after
the repair (`green.log`). No credentials, network service or Git LFS install is
needed by the regression.

Canonical `make validate` passes (`validate.log.gz`): formatting, vet, full tests,
race tests, coverage, script tests and module checks. Runtime Go code is
unchanged; existing performance and Kitty receipts are not relabeled as new
verification. Native/hosted Feature 006 release scenarios remain separate.

Publication and a fresh remote checkout with LFS unavailable are pending.
Kilo passing workspace setup must be observed separately on the published head.
The full review result and PR merge readiness are outside this startup receipt.
