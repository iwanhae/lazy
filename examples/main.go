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

	doubled := lazy.Map(ctx, nums, func(v int) (int, error) {
		if v > 5 {
			return 0, fmt.Errorf("v is greater than 5")
		}
		return v * 2, nil
	})

	if err := lazy.Consume(doubled, func(v int) error {
		fmt.Println(v)
		return nil
	}); err != nil {
		fmt.Println("err:", err)
	}
}
