package lazy_test

import (
	"context"
	"testing"

	"github.com/iwanhae/lazy"
	"go.uber.org/goleak"
)

func TestNewSlice_EmptyInput(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nums := lazy.NewSlice(ctx, []int{})
	id := lazy.Map(ctx, nums, func(v int) (int, error) { return v, nil })

	consumed := 0
	if err := lazy.Consume(id, func(v int) error {
		consumed++
		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if consumed != 0 {
		t.Fatalf("expected 0 items, got %d", consumed)
	}
}
