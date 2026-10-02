# Release acceptance records

These maintainer records keep local engineering, trusted candidate acceptance and
owner authorization separate. They are not application configuration. Creating an
`owner-approved` JSON file does not grant permission: record the owner's actual
session approval, including its exact action and identities, only after receiving
it. Fixture approvals are test data and authorize no real operation.

## Candidate and native acceptance

Freeze executable, packaging, verification and promotion tooling in the source
commit before the final main candidate run. Native acceptance consumes that run's
downloaded, signature-verified files. A new run or changed input needs fresh
affected acceptance. Later governance Markdown may use a separate documentation
SHA. Promotion checks both local history and current remote main for changed
build/acceptance inputs. Keep the tool checkout clean.

Retain an acceptance JSON file outside the source/candidate directories:

```json
{
  "schema": 1,
  "status": "accepted",
  "repository": "newbpydev/tusk",
  "version": "v0.3.0",
  "source_sha": "FULL_ACCEPTED_SOURCE_SHA",
  "manifest_sha256": "ACCEPTED_MANIFEST_SHA256",
  "cask_sha256": "ACCEPTED_CASK_SHA256",
  "run_id": 123,
  "prepublication_scenarios_reviewed": true,
  "gates": [
    {
      "id": "native/linux/amd64",
      "passed": true,
      "receipt": {"path": "/owned/native-linux-amd64.json", "sha256": "RECEIPT_SHA256"}
    }
  ],
  "reports": {
    "cli": {"path": "/owned/cli.json", "sha256": "REPORT_SHA256"},
    "tui": {"path": "/owned/tui.json", "sha256": "REPORT_SHA256"},
    "startup": {"path": "/owned/tui.json.startup.json", "sha256": "REPORT_SHA256"}
  }
}
```

The example is incomplete and deliberately unusable. Replace every placeholder
and supply exactly these nine gate IDs:

- `native/linux/amd64`, `native/linux/arm64`, `native/darwin/amd64`,
  `native/darwin/arm64`, `native/windows/amd64`;
- `cli-reference`, `tui-reference`;
- `cask/darwin/amd64`, `cask/darwin/arm64`.

Each gate receipt must contain `status: "accepted"`,
`scope: "trusted-candidate"`, its `gate` ID and matching `source_sha`, `version`,
`manifest_sha256` and numeric `run_id`. Include observed host/OS/architecture,
tool/terminal identity, commands/actions, UTC times, failures, binary/archive
digests and links to retained evidence. Native receipts cover the applicable
install, completion/manual, timezone/configuration, CLI/TUI, terminal, backup,
replacement/removal and unsigned-download scenarios in the verification plan.
The maintainer reviews those records; hashes bind their bytes and do not prove
physical host execution by themselves. An absent host leaves its gate pending.

CLI reports retain all 28 cases in each of three runs, five warmups and 100 samples
per case, the packaged executable/compiler identities and the reference profile.
Promotion recomputes the p90/p95/p99/max guards from the samples. The Feature 005
balanced-host allowance does not satisfy this release's reference gate. TUI
reports retain all 26 cases in each of three runs, sample/allocation/output data,
clean accepted source identity and compiler/binary identities. Preparation and
calendar samples must pass 16.7/33.3/50 ms p95/p99/max guards. Startup retains its
three complete sample/RSS runs; there is no invented startup latency SLA.

`prepublication_scenarios_reviewed` records maintainer review of the license,
rights, complete notices, applicable U1–U7 checks and their attached evidence.
Promotion independently re-verifies candidate signatures, run/attempt/artifact
expiry and cask integrity, then requires a public, owned, writable tap. A local
fixture or missing tap cannot substitute for that preflight.

## Separate draft and publication authority

After the owner approves the concrete result, record the exact action:

```json
{
  "schema": 1,
  "status": "owner-approved",
  "action": "draft",
  "repository": "newbpydev/tusk",
  "source_sha": "FULL_ACCEPTED_SOURCE_SHA",
  "version": "v0.3.0",
  "manifest_sha256": "ACCEPTED_MANIFEST_SHA256",
  "cask_sha256": "ACCEPTED_CASK_SHA256",
  "run_id": 123,
  "acceptance_sha256": "ACCEPTANCE_FILE_SHA256",
  "notes_sha256": "REVIEWED_NOTES_SHA256",
  "exclusive_release_window": true,
  "approval_reference": "REFERENCE_TO_ACTUAL_OWNER_APPROVAL"
}
```

Draft authority includes creating the named tag at the exact accepted SHA,
creating/resuming the private draft and uploading the nine accepted assets.
Publication needs a separate approval file with `action: "publish"`; draft
approval cannot authorize publication. Arrange an exclusive maintainer release
window: GitHub's APIs do not atomically guard against another privileged actor
publishing or retagging between a read and a write. The helper reads state before
writes and detects conflicts; it does not claim a server-side compare-and-swap.

Use full absolute paths and fresh retained output directories. All actions use
the same candidate/version/SHA/run/manifest selection:

```bash
make release-prepare CANDIDATE_DIR=/owned/candidate CANDIDATE_RUN_ID=RUN_ID CANDIDATE_MANIFEST_SHA256=ACCEPTED_SHA256 RELEASE_VERSION=v0.3.0 RELEASE_SHA=FULL_SOURCE_SHA RELEASE_PROMOTION_OUTPUT=/owned/prepare-new
make release-draft CANDIDATE_DIR=/owned/candidate CANDIDATE_RUN_ID=RUN_ID CANDIDATE_MANIFEST_SHA256=ACCEPTED_SHA256 RELEASE_VERSION=v0.3.0 RELEASE_SHA=FULL_SOURCE_SHA RELEASE_ACCEPTANCE=/owned/acceptance.json RELEASE_AUTHORIZATION=/owned/draft-approval.json RELEASE_NOTES=/owned/notes.md RELEASE_PROMOTION_OUTPUT=/owned/draft-new
make release-publish CANDIDATE_DIR=/owned/candidate CANDIDATE_RUN_ID=RUN_ID CANDIDATE_MANIFEST_SHA256=ACCEPTED_SHA256 RELEASE_VERSION=v0.3.0 RELEASE_SHA=FULL_SOURCE_SHA RELEASE_ACCEPTANCE=/owned/acceptance.json RELEASE_AUTHORIZATION=/owned/publish-approval.json RELEASE_NOTES=/owned/notes.md RELEASE_PROMOTION_OUTPUT=/owned/publish-new
make release-readback CANDIDATE_DIR=/owned/candidate CANDIDATE_RUN_ID=RUN_ID CANDIDATE_MANIFEST_SHA256=ACCEPTED_SHA256 RELEASE_VERSION=v0.3.0 RELEASE_SHA=FULL_SOURCE_SHA RELEASE_ACCEPTANCE=/owned/acceptance.json RELEASE_NOTES=/owned/notes.md RELEASE_PROMOTION_OUTPUT=/owned/readback-new
```

Prepare performs no GitHub request and labels the result `prepared-unaccepted`.
Readback performs no write and still requires fresh acceptance/verification.
Other actions refuse missing authorization. APIs receive structured JSON and
literal filenames. GitHub commands pin `GH_HOST=github.com` and remove inherited
HTTP debug settings to avoid routing or credential-log surprises. API requests
use literal `https://api.github.com/` and `https://uploads.github.com/` URLs so a
CLI `api_host` configuration cannot redirect release/metadata requests. Asset
upload passes the unchanged file through `--input` with an octet-stream header;
it has no filename-label parser or overwrite operation. See the
[GitHub CLI API contract](https://cli.github.com/manual/gh_api) and its
[pinned endpoint handling](https://github.com/cli/cli/blob/v2.102.0/pkg/cmd/api/http.go).

Existing tags must resolve to the accepted source. Existing releases must match
the exact version/SHA/name/notes and non-prerelease identity. Duplicate, unexpected,
starter or altered assets require explicit repair; there is no clobber/delete or
retag operation. Lost responses trigger state/download readback, never an automatic
second mutation. A fresh invocation can resume missing draft assets and skips
matching ones. Every complete draft/public result re-downloads and verifies all
nine files. Keep partial output and draft state after failure. Published defects
use a new version and retain the old bytes/receipts.

The promotion receipt closes only its recorded operation. Anonymous downloads,
live native tap installation, README activation, metadata/browser readback and
final governance settlement remain distinct V98–V102 gates.

## Repository metadata

`.github/repository-metadata.json` is the exact proposed About/topics payload.
Preparation writes reviewable requests and the real preview-image digest without
API access:

```bash
make prepare-repository-metadata METADATA_OUTPUT=/owned/metadata-preview-new
```

After publication and explicit owner approval of that payload and current clean
documentation SHA, record an authorization with exactly `schema: 1`,
`status: "owner-approved"`, `action: "metadata"`,
`repository: "newbpydev/tusk"`, `documentation_sha`, `payload_sha256`,
`release_receipt_sha256` and `approval_reference`:

```bash
make apply-repository-metadata METADATA_AUTHORIZATION=/owned/metadata-approval.json METADATA_RELEASE_RECEIPT=/owned/published-receipt.json METADATA_OUTPUT=/owned/metadata-apply-new
```

The helper confirms the published release identity and that remote main equals
the approved documentation SHA. It changes only description/homepage and topics,
compares current state first and confirms each operation through readback. A
partial topic failure preserves acknowledged About state; a fresh invocation
skips matching About values. Social preview remains an owner upload and observed
browser gate. License detection, private-security setting/contact accuracy,
rendered README at documentation SHA, live release/tap URLs and anonymous installs
remain separately recorded. No security setting is silently enabled.
