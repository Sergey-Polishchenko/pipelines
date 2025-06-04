// Package pipelines provides lightweight abstractions for building and executing
// data-processing pipelines in Go, where each node can have multiple inputs and outputs
// connected via channels.
package pipelines

import (
	"context"
	"sync"
)

// Runnable represents an entity that can be run with a context.
type Runnable interface {
	// Run starts processing and blocks until completion or error.
	// If the context is canceled, Run should exit promptly.
	Run(ctx context.Context) error
}

// Pipeline represents a collection of Runnables that are executed concurrently.
type Pipeline interface {
	Runnable

	// Add registers one or more Runnables (nodes) with this pipeline.
	// These will be started when Run is called.
	Add(...Runnable)
}

type pipeline struct {
	nodes []Runnable
}

// New creates and returns a new, empty Pipeline.
// Use Add to attach nodes, then call Run to execute.
func New() Pipeline {
	return &pipeline{}
}

func (p *pipeline) Add(n ...Runnable) {
	p.nodes = append(p.nodes, n...)
}

func (p *pipeline) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errChan := make(chan error, len(p.nodes))

	var wg sync.WaitGroup
	wg.Add(len(p.nodes))

	for _, node := range p.nodes {
		go func() {
			defer wg.Done()
			if err := node.Run(ctx); err != nil {
				select {
				case errChan <- err:
					cancel()
				default:
				}
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}
	return nil
}
