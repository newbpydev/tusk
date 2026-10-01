# Release procedure and public repository preview

No public release exists. The first proposed version is `v0.3.0`; the owner must
approve the actual version and source commit. Local engineering and commits do
not authorize pushing, workflow dispatch, tags, settings changes or publication.
Binary commands in [installation](install.md#binary-download-drafts) are draft
instructions until real URLs and native receipts exist.

## Candidate workflow

Keep Go 1.25 as the source floor; production builds use the compiler, GoReleaser
and action hashes pinned in `scripts/tool-versions.json`. Use the complete source
checkout with replacements and notices. Never run an ambient packager as proof
of the pinned release tool.

Before a candidate, run canonical validation and generated/module/license checks.
Local packaging will create immutable version input and five CGO-free archives,
source, notices, checksums and a manifest. The intended archive format is
`tusk_VERSION_OS_ARCH.tar.gz` on Unix and `.zip` on Windows, with root `tusk` or
`tusk.exe`, README, LICENSE, THIRD_PARTY_NOTICES, completions and manpages.
Linux ELF inspection must prove no interpreter/shared-library dependency; Darwin
and Windows may use normal OS libraries. Avoid a universal “fully static” claim.

The hosted candidate workflow is not implemented/activated yet. Its reviewed
contract is a trusted exact main SHA, no publishing/tap secrets and attested
payload/manifest hashes. Once authorized, build once and download those exact
bytes for native install, configuration, CLI/TUI, backup, upgrade, uninstall and
retained performance checks. A cross-build is not native execution. Required
hosts are Linux amd64/arm64, macOS Intel/ARM and native Windows amd64, including
an owned Windows 11 terminal. Missing access leaves the gate pending.

## Publication and failure recovery

Owner authorization must name version/SHA and accepted manifest/run before
creating a draft. Verify every asset, checksum, notice and provenance record in
the draft, then publish only after concrete final authorization. Promote accepted
files without rebuilding. No automatic public release occurs on push/PR.
Do not overwrite published bytes or retag. Reconcile API state after a lost
response before retrying; deprecate a faulty release and fix it under a new version.

The macOS cask destination proposal is `newbpydev/homebrew-tap`, `Casks/tusk.rb`.
The owner must establish destination/access. Validate Intel/ARM hashes, binary,
manuals and completions locally, then publish the reviewed cask after release
URLs work. Cask removal must preserve task data; no quarantine-removal hook.
A failed tap update leaves the phase pending. No tap push is authorized here.

## Proposed GitHub metadata

Apply these values only after owner authorization and retain before/after readback:

- Description: `Local task management in your terminal: a keyboard-driven TUI and scriptable CLI, backed by SQLite.`
- Homepage: `https://github.com/newbpydev/tusk#readme`
- Topics: `go`, `golang`, `cli`, `tui`, `task-manager`, `terminal`, `sqlite`,
  `bubbletea`, `productivity`, `offline`, `command-line`.
- Social preview: the inspected sample capture `docs/assets/tusk-tui.png`, with
  provenance in [assets](assets/README.md). Use manual upload if needed.
- Website/funding: leave absent optional values empty; no separate website exists.
- Security: private vulnerability reporting is currently disabled; do not claim
  it is active without verified setting readback and an updated policy.

After publication, verify anonymous downloads/install, complete release assets,
Homebrew, GitHub About/license/topics/homepage/social preview and actual rendered
README. Activate badges only when their real CI/release/license targets resolve.
Recheck the README's first-task examples against the published files. Preserve
native/hosted/publication evidence separately from local source checks.
