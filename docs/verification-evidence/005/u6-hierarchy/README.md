# U6 hierarchy and Darkmatter checkpoint

The owner approved connected tree guides, fractional direct-item progress and
the [Darkmatter palette](https://tweakcn.com/editor/theme?theme=darkmatter).
[checkpoint.json](checkpoint.json) records the implementation, evidence and
remaining CLI timing gate. This is not final plan acceptance.

- `red.log` and `green-final.log`: hierarchy/count behavior and SQLite refresh.
- `focused-build.log`: deep-tree clipping fix after the existing narrow test failed.
- `darkmatter-red.log` and `darkmatter-green.log`: exact backgrounds and shared controls.
- `darkmatter-validate.log`: canonical format, vet, tests, race, coverage, build and generated checks pass.
- `darkmatter-go125.log`: affected minimum-Go checks pass.
- `tui-latency.json`: all 78 cases pass; worst preparation p95 13.929 ms.
- `tui-latency.json.startup.json`: retained child-PTY startup observations.
- `cli-latency-first.json`: initial candidate passes 80/84.
- `control-cli.json`: previously accepted 03c5381 source passes 81/84 on this host.
- `cli-latency.json`: second candidate passes 77/84; final gate remains open.
- `host-*.json` and `idle-window.json`: retained host observations; sustained Chrome activity after Kilo quieted.
- `review/review.json`: scoped lite review complete, no findings; no independent review claim.
- `simplification.json`: reuse, quality and efficiency review, no additional changes required.

All timing samples and warmups remain intact. No thresholds changed. The control
checkout was removed after retaining its complete report. Earlier failed test
attempts are retained and explained in the checkpoint.

The screenshots and text captures show real app operation in owned Kitty window
2 with isolated databases: 80×24/120×40 in color/plain mode, hierarchy navigation,
collapse/expand, descendant completion, weighted counts, opaque calendar and
terminal restoration. Automated checks and timing stayed in Bash.

The original isolated database with owner-entered test tasks is preserved.
Feature 006 native/hosted release proof remains separate; nothing was published.
