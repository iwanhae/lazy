package main

import (
	"context"
	"fmt"

	"github.com/iwanhae/lazy"
)

func main() {
	a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nums := lazy.NewSlice(ctx, a)

	// Filter values to keep only <= 5, then map to double.
	filtered := lazy.Filter(ctx, nums, func(v int) (bool, error) {
		return v <= 5, nil
	})
	doubled := lazy.Map(ctx, filtered, func(v int) (int, error) {
		return v * 2, nil
	})

	if err := lazy.Consume(doubled, func(v int) error {
		fmt.Println(v)
		return nil
	}); err != nil {
		fmt.Println("err:", err)
	}
}
