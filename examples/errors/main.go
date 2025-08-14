package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/iwanhae/lazy"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Input values 1..5; stop mapping when value == 3
	nums := lazy.NewSlice(ctx, []int{1, 2, 3, 4, 5})
	mapped := lazy.Map(ctx, nums, func(v int) (int, error) {
		if v == 3 {
			return 0, errors.New("boom")
		}
		return v * 10, nil
	}, lazy.WithErrHandler(func(err error) lazy.Decision {
		// Stop the pipeline on first mapping error
		return lazy.DecisionStop
	}))

	fmt.Println("results (should stop before 3):")
	_ = lazy.Consume(mapped, func(v int) error {
		fmt.Println(v)
		return nil
	})
}
