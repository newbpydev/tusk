// Package cli provides user-related CLI functionality (disabled)
package cli

// Note: User commands have been disabled as requested.
// All user management is now available through the TUI interface only.

// All commands and init functions are commented out to prevent them from registering with the CLI
/*
import (
	"context"

	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users in the Tusk application",
	Long:  "Commands to create, update, delete, and list users in the Tusk application.",
}

// createUserCmd creates a new user in the Tusk application.
// It requires a username, email, and password as flags.
var createUserCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	Long:  "Command to create a new user in the Tusk application.",
	RunE: func(cmd *cobra.Command, args []string) error {
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			return err
		}

		email, err := cmd.Flags().GetString("email")
		if err != nil {
			return err
		}

		password, err := cmd.Flags().GetString("password")
		if err != nil {
			return err
		}

		u, err := userSvc.Create(context.Background(), username, email, password)
		if err != nil {
			return err
		}

		cmd.Printf("Created user %d (%s)\n", u.ID, u.Username)
		return nil
	},
}

// loginCmd logs in a user to the Tusk application.
// It requires a username and password as flags.
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login as a user",
	Long:  "Command to log in as a user in the Tusk application.",
	RunE: func(cmd *cobra.Command, args []string) error {
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			return err
		}

		password, err := cmd.Flags().GetString("password")
		if err != nil {
			return err
		}

		u, err := userSvc.Login(context.Background(), username, password)
		if err != nil {
			return err
		}

		cmd.Printf("Welcome back, %s!\n", u.Username)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(createUserCmd, loginCmd)

	// Define flags for createUserCmd
	createUserCmd.Flags().StringP("username", "u", "", "Username for the new user")
	createUserCmd.Flags().StringP("email", "e", "", "Email for the new user")
	createUserCmd.Flags().StringP("password", "p", "", "Password for the new user")
	createUserCmd.MarkFlagRequired("username")
	createUserCmd.MarkFlagRequired("email")
	createUserCmd.MarkFlagRequired("password")

	// Define flags for loginCmd
	loginCmd.Flags().StringP("username", "u", "", "Username for login")
	loginCmd.Flags().StringP("password", "p", "", "Password for login")
	loginCmd.MarkFlagRequired("username")
	loginCmd.MarkFlagRequired("password")
}
*/
