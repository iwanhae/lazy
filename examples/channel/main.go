package main

import (
	"context"
	"fmt"

	"github.com/iwanhae/lazy"
)

// Example: Wrap a user-provided channel using lazy.New, then Filter and Map.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// User-owned channel producing 1..6, then closed.
	in := make(chan int)
	go func() {
		defer close(in)
		for i := 1; i <= 6; i++ {
			in <- i
		}
	}()

	nums := lazy.New[int](ctx, in)

	// Keep odd numbers only, then multiply by 10.
	odds := lazy.Filter(ctx, nums, func(v int) (bool, error) { return v%2 == 1, nil })
	tens := lazy.Map(ctx, odds, func(v int) (int, error) { return v * 10, nil })

	_ = lazy.Consume(tens, func(v int) error {
		fmt.Println(v)
		return nil
	})
}
