# GitHub presentation, version controls and packaged candidate checks

The owner requested GitHub metadata/release readiness and version control. This
follow-up adds stable SemVer history checks, a changelog and a per-release
checklist. It also implements five-platform verification of unchanged packaged
candidate binaries. Public release acceptance remains pending.

[The receipt](acceptance.json) records canonical validation, focused Red/Green
controls and the current local review. All review lenses ran sequentially in
the root session under AGENTS.md. Claude's provider-capable independent review
failed authentication; no independent result is claimed.

The [preliminary candidate](candidate-jobs.json) at `31894ff` passed all nine
jobs. [Cryptographic verification](candidate-verification.json) and the
[Linux packaged smoke](linux-smoke.json) passed. The binary was inspected in an
owned CachyOS Kitty window with an isolated database; [the terminal receipt](kitty-receipt.json)
records navigation, help and restoration. These are basic observations, not
complete release acceptance. The final candidate must include the new tooling
and repeat the applicable gates after merge.

[The live history receipt](version-history.json) observed `v0.3.0` available
above existing stable tags/releases. That observation does not reserve a version.
Invalid versions, failed/malformed reads, newer history on later pages and
superseded draft publication are covered by retained controls.

[Repository metadata](repository-metadata.json) records About/homepage/topics,
private reporting, social preview, release immutability and sidebar readback.
The [settings capture](release-settings.jpg) shows the enabled immutable-release
setting and the actual Tusk preview. No tag, public release or package was created.

The initial lint failure and invalidated promotion-fixture upload count remain
as investigation logs; corrected controls and canonical validation passed.
Compressed logs are automated evidence. Terminal images show only task-created
sample data. Foreign physical consoles, full final-candidate measurements,
Homebrew and concrete publication authority remain pending.
