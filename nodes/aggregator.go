package nodes

import (
	"context"
	"fmt"

	"github.com/Sergey-Polishchenko/pipelines"
	"github.com/Sergey-Polishchenko/pipelines/pkg/utils"
)

var _ pipelines.Node[any, any] = &aggregator[any]{}

type aggregator[In any] struct {
	id uint64

	in   []<-chan In
	sink Sink[In]

	config Config
}

// NewResultAggregator creates a node that consumes inputs from multiple channels and applies the Sink function
// to each element. The node has no output channels; once all inputs are closed, Run returns. cfg.InBuffer
// controls the buffer size for the internal fan-in operation.
func NewResultAggregator[In any](sink Sink[In], cfg ...Config) pipelines.Node[In, any] {
	config := DefaultConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	return &aggregator[In]{
		id:     nextNodeID(),
		sink:   sink,
		config: config,
	}
}

func (n *aggregator[In]) ID() string {
	return fmt.Sprintf("result-aggregator-node-%d", n.id)
}

func (n *aggregator[In]) SetInput(in ...<-chan In) error {
	n.in = append(n.in, in...)
	return nil
}

func (n *aggregator[In]) Output() (chan any, error) {
	return nil, ErrHasNoOutput
}

func (n *aggregator[In]) Run(ctx context.Context) error {
	input, err := utils.FanIn(ctx, n.in, n.config.InBuffer)
	if err != nil {
		return err
	}

	for {
		select {
		case data, open := <-input:
			if !open {
				return nil
			}

			if err := n.sink(data); err != nil {
				return fmt.Errorf("%s: %w", n.ID(), err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
