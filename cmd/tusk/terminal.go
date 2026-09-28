package main

import (
	"os"

	"github.com/charmbracelet/x/term"
	"github.com/newbpydev/tusk/internal/cli"
)

func terminalFacts() cli.TerminalFacts {
	width, _, err := term.GetSize(os.Stdout.Fd())
	return cli.TerminalFacts{In: term.IsTerminal(os.Stdin.Fd()), Out: term.IsTerminal(os.Stdout.Fd()), Err: term.IsTerminal(os.Stderr.Fd()), Width: width, WidthError: err}
}
