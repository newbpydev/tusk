package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the Tusk task manager application",
	Long:  "Launch the interactive Task User Interface for managing your tasks, subtasks, and projects.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var userID int64

		// Show welcome intro - only once
		showWelcomeIntro()

		// Use a simpler direct terminal input method for authentication
		err := simpleTerminalAuth(ctx, &userID)
		if err != nil {
			if err == errAuthCancelled {
				fmt.Println("Authentication cancelled. Goodbye!")
				return nil // Return nil to avoid showing error message
			}
			return err
		}

		// Start TUI with authenticated user
		m := bubbletea.NewModel(ctx, taskSvc, userID)
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
	time.Sleep(1 * time.Second)
}

// readLine reads a line from stdin
func readLine(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	// Trim newline and carriage return
	return strings.TrimSpace(input), nil
}

// readPassword reads a password from stdin without echoing
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // Add a newline after the password
	if err != nil {
		return "", err
	}
	return string(bytePassword), nil
}

// errAuthCancelled is returned when the user cancels the authentication process
var errAuthCancelled = fmt.Errorf("authentication cancelled by user")

// simpleTerminalAuth handles authentication using direct terminal input
func simpleTerminalAuth(ctx context.Context, userID *int64) error {
	// Ask whether to login or create account
	fmt.Println("Choose an action:")
	fmt.Println("1) Login to existing account")
	fmt.Println("2) Create new account")
	fmt.Println("3) Cancel")

	choice, err := readLine("Enter choice (1-3): ")
	if err != nil {
		return fmt.Errorf("failed to read choice: %v", err)
	}

	// Convert choice to index
	idx := 0
	switch choice {
	case "1":
		idx = 0
	case "2":
		idx = 1
	case "3", "q", "exit", "cancel":
		idx = 2
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}

	// Handle cancellation
	if idx == 2 {
		return errAuthCancelled
	}

	// Username prompt
	username, err := readLine("Username: ")
	if err != nil {
		return fmt.Errorf("failed to read username: %v", err)
	}
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Password prompt
	password, err := readPassword("Password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %v", err)
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
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
		email, err := readLine("Email: ")
		if err != nil {
			return fmt.Errorf("failed to read email: %v", err)
		}
		if email == "" {
			return fmt.Errorf("email cannot be empty")
		}
		if !strings.Contains(email, "@") {
			return fmt.Errorf("invalid email format")
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
