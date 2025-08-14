package lazy

import "context"

// Filter passes through only values for which predicate returns true.
//
// Errors from predicate are handled via WithErrHandler. If the handler returns
// DecisionStop, the pipeline stops; if DecisionIgnore, the value is dropped and
// processing continues.
func Filter[T any](ctx context.Context, obj object[T], predicate func(v T) (bool, error), opts ...optionFunc) object[T] {
	opt := buildOpts(opts)
	ch := make(chan T, opt.size)

	go func() {
		defer recover()
		defer close(ch)
		for v := range obj.ch {
			ok, err := predicate(v)
			if err != nil {
				if decision := opt.onError(err); decision == DecisionStop {
					return
				}
				// DecisionIgnore: drop value and continue
				continue
			}
			if !ok {
				// Filtered out
				continue
			}

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
