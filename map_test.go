package lazy_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/iwanhae/lazy"
	"go.uber.org/goleak"
)

func TestMap_DefaultIgnoreError(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nums := lazy.NewSlice(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	doubled := lazy.Map(ctx, nums, func(v int) (int, error) {
		if v > 5 {
			return 0, errors.New("gt5")
		}
		return v * 2, nil
	})

	var got []int
	if err := lazy.Consume(doubled, func(v int) error {
		got = append(got, v)
		return nil
	}); err != nil {
		t.Fatalf("consume error: %v", err)
	}

	want := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result. got=%v want=%v", got, want)
	}
}

func TestMap_StopOnError(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nums := lazy.NewSlice(ctx, []int{1, 2, 3, 4, 5})
	mapped := lazy.Map(ctx, nums, func(v int) (int, error) {
		if v == 3 {
			return 0, errors.New("boom")
		}
		return v, nil
	}, lazy.WithErrHandler(func(err error) lazy.Decision {
		return lazy.DecisionStop
	}))

	var got []int
	if err := lazy.Consume(mapped, func(v int) error {
		got = append(got, v)
		return nil
	}); err != nil {
		t.Fatalf("consume error: %v", err)
	}

	// Should only get values before the error (1, 2)
	want := []int{1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result. got=%v want=%v", got, want)
	}
}

func TestConsume_PropagatesConsumerError(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nums := lazy.NewSlice(ctx, []int{1, 2, 3})
	id := lazy.Map(ctx, nums, func(v int) (int, error) { return v, nil })

	wantErr := errors.New("stop@2")
	err := lazy.Consume(id, func(v int) error {
		if v == 2 {
			return wantErr
		}
		return nil
	})
	if err == nil || err.Error() != wantErr.Error() {
		t.Fatalf("expected consumer error %v, got %v", wantErr, err)
	}
}

func TestOrderPreserved_WithIgnoredErrors(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nums := lazy.NewSlice(ctx, []int{1, 2, 3, 4, 5})
	mapped := lazy.Map(ctx, nums, func(v int) (int, error) {
		if v == 3 {
			return 0, errors.New("skip")
		}
		return v, nil
	})

	var got []int
	if err := lazy.Consume(mapped, func(v int) error {
		got = append(got, v)
		return nil
	}); err != nil {
		t.Fatalf("consume error: %v", err)
	}

	want := []int{1, 2, 4, 5}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result. got=%v want=%v", got, want)
	}
}
