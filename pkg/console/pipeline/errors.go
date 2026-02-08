package pipeline

import "errors"

var (
	// ErrAborted is returned when pipeline execution is aborted via Context.Abort.
	ErrAborted = errors.New("pipeline: execution aborted")

	// ErrNilHandler is returned when a nil handler is provided.
	ErrNilHandler = errors.New("pipeline: nil handler")
)
