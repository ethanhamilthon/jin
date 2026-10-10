package update

import (
	"context"
	"io"
)

// Control stops the running daemon before an update replaces the binary the
// daemon runs, and starts one again afterwards. Client and daemon share one
// executable, so a daemon left running would make every later command fail
// with a version mismatch. A zero Control skips both steps.
type Control struct {
	Stop  func(ctx context.Context, force bool, out io.Writer) error
	Start func(ctx context.Context, out io.Writer) error
}

func (c Control) stop(ctx context.Context, force bool, out io.Writer) error {
	if c.Stop == nil {
		return nil
	}
	return c.Stop(ctx, force, out)
}

func (c Control) start(ctx context.Context, out io.Writer) error {
	if c.Start == nil {
		return nil
	}
	return c.Start(ctx, out)
}
