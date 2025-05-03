package services

import (
	"time"

	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/core/task"
)

// SectionMap is a map of section types to sections
type SectionMap map[hooks.SectionType]Section

// Section represents a section with tasks
type Section struct {
	Title   string
	Items   []task.Task
	IsOpen  bool
}

// TaskCategorizationService handles categorization of tasks for UI display
type TaskCategorizationService struct{}

// NewTaskCategorizationService creates a new TaskCategorizationService
func NewTaskCategorizationService() *TaskCategorizationService {
	return &TaskCategorizationService{}
}

// CategorizeTasks organizes tasks into todo, projects, and completed categories
func (s *TaskCategorizationService) CategorizeTasks(tasks []task.Task) ([]task.Task, []task.Task, []task.Task) {
	var todoTasks, projectTasks, completedTasks []task.Task

	for _, t := range tasks {
		if t.Status == task.StatusDone {
			completedTasks = append(completedTasks, t)
		} else if t.ParentID != nil {
			projectTasks = append(projectTasks, t)
		} else {
			todoTasks = append(todoTasks, t)
		}
	}

	return todoTasks, projectTasks, completedTasks
}

// CategorizeTimelineTasks organizes tasks into overdue, today, and upcoming categories
func (s *TaskCategorizationService) CategorizeTimelineTasks(tasks []task.Task) ([]task.Task, []task.Task, []task.Task) {
	var overdueTasks, todayTasks, upcomingTasks []task.Task
	now := time.Now()

	for _, t := range tasks {
		// Skip tasks without due dates or completed tasks
		if t.DueDate == nil || t.Status == task.StatusDone {
			continue
		}

		if s.isBeforeDay(*t.DueDate, now) {
			// Task is overdue (due date is before today)
			overdueTasks = append(overdueTasks, t)
		} else if s.isSameDay(*t.DueDate, now) {
			// Task is due today
			todayTasks = append(todayTasks, t)
		} else {
			// Task is upcoming (due date is after today)
			upcomingTasks = append(upcomingTasks, t)
		}
	}

	return overdueTasks, todayTasks, upcomingTasks
}

// CreateCategorizedSections builds section data for the collapsible manager based on task categories
func (s *TaskCategorizationService) CreateCategorizedSections(
	todoTasks, projectTasks, completedTasks []task.Task,
) SectionMap {
	sections := make(SectionMap)
	sections[hooks.SectionTypeTodo] = Section{
		Title:   "Todo",
		Items:   todoTasks,
		IsOpen:  true,
	}
	sections[hooks.SectionTypeProjects] = Section{
		Title:   "Projects",
		Items:   projectTasks,
		IsOpen:  true,
	}
	sections[hooks.SectionTypeCompleted] = Section{
		Title:   "Completed",
		Items:   completedTasks,
		IsOpen:  true,
	}

	return sections
}

// CreateTimelineSections builds section data for the timeline view
func (s *TaskCategorizationService) CreateTimelineSections(
	overdueTasks, todayTasks, upcomingTasks []task.Task,
) SectionMap {
	sections := make(SectionMap)
	sections[hooks.SectionTypeOverdue] = Section{
		Title:   "Overdue",
		Items:   overdueTasks,
		IsOpen:  true,
	}
	sections[hooks.SectionTypeToday] = Section{
		Title:   "Today",
		Items:   todayTasks,
		IsOpen:  true,
	}
	sections[hooks.SectionTypeUpcoming] = Section{
		Title:   "Upcoming",
		Items:   upcomingTasks,
		IsOpen:  true,
	}

	return sections
}

// isSameDay checks if two times occur on the same calendar day
func (s *TaskCategorizationService) isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// isBeforeDay checks if t1 is on a calendar day before t2
func (s *TaskCategorizationService) isBeforeDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	
	// Compare years first
	if y1 < y2 {
		return true
	}
	if y1 > y2 {
		return false
	}
	
	// Same year, compare months
	if m1 < m2 {
		return true
	}
	if m1 > m2 {
		return false
	}
	
	// Same year and month, compare days
	return d1 < d2
}

// isAfterDay checks if t1 is on a calendar day after t2
func (s *TaskCategorizationService) isAfterDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	
	// Compare years first
	if y1 > y2 {
		return true
	}
	if y1 < y2 {
		return false
	}
	
	// Same year, compare months
	if m1 > m2 {
		return true
	}
	if m1 < m2 {
		return false
	}
	
	// Same year and month, compare days
	return d1 > d2
}
