package tea

// Tusk compatibility patch: the upstream package initializer queried global
// terminal colors before command admission. Tusk supplies per-session renderer
// settings; importing this package must not access stdin/stdout or the terminal.
