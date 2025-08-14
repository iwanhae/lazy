package main

import (
	"context"
	"fmt"
	"time"

	"github.com/iwanhae/lazy"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Large input to illustrate cancellation mid-stream
	var big []int
	for i := 1; i <= 1000; i++ {
		big = append(big, i)
	}

	nums := lazy.NewSlice(ctx, big)
	id := lazy.Map(ctx, nums, func(v int) (int, error) { return v, nil })

	consumed := 0
	done := make(chan struct{})
	go func() {
		_ = lazy.Consume(id, func(v int) error {
			consumed++
			if consumed == 5 {
				// Cancel after consuming a few items
				cancel()
			}
			return nil
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		fmt.Println("timeout waiting for cancellation")
	}

	fmt.Printf("consumed before cancel: %d\n", consumed)
}
