package expressions

func Select[T any](condition bool, thenValue T, elseValue T) T {
	if condition {
		return thenValue
	} else {
		return elseValue
	}
}
