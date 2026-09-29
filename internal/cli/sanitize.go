package cli

import "github.com/newbpydev/tusk/internal/terminaltext"

func sanitize(text string) string {
	return terminaltext.Scalar(text)
}
