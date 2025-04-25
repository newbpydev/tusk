package user

import (
	"context"

	"github.com/newbpydev/tusk/internal/core/user"
)

// Service is the interface that defines the methods for managing users.
// It includes methods for creating and logging in users.
type Service interface {
	Create(ctx context.Context, username, email, password string) (user.User, error)
	Login(ctx context.Context, username, password string) (user.User, error)
}
