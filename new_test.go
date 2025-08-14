package lazy_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/iwanhae/lazy"
	"go.uber.org/goleak"
)

func TestNew_WrapsUserChannel(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan int)
	go func() {
		defer close(in)
		for i := 1; i <= 5; i++ {
			in <- i
		}
	}()

	nums := lazy.New[int](ctx, in)
	id := lazy.Map(ctx, nums, func(v int) (int, error) { return v, nil })

	var got []int
	if err := lazy.Consume(id, func(v int) error {
		got = append(got, v)
		return nil
	}); err != nil {
		t.Fatalf("consume error: %v", err)
	}

	want := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result. got=%v want=%v", got, want)
	}
}
