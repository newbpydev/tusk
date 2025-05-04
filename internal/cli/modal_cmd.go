package cli

import (
	"github.com/newbpydev/tusk/internal/examples/modal"
	"github.com/spf13/cobra"
)

// modalCmd represents the command to run the modal examples
var modalCmd = &cobra.Command{
	Use:   "modal",
	Short: "Run the modal example to test modal functionality",
	Long: `Runs a demo application that shows how modals work in the Tusk application.
This lets you visualize and test both full-screen and content-area modals.`,
	Run: func(cmd *cobra.Command, args []string) {
		modal.RunModalExample()
	},
}

func init() {
	rootCmd.AddCommand(modalCmd)
}
