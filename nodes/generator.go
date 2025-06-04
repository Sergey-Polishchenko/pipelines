package nodes

import (
	"context"
	"fmt"

	"github.com/Sergey-Polishchenko/pipelines"
	"github.com/Sergey-Polishchenko/pipelines/pkg/utils"
)

var _ pipelines.Node[any, any] = &generator[any]{}

type generator[Out any] struct {
	id uint64

	out      []chan<- Out
	generate Generator[Out]
}

// NewGenerator creates a node that produces elements using the provided Generator function.
// The Generator receives a context for cancellation and returns a receive-only channel of outputs.
// The node can have multiple output channels; each output is created with buffer size cfg.Buffer.
func NewGenerator[Out any](gen Generator[Out]) pipelines.Node[any, Out] {
	return &generator[Out]{
		id:       nextNodeID(),
		generate: gen,
	}
}

func (n *generator[Out]) ID() string {
	return fmt.Sprintf("generator-node-%d", n.id)
}

func (n *generator[Out]) SetInput(in ...<-chan any) error {
	return ErrHasNoInput
}

func (n *generator[Out]) Output() (chan Out, error) {
	out := make(chan Out, 1)
	n.out = append(n.out, out)
	return out, nil
}

func (n *generator[Out]) Run(ctx context.Context) error {
	defer utils.CloseChannels(n.out)

	ch, err := n.generate(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", n.ID(), err)
	}

	for data := range ch {
		if err := utils.Broadcast(ctx, n.out, data); err != nil {
			return err
		}
	}
	return nil
}
