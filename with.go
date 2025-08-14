package lazy

type option struct {
	size    int
	onError errHandlerFunc
}

type optionFunc func(opts *option)

func buildOpts(opts []optionFunc) option {
	opt := option{
		size:    0,
		onError: IgnoreErrorHandler,
	}
	for _, f := range opts {
		f(&opt)
	}
	return opt
}

func WithSize(size int) optionFunc {
	return func(opts *option) {
		opts.size = size
	}
}

type Decision string

const (
	DecisionStop   = "stop"
	DecisionIgnore = "ignore"
)

type errHandlerFunc func(err error) Decision

var (
	IgnoreErrorHandler errHandlerFunc = func(err error) Decision {
		return DecisionIgnore
	}
)

func WithErrHandler(handler errHandlerFunc) optionFunc {
	return func(opts *option) {
		opts.onError = handler
	}
}
