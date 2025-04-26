package cli

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea"
	"github.com/spf13/cobra"
)

var (
	// Flag for user ID in TUI mode
	tuiUserID int64
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the TUI (Text User Interface) application",
	RunE: func(cmd *cobra.Command, args []string) error {
		if tuiUserID == 0 {
			return fmt.Errorf("user ID is required")
		}

		ctx := context.Background()
		m := bubbletea.NewModel(ctx, taskSvc, tuiUserID)

		// Configure with proper options for terminal programs
		p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
		return p.Start()
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)

	// Add user ID flag
	tuiCmd.Flags().Int64VarP(&tuiUserID, "user", "u", 0, "User ID (required)")
	tuiCmd.MarkFlagRequired("user")
}
