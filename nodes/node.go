// Package nodes provides implementations of various Node types for pipelines,
// including basic processing node, worker pool, result aggregator, generator, and zip node.
package nodes

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/Sergey-Polishchenko/pipelines"
	"github.com/Sergey-Polishchenko/pipelines/pkg/utils"
)

var _ pipelines.Node[any, any] = &node[any, any]{}

type node[In, Out any] struct {
	id uint64

	in      []<-chan In
	out     []chan<- Out
	process Processor[In, Out]

	isRunning atomic.Bool
	config    Config
}

// NewNode creates a basic node that applies the provided Processor function to each input element.
// The node can have zero or more input channels and may have multiple output channels.
// Each output channel is created with a buffer size defined in cfg.Buffer (or using DefaultConfig if none provided).
// The Processor function is called for each element received via a fan-in of the inputs.
func NewNode[In, Out any](proc Processor[In, Out], cfg ...Config) pipelines.Node[In, Out] {
	config := DefaultConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	return &node[In, Out]{
		id:      nextNodeID(),
		process: proc,
		config:  config,
	}
}

func (n *node[In, Out]) ID() string {
	return fmt.Sprintf("node-%d", n.id)
}

func (n *node[In, Out]) SetInput(in ...<-chan In) error {
	if n.isRunning.Load() {
		return ErrAccessRunningNode
	}

	n.in = append(n.in, in...)

	return nil
}

func (n *node[In, Out]) Output() (chan Out, error) {
	if n.isRunning.Load() {
		return nil, ErrAccessRunningNode
	}

	out := make(chan Out, n.config.Buffer)
	n.out = append(n.out, out)

	return out, nil
}

// node[In, Out].Run(ctx) error
func (n *node[In, Out]) Run(ctx context.Context) error {
	if n.isRunning.Load() {
		return ErrNodeRunning
	}

	inChan, err := utils.FanIn(ctx, n.in, n.config.InBuffer)
	if err != nil {
		return err
	}
	defer utils.CloseChannels(n.out)

	// not safe for concurrent use
	n.isRunning.Store(true)
	defer n.isRunning.Swap(false)

	for {
		select {
		case data, open := <-inChan:
			if !open {
				return nil
			}

			result, err := n.process(data)
			if err != nil {
				return fmt.Errorf("%s: %w", n.ID(), err)
			}

			if err := utils.Broadcast(ctx, n.out, result); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
