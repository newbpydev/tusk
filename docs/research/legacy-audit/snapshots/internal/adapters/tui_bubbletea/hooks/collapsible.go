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

package hooks

// SectionType represents a category of collapsible section
type SectionType string

// Section types
const (
	// Task list sections
	SectionTypeTodo      SectionType = "todo"
	SectionTypeProjects  SectionType = "projects"
	SectionTypeCompleted SectionType = "completed"
	
	// Timeline sections
	SectionTypeOverdue   SectionType = "overdue"
	SectionTypeToday     SectionType = "today"
	SectionTypeUpcoming  SectionType = "upcoming"
)

// Section represents a collapsible section in the task list
type Section struct {
	Type       SectionType
	Title      string
	ItemCount  int
	StartIndex int
	IsExpanded bool
}

// CollapsibleManager manages the state of collapsible sections
type CollapsibleManager struct {
	Sections         []Section
	FlatCursorPos    int // Cursor position in the flattened task list
	VisibleStartIdx  int // Index where visible tasks start (for scrolling)
	expandedSections map[SectionType]bool
}

// NewCollapsibleManager creates a new CollapsibleManager with default settings
func NewCollapsibleManager() *CollapsibleManager {
	cm := &CollapsibleManager{
		Sections: make([]Section, 0),
	}

	// Set default expanded states
	cm.expandedSections = map[SectionType]bool{
		SectionTypeTodo:      true,  // Todo section expanded by default
		SectionTypeProjects:  false, // Projects collapsed by default
		SectionTypeCompleted: true,  // Completed section expanded by default
	}

	return cm
}

// AddSection adds a new section to the manager
func (cm *CollapsibleManager) AddSection(sectionType SectionType, title string, itemCount int, startIndex int) {
	// Get default expanded state from the map
	isExpanded := cm.expandedSections[sectionType]

	section := Section{
		Type:       sectionType,
		Title:      title,
		ItemCount:  itemCount,
		StartIndex: startIndex,
		IsExpanded: isExpanded,
	}

	cm.Sections = append(cm.Sections, section)
}

// ClearSections removes all sections
func (cm *CollapsibleManager) ClearSections() {
	cm.Sections = []Section{}
}

// ToggleSection toggles the expanded state of a section
func (cm *CollapsibleManager) ToggleSection(sectionType SectionType) {
	for i := range cm.Sections {
		if cm.Sections[i].Type == sectionType {
			cm.Sections[i].IsExpanded = !cm.Sections[i].IsExpanded
			cm.expandedSections[sectionType] = cm.Sections[i].IsExpanded
			break
		}
	}
}

// GetSectionAtIndex returns the section at the given index
// Returns nil if the index is not a section
func (cm *CollapsibleManager) GetSectionAtIndex(index int) *Section {
	var currentIndex int = 0

	for i, section := range cm.Sections {
		if currentIndex == index {
			return &cm.Sections[i]
		}
		currentIndex++

		// Skip section items if expanded
		if section.IsExpanded {
			currentIndex += section.ItemCount
		}
	}

	return nil
}

// IsSectionHeader determines if the given index points to a section header
func (cm *CollapsibleManager) IsSectionHeader(index int) bool {
	return cm.GetSectionAtIndex(index) != nil
}

// GetItemCount returns the total number of visible items (sections + their visible items)
func (cm *CollapsibleManager) GetItemCount() int {
	count := 0
	for _, section := range cm.Sections {
		count++ // Count the section header
		if section.IsExpanded {
			count += section.ItemCount
		}
	}
	return count
}

// GetNextCursorPosition returns the next cursor position based on the current position
// and direction (1 for down, -1 for up)
func (cm *CollapsibleManager) GetNextCursorPosition(currentPos, direction int) int {
	totalItems := cm.GetItemCount()
	nextPos := currentPos + direction

	if nextPos < 0 {
		return 0 // Don't go beyond the first item
	}

	if nextPos >= totalItems {
		return totalItems - 1 // Don't go beyond the last item
	}

	return nextPos
}

// GetActualTaskIndex translates a visible cursor position to an actual task index
// Returns -1 if the cursor is on a section header
func (cm *CollapsibleManager) GetActualTaskIndex(visibleIndex int) int {
	if visibleIndex < 0 {
		return -1
	}

	var currentVisibleIndex int = 0
	var actualTaskOffset int = 0

	for _, section := range cm.Sections {
		// If we're pointing at a section header
		if currentVisibleIndex == visibleIndex {
			return -1
		}

		currentVisibleIndex++ // Skip past section header

		// If section is expanded and our target index is within this section
		if section.IsExpanded && visibleIndex < currentVisibleIndex+section.ItemCount {
			// Calculate the index within the section
			withinSectionIndex := visibleIndex - currentVisibleIndex
			return section.StartIndex + withinSectionIndex
		}

		// Account for expanded section items in the visible index
		if section.IsExpanded {
			currentVisibleIndex += section.ItemCount
		}

		// Keep track of the actual task index offset
		actualTaskOffset += section.ItemCount
	}

	// Index not found
	return -1
}

// GetVisibleIndexFromTaskIndex translates an actual task index to a visible cursor position
// Useful for setting cursor after operations that change the task list
func (cm *CollapsibleManager) GetVisibleIndexFromTaskIndex(taskIndex int) int {
	var visibleIndex int = 0

	for _, section := range cm.Sections {
		visibleIndex++ // Section header

		// Check if the task is in this section and if the section is expanded
		if taskIndex >= section.StartIndex && taskIndex < section.StartIndex+section.ItemCount {
			if section.IsExpanded {
				return visibleIndex + (taskIndex - section.StartIndex)
			}
			// If section is collapsed, return the section header index
			return visibleIndex - 1
		}

		// Add expanded section items to the visible index
		if section.IsExpanded {
			visibleIndex += section.ItemCount
		}
	}

	// Task not found in any section
	return 0
}
