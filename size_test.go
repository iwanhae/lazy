package lazy

import (
	"context"
	"go.uber.org/goleak"
	"testing"
)

// Whitebox tests that inspect internal channel buffer sizes.
func TestWithSize_NewSliceBuffer(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obj := NewSlice[int](ctx, []int{1, 2, 3}, WithSize(3))
	if cap(obj.ch) != 3 {
		t.Fatalf("expected NewSlice buffer=3, got %d", cap(obj.ch))
	}
}

func TestWithSize_MapBuffer(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := NewSlice[int](ctx, []int{1}, WithSize(1))
	out := Map[int, int](ctx, in, func(v int) (int, error) { return v, nil }, WithSize(4))
	if cap(out.ch) != 4 {
		t.Fatalf("expected Map buffer=4, got %d", cap(out.ch))
	}
}
