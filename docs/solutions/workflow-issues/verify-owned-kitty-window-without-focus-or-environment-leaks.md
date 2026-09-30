---
title: "Verify the owned Kitty window without relying on desktop focus"
date: "2026-09-30"
category: "workflow-issues"
module: "Terminal verification"
problem_type: "workflow_issue"
component: "development_workflow"
severity: "medium"
applies_when:
  - "Operating the real CLI or TUI in an owned Kitty window"
  - "Capturing terminal images while other desktop applications are open"
  - "Inspecting Kitty remote-control inventory"
tags: ["kitty", "tui", "terminal-verification", "window-identity", "ui-readiness", "metadata"]
---

# Verify the owned Kitty window without relying on desktop focus

## Context

During Feature 005 hierarchy and Darkmatter verification, an active-window
screenshot captured the user's browser after desktop focus changed. That image
was discarded. Capturing the owned Kitty OS window by its explicit X11 window
ID produced the intended app image, which was inspected before retention.

Another inventory call displayed the full `kitty @ ls` response, including
process environment metadata. Closely spaced input and capture also produced
stale frames; sending consecutive Escape keys without observing each transition
left a form active when subsequent text was sent. These were verification
workflow failures, not evidence of an application defect.

## Guidance

1. Record the remote-control socket and Kitty window ID when creating the owned
   verification window. Use an isolated temporary database. Target that socket
   and exact window ID for every action; do not infer ownership from desktop
   focus, the newest window, or a matching title.
2. Capture the inventory response inside the process that parses it. Emit only
   the matching window ID, platform window ID and terminal dimensions. Local
   Kitty help states that `ls` includes the process environment; omitting
   `--all-env-vars` still includes differing variables. Filtering after raw JSON
   has already reached tool output does not prevent exposure.
3. Send an action, then observe its expected transition before sending dependent
   input. Use targeted `get-text --extent screen` with bounded polling to check
   for the expected view, dialog or prompt. A fixed sleep or a successful
   remote-control response does not prove the app has handled the input. Stop
   and inspect on timeout instead of sending the next action blindly.
4. Capture the identified OS window using a backend that supports its platform
   ID. On the verified X11/Xwayland host, `import -window` worked. Inspect the
   image for the correct app, state and dimensions before adding it to evidence.
   Review visible content as well as ownership before retaining it.
5. Keep visual evidence distinct from automated checks. Tests, builds and
   latency measurements belong in Bash. If desktop capture is unavailable,
   leave the required visual check pending; text dumps or synthetic terminal
   tests cannot substitute for inspecting the app.

## Why This Matters

Desktop focus is shared with the user and can change independently of the
verification workflow. A successful screenshot command can therefore produce
an image of the wrong application. Window identity avoids that race, while
observed UI transitions avoid capturing an old frame or typing into the wrong
screen. Filtering inventory also keeps unrelated environment data out of logs
and receipts. Terminal text helps establish readiness, but the screenshot is
still needed to inspect padding, tree guides, colors and opaque backgrounds.

## When to Apply

- Operating or capturing the real CLI/TUI while other desktop apps are open.
- Checking dialogs, navigation, resizing or terminal restoration in Kitty.
- Querying Kitty inventory for verification metadata.

## Examples

Given the socket and window ID recorded when creating the owned window, this
query retains only the metadata needed to target and describe a capture:

```bash
python3 - "$owned_socket" "$owned_window_id" <<'PY'
import json
import subprocess
import sys

socket, target = sys.argv[1], int(sys.argv[2])
inventory = json.loads(subprocess.check_output(
    ["kitty", "@", "--to", socket, "ls", "--match", f"id:{target}"],
    text=True,
))
matches = [
    (os_window, window)
    for os_window in inventory
    for tab in os_window["tabs"]
    for window in tab["windows"]
    if window["id"] == target
]
if len(matches) != 1:
    raise SystemExit("Expected exactly one owned Kitty window")
os_window, window = matches[0]
print(json.dumps({
    "id": window["id"],
    "platform_window_id": os_window.get("platform_window_id"),
    "columns": window["columns"],
    "lines": window["lines"],
}))
PY
```

Use the same socket and ID to observe the current screen after each action:

```bash
kitty @ --to "$owned_socket" get-text \
  --match "id:$owned_window_id" --extent screen
```

For example, after closing the calendar, confirm that the calendar is gone
before sending another Escape. Confirm a shell prompt before sending a CLI
command. Retain only relevant screen text, not unrestricted scrollback.

When the returned platform ID is confirmed to identify the owned X11 window,
capture it directly rather than using an active-window screenshot:

```bash
import -window "$platform_window_id" "$capture_path"
```

Platform IDs are backend dependent. This command does not establish a portable
capture method for native Wayland; use a supported capture backend there or
record the visual check as pending. Re-query if the window is recreated.

## Related

- [Terminal verification contract](../../../AGENTS.md): Bash checks, owned
  Kitty app inspection and isolated databases.
- [Hierarchy and Darkmatter evidence](../../verification-evidence/005/u6-hierarchy/README.md):
  inspected real-app captures across sizes, color/plain mode and calendar use.
- [Local acceptance receipt](../../verification-evidence/005/u6-hierarchy/acceptance.json):
  distinct verification tiers and retained source identity.
