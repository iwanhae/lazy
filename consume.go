package lazy

// Consume drains the object and applies consumer to each value.
//
// Input: object[T], consumer func(T) error
// Output: error (first consumer error, if any)
// Order: consumes values in upstream order
// Cancellation: N/A; respects upstream closure
// Errors: returns the first error from consumer
// Buffering: N/A
func Consume[IN any](obj object[IN], consumer func(v IN) error) error {
	for v := range obj.ch {
		if err := consumer(v); err != nil {
			return err
		}
	}
	return nil
}
