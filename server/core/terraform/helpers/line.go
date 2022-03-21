package helpers

// Line represents a line that was output from a terraform command.
type Line struct {
	// Line is the contents of the line (without the newline).
	Line string
	// Err is set if there was an error.
	Err error
}
