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

package task

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/newbpydev/tusk/internal/config"
	domainerrors "github.com/newbpydev/tusk/internal/core/errors"
	"github.com/newbpydev/tusk/internal/core/task"
	"github.com/newbpydev/tusk/internal/ports/output"
	"github.com/newbpydev/tusk/internal/util/logging"
)

// Setup a mock repository that implements the TaskRepository interface
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(ctx context.Context, t task.Task) (task.Task, error) {
	args := m.Called(ctx, t)
	return args.Get(0).(task.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, t task.Task) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id int64) (task.Task, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(task.Task), args.Error(1)
}

func (m *MockTaskRepository) ListRootTasks(ctx context.Context, userID int64) ([]task.Task, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]task.Task), args.Error(1)
}

func (m *MockTaskRepository) ListSubTasks(ctx context.Context, parentID int64) ([]task.Task, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]task.Task), args.Error(1)
}

func (m *MockTaskRepository) GetTaskTree(ctx context.Context, rootID int64) (task.Task, error) {
	args := m.Called(ctx, rootID)
	return args.Get(0).(task.Task), args.Error(1)
}

func (m *MockTaskRepository) ReorderTask(ctx context.Context, taskID int64, newOrder int) error {
	args := m.Called(ctx, taskID, newOrder)
	return args.Error(0)
}

func (m *MockTaskRepository) SearchTasksByTitle(ctx context.Context, userID int64, titlePattern string) ([]task.Task, error) {
	args := m.Called(ctx, userID, titlePattern)
	return args.Get(0).([]task.Task), args.Error(1)
}

func (m *MockTaskRepository) SearchTasksByTag(ctx context.Context, userID int64, tag string) ([]task.Task, error) {
	args := m.Called(ctx, userID, tag)
	return args.Get(0).([]task.Task), args.Error(1)
}

func (m *MockTaskRepository) ListTasksByStatus(ctx context.Context, userID int64, status task.Status) ([]task.Task, error) {
	args := m.Called(ctx, userID, status)
	return args.Get(0).([]task.Task), args.Error(1)
}

func (m *MockTaskRepository) ListTasksByPriority(ctx context.Context, userID int64, priority task.Priority) ([]task.Task, error) {
	args := m.Called(ctx, userID, priority)
	return args.Get(0).([]task.Task), args.Error(1)
}

func (m *MockTaskRepository) ListTasksDueToday(ctx context.Context, userID int64) ([]task.Task, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]task.Task), args.Error(1)
}

func (m *MockTaskRepository) ListTasksDueSoon(ctx context.Context, userID int64) ([]task.Task, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]task.Task), args.Error(1)
}

func (m *MockTaskRepository) ListOverdueTasks(ctx context.Context, userID int64) ([]task.Task, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]task.Task), args.Error(1)
}

func (m *MockTaskRepository) GetTaskCountsByStatus(ctx context.Context, userID int64) (output.TaskStatusCounts, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(output.TaskStatusCounts), args.Error(1)
}

func (m *MockTaskRepository) GetTaskCountsByPriority(ctx context.Context, userID int64) (output.TaskPriorityCounts, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(output.TaskPriorityCounts), args.Error(1)
}

func (m *MockTaskRepository) GetRecentlyCompletedTasks(ctx context.Context, userID int64, limit int32) ([]task.Task, error) {
	args := m.Called(ctx, userID, limit)
	return args.Get(0).([]task.Task), args.Error(1)
}

func (m *MockTaskRepository) BulkUpdateTaskStatus(ctx context.Context, taskIDs []int32, status task.Status, isCompleted bool) error {
	args := m.Called(ctx, taskIDs, status, isCompleted)
	return args.Error(0)
}

func (m *MockTaskRepository) GetAllTagsForUser(ctx context.Context, userID int64) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

// Initialize logging to prevent panics during tests
func init() {
	// Create a basic config for testing
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
		// Fallback to a test logger if we can't initialize proper logging
		logging.Logger = zaptest.NewLogger(nil)
		logging.ServiceLogger = logging.Logger.Named("service-test")
	}
}

// Helper function to create a new taskService with a mock repository
func newTestTaskService(repo output.TaskRepository) Service {
	return NewTaskService(repo)
}

func TestCreateTask(t *testing.T) {
	// Test cases for Create function
	testCases := []struct {
		name           string
		userID         int64
		parentID       *int64
		title          string
		description    string
		dueDate        *time.Time
		priority       task.Priority
		tags           []string
		mockSetup      func(*MockTaskRepository)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:        "Valid task creation",
			userID:      1,
			title:       "Test Task",
			description: "Test Description",
			priority:    task.PriorityMedium,
			tags:        []string{"test"},
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("task.Task")).Return(task.Task{
					ID:          1,
					UserID:      1,
					Title:       "Test Task",
					Description: stringPtr("Test Description"),
					Priority:    task.PriorityMedium,
					Status:      task.StatusTodo,
					IsCompleted: false,
					Tags:        []task.Tag{{Name: "test"}},
				}, nil)
			},
			expectedError: false,
		},
		{
			name:           "Invalid user ID",
			userID:         -1, // Invalid user ID
			title:          "Test Task",
			description:    "Test Description",
			priority:       task.PriorityMedium,
			tags:           []string{"test"},
			mockSetup:      func(mockRepo *MockTaskRepository) {},
			expectedError:  true,
			expectedErrMsg: "user ID must be positive",
		},
		{
			name:           "Empty title",
			userID:         1,
			title:          "", // Empty title
			description:    "Test Description",
			priority:       task.PriorityMedium,
			tags:           []string{"test"},
			mockSetup:      func(mockRepo *MockTaskRepository) {},
			expectedError:  true,
			expectedErrMsg: "title is required",
		},
		{
			name:        "Repository error",
			userID:      1,
			title:       "Test Task",
			description: "Test Description",
			priority:    task.PriorityMedium,
			tags:        []string{"test"},
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("task.Task")).Return(task.Task{}, errors.New("repository error"))
			},
			expectedError:  true,
			expectedErrMsg: "repository error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock repository
			mockRepo := new(MockTaskRepository)
			// Set up mock expectations
			tc.mockSetup(mockRepo)

			// Create the task service with the mock repo
			taskService := newTestTaskService(mockRepo)

			// Call the Create function
			createdTask, err := taskService.Create(context.Background(), tc.userID, tc.parentID, tc.title, tc.description, tc.dueDate, tc.priority, tc.tags)

			// Check error
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.title, createdTask.Title)
			}

			// Verify all expected mock calls were made
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestShowTask(t *testing.T) {
	// Test cases for Show function
	testCases := []struct {
		name           string
		taskID         int64
		mockSetup      func(*MockTaskRepository)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:   "Valid task retrieval",
			taskID: 1,
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("GetTaskTree", mock.Anything, int64(1)).Return(task.Task{
					ID:     1,
					UserID: 1,
					Title:  "Test Task",
					Status: task.StatusTodo,
				}, nil)
			},
			expectedError: false,
		},
		{
			name:           "Invalid task ID",
			taskID:         0, // Invalid task ID
			mockSetup:      func(mockRepo *MockTaskRepository) {},
			expectedError:  true,
			expectedErrMsg: "task ID must be positive",
		},
		{
			name:   "Task not found",
			taskID: 999,
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("GetTaskTree", mock.Anything, int64(999)).Return(task.Task{}, domainerrors.NotFound("task not found"))
			},
			expectedError:  true,
			expectedErrMsg: "task not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock repository
			mockRepo := new(MockTaskRepository)
			// Set up mock expectations
			tc.mockSetup(mockRepo)

			// Create the task service with the mock repo
			taskService := newTestTaskService(mockRepo)

			// Call the Show function
			retrievedTask, err := taskService.Show(context.Background(), tc.taskID)

			// Check error
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, retrievedTask.ID)
			}

			// Verify all expected mock calls were made
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestListTasks(t *testing.T) {
	// Test cases for List function
	testCases := []struct {
		name           string
		userID         int64
		mockSetup      func(*MockTaskRepository)
		expectedError  bool
		expectedErrMsg string
		expectedCount  int
	}{
		{
			name:   "Valid task listing with root tasks",
			userID: 1,
			mockSetup: func(mockRepo *MockTaskRepository) {
				rootTasks := []task.Task{
					{ID: 1, UserID: 1, Title: "Task 1"},
					{ID: 2, UserID: 1, Title: "Task 2"},
				}
				mockRepo.On("ListRootTasks", mock.Anything, int64(1)).Return(rootTasks, nil)

				// Set up expectations for GetTaskTree for each root task
				mockRepo.On("GetTaskTree", mock.Anything, int64(1)).Return(
					task.Task{ID: 1, UserID: 1, Title: "Task 1", SubTasks: []task.Task{}}, nil)
				mockRepo.On("GetTaskTree", mock.Anything, int64(2)).Return(
					task.Task{ID: 2, UserID: 1, Title: "Task 2", SubTasks: []task.Task{}}, nil)
			},
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:           "Invalid user ID",
			userID:         -1, // Invalid user ID
			mockSetup:      func(mockRepo *MockTaskRepository) {},
			expectedError:  true,
			expectedErrMsg: "user ID must be positive",
		},
		{
			name:   "Repository error",
			userID: 1,
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("ListRootTasks", mock.Anything, int64(1)).Return([]task.Task{}, errors.New("repository error"))
			},
			expectedError:  true,
			expectedErrMsg: "repository error",
		},
		{
			name:   "Error retrieving task tree",
			userID: 1,
			mockSetup: func(mockRepo *MockTaskRepository) {
				rootTasks := []task.Task{
					{ID: 1, UserID: 1, Title: "Task 1"},
				}
				mockRepo.On("ListRootTasks", mock.Anything, int64(1)).Return(rootTasks, nil)
				mockRepo.On("GetTaskTree", mock.Anything, int64(1)).Return(task.Task{}, errors.New("error retrieving tree"))
			},
			expectedError: false, // The function continues even if it can't get the tree for a task
			expectedCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock repository
			mockRepo := new(MockTaskRepository)
			// Set up mock expectations
			tc.mockSetup(mockRepo)

			// Create the task service with the mock repo
			taskService := newTestTaskService(mockRepo)

			// Call the List function
			tasks, err := taskService.List(context.Background(), tc.userID)

			// Check error
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Len(t, tasks, tc.expectedCount)
			}

			// Verify all expected mock calls were made
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteTask(t *testing.T) {
	// Test cases for Delete function
	testCases := []struct {
		name           string
		taskID         int64
		mockSetup      func(*MockTaskRepository)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:   "Valid task deletion",
			taskID: 1,
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("GetByID", mock.Anything, int64(1)).Return(task.Task{
					ID:     1,
					UserID: 1,
					Title:  "Test Task",
				}, nil)
				mockRepo.On("Delete", mock.Anything, int64(1)).Return(nil)
			},
			expectedError: false,
		},
		{
			name:           "Invalid task ID",
			taskID:         0, // Invalid task ID
			mockSetup:      func(mockRepo *MockTaskRepository) {},
			expectedError:  true,
			expectedErrMsg: "task ID must be positive",
		},
		{
			name:   "Task not found",
			taskID: 999,
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("GetByID", mock.Anything, int64(999)).Return(task.Task{}, domainerrors.NotFound("task not found"))
			},
			expectedError:  true,
			expectedErrMsg: "task not found",
		},
		{
			name:   "Delete error",
			taskID: 1,
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("GetByID", mock.Anything, int64(1)).Return(task.Task{
					ID:     1,
					UserID: 1,
					Title:  "Test Task",
				}, nil)
				mockRepo.On("Delete", mock.Anything, int64(1)).Return(errors.New("delete error"))
			},
			expectedError:  true,
			expectedErrMsg: "delete error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock repository
			mockRepo := new(MockTaskRepository)
			// Set up mock expectations
			tc.mockSetup(mockRepo)

			// Create the task service with the mock repo
			taskService := newTestTaskService(mockRepo)

			// Call the Delete function
			err := taskService.Delete(context.Background(), tc.taskID)

			// Check error
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expected mock calls were made
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCompleteTask(t *testing.T) {
	// Test cases for Complete function
	testCases := []struct {
		name           string
		taskID         int64
		mockSetup      func(*MockTaskRepository)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:   "Valid task completion",
			taskID: 1,
			mockSetup: func(mockRepo *MockTaskRepository) {
				// Mock getting the task
				mockRepo.On("GetByID", mock.Anything, int64(1)).Return(task.Task{
					ID:          1,
					UserID:      1,
					Title:       "Test Task",
					Status:      task.StatusTodo,
					IsCompleted: false,
				}, nil).Once()

				// Mock updating the task
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(t task.Task) bool {
					return t.ID == 1 && t.Status == task.StatusDone && t.IsCompleted == true
				})).Return(nil)

				// Mock getting the updated task
				mockRepo.On("GetByID", mock.Anything, int64(1)).Return(task.Task{
					ID:          1,
					UserID:      1,
					Title:       "Test Task",
					Status:      task.StatusDone,
					IsCompleted: true,
				}, nil).Once()
			},
			expectedError: false,
		},
		{
			name:           "Invalid task ID",
			taskID:         0, // Invalid task ID
			mockSetup:      func(mockRepo *MockTaskRepository) {},
			expectedError:  true,
			expectedErrMsg: "task ID must be positive",
		},
		{
			name:   "Task not found",
			taskID: 999,
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("GetByID", mock.Anything, int64(999)).Return(task.Task{}, domainerrors.NotFound("task not found"))
			},
			expectedError:  true,
			expectedErrMsg: "task not found",
		},
		{
			name:   "Update error",
			taskID: 1,
			mockSetup: func(mockRepo *MockTaskRepository) {
				// Mock getting the task
				mockRepo.On("GetByID", mock.Anything, int64(1)).Return(task.Task{
					ID:          1,
					UserID:      1,
					Title:       "Test Task",
					Status:      task.StatusTodo,
					IsCompleted: false,
				}, nil).Once()

				// Mock update failure
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update error"))
			},
			expectedError:  true,
			expectedErrMsg: "update error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock repository
			mockRepo := new(MockTaskRepository)
			// Set up mock expectations
			tc.mockSetup(mockRepo)

			// Create the task service with the mock repo
			taskService := newTestTaskService(mockRepo)

			// Call the Complete function
			completedTask, err := taskService.Complete(context.Background(), tc.taskID)

			// Check error
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.True(t, completedTask.IsCompleted)
				assert.Equal(t, task.StatusDone, completedTask.Status)
			}

			// Verify all expected mock calls were made
			mockRepo.AssertExpectations(t)
		})
	}
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
