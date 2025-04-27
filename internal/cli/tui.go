package cli

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/manifoldco/promptui"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea"
	"github.com/newbpydev/tusk/internal/core/errors"
	"github.com/spf13/cobra"
)

var (
	// Authentication flags for TUI mode
	tuiUsername string
	tuiPassword string
	createUser  bool
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the TUI (Text User Interface) application",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var userID int64
		var err error

		// First attempt to authenticate with provided credentials
		if tuiUsername != "" && tuiPassword != "" {
			user, err := userSvc.Login(ctx, tuiUsername, tuiPassword)
			if err != nil {
				if createUser {
					// If login failed but create flag is set, prompt for email
					prompt := promptui.Prompt{
						Label: "Email for new account",
						Validate: func(input string) error {
							if input == "" {
								return fmt.Errorf("email cannot be empty")
							}
							if !strings.Contains(input, "@") {
								return fmt.Errorf("invalid email format")
							}
							return nil
						},
					}

					email, err := prompt.Run()
					if err != nil {
						return fmt.Errorf("failed to read email: %v", err)
					}

					// Create new user
					user, err = userSvc.Create(ctx, tuiUsername, email, tuiPassword)
					if err != nil {
						return fmt.Errorf("failed to create user: %v", err)
					}
					fmt.Printf("Created new user: %s\n", user.Username)
				} else {
					// Not creating a new user, offer options
					if errors.IsUnauthorized(err) {
						return fmt.Errorf("invalid username or password. Use --create flag to create a new account")
					}
					return fmt.Errorf("authentication failed: %v", err)
				}
			}

			userID = int64(user.ID)
			fmt.Printf("Welcome, %s!\n", user.Username)
		} else {
			// Interactive mode if no credentials provided
			err = interactiveAuth(ctx, &userID)
			if err != nil {
				return err
			}
		}

		// Start TUI with authenticated user
		m := bubbletea.NewModel(ctx, taskSvc, userID)

		// Configure with proper options for terminal programs
		p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
		return p.Start()
	},
}

// interactiveAuth handles interactive authentication flow
func interactiveAuth(ctx context.Context, userID *int64) error {
	// Ask whether to login or create account
	promptSelect := promptui.Select{
		Label: "Choose an action",
		Items: []string{"Login to existing account", "Create new account"},
	}

	idx, _, err := promptSelect.Run()
	if err != nil {
		return fmt.Errorf("prompt failed: %v", err)
	}

	// Username prompt
	usernamePrompt := promptui.Prompt{
		Label: "Username",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("username cannot be empty")
			}
			return nil
		},
	}

	username, err := usernamePrompt.Run()
	if err != nil {
		return fmt.Errorf("prompt failed: %v", err)
	}

	// Password prompt
	passwordPrompt := promptui.Prompt{
		Label: "Password",
		Mask:  '*',
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("password cannot be empty")
			}
			return nil
		},
	}

	password, err := passwordPrompt.Run()
	if err != nil {
		return fmt.Errorf("prompt failed: %v", err)
	}

	if idx == 0 {
		// Login flow
		user, err := userSvc.Login(ctx, username, password)
		if err != nil {
			return fmt.Errorf("login failed: %v", err)
		}
		*userID = int64(user.ID)
		fmt.Printf("Welcome back, %s!\n", user.Username)
	} else {
		// Create account flow
		emailPrompt := promptui.Prompt{
			Label: "Email",
			Validate: func(input string) error {
				if input == "" {
					return fmt.Errorf("email cannot be empty")
				}
				if !strings.Contains(input, "@") {
					return fmt.Errorf("invalid email format")
				}
				return nil
			},
		}

		email, err := emailPrompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %v", err)
		}

		user, err := userSvc.Create(ctx, username, email, password)
		if err != nil {
			return fmt.Errorf("failed to create account: %v", err)
		}
		*userID = int64(user.ID)
		fmt.Printf("Account created! Welcome, %s!\n", user.Username)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(tuiCmd)

	// Add authentication flags
	tuiCmd.Flags().StringVarP(&tuiUsername, "username", "u", "", "Username for authentication")
	tuiCmd.Flags().StringVarP(&tuiPassword, "password", "p", "", "Password for authentication")
	tuiCmd.Flags().BoolVarP(&createUser, "create", "c", false, "Create a new user if login fails")
}
