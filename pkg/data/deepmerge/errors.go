package deepmerge

import "errors"

var (
	// ErrMaxDepthExceeded is returned when merge recursion exceeds limit
	ErrMaxDepthExceeded = errors.New("deepmerge: maximum recursion depth exceeded")
)
