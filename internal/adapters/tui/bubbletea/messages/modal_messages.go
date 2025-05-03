package messages

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ShowModalMsg is sent when a modal needs to be shown
type ShowModalMsg struct {
	Content tea.Model
	Width   int
	Height  int
}

// HideModalMsg is sent when a modal needs to be hidden
type HideModalMsg struct{}
