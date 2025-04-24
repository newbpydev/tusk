package user

import "time"

// User represents a user in the system.
// It includes fields for the user's ID, username, email, password hash,
// created and updated timestamps, last login timestamp, and active status.
// The password hash is used for authentication and should not be stored in plain text.
type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"` // Exclude from JSON response
	// PasswordHash is used for authentication and should not be stored in plain text.
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty"` // nil means never logged in
	IsActive  bool       `json:"is_active"`
}
