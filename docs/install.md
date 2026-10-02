# Install and manage Tusk

## Source installation

There is no public release yet. Keep the full Git checkout, including
`third_party/bubbletea` and `third_party/glamour`: `go.mod` uses local patched
replacements. A versioned Go install from a module proxy cannot resolve them.

Install Git, Go 1.25.0 or newer, **GNU Make** and **Bash**. `make setup` checks Go;
it does not install the other prerequisites. `make validate` also requires a
native C compiler for Go's race detector and jq 1.7+ for script fixtures. Linux uses
GCC; macOS uses the Command Line Tools compiler. Windows developers use native
Go, Git Bash, GNU Make and native GCC, with Make/GCC on PATH. WSL checks do not
establish native Windows behavior. Python, Kitty and jq are not runtime requirements.

From Bash on Linux/macOS:

```sh
git clone https://github.com/newbpydev/tusk.git
cd tusk
make setup build
./bin/tusk --version
mkdir -p "$HOME/.local/bin"
cp bin/tusk "$HOME/.local/bin/tusk"
export PATH="$HOME/.local/bin:$PATH"
command -v tusk
tusk --version
```

`make build` uses `CGO_ENABLED=0` and builds the whole `./cmd/tusk` package.
No daemon or external database is needed. Add the PATH line to your chosen shell's
startup file yourself if desired; Tusk never edits it.

For Windows source installation, run in **Git Bash**:

```sh
git clone https://github.com/newbpydev/tusk.git
cd tusk
make setup build BUILD_OUTPUT=bin/tusk.exe
./bin/tusk.exe --version
```

Then use PowerShell to copy `bin\tusk.exe` to `$env:USERPROFILE\Apps\Tusk`, add
that directory to PATH using the commands in the draft below and check
`Get-Command tusk`. Native Windows runtime/console acceptance remains pending.

## Binary download drafts

**Pending release:** these commands describe the intended archive format. Do not
run them until the selected version and matching assets are published and their
native checks pass. No unsigned download, Homebrew or older OS route is claimed
accepted yet. Prebuilt-binary users need no Go, Make, Kitty or jq.

| System | Architecture | Asset |
| --- | --- | --- |
| Linux | x86_64 / amd64 | `tusk_VERSION_linux_amd64.tar.gz` |
| Linux | aarch64 / arm64 | `tusk_VERSION_linux_arm64.tar.gz` |
| macOS | x86_64 / Intel | `tusk_VERSION_darwin_amd64.tar.gz` |
| macOS | arm64 / Apple Silicon | `tusk_VERSION_darwin_arm64.tar.gz` |
| Windows | x64 / amd64 | `tusk_VERSION_windows_amd64.zip` |

Replace `VERSION` with the numeric released version, without `v`. Downloads and
`checksums.txt` must come from the same GitHub release. The first proposed version
is 0.3.0; it is not a published version.

Linux/macOS draft (POSIX shell, `curl`, `tar`, `awk` and a SHA256 utility):

```sh
version=0.3.0                       # Replace with an actual published version
os=linux                            # darwin on macOS
arch=amd64                          # arm64 for aarch64 / Apple Silicon
asset="tusk_${version}_${os}_${arch}.tar.gz"
base="https://github.com/newbpydev/tusk/releases/download/v${version}"
mkdir -p "$HOME/Downloads/tusk-$version"
cd "$HOME/Downloads/tusk-$version"
curl --fail --location --output "$asset" "$base/$asset"
curl --fail --location --output checksums.txt "$base/checksums.txt"
expected=$(awk -v asset="$asset" '$2 == asset { print $1 }' checksums.txt)
test "${#expected}" = 64 || exit 1
if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$asset" | awk '{print $1}')
else
    actual=$(shasum -a 256 "$asset" | awk '{print $1}')
fi
test "$actual" = "$expected" || exit 1
mkdir unpacked
tar -xzf "$asset" -C unpacked
mkdir -p "$HOME/.local/bin"
cp unpacked/tusk "$HOME/.local/bin/tusk"
chmod u+x "$HOME/.local/bin/tusk"
export PATH="$HOME/.local/bin:$PATH"
command -v tusk
tusk --version
```

Windows draft (native **PowerShell**, amd64):

```powershell
$version = '0.3.0' # Replace with an actual published version
$asset = "tusk_${version}_windows_amd64.zip"
$base = "https://github.com/newbpydev/tusk/releases/download/v$version"
$download = Join-Path $env:USERPROFILE "Downloads\tusk-$version"
New-Item -ItemType Directory -Force $download | Out-Null
Set-Location $download
Invoke-WebRequest "$base/$asset" -OutFile $asset
Invoke-WebRequest "$base/checksums.txt" -OutFile checksums.txt
$lines = @(Get-Content checksums.txt | Where-Object { ($_ -split '\s+')[1] -eq $asset })
if ($lines.Count -ne 1) { throw 'Missing or duplicate checksum' }
$expected = ($lines[0] -split '\s+')[0]
$actual = (Get-FileHash $asset -Algorithm SHA256).Hash.ToLowerInvariant()
if ($actual -ne $expected) { throw 'Checksum mismatch; stop before extraction' }
Expand-Archive -LiteralPath $asset -DestinationPath unpacked
$install = Join-Path $env:USERPROFILE 'Apps\Tusk'
New-Item -ItemType Directory -Force $install | Out-Null
Copy-Item unpacked\tusk.exe (Join-Path $install 'tusk.exe')
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $install) {
    [Environment]::SetEnvironmentVariable('Path', "$install;$userPath", 'User')
}
$env:Path = "$install;$env:Path"
Get-Command tusk
tusk --version
```

These drafts do not disable Gatekeeper, quarantine or Windows security controls.
Unsigned binaries may trigger warnings. If normal installation is blocked, stop
and report the route; use the verified source route rather than bypassing controls.
Homebrew cask configuration is prepared locally for macOS Intel and ARM. Delivery
remains pending native Homebrew checks, direct-download acceptance and an
owner-controlled tap. See [the release procedure](releasing.md#local-macos-cask-preparation).

## Completion and manuals

From an installed binary, generate the script for your shell. Save it to a path
you choose before sourcing; do not blindly append generated text to a startup file.

```sh
mkdir -p "$HOME/.config/tusk"
tusk completion bash > "$HOME/.config/tusk/tusk.bash"
# Load your installed bash-completion framework, then:
. "$HOME/.config/tusk/tusk.bash"
```

Bash needs the `bash-completion` framework. For Zsh, initialize `compinit`, run
`tusk completion zsh > "$HOME/.config/tusk/tusk.zsh"`, then source that file.
For Fish use `tusk completion fish > "$HOME/.config/tusk/tusk.fish"` and source
it in Fish. These commands work without task storage and offer static commands,
flags and enums, not your task IDs or tags. No shell startup file is changed.

Manuals are in `docs/man/` for source checkouts and `man/` in planned archives.
Inspect with `man -l docs/man/tusk.1` on a system with `man`; install to your
user manual directory if desired. Windows can use `tusk help COMMAND` without man.

## Upgrade, remove and troubleshoot

Stop Tusk processes, back up data, and replace only the executable. Run
`tusk --version` and read `tusk list --all --json` against the same path afterward.
Migration is transactional; a database with a newer schema is rejected without
fallback or automatic downgrade. Do not remove data to make a downgrade work.

To uninstall, remove the executable you installed and any completion/manual files
you deliberately installed. On Windows remove your owned PATH entry if desired.
**Keep the data directory and DB/WAL/SHM** unless separately intending data deletion.
There is no automatic updater or installer that owns your task data.

- Command not found: inspect `command -v tusk` / `Get-Command tusk`, PATH and stale
  copies. Open a new shell after a persistent PATH change.
- Wrong architecture / exec format: compare the asset table with `uname -m` or
  `$env:PROCESSOR_ARCHITECTURE`. Use amd64 for x86_64, arm64 for aarch64.
- Permission denied: use your own writable install directory; on Unix confirm
  the executable permission. Do not run task commands with sudo to repair PATH.
- Download or checksum failure: stop before extraction/replacement. Preserve the
  existing executable and obtain a matching release/checksum pair again.
- Storage/configuration failure: inspect the explicit path and timezone. Tusk
  never silently falls back. Preserve files; follow [CLI recovery](cli.md#exit-codes-and-recovery).

## Backup and restore

SQLite uses WAL mode and `synchronous=NORMAL`. A transaction can survive process
failure while recent commits may be lost on power/OS failure. This is not a
power-loss durability guarantee or a synchronized cloud backup.

Stop **all** CLI/TUI processes and any other owner of the database before copying.
Do not copy only a live `.db` file. Preserve the closed database and any remaining
`-wal` / `-shm` sidecars as one set. Find the actual path from your selected
configuration; no command here deletes or checkpoints a live database.

Example after every owner has stopped (Bash, replacing the path if necessary):

```sh
db="$HOME/.local/share/tusk/tusk.db"
backup="$HOME/tusk-backup-$(date +%Y%m%d-%H%M%S)"
mkdir "$backup"
cp "$db" "$backup/tusk.db"
for suffix in -wal -shm; do
    if [ -f "$db$suffix" ]; then cp "$db$suffix" "$backup/tusk.db$suffix"; fi
done
restore="$HOME/tusk-restore-check"
mkdir "$restore"
cp "$backup/"* "$restore/"
TUSK_DB_PATH="$restore/tusk.db" tusk list --all --json
TUSK_DB_PATH="$restore/tusk.db" tusk tree --json
```

Record full task IDs and compare `tusk history FULL_ID --json` too. A SQLite
inspection tool may run `PRAGMA integrity_check` on the isolated restored copy;
it should return `ok`. It is optional tooling, not required to run Tusk. Use
PowerShell `Copy-Item` for the same closed file set on Windows. Keep the original
untouched until the restored copy's tasks/events and integrity are confirmed.
For a committed or unknown write outcome, read fresh data before retrying; backup
cannot reconstruct deleted history or serve as undo.
