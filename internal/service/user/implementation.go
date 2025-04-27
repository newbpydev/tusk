package user

import (
	"context"
	"time"

	"github.com/newbpydev/tusk/internal/core/errors"
	"github.com/newbpydev/tusk/internal/core/user"
	repo "github.com/newbpydev/tusk/internal/ports/output"
	"github.com/newbpydev/tusk/internal/util/logging"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UserService is the interface that defines the methods for managing users.
// It includes methods for creating and logging in users.
type userService struct {
	repo repo.UserRepository
	log  *zap.Logger
}

// NewUserService creates a new instance of UserService with the given UserRepository.
// It returns a pointer to the userService struct.
func NewUserService(r repo.UserRepository) Service {
	return &userService{
		repo: r,
		log:  logging.GetFileOnlyLogger("service.user"),
	}
}

// Create creates a new user with the given username, email, and password.
// It returns the created user or an error if the user could not be created.
func (s *userService) Create(ctx context.Context, username, email, password string) (user.User, error) {
	// Validate inputs
	if username == "" {
		s.log.Error("Failed to create user: username is required")
		return user.User{}, errors.InvalidInput("username is required")
	}
	if email == "" {
		s.log.Error("Failed to create user: email is required")
		return user.User{}, errors.InvalidInput("email is required")
	}
	if password == "" {
		s.log.Error("Failed to create user: password is required")
		return user.User{}, errors.InvalidInput("password is required")
	}

	s.log.Info("Attempting to create new user",
		zap.String("username", username),
		// Do NOT log the password or full email
		// Only log email domain for troubleshooting patterns
		zap.String("email_domain", emailDomain(email)))

	// Check if user already exists with the same username
	existingUser, err := s.repo.GetByUsername(ctx, username)
	if err == nil && existingUser.ID != 0 {
		s.log.Warn("Username already exists",
			zap.String("username", username))
		return user.User{}, errors.Conflict("username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("Failed to hash password",
			zap.Error(err))
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
		s.log.Error("Failed to create user in repository",
			zap.String("username", username),
			zap.Error(err))
		return user.User{}, err
	}

	s.log.Info("User created successfully",
		zap.Int32("user_id", createdUser.ID),
		zap.String("username", createdUser.Username))

	return createdUser, nil
}

// Login authenticates a user with the given username and password.
// It returns the user if authentication is successful or an error if not.
func (s *userService) Login(ctx context.Context, username, password string) (user.User, error) {
	// Validate inputs
	if username == "" {
		s.log.Error("Login attempt with empty username")
		return user.User{}, errors.InvalidInput("username is required")
	}
	if password == "" {
		s.log.Error("Login attempt with empty password",
			zap.String("username", username))
		return user.User{}, errors.InvalidInput("password is required")
	}

	s.log.Debug("Login attempt",
		zap.String("username", username))

	// Get the user by username
	foundUser, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		// Don't reveal if the username exists or not in the logs
		// to prevent user enumeration
		s.log.Info("Failed login attempt: invalid username",
			zap.String("attempted_username", username))
		return user.User{}, errors.Unauthorized("invalid username or password")
	}

	// Compare the password hash
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password))
	if err != nil {
		s.log.Warn("Failed login attempt: invalid password",
			zap.String("username", username),
			zap.Int32("user_id", foundUser.ID))
		return user.User{}, errors.Unauthorized("invalid username or password")
	}

	// Update the last login time
	now := time.Now()
	foundUser.LastLogin = &now
	foundUser.UpdatedAt = now

	// Update the user in the repository
	updatedUser, err := s.repo.Update(ctx, foundUser)
	if err != nil {
		s.log.Error("Failed to update last login timestamp",
			zap.String("username", username),
			zap.Int32("user_id", foundUser.ID),
			zap.Error(err))
		return user.User{}, err
	}

	s.log.Info("User logged in successfully",
		zap.String("username", username),
		zap.Int32("user_id", updatedUser.ID))

	return updatedUser, nil
}

// emailDomain extracts domain part from an email address
// Returns empty string if invalid email format
func emailDomain(email string) string {
	for i := 0; i < len(email); i++ {
		if email[i] == '@' && i+1 < len(email) {
			return email[i+1:]
		}
	}
	return ""
}
