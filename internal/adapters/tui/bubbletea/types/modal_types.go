// Package types contains shared type definitions used across the UI components
package types

// ModalDisplayMode determines how the modal is displayed
type ModalDisplayMode string

const (
	// FullScreen shows the modal centered on the entire screen
	FullScreen ModalDisplayMode = "fullscreen"
	// ContentArea shows the modal only in the main content area, preserving header and footer
	ContentArea ModalDisplayMode = "content-area"
)
