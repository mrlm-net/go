package console

import "errors"

var (
	// ErrCommandNotFound is returned when no command matches the given arguments.
	ErrCommandNotFound = errors.New("console: command not found")

	// ErrFlagNotFound is returned when a requested flag is not registered.
	ErrFlagNotFound = errors.New("console: flag not found")

	// ErrFlagParse is returned when a flag value cannot be parsed to the expected type.
	ErrFlagParse = errors.New("console: flag parse error")

	// ErrNoHandler is returned when a command has no handler assigned.
	ErrNoHandler = errors.New("console: command has no handler")

	// ErrHelpRequested signals that help was displayed. This is not a true error
	// but a control flow signal. Callers can check for this with errors.Is().
	ErrHelpRequested = errors.New("console: help requested")

	// ErrFlagRequired is returned when a required flag is not provided.
	ErrFlagRequired = errors.New("console: required flag not provided")
)
