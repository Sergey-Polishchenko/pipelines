package pipelines

// Connect links the output of one node to the input of another (1->1).
// The output channel(s) of 'from' are connected to the input channel(s) of 'to'.
// Returns an error if retrieving or setting channels fails.
func Connect[In, Mid, Out any](from Node[In, Mid], to Node[Mid, Out]) error {
	out, err := from.Output()
	if err != nil {
		return err
	}

	return to.SetInput(out)
}

// ConnectToMany links the output of 'from' to multiple target nodes.
// Creates a separate output channel for each target and sets each as input.
// Note: In the current version, workerPool nodes only support a single output.
func ConnectToMany[In, Mid, Out any](from Node[In, Mid], targets ...Node[Mid, Out]) error {
	var outs []chan Mid
	for range targets {
		out, err := from.Output()
		if err != nil {
			return err
		}
		outs = append(outs, out)
	}

	for i, tgt := range targets {
		if err := tgt.SetInput(outs[i]); err != nil {
			return err
		}
	}

	return nil
}

// ConnectFromMany merges outputs from multiple source nodes into a single target node input.
// Uses a fan-in mechanism to combine channels.
func ConnectFromMany[In, Mid, Out any](sources []Node[In, Mid], to Node[Mid, Out]) error {
	var inputs []<-chan Mid
	for _, source := range sources {
		out, err := source.Output()
		if err != nil {
			return err
		}
		inputs = append(inputs, out)
	}
	return to.SetInput(inputs...)
}
