package lazy_test

import (
    "context"
    "testing"
    "time"

    "github.com/iwanhae/lazy"
)

func TestContextCancellationStopsPipeline(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Large input to ensure there would be more to process after cancel.
    var big []int
    for i := 0; i < 100000; i++ {
        big = append(big, i)
    }

    nums := lazy.NewSlice(ctx, big)
    id := lazy.Map(ctx, nums, func(v int) (int, error) { return v, nil })

    done := make(chan struct{})
    count := 0
    go func() {
        _ = lazy.Consume(id, func(v int) error {
            count++
            if count == 5 {
                cancel()
            }
            return nil
        })
        close(done)
    }()

    select {
    case <-done:
        // ok
    case <-time.After(2 * time.Second):
        t.Fatal("pipeline did not stop after cancellation")
    }

    if count < 5 {
        t.Fatalf("expected to consume at least 5 before cancel, got %d", count)
    }
}

