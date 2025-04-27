package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/manifoldco/promptui"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the Tusk task manager application",
	Long:  "Launch the interactive Task User Interface for managing your tasks, subtasks, and projects.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var userID int64

		// Show welcome intro
		showWelcomeIntro()

		// Always use interactive authentication
		err := interactiveAuth(ctx, &userID)
		if err != nil {
			if err == errAuthCancelled {
				fmt.Println("Authentication cancelled. Goodbye!")
				return nil // Return nil to avoid showing error message
			}
			return err
		}

		// Start TUI with authenticated user
		m := bubbletea.NewModel(ctx, taskSvc, userID)

		// Configure with proper options for terminal programs
		p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
		return p.Start()
	},
}

// showWelcomeIntro displays a friendly introduction to the Tusk application
func showWelcomeIntro() {
	// Clear the terminal screen with ANSI escape code
	fmt.Print("\033[H\033[2J")

	// Improved ASCII art logo that clearly spells "TUSK"
	logo := `
  _____  _    _  _____  _  __
 |_   _|| |  | |/ ____|| |/ /
   | |  | |  | | (___  | ' / 
   | |  | |  | |\___ \ |  <  
   | |  | |__| |____) || . \ 
   |_|   \____/|_____/ |_|\_\
                             
`
	fmt.Println(logo)
	fmt.Println("Welcome to Tusk - Your Personal Task Management System")
	fmt.Println("====================================================")
	fmt.Println()
	fmt.Println("Tusk helps you organize your tasks, track progress, and boost productivity.")
	fmt.Println("Features include:")
	fmt.Println("• Create and manage tasks with priorities and deadlines")
	fmt.Println("• Organize tasks into projects")
	fmt.Println("• Track your productivity with built-in reports")
	fmt.Println("• Simple and intuitive interface")
	fmt.Println()
	fmt.Println("Let's get started!")
	fmt.Println()

	// Give users a moment to read the introduction
	time.Sleep(2 * time.Second)
}

// errAuthCancelled is returned when the user cancels the authentication process
var errAuthCancelled = fmt.Errorf("authentication cancelled by user")

// interactiveAuth handles interactive authentication flow
func interactiveAuth(ctx context.Context, userID *int64) error {
	// Ask whether to login or create account
	promptSelect := promptui.Select{
		Label: "Choose an action",
		Items: []string{"Login to existing account", "Create new account", "Cancel"},
	}

	idx, _, err := promptSelect.Run()
	if err != nil {
		return fmt.Errorf("prompt failed: %v", err)
	}

	// Handle cancellation
	if idx == 2 {
		return errAuthCancelled
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
	// No flags added - we're using interactive authentication only
}
