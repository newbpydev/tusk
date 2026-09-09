package ports

const (
	ErrInvalidCommand        Error = "service: invalid command"
	ErrInvalidText           Error = "service: invalid text"
	ErrInvalidDate           Error = "service: invalid date"
	ErrInvalidReferenceTime  Error = "service: invalid reference time"
	ErrIdentityGeneration    Error = "service: identity generation failed"
	ErrConflict              Error = "service: task changed"
	ErrConfirmationRequired  Error = "service: delete confirmation required"
	ErrInvalidServiceOptions Error = "service: invalid options"
)
