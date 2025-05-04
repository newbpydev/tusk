// Package messages contains message types for the BubbleTea TUI
package messages

import (
	"github.com/newbpydev/tusk/internal/core/task"
)

// ModalFormCloseMsg is sent when the modal form is closed without submitting
type ModalFormCloseMsg struct{}

// ModalFormSubmitMsg is sent when the form is submitted successfully
type ModalFormSubmitMsg struct {
	Task task.Task
}
