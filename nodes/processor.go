package nodes

import "context"

// Processor defines a function that processes an input of type In and produces an output of type Out.
// Returns an error if processing fails.
type Processor[In, Out any] func(In) (Out, error)

// ZipProcessor defines a function that processes a slice of inputs (one from each input channel)
// and produces a single output of type Out. Returns an error if processing fails.
type ZipProcessor[In, Out any] func([]In) (Out, error)

// Sink defines a function that consumes an input of type In, typically producing a side effect
// (e.g., writing to a file or printing). Returns an error if consumption fails.
type Generator[Out any] func(context.Context) (<-chan Out, error)

// Generator defines a function that produces a channel of outputs of type Out.
// It receives a context for cancellation. Returns an error if generation fails.
type Sink[In any] func(In) error
