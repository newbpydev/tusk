package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
)

// HandleTimelinePanelKeys processes keyboard input when the timeline panel is active
func HandleTimelinePanelKeys(m AppModel, msg tea.KeyMsg) (tea.Cmd, bool) {
	// Make sure the timeline collapsible manager is initialized
	// This would need to be extended to the AppModel interface
	
	switch msg.String() {
	case "tab", "right", "l":
		// Circular navigation - go to first panel
		m.SetActivePanel(0)
		return nil, true

	case "shift+tab", "left", "h":
		// Move to previous panel if available
		m.SetActivePanel(1)
		return nil, true

	case "j", "down":
		// Need to add timeline navigation to AppModel interface
		// m.NavigateTimelineDown()
		return nil, true

	case "k", "up":
		// Need to add timeline navigation to AppModel interface
		// m.NavigateTimelineUp()
		return nil, true

	case "g":
		// Jump to top of timeline
		// Need to add timeline navigation to AppModel interface
		// m.NavigateTimelineToTop()
		return nil, true

	case "G":
		// Jump to bottom of timeline
		// Need to add timeline navigation to AppModel interface
		// m.NavigateTimelineToBottom()
		return nil, true

	case "enter", "d":
		// Toggle section or view task details
		// Need to add handling for this in the AppModel interface
		// if m.GetTimelineCursorOnHeader() {
		//     return m.ToggleTimelineSection(), true
		// }
		return nil, true

	case " ":
		// Toggle task completion status from timeline view
		// Need to add handling for this in the AppModel interface
		// if !m.GetTimelineCursorOnHeader() && m.GetTimelineCursor() < len(m.GetTimelineTasks()) {
		//     return m.ToggleTimelineTaskCompletion(), true
		// }
		return nil, true

	case "r":
		// Refresh tasks
		m.SetLoadingStatus("Refreshing tasks...")
		return m.RefreshTasks(), true
	}

	return nil, false
}

// HandleDetailsPanelKeys processes keyboard input when the task details panel is active
func HandleDetailsPanelKeys(m AppModel, msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "tab", "right", "l":
		// Move to next panel if available
		// This needs to check if timeline is available
		m.SetActivePanel(2)
		return nil, true

	case "shift+tab", "left", "h", "esc":
		// Move to previous panel
		m.SetActivePanel(0)
		return nil, true

	case "j", "down":
		// Scroll down in task details
		// Need to add method to AppModel for scrolling details
		// m.ScrollTaskDetailsDown()
		return nil, true

	case "k", "up":
		// Scroll up in task details
		// Need to add method to AppModel for scrolling details
		// m.ScrollTaskDetailsUp()
		return nil, true

	case "e":
		// Edit current task
		if !m.GetCursorOnHeader() && m.GetCursor() < len(m.GetTasks()) {
			m.SetViewMode("edit")
			// Load current task into form
			m.LoadTaskIntoForm(m.GetTasks()[m.GetCursor()])
			return nil, true
		}
		return nil, true

	case "r":
		// Refresh tasks
		m.SetLoadingStatus("Refreshing tasks...")
		return m.RefreshTasks(), true
	}

	return nil, false
}

// HandlePanelVisibilityKeys processes keyboard input for panel visibility toggles
func HandlePanelVisibilityKeys(m AppModel, msg tea.KeyMsg) (tea.Cmd, bool) {
	// Need to add methods to toggle panel visibility to AppModel
	switch msg.String() {
	case "1":
		// Toggle task list panel
		// m.ToggleTaskList()
		return nil, true
	case "2":
		// Toggle task details panel
		// m.ToggleTaskDetails()
		return nil, true
	case "3":
		// Toggle timeline panel
		// m.ToggleTimeline()
		return nil, true
	}
	
	return nil, false
}
