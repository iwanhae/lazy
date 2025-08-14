package main

import (
	"fmt"

	"github.com/iwanhae/lazy"
)

func main() {
	a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	nums := lazy.NewSlice(a, lazy.WithSize(5))

	doubled := lazy.Map(nums, func(v int) (int, error) {
		return v * 2, nil
	}, lazy.WithSize(1))

	_ = lazy.Consume(doubled, func(v int) error {
		fmt.Println(v)
		return nil
	})
}
