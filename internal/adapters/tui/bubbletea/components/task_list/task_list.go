// Package task_list provides a component for rendering task lists 
// with collapsible sections in the TUI application.
package task_list

import (
	"strings"

	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/section_header"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/task_item"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/styles"
	"github.com/newbpydev/tusk/internal/core/task"
)

// getTasks returns the tasks associated with a section
func getTasks(section hooks.Section) []task.Task {
	// This is a temporary implementation until we have a proper task storage mechanism
	// In the future, tasks should be directly accessible from the section
	return []task.Task{}
}

// SectionMap maps section types to sections
type SectionMap map[hooks.SectionType]hooks.Section

// TaskListOptions configures how a task list should be rendered
type TaskListOptions struct {
	// Width is the available width for rendering
	Width int
	
	// Height is the available height for rendering
	Height int
	
	// Offset is the vertical scroll position
	Offset int
	
	// Cursor is the current cursor position
	Cursor int
	
	// CursorOnHeader indicates if the cursor is on a header or a task
	CursorOnHeader bool
}

// DefaultTaskListOptions returns the default options for rendering a task list
func DefaultTaskListOptions() TaskListOptions {
	return TaskListOptions{
		Width:         80,
		Height:        24,
		Offset:        0,
		Cursor:        0,
		CursorOnHeader: false,
	}
}

// RenderTaskList creates a string representation of a task list for display
// This is a pure function with no side effects, making it easy to test and reuse
func RenderTaskList(
	sections SectionMap, 
	collapsibleManager *hooks.CollapsibleManager,
	s *styles.Styles, 
	opts TaskListOptions,
) string {
	var result strings.Builder
	
	// If there are no sections to render, return early
	if len(sections) == 0 {
		return "No tasks found."
	}
	
	// Get the visual indices for the collapsible sections
	visualIdx := 0
	linesToRender := make([]string, 0)
	
	// Process each section
	sectionTypes := []hooks.SectionType{
		hooks.SectionTypeTodo,
		hooks.SectionTypeProjects,
		hooks.SectionTypeCompleted,
	}
	
	for _, sectionType := range sectionTypes {
		section, ok := sections[sectionType]
		if !ok {
			continue
		}
		
		// Skip rendering tasks if section is not expanded
		if !section.IsExpanded {
			visualIdx++
			continue
		}
		
		// Create section data for the header component
		sectionData := section_header.SectionData{
			Type:       sectionType,
			Title:      section.Title,
			Items:      getTasks(section), // Get tasks for this section
			IsExpanded: section.IsExpanded,
		}
		
		// Render section header
		headerOpts := section_header.SectionHeaderOptions{
			IsSelected: opts.CursorOnHeader && (visualIdx == opts.Cursor),
		}
		headerLine := section_header.RenderSectionHeader(sectionData, headerOpts)
		linesToRender = append(linesToRender, headerLine)
		visualIdx++
		
		// Render its tasks
		for _, t := range sectionData.Items {
			taskOpts := task_item.DefaultTaskItemOptions()
			taskOpts.IsSelected = !opts.CursorOnHeader && (visualIdx == opts.Cursor)
			taskOpts.Width = opts.Width
			
			taskLine := task_item.RenderTaskItem(t, s, taskOpts)
			linesToRender = append(linesToRender, taskLine)
			visualIdx++
		}
	}
	
	// Apply viewport scrolling to only show the visible portion
	start := max(0, opts.Offset)
	end := min(len(linesToRender), start+opts.Height)
	
	for i := start; i < end; i++ {
		result.WriteString(linesToRender[i])
		result.WriteString("\n")
	}
	
	return result.String()
}

// CalculateVisualMapping creates a mapping between task indices and visual indices
// This is useful for navigating through the task list
func CalculateVisualMapping(
	sections SectionMap, 
	collapsibleManager *hooks.CollapsibleManager,
) map[int]struct{int; hooks.SectionType; bool; int32} {
	mapping := make(map[int]struct{int; hooks.SectionType; bool; int32})
	visualIdx := 0
	taskIdx := 0
	
	// Process each section
	sectionTypes := []hooks.SectionType{
		hooks.SectionTypeTodo,
		hooks.SectionTypeProjects,
		hooks.SectionTypeCompleted,
	}
	
	for _, sectionType := range sectionTypes {
		section, ok := sections[sectionType]
		if !ok {
			continue
		}
		
		// Map section header
		mapping[visualIdx] = struct{
			int
			hooks.SectionType
			bool
			int32
		}{
			-1,                // taskIdx (not a task)
			sectionType,       // section type
			true,              // isHeader
			0,                 // taskID (n/a for headers)
		}
		visualIdx++
		
		// Add all tasks in section to visual mapping
		tasks := getTasks(section)
		for _, taskItem := range tasks {
			mapping[visualIdx] = struct{
				int
				hooks.SectionType
				bool
				int32
			}{
				taskIdx,    // taskIdx
				sectionType, // section type
				false,      // isHeader
				taskItem.ID, // taskID
			}
			visualIdx++
			taskIdx++
		}
	}
	
	return mapping
}

// FindVisualIndex finds the visual index for a task by its ID
func FindVisualIndex(
	taskID int32,
	tasks []task.Task,
	sections SectionMap,
	collapsibleManager *hooks.CollapsibleManager,
) int {
	mapping := CalculateVisualMapping(sections, collapsibleManager)
	
	// Search through the mapping for the task ID
	for visualIdx, info := range mapping {
		if !info.bool && info.int32 == taskID {
			return visualIdx
		}
	}
	
	return 0 // Default to first item if not found
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
