package output

import (
	"context"

	"github.com/newbpydev/tusk/internal/core/user"
)

// UserRepository defines the interface for user-related database operations.
// It provides methods for creating, updating, and retrieving users from the database.
type UserRepository interface {
	// Create creates a new user in the database.
	// It returns the created user or an error if the user could not be created.
	Create(ctx context.Context, user user.User) (user.User, error)

	// Update updates an existing user in the database.
	// It returns the updated user or an error if the user could not be updated.
	Update(ctx context.Context, user user.User) (user.User, error)

	// GetByID retrieves a user by their ID from the database.
	// It returns the user or an error if the user could not be found.
	GetByID(ctx context.Context, id int64) (user.User, error)

	// GetByUsername retrieves a user by their username from the database.
	// It returns the user or an error if the user could not be found.
	GetByUsername(ctx context.Context, username string) (user.User, error)
}
