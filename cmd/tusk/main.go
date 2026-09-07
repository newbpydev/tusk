package main

import (
	"fmt"
	"io"
	"os"
)

// Version defines the current application release version.
const Version = "0.2.0-reboot"

func run(args []string, stdout io.Writer) int {
	if len(args) > 1 {
		switch args[1] {
		case "-v", "--version", "version":
			fmt.Fprintf(stdout, "tusk version %s\n", Version)
			return 0
		case "-h", "--help", "help":
			printHelp(stdout)
			return 0
		}
	}
	printHelp(stdout)
	return 0
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "Tusk - Zero-friction terminal task management system")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  tusk [command] [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Available Commands:")
	fmt.Fprintln(w, "  help       Help about any command")
	fmt.Fprintln(w, "  version    Print the version of tusk")
}

func main() {
	os.Exit(run(os.Args, os.Stdout))
}
