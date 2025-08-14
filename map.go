package lazy

import "context"

func Map[IN any, OUT any](ctx context.Context, obj object[IN], mapper func(v IN) (OUT, error), opts ...optionFunc) object[OUT] {
	opt := buildOpts(opts)
	ch := make(chan OUT, opt.size)

	go func() {
		defer recover()
		defer close(ch)
		for v := range obj.ch {
			result, err := mapper(v)
			if err != nil {
				if decision := opt.onError(err); decision == DecisionStop {
					return
				}
				// DecisionIgnore: drop value and continue
				continue
			}
			// Respect cancellation when forwarding results to the next stage
			select {
			case <-ctx.Done():
				return
			case ch <- result:
			}
		}
	}()

	return object[OUT]{
		ch: ch,
	}
}
