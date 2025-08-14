package lazy

import "context"

type object[T any] struct {
	ch chan T
}

func NewSlice[T any](ctx context.Context, slice []T, opts ...optionFunc) object[T] {
	opt := buildOpts(opts)
	ch := make(chan T, opt.size)
	go func() {
		defer recover()
		defer close(ch)
		for _, v := range slice {
			select {
			case <-ctx.Done():
				return
			case ch <- v:
			}
		}
	}()
	return object[T]{
		ch: ch,
	}
}
