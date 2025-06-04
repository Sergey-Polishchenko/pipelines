package nodes

import (
	"context"
	"fmt"

	"github.com/Sergey-Polishchenko/pipelines"
	"github.com/Sergey-Polishchenko/pipelines/pkg/utils"
)

var _ pipelines.Node[any, any] = &zip[any, any]{}

type zip[In, Out any] struct {
	id uint64

	in      []<-chan In
	out     []chan<- Out
	process ZipProcessor[In, Out]

	config Config
}

// NewZip creates a node that reads one element from each of its input channels, collects them into
// a slice, and applies the ZipProcessor function to produce a single output. The node can have multiple
// output channels, each created with buffer size cfg.Buffer. If any input channel is closed prematurely,
// Run returns an error.
func NewZip[In, Out any](proc ZipProcessor[In, Out], cfg ...Config) pipelines.Node[In, Out] {
	config := DefaultConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	return &zip[In, Out]{
		id:      nextNodeID(),
		process: proc,
		config:  config,
	}
}

func (n *zip[In, Out]) ID() string {
	return fmt.Sprintf("zip-node-%d", n.id)
}

func (n *zip[In, Out]) SetInput(in ...<-chan In) error {
	n.in = append(n.in, in...)

	return nil
}

func (n *zip[In, Out]) Output() (chan Out, error) {
	out := make(chan Out, n.config.Buffer)
	n.out = append(n.out, out)
	return out, nil
}

func (n *zip[In, Out]) Run(ctx context.Context) error {
	if len(n.in) == 0 {
		return ErrZipNodeNoInput
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		inputs := make([]In, len(n.in))
		for i, ch := range n.in {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case data, ok := <-ch:
				if !ok {
					return ErrZipNodeClosedInput
				}
				inputs[i] = data
			}
		}

		res, err := n.process(inputs)
		if err != nil {
			return fmt.Errorf("%s: %w", n.ID(), err)
		}

		if err := utils.Broadcast(ctx, n.out, res); err != nil {
			return err
		}
	}
}
