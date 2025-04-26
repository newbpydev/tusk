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
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	sqlcgen "github.com/newbpydev/tusk/internal/adapters/db/sqlc"
	"github.com/newbpydev/tusk/internal/config"
	"github.com/newbpydev/tusk/internal/core/task"
	"github.com/newbpydev/tusk/internal/util/logging"
)

var (
	testDBPool *pgxpool.Pool
	testRepo   *SQLTaskRepository
	ctx        context.Context
)

// Test setup helper functions

func setupTestDB(t *testing.T) {
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
		if password == "" {
			password = "postgres" // Default password, change this if needed
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

	ctx = context.Background()

	// Connect to test database
	// Reuse the existing err variable instead of redeclaring it
	testDBPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Skipping database tests - could not connect to test database: %v", err)
		return
	}

	// Clean up database before tests
	cleanDB(t)

	// Create repositories
	testRepo = NewSQLTaskRepository(testDBPool)

	// If not created properly, create a basic instance for testing
	if testRepo == nil {
		queries := sqlcgen.New(testDBPool)
		testRepo = &SQLTaskRepository{
			q:   queries,
			log: zaptest.NewLogger(t).Named("task-test"),
		}
	}
}

func cleanDB(t *testing.T) {
	// Delete all tasks and reset sequences
	_, err := testDBPool.Exec(ctx, "DELETE FROM tasks")
	require.NoError(t, err)
	_, err = testDBPool.Exec(ctx, "ALTER SEQUENCE tasks_id_seq RESTART WITH 1")
	require.NoError(t, err)
}

func teardownTestDB() {
	if testDBPool != nil {
		testDBPool.Close()
	}
}

// createTestUser creates a test user in the database
func createTestUser(t *testing.T) int32 {
	result, err := testDBPool.Exec(ctx, `
		INSERT INTO users (username, email, password_hash, created_at, updated_at)
		VALUES ('testuser', 'test@example.com', 'hashedpassword', NOW(), NOW())
		RETURNING id
	`)
	require.NoError(t, err)

	rows := result.RowsAffected()
	require.Equal(t, int64(1), rows)

	var userID int32
	err = testDBPool.QueryRow(ctx, "SELECT lastval()").Scan(&userID)
	require.NoError(t, err)

	return userID
}

// Test cases for Task Repository

func TestTaskRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	setupTestDB(t)
	defer teardownTestDB()

	userID := createTestUser(t)

	testCases := []struct {
		name          string
		taskToCreate  task.Task
		expectedError bool
	}{
		{
			name: "Create basic task",
			taskToCreate: task.Task{
				UserID:      userID,
				Title:       "Test Task",
				Description: stringPtr("Test Description"),
				Status:      task.StatusTodo,
				Priority:    task.PriorityMedium,
				IsCompleted: false,
				Tags:        []task.Tag{{Name: "test"}, {Name: "integration"}},
			},
			expectedError: false,
		},
		{
			name: "Create task with due date",
			taskToCreate: task.Task{
				UserID:      userID,
				Title:       "Task with due date",
				DueDate:     timePtr(time.Now().Add(24 * time.Hour)),
				Status:      task.StatusTodo,
				Priority:    task.PriorityHigh,
				IsCompleted: false,
			},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createdTask, err := testRepo.Create(ctx, tc.taskToCreate)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, createdTask.ID)
				assert.Equal(t, tc.taskToCreate.Title, createdTask.Title)
				assert.Equal(t, tc.taskToCreate.UserID, createdTask.UserID)
				assert.Equal(t, tc.taskToCreate.Status, createdTask.Status)
				assert.Equal(t, tc.taskToCreate.Priority, createdTask.Priority)
				assert.NotZero(t, createdTask.CreatedAt)
				assert.NotZero(t, createdTask.UpdatedAt)

				if tc.taskToCreate.Description != nil {
					assert.Equal(t, *tc.taskToCreate.Description, *createdTask.Description)
				}

				if tc.taskToCreate.DueDate != nil {
					assert.WithinDuration(t, *tc.taskToCreate.DueDate, *createdTask.DueDate, time.Second)
				}

				if len(tc.taskToCreate.Tags) > 0 {
					assert.Equal(t, len(tc.taskToCreate.Tags), len(createdTask.Tags))
					for i, tag := range tc.taskToCreate.Tags {
						assert.Equal(t, tag.Name, createdTask.Tags[i].Name)
					}
				}
			}
		})
	}
}

func TestTaskRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	setupTestDB(t)
	defer teardownTestDB()

	userID := createTestUser(t)

	// Create a task first
	taskToCreate := task.Task{
		UserID:      userID,
		Title:       "Test Task for Get",
		Description: stringPtr("Test Description for Get"),
		Status:      task.StatusTodo,
		Priority:    task.PriorityMedium,
		IsCompleted: false,
		Tags:        []task.Tag{{Name: "get-test"}},
	}

	createdTask, err := testRepo.Create(ctx, taskToCreate)
	require.NoError(t, err)
	require.NotZero(t, createdTask.ID)

	// Test getting the task
	retrievedTask, err := testRepo.GetByID(ctx, int64(createdTask.ID))
	assert.NoError(t, err)
	assert.Equal(t, createdTask.ID, retrievedTask.ID)
	assert.Equal(t, createdTask.Title, retrievedTask.Title)
	assert.Equal(t, *createdTask.Description, *retrievedTask.Description)
	assert.Equal(t, createdTask.Status, retrievedTask.Status)
	assert.Equal(t, createdTask.Tags[0].Name, retrievedTask.Tags[0].Name)

	// Test getting a non-existent task
	_, err = testRepo.GetByID(ctx, int64(999999))
	assert.Error(t, err)
}

func TestTaskRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	setupTestDB(t)
	defer teardownTestDB()

	userID := createTestUser(t)

	// Create a task first
	initialTask := task.Task{
		UserID:      userID,
		Title:       "Initial Title",
		Description: stringPtr("Initial Description"),
		Status:      task.StatusTodo,
		Priority:    task.PriorityMedium,
		IsCompleted: false,
	}

	createdTask, err := testRepo.Create(ctx, initialTask)
	require.NoError(t, err)

	// Update the task
	updatedTaskData := createdTask
	updatedTaskData.Title = "Updated Title"
	updatedTaskData.Description = stringPtr("Updated Description")
	updatedTaskData.Status = task.StatusInProgress
	updatedTaskData.Priority = task.PriorityHigh
	updatedTaskData.Tags = []task.Tag{{Name: "updated"}}

	err = testRepo.Update(ctx, updatedTaskData)
	assert.NoError(t, err)

	// Retrieve the task to verify updates
	retrievedTask, err := testRepo.GetByID(ctx, int64(createdTask.ID))
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", retrievedTask.Title)
	assert.Equal(t, "Updated Description", *retrievedTask.Description)
	assert.Equal(t, task.StatusInProgress, retrievedTask.Status)
	assert.Equal(t, task.PriorityHigh, retrievedTask.Priority)
	assert.Equal(t, "updated", retrievedTask.Tags[0].Name)
}

func TestTaskRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	setupTestDB(t)
	defer teardownTestDB()

	userID := createTestUser(t)

	// Create a task first
	taskToCreate := task.Task{
		UserID: userID,
		Title:  "Task To Delete",
		Status: task.StatusTodo,
	}

	createdTask, err := testRepo.Create(ctx, taskToCreate)
	require.NoError(t, err)

	// Delete the task
	err = testRepo.Delete(ctx, int64(createdTask.ID))
	assert.NoError(t, err)

	// Try to retrieve the deleted task
	_, err = testRepo.GetByID(ctx, int64(createdTask.ID))
	assert.Error(t, err) // Should error as task should be deleted
}

func TestTaskRepository_ListRootTasks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	setupTestDB(t)
	defer teardownTestDB()

	userID := createTestUser(t)

	// Create several root tasks
	for i := 0; i < 3; i++ {
		task := task.Task{
			UserID: userID,
			Title:  "Root Task " + string(rune('A'+i)),
			Status: task.StatusTodo,
		}
		_, err := testRepo.Create(ctx, task)
		require.NoError(t, err)
	}

	// Create a task for a different user
	otherUserID := userID + 1
	otherUserTask := task.Task{
		UserID: otherUserID,
		Title:  "Other User Task",
		Status: task.StatusTodo,
	}
	_, err := testRepo.Create(ctx, otherUserTask)
	require.NoError(t, err)

	// List root tasks for the first user
	rootTasks, err := testRepo.ListRootTasks(ctx, int64(userID))
	assert.NoError(t, err)
	assert.Len(t, rootTasks, 3)

	// List root tasks for the other user
	otherUserTasks, err := testRepo.ListRootTasks(ctx, int64(otherUserID))
	assert.NoError(t, err)
	assert.Len(t, otherUserTasks, 1)
}

// Helper functions for testing
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
