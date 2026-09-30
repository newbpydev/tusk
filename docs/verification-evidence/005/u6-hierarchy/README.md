# U6 hierarchy and Darkmatter acceptance

The owner approved connected tree guides, fractional direct-item progress and
the [Darkmatter palette](https://tweakcn.com/editor/theme?theme=darkmatter).
[acceptance.json](acceptance.json) closes local acceptance after the owner approved
the final app checks and reported Cline CLI shut down. All 84 final CLI cases pass
(query/help worst p90 14.291/4.112 ms), as does fresh canonical validation.
The application source is unchanged from the reviewed implementation.
[checkpoint.json](checkpoint.json) preserves the earlier open gate.

- `red.log` and `green-final.log`: hierarchy/count behavior and SQLite refresh.
- `focused-build.log`: deep-tree clipping fix after the existing narrow test failed.
- `darkmatter-red.log` and `darkmatter-green.log`: exact backgrounds and shared controls.
- `darkmatter-validate.log`: canonical format, vet, tests, race, coverage, build and generated checks pass.
- `darkmatter-go125.log`: affected minimum-Go checks pass.
- `tui-latency.json`: all 78 cases pass; worst preparation p95 13.929 ms.
- `tui-latency.json.startup.json`: retained child-PTY startup observations.
- `cli-latency-first.json`: initial candidate passes 80/84.
- `control-cli.json`: previously accepted 03c5381 source passes 81/84 on this host.
- `cli-latency.json`: historical second candidate passes 77/84.
- `cli-latency-after-cline.json`: final current-candidate matrix passes 84/84.
- `final-validate.log`: fresh canonical validation, build and generated checks pass.
- `host-after-cline-shutdown.json`: lower host activity before final measurement.
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

The 78-case TUI timing report, Go 1.25 and Kitty receipts are carried forward
for identical source hashes. They are not relabelled as newly executed checks.
The CLI report records the rebuilt binary at implementation commit 246db8c.
