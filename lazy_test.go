package lazy_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/iwanhae/lazy"
	"go.uber.org/goleak"
)

func TestLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// given
	nums := lazy.NewSlice(ctx, []int{1, 2, 3, 4, 5})

	// when
	errOccur := false
	mapped := lazy.Map(ctx, nums, func(i int) (int, error) {
		if i == 3 {
			errOccur = true
			return 0, errors.New("error")
		}
		return i, nil
	}, lazy.WithErrHandler(func(err error) lazy.Decision {
		return lazy.DecisionStop
	}))

	// then
	_ = lazy.Consume(mapped, func(i int) error {
		return nil
	})

	if !errOccur {
		t.Error("error should occur")
	}
	time.Sleep(100 * time.Millisecond) // wait for goroutine to exit
}
