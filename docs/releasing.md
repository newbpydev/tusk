# Release procedure and public repository preview

No public release exists. The first proposed version is `v0.3.0`; the owner must
approve the actual version and source commit. Local engineering and commits do
not authorize pushing, workflow dispatch, tags, settings changes or publication.
Binary commands in [installation](install.md#binary-download-drafts) are draft
instructions until real URLs and native receipts exist.

## Version policy

Tusk follows [Semantic Versioning 2.0.0](https://semver.org/). The public
compatibility surface includes documented commands and flags, JSON fields and
types, exit codes, configuration, database compatibility and backup/upgrade
behavior. The maintainer chooses the bump from the actual change; commits do not
automatically decide a version.

| Change | Before 1.0 | From 1.0 onward |
| --- | --- | --- |
| Compatible fixes | Patch: `0.3.0` to `0.3.1` | Patch: `1.0.0` to `1.0.1` |
| Compatible features or deprecations | Minor: `0.3.0` to `0.4.0` | Minor: `1.0.0` to `1.1.0` |
| Incompatible behavior or formats | Minor, with migration notes | Major, with migration notes: `1.0.0` to `2.0.0` |

Reset the patch when increasing the minor, and both minor and patch when
increasing the major. A `0.x` release remains in initial development; the
pre-1.0 bump policy above is Tusk's convention. Declare 1.0 only when the public
contracts and support scope are ready for that commitment.

The Git tag `vMAJOR.MINOR.PATCH` is the published version identity. Release
construction generates an immutable constant from that selection and checks the
binary's reported version against the manifest. Ordinary source builds report
`dev`; local snapshots report the intended version with `-dev`. There is no
second mutable application version file to keep in sync. The current pipeline
supports stable versions only; prerelease/build-metadata inputs are refused.

Before dispatching a candidate, select a version and inspect live history:

```bash
make check-release-version RELEASE_VERSION=v0.3.0 RELEASE_VERSION_OUTPUT=/tmp/tusk-version-history-unique
```

This read-only check uses authenticated GitHub access and a bounded client. It
reads every page of repository tags and releases, including visible drafts,
and requires the selected stable version to be greater than every existing
stable version. A failed or malformed read refuses release preparation. Unrelated
legacy tag names do not establish a stable-version floor. A passing receipt is
an observation, not a reservation; use an exclusive maintainer release window.

The hosted candidate workflow runs this check before its CI/build jobs.
Promotion repeats the check before draft/publication, permitting the selected
version only for exact resumption; the existing source, notes and asset identity
guards still apply. Readback of historical releases remains available. Local
archive construction is offline engineering evidence and does not check or
reserve a remote version. Never retag, overwrite published assets or reuse a
released version for changed contents. Ship a new version for a correction.

## Checklist for every release

1. Review [Unreleased changes](../CHANGELOG.md) and compatibility/migration
   effects. Select the SemVer bump, full source SHA and supported platforms.
2. Pass canonical validation, generated/module/license checks, native source CI
   and the live version-history check. Freeze build and acceptance tooling.
3. Build one trusted candidate. Review its manifest, nine public assets,
   checksums, notices and signed provenance. Verify before executing downloads.
4. Test those unchanged packaged bytes. Retain platform smoke checks, required
   terminal/install checks, backup/upgrade checks, reference measurements and
   applicable cask evidence. Settle unavailable gates explicitly; automation
   receipts do not claim physical console observations.
5. Write reviewed release notes from the changelog: user-visible changes,
   breaking changes, migration/backup instructions, install/verification routes,
   supported platforms and known limitations. Bind the notes and accepted gates
   to the concrete candidate in the owner-approved action record.
6. Create or resume the exact draft, upload and read back all nine assets, then
   publish under the existing promotion gates. Repository release immutability
   is enabled; published tags and assets are locked. Complete uploads in draft.
7. Verify the public tag/source, release and asset hashes, anonymous installation,
   provenance and applicable package-manager route. Date the changelog using the
   actual publication date; refresh supported versions/security policy, README
   and public metadata. Retain receipts and start the next Unreleased section.

GitHub [immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases)
and [release integrity verification](https://docs.github.com/en/code-security/how-tos/secure-your-supply-chain/secure-your-dependencies/verify-release-integrity)
complement the existing candidate provenance checks. Stable publication remains
pending until this checklist's acceptance and authorization gates are met.

## Candidate workflow

Keep Go 1.25 as the source floor; production builds use the compiler, GoReleaser
and action hashes pinned in `scripts/tool-versions.json`. Use the complete source
checkout with replacements and notices. Never run an ambient packager as proof
of the pinned release tool.

Before a candidate, run canonical validation and generated/module/license checks.
Local packaging creates immutable version input and five CGO-free archives,
source, notices, checksums and a manifest. The intended archive format is
`tusk_VERSION_OS_ARCH.tar.gz` on Unix and `.zip` on Windows, with root `tusk` or
`tusk.exe`, README, LICENSE, THIRD_PARTY_NOTICES, completions and manpages.
Linux ELF inspection must prove no interpreter/shared-library dependency; Darwin
and Windows may use normal OS libraries. Avoid a universal “fully static” claim.

Use Bash and GNU Make with `jq`, Git, `curl` and the pinned Go toolchain available.
Windows requires jq 1.7+ with binary output support, checked by setup/preflight.
Install the verified packager once, then check the current source prerequisites:

```bash
make setup-sqlc setup-release
make test-release release-check
```

Select a full commit SHA containing the release tooling. Use a new output directory
whose parent exists, with a separate directory for each retained run:

```bash
make release-candidate RELEASE_VERSION=v0.3.0 RELEASE_SHA=FULL_COMMIT_SHA RELEASE_OUTPUT=/tmp/tusk-candidate-unique
make verify-release RELEASE_ASSETS=/tmp/tusk-candidate-unique/assets
make release-snapshot RELEASE_VERSION=v0.3.0 RELEASE_SHA=FULL_COMMIT_SHA RELEASE_OUTPUT=/tmp/tusk-snapshot-unique
```

`FULL_COMMIT_SHA` and the output names are placeholders. Candidate binaries report
`0.3.0`; snapshots report `0.3.0-dev`; ordinary source builds report `dev`. Local
candidate names express intended version identity and do not prove publication.
The helper creates a detached temporary checkout and local tag, runs canonical
release prerequisites there, applies a generated constant overlay outside the checkout, and
invokes GoReleaser with publication disabled and an environment without publishing
credentials. The caller's source, tags, tasks and user configuration are preserved.
An existing conflicting tag or retained output path is rejected.

Git source-archive modes use a fixed 0022 tar umask and tracked executable bits.
GoReleaser's `dist` setting is literal, so the helper generates an external runtime
config changing only that field to owned temporary storage. The canonical config
hash remains stable in the manifest; `config-input.json` retains the actual path
override as per-run diagnostics.

The `assets` directory contains exactly five platform archives, source, standalone
notices, checksums and a manifest. The manifest binds source, compiler, tool/config,
module/license/patch inputs, generated constant, executable and member hashes.
Checksums cover the seven payload/notice files; the manifest records their hashes
and the checksum-file hash, avoiding a self-hash cycle. Hosted run/provenance
identity belongs in a separate receipt. Binary archives include guides and the
sample screenshot; source archives include the complete engineering documentation
and patched module sources. Relative engineering links in the contribution guides
can be followed in the source archive or repository.

`build.log`, `version-input.json`, `config-input.json` and `result.json` remain beside successful assets.
Failed runs keep the available diagnostics and never replace an earlier run. Only
the invocation-created temporary build storage is cleaned. Compare executable and
archive hashes from two locations when accepting reproducibility; do not infer it
from stable filenames.

The manual hosted workflow is published on main; its first `v0.3.0` candidate
run passed all nine jobs, including build and provenance. A fresh candidate will
be required after this tooling follow-up merges. Its source SHA must equal the main dispatch and workflow SHA. Six
reusable canonical CI jobs must finish before one build uploads a uniquely named
artifact retained for 90 days. Only the separate provenance job can request OIDC
and write attestations; no job can publish a release or access tap secrets.
Canceled/failed upload or attestation leaves the entire run unaccepted.

After authorized workflow publication/dispatch completes successfully, download
`tusk-candidate-RUN_ID-ATTEMPT` from that run into a new owned directory. It contains
`assets/`, `candidate-run.json` and the reviewed `homebrew/Casks/tusk.rb`. Record the maintainer-selected manifest SHA256
before verification; the digest cannot be inferred as accepted from downloaded
text alone. Use GitHub CLI with attestation support (tested 2.102.0), authenticated
read access, pinned Go/Bash/Make and a fresh verification output directory:

```bash
make verify-candidate CANDIDATE_DIR=/tmp/downloaded-candidate CANDIDATE_VERIFICATION_DIR=/tmp/candidate-verification-unique CANDIDATE_RUN_ID=RUN_ID CANDIDATE_MANIFEST_SHA256=ACCEPTED_SHA256 RELEASE_VERSION=v0.3.0 RELEASE_SHA=FULL_COMMIT_SHA
```

Run/attempt/repository/default-branch/artifact readbacks must match. Every payload,
manifest, checksums file and run receipt must have a verified signature from this
repository's main release workflow, the exact source/builder SHA and the same run
invocation. The generated cask is also digest-bound in the run receipt and attested;
it stays separate from the nine public release assets. This uses signed certificate fields rather than self-reported predicate
metadata; see the [GitHub CLI verification contract](https://cli.github.com/manual/gh_attestation_verify).
Receipts retain API artifact ID/digest and complete cryptographic verification.
The API transport ZIP digest is not represented as an independently computed hash.
A failed read retains its partial directory without acceptance; retry readback in
a new directory. An expired/incomplete candidate needs a new workflow run and new
native acceptance. Verification never writes into candidate or Git storage.

Build once and download those exact bytes for native install, configuration,
CLI/TUI, backup, upgrade, uninstall and retained performance checks. A cross-build is not native execution. Required
hosts are Linux amd64/arm64, macOS Intel/ARM and native Windows amd64, including
an owned Windows 11 terminal. Missing access leaves the gate pending.

## Native smoke and retained measurements

The separate `Native candidate verification` workflow, once published, downloads the selected
completed candidate run/attempt. Each Linux amd64/arm64, macOS Intel/ARM and
Windows amd64 job authenticates every payload, manifest and run identity before
extracting and executing its native binary. The unchanged binary runs the CLI
process suite and Linux child-PTY lifecycle checks where available. Reports are
retained for every job; their status is `automated-passed`, with manual terminal
inspection still pending. This workflow does not rebuild the selected executable,
publish, create tags or change the Homebrew tap.

After trusted verification, extract the native archive into an owned directory.
Keep evidence in a new directory outside both candidate storage and the source
checkout. The drivers verify manifest and executable identities, consume the
selected executable without building it, and retain owned DB/WAL/SHM fixtures:

```bash
make release-smoke RELEASE_BINARY=/tmp/extracted/tusk RELEASE_MANIFEST=/tmp/downloaded-candidate/assets/release-manifest.json CANDIDATE_MANIFEST_SHA256=ACCEPTED_SHA256 CANDIDATE_VERIFICATION_RECEIPT=/tmp/candidate-verification-unique/receipt.json RELEASE_SMOKE_OUTPUT=/tmp/native-smoke-unique
make bench-cli-release RELEASE_BINARY=/tmp/extracted/tusk RELEASE_MANIFEST=/tmp/downloaded-candidate/assets/release-manifest.json CANDIDATE_MANIFEST_SHA256=ACCEPTED_SHA256 CANDIDATE_VERIFICATION_RECEIPT=/tmp/candidate-verification-unique/receipt.json CLI_BENCH_OUTPUT=/tmp/release-cli-unique.json
make bench-tui-release RELEASE_BINARY=/tmp/extracted/tusk RELEASE_MANIFEST=/tmp/downloaded-candidate/assets/release-manifest.json CANDIDATE_MANIFEST_SHA256=ACCEPTED_SHA256 CANDIDATE_VERIFICATION_RECEIPT=/tmp/candidate-verification-unique/receipt.json TUI_BENCH_OUTPUT=/tmp/release-tui-unique.json
```

Run timings sequentially without concurrent test/build load. CLI measurements use
the reference profile. TUI model measurements require a clean checkout at the
candidate SHA on native Linux amd64; startup runs the packaged executable.
Binary compiler and measurement compiler identities are recorded separately.
All three targets refuse existing outputs and preserve the supplied binary hash.
`RELEASE_EVIDENCE_SCOPE=local-fixture` permits preliminary local engineering checks;
that explicit label cannot close any trusted/native release scenario.

On native Windows amd64, invoke `scripts/release_smoke.ps1` with `-Binary`,
`-Manifest`, `-ManifestSHA256`, `-VerificationReceipt` and a new `-OutputDirectory`.
GNU Make/Git Bash and the maintainer toolchain are QA prerequisites, not application
runtime dependencies. The PowerShell entry point has not been executed on Windows.
Automated Linux child PTY restoration checks remain separate from actual owned
Kitty interactions; Windows console/visual checks and macOS/Linux ARM native
records must be supplied on their native hosts. Missing hosts leave gates open.

## Local macOS cask preparation

The pinned [GoReleaser cask configuration](https://goreleaser.com/customization/publish/homebrew_casks/)
uses separate macOS build/archive IDs, `skip_upload: true`, both architectures,
static shell completions and all manuals. The generated cask declares macOS only.
A strict statement-by-statement audit compares URLs, archive hashes and installed
files with the verified manifest, without evaluating downloaded Ruby. It rejects
hooks, zap/uninstall actions, arbitrary code and security-control bypasses.

```bash
make test-homebrew
make check-homebrew RELEASE_MANIFEST=/tmp/downloaded-candidate/assets/release-manifest.json HOMEBREW_CASK=/tmp/downloaded-candidate/homebrew/Casks/tusk.rb
make homebrew-candidate RELEASE_MANIFEST=/tmp/downloaded-candidate/assets/release-manifest.json HOMEBREW_CASK=/tmp/new-owned-directory/tusk.rb
make homebrew-destination
```

The render target refuses existing outputs and candidate/source storage. It can
reproduce the cask for review; trusted acceptance uses the attested cask and run
receipt. The destination target only reads the proposed tap and refuses missing,
private, archived or inaccessible ownership/write access. No target pushes a tap.
Failed cask audit retains the actual generated file beside failed packaging logs.

On native Intel and ARM macOS, separately run Ruby syntax checking, Homebrew
style/audit/install/version, shell completion/man inspection, replacement/removal
and database preservation, with owned fixture data. Follow the
[Homebrew cask artifact contract](https://docs.brew.sh/Cask-Cookbook).
Linux has no native macOS/Homebrew evidence, and this host has neither Ruby nor
Homebrew installed. The tap lookup remains missing/inaccessible. These checks,
public URLs and signed-download behavior stay pending; no Linuxbrew route or
quarantine-removal instructions are advertised.

## Publication and failure recovery

Local promotion and metadata helpers are implemented, with no hosted write run.
See [acceptance records and exact commands](release-acceptance.md) for the native
gate inventory, retained reports, separate draft/tag versus publication approval,
hash-bound notes and current-source checks. `make release-prepare` and
`make prepare-repository-metadata` produce reviewable unaccepted/unapplied plans
without API requests. A promotion receipt leaves anonymous/tap/browser/final
settlement gates open.

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
