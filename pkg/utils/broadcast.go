package utils

import (
	"context"
)

// Broadcast sends the given data to all output channels.
// It returns an error if the context is canceled.
func Broadcast[Out any](ctx context.Context, outChans []chan<- Out, data Out) error {
	for _, out := range outChans {
		select {
		case out <- data:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
