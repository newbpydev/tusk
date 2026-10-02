// releasecheck is build tooling; it is not linked into the Tusk application.
package main

import (
	"fmt"
	"io"
	"os"
)

func run(args []string, stderr io.Writer) int {
	var err error
	switch {
	case len(args) == 5 && args[0] == "overlay" && (args[4] == "candidate" || args[4] == "snapshot"):
		var info overlayInfo
		info, err = createOverlay(args[1], args[2], args[3], args[4] == "snapshot")
		if err == nil {
			err = writeJSON(args[2]+string(os.PathSeparator)+"input.json", info)
		}
	case len(args) == 8 && args[0] == "finalize":
		err = finalize(args[1], args[2], args[3], args[4], args[5], args[6], args[7])
	case len(args) == 2 && args[0] == "verify":
		err = verifyInventory(args[1])
	default:
		err = fmt.Errorf("expected overlay ROOT NEW_DIRECTORY TAG candidate|snapshot; finalize ROOT DIST ASSETS INPUT SHA COMPILER MODE; verify ASSETS")
	}
	if err != nil {
		fmt.Fprintf(stderr, "Release check: %v\n", err)
		return 1
	}
	return 0
}

func main() { os.Exit(run(os.Args[1:], os.Stderr)) }
