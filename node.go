package pipelines

// Node represents a processing stage in the pipeline that transforms inputs of type In
// into outputs of type Out, possibly handling multiple input or output channels.
type Node[In, Out any] interface {
	Runnable

	// ID returns a unique identifier for this node.
	ID() string

	// SetInput assigns one or more input channels to this node.
	// If the node does not accept inputs (e.g., a generator), returns ErrHasNoInput.
	SetInput(in ...<-chan In) error

	// Output creates and returns a new output channel for this node.
	// If the node does not produce outputs (e.g., an aggregator), returns ErrHasNoOutput.
	Output() (chan Out, error)
}
