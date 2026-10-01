# Contributing to Tusk

Open an issue with the problem, expected behavior and a small sanitized example.
Include `tusk --version`, OS/architecture and terminal details when relevant.
Do not attach personal task databases or unredacted notes. For security reports,
follow [SECURITY.md](SECURITY.md).

Clone the complete repository. Install Go 1.25+, Bash, GNU Make and a native C
compiler for race verification; install jq for the automation fixtures. Windows
uses native Go with Git Bash and GNU Make/GCC (not WSL native acceptance).
See [installation](docs/install.md#source-installation).

```sh
make setup
make validate build check-generated check-docs test-docs check-notices
```

`make validate` runs formatting, vet, tests, race, coverage, script fixtures and
module checks. It is required before commits. Use Red → Green → Refactor:
write a failing behavior test, observe it, then implement the smallest fix.
[AGENTS.md](AGENTS.md) defines boundaries, coverage and active-unit governance;
[MASTERPLAN.md](MASTERPLAN.md) defines current implementation order. Keep the
feature plan, verification plan and workorder synchronized when they change.

Generate static docs with `make generate-docs`; check drift with `make check-docs`.
`make test-completions` additionally needs Bash completion, Zsh and Fish. On
macOS set `BASH_COMPLETION_SOURCE` to your installed framework. Dependency
changes require a reviewed license inventory update and `make generate-notices`.
Do not edit vendored replacements without updating their `TUSK-PATCH.json`
provenance and compatibility tests.

Run canonical tests/builds/measurements in Bash. Operate the real CLI/TUI in an
owned Kitty window with isolated temporary data for Linux/macOS visual checks.
Windows terminal verification uses an owned native Windows Terminal/PowerShell
session. Record source/binary identity, observations and cleanup separately from
test and timing output. Never reuse or close another person's terminal session.

Keep a PR focused. Explain behavior, Red/Green evidence, canonical verification
and any unexecuted native checks. Pushes, hosted dispatch, releases and repository
settings require their concrete authorization. See [release procedure](docs/releasing.md).
