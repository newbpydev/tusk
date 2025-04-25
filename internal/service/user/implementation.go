package user

import (
	"context"
	"time"

	"github.com/newbpydev/tusk/internal/core/errors"
	"github.com/newbpydev/tusk/internal/core/user"
	repo "github.com/newbpydev/tusk/internal/ports/output"
	"golang.org/x/crypto/bcrypt"
)

// UserService is the interface that defines the methods for managing users.
// It includes methods for creating and logging in users.
type userService struct {
	repo repo.UserRepository
}

// NewUserService creates a new instance of UserService with the given UserRepository.
// It returns a pointer to the userService struct.
func NewUserService(r repo.UserRepository) Service {
	return &userService{repo: r}
}

// Create creates a new user with the given username, email, and password.
// It returns the created user or an error if the user could not be created.
func (s *userService) Create(ctx context.Context, username, email, password string) (user.User, error) {
	// Validate inputs
	if username == "" {
		return user.User{}, errors.InvalidInput("username is required")
	}
	if email == "" {
		return user.User{}, errors.InvalidInput("email is required")
	}
	if password == "" {
		return user.User{}, errors.InvalidInput("password is required")
	}

	// Check if user already exists with the same username
	existingUser, err := s.repo.GetByUsername(ctx, username)
	if err == nil && existingUser.ID != 0 {
		return user.User{}, errors.Conflict("username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return user.User{}, errors.InternalError("failed to hash password")
	}

	now := time.Now()
	newUser := user.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	// Create the user in the repository
	createdUser, err := s.repo.Create(ctx, newUser)
	if err != nil {
		return user.User{}, err
	}

	return createdUser, nil
}

// Login authenticates a user with the given username and password.
// It returns the user if authentication is successful or an error if not.
func (s *userService) Login(ctx context.Context, username, password string) (user.User, error) {
	// Validate inputs
	if username == "" {
		return user.User{}, errors.InvalidInput("username is required")
	}
	if password == "" {
		return user.User{}, errors.InvalidInput("password is required")
	}

	// Get the user by username
	foundUser, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return user.User{}, errors.Unauthorized("invalid username or password")
	}

	// Compare the password hash
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password))
	if err != nil {
		return user.User{}, errors.Unauthorized("invalid username or password")
	}

	// Update the last login time
	now := time.Now()
	foundUser.LastLogin = &now
	foundUser.UpdatedAt = now

	// Update the user in the repository
	updatedUser, err := s.repo.Update(ctx, foundUser)
	if err != nil {
		return user.User{}, err
	}

	return updatedUser, nil
}
