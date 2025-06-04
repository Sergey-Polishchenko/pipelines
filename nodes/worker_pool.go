package nodes

import (
	"context"
	"fmt"
	"sync"

	"github.com/Sergey-Polishchenko/pipelines"
)

var _ pipelines.Node[any, any] = &workerPool[any, any]{}

type workerPool[In, Out any] struct {
	id uint64

	in      <-chan In
	out     chan<- Out
	process Processor[In, Out]

	config Config
}

// NewWorkerPool creates a node that processes inputs using a pool of worker goroutines.
// It accepts exactly one input channel, and spawns cfg.Workers concurrent goroutines,
// each applying the Processor function to incoming elements. Results are sent to a single output channel
// buffered with size cfg.Workers. If more than one input channel is set via SetInput, returns an error.
func NewWorkerPool[In, Out any](
	proc Processor[In, Out],
	cfg ...Config,
) pipelines.Node[In, Out] {
	config := DefaultConfig()
	if len(cfg) > 0 && cfg[0].Workers > 0 {
		config = cfg[0]
	}
	return &workerPool[In, Out]{
		id:      nextNodeID(),
		process: proc,
		config:  config,
	}
}

func (n *workerPool[In, Out]) ID() string {
	return fmt.Sprintf("worker-pool-node-%d", n.id)
}

func (n *workerPool[In, Out]) SetInput(in ...<-chan In) error {
	if len(in) != 1 {
		return ErrOnlyOneInput
	}
	n.in = in[0]
	return nil
}

func (n *workerPool[In, Out]) Output() (chan Out, error) {
	out := make(chan Out, n.config.Workers)
	n.out = out
	return out, nil
}

func (n *workerPool[In, Out]) Run(ctx context.Context) error {
	defer close(n.out)

	errChan := make(chan error, n.config.Workers)

	var wg sync.WaitGroup
	wg.Add(n.config.Workers)

	for i := 0; i < n.config.Workers; i++ {
		go n.runWorker(ctx, errChan, &wg)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("%s: %w", n.ID(), err)
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (n *workerPool[In, Out]) runWorker(
	ctx context.Context,
	errChan chan<- error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case data, ok := <-n.in:
			if !ok {
				return
			}

			result, err := n.process(data)
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}

			select {
			case n.out <- result:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			return
		}
	}
}
