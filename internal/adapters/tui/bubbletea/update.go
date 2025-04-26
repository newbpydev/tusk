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

package bubbletea

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles user input and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	default:
		return m, nil
	}
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.viewMode {
	case "list":
		return m.handleListViewKeys(msg)
	case "detail":
		return m.handleDetailViewKeys(msg)
	case "edit":
		return m.handleEditViewKeys(msg)
	default:
		return m, nil
	}
}

// handleListViewKeys processes keyboard input in list view
func (m Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "j":
		if m.cursor < len(m.tasks)-1 {
			m.cursor++
		}
		return m, nil

	case "enter":
		if len(m.tasks) > 0 {
			m.viewMode = "detail"
			return m, nil
		}

	case "c":
		// Toggle completion status
		if len(m.tasks) > 0 {
			return m, m.toggleTaskCompletion
		}

	case "r":
		// Refresh task list
		return m, m.refreshTasks

	case "s":
		// Change sort order (not implemented yet)
		return m, nil

	case "f":
		// Filter tasks (not implemented yet)
		return m, nil
	}

	return m, nil
}

// handleDetailViewKeys processes keyboard input in detail view
func (m Model) handleDetailViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.viewMode = "list"
		return m, nil

	case "e":
		m.viewMode = "edit"
		return m, nil

	case "d":
		// Delete task
		return m, m.deleteCurrentTask

	case "c":
		// Toggle completion status
		return m, m.toggleTaskCompletion
	}

	return m, nil
}

// handleEditViewKeys processes keyboard input in edit view
func (m Model) handleEditViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = "detail"
		return m, nil

	case "enter":
		// Save edits (not fully implemented)
		m.viewMode = "detail"
		return m, nil
	}

	return m, nil
}
