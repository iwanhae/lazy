package lazy_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/iwanhae/lazy"
	"go.uber.org/goleak"
)

func TestFilter_Evens(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nums := lazy.NewSlice(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	evens := lazy.Filter(ctx, nums, func(v int) (bool, error) {
		return v%2 == 0, nil
	})

	var got []int
	if err := lazy.Consume(evens, func(v int) error {
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

func TestFilter_DefaultIgnoreError(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nums := lazy.NewSlice(ctx, []int{1, 2, 3, 4, 5, 6})
	filtered := lazy.Filter(ctx, nums, func(v int) (bool, error) {
		if v > 4 {
			return false, errors.New("gt4")
		}
		return v%2 == 0, nil
	})

	var got []int
	if err := lazy.Consume(filtered, func(v int) error {
		got = append(got, v)
		return nil
	}); err != nil {
		t.Fatalf("consume error: %v", err)
	}

	// Values >4 cause predicate error and are dropped by default; evens <=4 remain
	want := []int{2, 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result. got=%v want=%v", got, want)
	}
}

func TestFilter_StopOnError(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nums := lazy.NewSlice(ctx, []int{1, 2, 3, 4, 5})
	filtered := lazy.Filter(ctx, nums, func(v int) (bool, error) {
		if v == 3 {
			return false, errors.New("boom")
		}
		return true, nil
	}, lazy.WithErrHandler(func(err error) lazy.Decision { return lazy.DecisionStop }))

	var got []int
	if err := lazy.Consume(filtered, func(v int) error {
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
