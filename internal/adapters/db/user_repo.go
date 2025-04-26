package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/newbpydev/tusk/internal/adapters/db/sqlc"
	"github.com/newbpydev/tusk/internal/core/errors"
	"github.com/newbpydev/tusk/internal/core/user"
	"github.com/newbpydev/tusk/internal/util/logging"
	"go.uber.org/zap"
)

// SQLUserRepo is a struct that implements the UserRepository interface
// and provides methods to interact with the user table in the database.
type SQLUserRepo struct {
	q   *sqlc.Queries
	log *zap.Logger
}

// NewSQLUserRepo creates a new instance of SQLUserRepo with the given database connection pool.
// It initializes the SQL queries using sqlc and returns a pointer to the SQLUserRepo instance.
func NewSQLUserRepo(pool *pgxpool.Pool) *SQLUserRepo {
	return &SQLUserRepo{
		q:   sqlc.New(pool),
		log: logging.DBLogger.Named("user_repo"),
	}
}

// Create creates a new user in the database using the provided user.User struct.
// It returns the created user.User struct or an error if the operation fails.
func (r *SQLUserRepo) Create(ctx context.Context, u user.User) (user.User, error) {
	// Only log username - avoid logging email or password hash for privacy
	r.log.Debug("Creating new user in database",
		zap.String("username", u.Username),
		zap.String("email_domain", emailDomain(u.Email)))

	startTime := time.Now()
	row, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
	})
	queryDuration := time.Since(startTime)

	if err != nil {
		r.log.Error("Failed to create user in database",
			zap.String("username", u.Username),
			zap.Duration("duration_ms", queryDuration),
			zap.Error(err))
		return user.User{}, errors.InternalError(fmt.Sprintf("failed to create user: %v", err))
	}

	r.log.Info("User created successfully in database",
		zap.Int32("user_id", row.ID),
		zap.String("username", u.Username),
		zap.Duration("duration_ms", queryDuration))

	return user.User{
		ID:           row.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

// Update updates an existing user in the database using the provided user.User struct.
// It returns the updated user.User struct or an error if the operation fails.
func (r *SQLUserRepo) Update(ctx context.Context, u user.User) (user.User, error) {
	// Avoid logging sensitive data
	r.log.Debug("Updating user in database",
		zap.Int32("user_id", u.ID),
		zap.String("username", u.Username),
		zap.String("email_domain", emailDomain(u.Email)),
		zap.Bool("is_active", u.IsActive))

	startTime := time.Now()
	row, err := r.q.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		IsActive: pgtype.Bool{
			Bool:  u.IsActive,
			Valid: true,
		},
	})
	queryDuration := time.Since(startTime)

	if err != nil {
		r.log.Error("Failed to update user in database",
			zap.Int32("user_id", u.ID),
			zap.String("username", u.Username),
			zap.Duration("duration_ms", queryDuration),
			zap.Error(err))
		return user.User{}, errors.InternalError(fmt.Sprintf("failed to update user: %v", err))
	}

	r.log.Info("User updated successfully in database",
		zap.Int32("user_id", row.ID),
		zap.String("username", u.Username),
		zap.Duration("duration_ms", queryDuration))

	return user.User{
		ID:       row.ID,
		Username: row.Username,
		Email:    row.Email,
		// PasswordHash field is not present in UpdateUserRow, so it is omitted.
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		IsActive:  row.IsActive.Bool,
	}, nil
}

// GetByUsername retrieves a user from the database by username.
// It returns the user.User struct or an error if the user is not found or if the operation fails.
func (r *SQLUserRepo) GetByUsername(ctx context.Context, username string) (user.User, error) {
	r.log.Debug("Fetching user by username",
		zap.String("username", username))

	startTime := time.Now()
	row, err := r.q.GetUserByUsername(ctx, username)
	queryDuration := time.Since(startTime)

	if err != nil {
		r.log.Warn("User not found by username",
			zap.String("username", username),
			zap.Duration("duration_ms", queryDuration),
			zap.Error(err))
		return user.User{}, errors.NotFound(fmt.Sprintf("user with username %s not found", username))
	}

	r.log.Debug("User fetched successfully by username",
		zap.Int32("user_id", row.ID),
		zap.String("username", username),
		zap.Duration("duration_ms", queryDuration))

	return user.User{
		ID:        row.ID,
		Username:  row.Username,
		Email:     row.Email,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		IsActive:  row.IsActive.Bool,
		LastLogin: &row.LastLogin.Time,
	}, nil
}

// GetByID retrieves a user from the database by ID.
// It returns the user.User struct or an error if the user is not found or if the operation fails.
func (r *SQLUserRepo) GetByID(ctx context.Context, id int64) (user.User, error) {
	r.log.Debug("Fetching user by ID",
		zap.Int64("user_id", id))

	startTime := time.Now()
	row, err := r.q.GetUserById(ctx, int32(id))
	queryDuration := time.Since(startTime)

	if err != nil {
		r.log.Warn("User not found by ID",
			zap.Int64("user_id", id),
			zap.Duration("duration_ms", queryDuration),
			zap.Error(err))
		return user.User{}, errors.NotFound(fmt.Sprintf("user with id %d not found", id))
	}

	r.log.Debug("User fetched successfully by ID",
		zap.Int64("user_id", id),
		zap.String("username", row.Username),
		zap.Duration("duration_ms", queryDuration))

	return user.User{
		ID:        row.ID,
		Username:  row.Username,
		Email:     row.Email,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		IsActive:  row.IsActive.Bool,
		LastLogin: &row.LastLogin.Time,
	}, nil
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
