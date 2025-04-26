// Copyright (C) 2025 Juan Antonio Gomez Pena
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package db

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	sqlcgen "github.com/newbpydev/tusk/internal/adapters/db/sqlc"
	"github.com/newbpydev/tusk/internal/config"
	domainerrors "github.com/newbpydev/tusk/internal/core/errors"
	"github.com/newbpydev/tusk/internal/core/user"
	"github.com/newbpydev/tusk/internal/util/logging"
)

var (
	userTestDBPool *pgxpool.Pool
	userTestRepo   *SQLUserRepo
	userTestCtx    context.Context
)

// Test setup helper functions

func setupUserTestDB(t *testing.T) {
	// Initialize logging with a test configuration
	cfg := &config.Config{
		AppEnv: "test",
		DBURL:  "test-db-url",
	}

	// Set up log directory for tests
	testLogDir := os.TempDir()
	os.Setenv("LOG_DIR", testLogDir)

	// Initialize logging with the test config
	err := logging.Init(cfg)
	if err != nil {
		// Fallback to test logger if initialization fails
		logging.Logger = zaptest.NewLogger(t)
		// Create DBLogger if it doesn't exist
		if logging.DBLogger == nil {
			logging.DBLogger = logging.Logger.Named("db-test")
		}
	}

	// Get test database URL from environment variable with more flexible fallback options
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		// Try to build connection string from individual components if available
		username := os.Getenv("TEST_DB_USER")
		if username == "" {
			username = "postgres"
		}

		password := os.Getenv("TEST_DB_PASSWORD")
		// Don't use default password as it's likely not correct in most environments
		if password == "" {
			// Check if the user is running this on their local machine
			t.Log("Warning: TEST_DB_PASSWORD environment variable not set. Please set it to run the tests.")
			t.Skip("Skipping test as TEST_DB_PASSWORD is not set")
			return
		}

		host := os.Getenv("TEST_DB_HOST")
		if host == "" {
			host = "localhost"
		}

		port := os.Getenv("TEST_DB_PORT")
		if port == "" {
			port = "5432"
		}

		dbName := os.Getenv("TEST_DB_NAME")
		if dbName == "" {
			dbName = "tusk_test"
		}

		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			username, password, host, port, dbName)

		t.Logf("Using constructed database URL: %s",
			fmt.Sprintf("postgres://%s:***@%s:%s/%s?sslmode=disable",
				username, host, port, dbName))
	}

	userTestCtx = context.Background()

	// Connect to test database
	userTestDBPool, err = pgxpool.New(userTestCtx, dbURL)
	if err != nil {
		t.Skipf("Skipping database tests - could not connect to test database: %v", err)
		return
	}

	// Clean up database before tests
	cleanUserDB(t)

	// Create repositories
	userTestRepo = NewSQLUserRepo(userTestDBPool)

	// If not created properly, create a basic instance for testing
	if userTestRepo == nil {
		queries := sqlcgen.New(userTestDBPool)
		userTestRepo = &SQLUserRepo{
			q:   queries,
			log: zaptest.NewLogger(t).Named("user-test"),
		}
	}
}

func cleanUserDB(t *testing.T) {
	// Delete all users and reset sequences
	_, err := userTestDBPool.Exec(userTestCtx, "DELETE FROM users")
	require.NoError(t, err)
	_, err = userTestDBPool.Exec(userTestCtx, "ALTER SEQUENCE users_id_seq RESTART WITH 1")
	require.NoError(t, err)
}

func teardownUserTestDB() {
	if userTestDBPool != nil {
		userTestDBPool.Close()
	}
}

// Test cases for User Repository

func TestUserRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	setupUserTestDB(t)
	defer teardownUserTestDB()

	testCases := []struct {
		name          string
		userToCreate  user.User
		expectedError bool
	}{
		{
			name: "Create basic user",
			userToCreate: user.User{
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: "hashedpassword",
			},
			expectedError: false,
		},
		{
			name: "Create user with active status",
			userToCreate: user.User{
				Username:     "activeuser",
				Email:        "active@example.com",
				PasswordHash: "hashedpassword",
				IsActive:     true,
			},
			expectedError: false,
		},
		{
			name: "Create user with duplicate username",
			userToCreate: user.User{
				Username:     "duplicateuser",
				Email:        "duplicate@example.com",
				PasswordHash: "hashedpassword",
			},
			expectedError: false, // First creation should succeed
		},
		{
			name: "Create duplicate user",
			userToCreate: user.User{
				Username:     "duplicateuser",
				Email:        "another@example.com",
				PasswordHash: "hashedpassword",
			},
			expectedError: true, // Second creation with same username should fail
		},
		{
			name: "Create user with duplicate email",
			userToCreate: user.User{
				Username:     "anotherduplicate",
				Email:        "duplicate@example.com", // Same email as above
				PasswordHash: "hashedpassword",
			},
			expectedError: true, // Creation with duplicate email should fail
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createdUser, err := userTestRepo.Create(userTestCtx, tc.userToCreate)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, createdUser.ID)
				assert.Equal(t, tc.userToCreate.Username, createdUser.Username)
				assert.Equal(t, tc.userToCreate.Email, createdUser.Email)
				assert.Equal(t, tc.userToCreate.PasswordHash, createdUser.PasswordHash)
				assert.NotZero(t, createdUser.CreatedAt)
				assert.NotZero(t, createdUser.UpdatedAt)
			}
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	setupUserTestDB(t)
	defer teardownUserTestDB()

	// Create a user first
	userToCreate := user.User{
		Username:     "getuser",
		Email:        "get@example.com",
		PasswordHash: "hashedpassword",
	}

	createdUser, err := userTestRepo.Create(userTestCtx, userToCreate)
	require.NoError(t, err)
	require.NotZero(t, createdUser.ID)

	// Test getting the user
	retrievedUser, err := userTestRepo.GetByUsername(userTestCtx, createdUser.Username)
	assert.NoError(t, err)
	assert.Equal(t, createdUser.ID, retrievedUser.ID)
	assert.Equal(t, createdUser.Username, retrievedUser.Username)
	assert.Equal(t, createdUser.Email, retrievedUser.Email)

	// Test getting a non-existent user
	_, err = userTestRepo.GetByUsername(userTestCtx, "nonexistentuser")
	assert.Error(t, err)
	assert.True(t, domainerrors.IsNotFound(err))
}

func TestUserRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	setupUserTestDB(t)
	defer teardownUserTestDB()

	// Create a user first
	userToCreate := user.User{
		Username:     "iduser",
		Email:        "id@example.com",
		PasswordHash: "hashedpassword",
	}

	createdUser, err := userTestRepo.Create(userTestCtx, userToCreate)
	require.NoError(t, err)
	require.NotZero(t, createdUser.ID)

	// Test getting the user
	retrievedUser, err := userTestRepo.GetByID(userTestCtx, int64(createdUser.ID))
	assert.NoError(t, err)
	assert.Equal(t, createdUser.ID, retrievedUser.ID)
	assert.Equal(t, createdUser.Username, retrievedUser.Username)
	assert.Equal(t, createdUser.Email, retrievedUser.Email)

	// Test getting a non-existent user
	_, err = userTestRepo.GetByID(userTestCtx, int64(99999))
	assert.Error(t, err)
	assert.True(t, domainerrors.IsNotFound(err))
}

func TestUserRepository_GetByEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	setupUserTestDB(t)
	defer teardownUserTestDB()

	// Create a user first
	userToCreate := user.User{
		Username:     "emailuser",
		Email:        "find@example.com",
		PasswordHash: "hashedpassword",
	}

	createdUser, err := userTestRepo.Create(userTestCtx, userToCreate)
	require.NoError(t, err)
	require.NotZero(t, createdUser.ID)

	// Test getting the user by email
	retrievedUser, err := userTestRepo.GetByEmail(userTestCtx, createdUser.Email)
	assert.NoError(t, err)
	assert.Equal(t, createdUser.ID, retrievedUser.ID)
	assert.Equal(t, createdUser.Username, retrievedUser.Username)
	assert.Equal(t, createdUser.Email, retrievedUser.Email)

	// Test getting a non-existent user by email
	_, err = userTestRepo.GetByEmail(userTestCtx, "nonexistent@example.com")
	assert.Error(t, err)
	assert.True(t, domainerrors.IsNotFound(err))
}

func TestUserRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	setupUserTestDB(t)
	defer teardownUserTestDB()

	// Create a user first
	initialUser := user.User{
		Username:     "updateuser",
		Email:        "update@example.com",
		PasswordHash: "hashedpassword",
	}

	createdUser, err := userTestRepo.Create(userTestCtx, initialUser)
	require.NoError(t, err)

	// Update the user
	updatedUserData := createdUser
	updatedUserData.Username = "updatedusername"
	updatedUserData.Email = "updated@example.com"
	updatedUserData.PasswordHash = "newhashpassword"
	updatedUserData.IsActive = true

	updatedUser, err := userTestRepo.Update(userTestCtx, updatedUserData)
	assert.NoError(t, err)

	// Verify the returned updated user has the correct values
	assert.Equal(t, updatedUserData.Username, updatedUser.Username)
	assert.Equal(t, updatedUserData.Email, updatedUser.Email)
	// Password hash is not returned from the Update method as per implementation
	assert.Equal(t, updatedUserData.IsActive, updatedUser.IsActive)

	// Retrieve the user to verify updates in the database
	retrievedUser, err := userTestRepo.GetByID(userTestCtx, int64(createdUser.ID))
	assert.NoError(t, err)
	assert.Equal(t, "updatedusername", retrievedUser.Username)
	assert.Equal(t, "updated@example.com", retrievedUser.Email)
	assert.True(t, retrievedUser.IsActive)

	// Test updating a non-existent user
	nonExistentUser := user.User{
		ID:           99999,
		Username:     "nonexistent",
		Email:        "nonexistent@example.com",
		PasswordHash: "hashedpassword",
	}
	_, err = userTestRepo.Update(userTestCtx, nonExistentUser)
	assert.Error(t, err)
	// This should be an internal error, not a not found error based on the implementation
	assert.Contains(t, err.Error(), "failed to update user")
}

func TestUserRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	setupUserTestDB(t)
	defer teardownUserTestDB()

	// Create a user first
	userToCreate := user.User{
		Username:     "deleteuser",
		Email:        "delete@example.com",
		PasswordHash: "hashedpassword",
	}

	createdUser, err := userTestRepo.Create(userTestCtx, userToCreate)
	require.NoError(t, err)
	require.NotZero(t, createdUser.ID)

	// Test deleting the user
	err = userTestRepo.Delete(userTestCtx, int64(createdUser.ID))
	assert.NoError(t, err)

	// Verify the user was deleted by trying to retrieve it
	_, err = userTestRepo.GetByID(userTestCtx, int64(createdUser.ID))
	assert.Error(t, err)
	assert.True(t, domainerrors.IsNotFound(err))

	// Test deleting a non-existent user
	err = userTestRepo.Delete(userTestCtx, int64(99999))
	assert.Error(t, err)
	assert.True(t, domainerrors.IsNotFound(err))
}
