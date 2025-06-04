// Package utils provides primitives for channel operations used in custom node implementations.
// These functions handle context cancellation safely and simplify complex concurrency patterns.
package utils

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrEmptyInChan     = errors.New("empty in chan")
	ErrNegativeBufSize = errors.New("negative buffer size")
)

// FanIn merges multiple input channels into a single output channel with the specified buffer size.
// It returns the merged channel or an error if parameters are invalid.
func FanIn[T any](ctx context.Context, in []<-chan T, buf int) (<-chan T, error) {
	if len(in) == 0 {
		return nil, ErrEmptyInChan
	}

	if buf < 0 {
		return nil, ErrNegativeBufSize
	}

	out := make(chan T, buf)

	var wg sync.WaitGroup
	wg.Add(len(in))

	for _, ch := range in {
		go func() {
			defer wg.Done()

			for {
				select {
				case data, open := <-ch:
					if !open {
						return
					}

					select {
					case out <- data:
					case <-ctx.Done():
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out, nil
}
