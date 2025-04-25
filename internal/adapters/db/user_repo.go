package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/newbpydev/tusk/internal/adapters/db/sqlc"
	"github.com/newbpydev/tusk/internal/core/errors"
	"github.com/newbpydev/tusk/internal/core/user"
)

// SQLUserRepo is a struct that implements the UserRepository interface
// and provides methods to interact with the user table in the database.
type SQLUserRepo struct {
	q *sqlc.Queries
}

// NewSQLUserRepo creates a new instance of SQLUserRepo with the given database connection pool.
// It initializes the SQL queries using sqlc and returns a pointer to the SQLUserRepo instance.
func NewSQLUserRepo(pool *pgxpool.Pool) *SQLUserRepo {
	return &SQLUserRepo{
		q: sqlc.New(pool),
	}
}

// Create creates a new user in the database using the provided user.User struct.
// It returns the created user.User struct or an error if the operation fails.
func (r *SQLUserRepo) Create(ctx context.Context, u user.User) (user.User, error) {
	row, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
	})
	if err != nil {
		return user.User{}, errors.InternalError(fmt.Sprintf("failed to create user: %v", err))
	}

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
	if err != nil {
		return user.User{}, errors.InternalError(fmt.Sprintf("failed to update user: %v", err))
	}

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
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		return user.User{}, errors.NotFound(fmt.Sprintf("user with username %s not found: %v", username, err))
	}

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
	row, err := r.q.GetUserById(ctx, int32(id))
	if err != nil {
		return user.User{}, errors.NotFound(fmt.Sprintf("user with id %d not found: %v", id, err))
	}

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
