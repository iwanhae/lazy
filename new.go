package lazy

import "context"

type object[T any] struct {
	ch chan T
}

// NewSlice creates a source object from a slice.
//
// Input: slice []T
// Output: object[T]
// Order: preserves input order for emitted values
// Cancellation: stops emission when ctx.Done()
// Errors: none
// Buffering: output channel capacity via WithSize
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

// New wraps a user-provided channel as a source object.
//
// Input: in <-chan T (receive-only, user-provided)
// Output: object[T] (forwards values from in)
// Order: preserves input order for emitted values
// Cancellation: stops forwarding when ctx.Done()
// Errors: none
// Buffering: output channel capacity via WithSize
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
