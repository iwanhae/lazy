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

// New wraps a user-provided input channel as a lazy object.
//
// Input: receive-only `<-chan T` provided by the caller.
// Output: `object[T]` which forwards values from the input channel.
// Order is preserved. Cancellation is respected before forwarding each value.
// Buffering: the returned object's output channel uses `WithSize`.
func New[T any](ctx context.Context, in <-chan T, opts ...optionFunc) object[T] {
	opt := buildOpts(opts)
	ch := make(chan T, opt.size)
	go func() {
		defer recover()
		defer close(ch)
		for v := range in {
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
