# Tusk

A local task manager with a keyboard-driven terminal workspace and a scriptable
CLI. Tasks, subtasks and activity history live in embedded SQLite. One binary;
no daemon, account or network service.

```sh
make setup
make build
./bin/tusk add 'Plan the release' --priority high --due tomorrow
./bin/tusk tui
```

The TUI uses spacious task rows, a details pane, searchable filters and focused
forms. Use a terminal of at least 80×24. Press `?` for help, `a` to create a task,
`e` to edit, and `q` to quit browsing. A smaller window preserves your draft and
asks you to resize.

- [TUI guide](docs/tui.md): keys, forms, conflicts, deletion and recovery.
- [CLI guide](docs/cli.md): commands, JSON, configuration and scripting.
- [Masterplan](MASTERPLAN.md): implementation status and verification evidence.

Data defaults to `~/.local/share/tusk/tusk.db`. Set `TUSK_DB_PATH` to choose a
different database. TUI and CLI commands share the same data and service rules.
For scripts, use commands such as `./bin/tusk list --all --json`.

Development requires Go 1.25 or newer. Run `make validate` for formatting, vet,
tests, race detection, coverage and module checks; `make check-generated` checks
generated storage code. Run `make bench-tui` and `make bench-cli` separately from
other builds/tests for retained measurements. Kitty is used to operate the real
application; automated verification runs in Bash.

Native macOS/Windows terminal acceptance and release packaging are tracked in
Feature 006. Cross-compilation alone does not establish native terminal support.
