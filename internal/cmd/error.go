package cmd

// CommandError wraps an exit code with a message
type CommandError struct {
	Code int
	Msg  string
}

// Error implements the error interface
func (e *CommandError) Error() string {
	return e.Msg
}
