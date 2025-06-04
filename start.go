package pipelines

import "context"

// For Starting Nodes without pipeline
func Start[In, Out any](ctx context.Context, n Node[In, Out]) error {
	errChan := make(chan error)

	go func() {
		errChan <- n.Run(ctx)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
