package lazy

import "context"

func Map[IN any, OUT any](ctx context.Context, obj object[IN], mapper func(v IN) (OUT, error), opts ...optionFunc) object[OUT] {
	opt := buildOpts(opts)
	ch := make(chan OUT, opt.size)

	go func() {
		defer recover()
		defer close(ch)
		for v := range obj.ch {
			select {
			case <-ctx.Done():
				return
			default:
				result, err := mapper(v)
				if err != nil {
					if decision := opt.onError(err); decision == DecisionStop {
						return
					}
				} else {
					ch <- result
				}
			}
		}
	}()

	return object[OUT]{
		ch: ch,
	}
}
