package nodes

import "errors"

var (
	ErrAccessRunningNode = errors.New("access attempt to running node")
	ErrNodeRunning       = errors.New("node is already running")
	ErrOnlyOneInput      = errors.New("worker pool requires exactly one input")

	ErrHasNoOutput = errors.New("this node has not outputs")
	ErrHasNoInput  = errors.New("this node has not inputs")

	ErrZipNodeNoInput     = errors.New("zipNode: no input channels")
	ErrZipNodeClosedInput = errors.New("zipNode: one of the input channels was closed")
)
