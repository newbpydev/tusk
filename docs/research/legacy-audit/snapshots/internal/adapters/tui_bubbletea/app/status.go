package app

import "time"

const (
	statusTypeSuccess = "success"
	statusTypeError   = "error"
	statusTypeInfo    = "info"
	statusTypeLoading = "loading"
)

// setStatusMessage sets a status message with a type and expiry duration.
// This is the base function for setting any kind of status.
func (m *Model) setStatusMessage(msg, msgType string, duration time.Duration) {
	m.statusMessage = msg
	m.statusType = msgType
	if duration > 0 {
		m.statusExpiry = time.Now().Add(duration)
	} else {
		m.statusExpiry = time.Time{} // No expiry
	}
	m.isLoading = (msgType == statusTypeLoading) // Set isLoading only if type is loading
	m.err = nil                                  // Clear previous error when setting a new status
}

// setSuccessStatus is a helper to set success status messages with a default duration.
func (m *Model) setSuccessStatus(msg string) {
	m.setStatusMessage(msg, statusTypeSuccess, 5*time.Second)
}

// setErrorStatus is a helper to set error status messages with a default duration.
// It also stores the error itself in m.err.
func (m *Model) setErrorStatus(msg string) {
	// Optionally, extract the underlying error if msg contains structured error info
	// For now, just display the message. Storing the error happens in the Update loop.
	m.setStatusMessage(msg, statusTypeError, 10*time.Second)
}

// setLoadingStatus sets the app in loading state with a message and indefinite duration (cleared manually).
func (m *Model) setLoadingStatus(msg string) {
	m.setStatusMessage(msg, statusTypeLoading, 0) // Duration 0 means no auto-expiry
}

// clearLoadingStatus clears the loading state and any associated status message.
func (m *Model) clearLoadingStatus() {
	if m.isLoading { // Only clear if currently loading
		m.isLoading = false
		if m.statusType == statusTypeLoading {
			m.statusMessage = ""
			m.statusType = ""
			m.statusExpiry = time.Time{}
		}
	}
}
