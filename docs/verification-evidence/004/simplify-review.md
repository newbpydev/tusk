# Feature 004 branch simplification evaluation

Scope: `6a8aff0..8e6380d` on `feat/cli-interface-and-scripting`.
The user explicitly requested this post-acceptance review. Reuse, quality and
efficiency lenses ran sequentially in the main context per the repository tool
mapping; this is not independent multi-agent review.

## Outcome

No production or test changes are retained. The accepted implementation remains
byte-for-byte identical to `8e6380d`; this follow-up records the experiment only.
Feature 005 planning remains next, with no later-phase implementation started.

- Reuse: 0 retained. Combining command handlers/result helpers would obscure
  different argument, mutation and committed-outcome semantics for little gain.
- Quality: 0 retained. Explicit JSON encoding, schema validation and platform
  input lifetime handling have contract/performance reasons and remain intact.
- Efficiency: 1 trial, 0 retained. Human tree formatting used one shared builder
  instead of returning and copying a string at each recursion level. Direct
  writes also removed temporary concatenated line strings. The proposed patch
  is saved in `simplify-proposed.patch.gz` and is not applied.
- Skipped: 2 low-value alternatives (generic command handlers; replacing measured
  explicit encoding/validation with broad helpers). Reverted: 1 measured trial.

## Evidence and decision

The allocation regression first failed: 1,628,627 bytes allocated to render
119,100 output bytes, against an 8x budget. The final trial used 807,317 bytes
(about 50% less); the race run used 906,142 and Go 1.25 used 807,230. A first
shared-buffer version still exceeded the same budget under race, so the final
trial removed line concatenation without changing the assertion.

Exact expected output for 1,000 nodes at ten levels and a nested-invalid-node
check passed. Formatting/Unicode/width/color tests passed. `make validate build`
passed on the trial. Go 1.25 focused formatter tests passed. Old/new output in an
owned Kitty window had identical transcripts (`simplify-kitty-before.txt` and
`simplify-kitty-after.txt`); the final screenshot was visually inspected. Both
owned windows and temporary databases were released. Screenshot attempts that
captured the editor rather than Kitty were excluded.

The trial's full three-run reference matrix passed 83/84 case-runs. The sole
failure was run 3, 1,000-task JSON tree: p90 15.125029 ms (limit <15 ms), 11/100
target misses. All affected human-tree cases passed (p90 11.233032, 10.686991,
10.218686 ms). JSON formatting was not changed by the patch.

A control rebuilt the exact pre-change production source using Go's source
overlay and the canonical `make bench-cli` target. All 84 control case-runs
passed; JSON tree p90 was 11.903000, 12.528229 and 13.844273 ms. This comparison
does not establish that the trial caused the timing miss, but it also does not
confirm preservation of the branch's fixed latency gate. Per ce-simplify-code's
verification rule, the trial was reverted. Thresholds stayed fixed and all
samples were retained. The candidate was not retried to obtain a green result.

The original U6 acceptance report remains valid for its recorded source. These
reports are additional measurements, not replacements for that history.

## Files

- `simplify-red.log.gz`: allocation red and output checks.
- `simplify-green.log.gz`: first focused green.
- `simplify-race-red.log.gz`: allocation failure under race before line-copy fix.
- `simplify-validate.log.gz`: final trial canonical validation/build, exit 0.
- `simplify-minimum.log.gz`: Go 1.25 formatter tests, exit 0.
- `simplify-latency.json` / `simplify-latency.log.gz`: trial, exit 2 from Make.
- `simplify-control-latency.json` / `simplify-control-latency.log.gz`: control, exit 0.
- `simplify-restored-validate.log.gz`: canonical validation after restoring source.
- `simplify-kitty-before.png`, `simplify-kitty-final.png`: inspected terminal views.
- `simplify-proposed.patch.gz`: reversible trial, not production code.

Restored-source `make validate build` completed with exit 0. All source hashes
match the original U6 acceptance receipt. The rebuilt binary contains updated
Go VCS metadata for `8e6380d`, so its hash differs from the original pre-commit
acceptance build.
