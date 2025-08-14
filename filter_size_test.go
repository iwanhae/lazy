package lazy

import (
	"context"
	"testing"

	"go.uber.org/goleak"
)

func TestWithSize_FilterBuffer(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := NewSlice[int](ctx, []int{1, 2, 3}, WithSize(1))
	out := Filter[int](ctx, in, func(v int) (bool, error) { return true, nil }, WithSize(5))
	if cap(out.ch) != 5 {
		t.Fatalf("expected Filter buffer=5, got %d", cap(out.ch))
	}
}
