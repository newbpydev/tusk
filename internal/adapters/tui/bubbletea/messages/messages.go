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

package messages

import (
	"time"
	"github.com/newbpydev/tusk/internal/core/task"
)

// TickMsg is a tick from the timer
// It wraps time.Time
type TickMsg time.Time

// ErrorMsg represents a generic error
// It wraps the built-in error
type ErrorMsg error

// StatusUpdateErrorMsg represents an error during task status update
// Holds the index and title of the task, and the error
type StatusUpdateErrorMsg struct {
	TaskIndex int
	TaskTitle string
	Err       error
}

// StatusUpdateSuccessMsg represents a successful task status update
// Contains the updated task and a message
type StatusUpdateSuccessMsg struct {
	Task    task.Task
	Message string
}

// TasksRefreshedMsg represents refreshed tasks from a background operation
// Contains the list of tasks
type TasksRefreshedMsg struct {
	Tasks []task.Task
}