package lazy

func Consume[IN any](obj object[IN], consumer func(v IN) error) error {
	for v := range obj.ch {
		if err := consumer(v); err != nil {
			return err
		}
	}
	return nil
}
