package utils

// This function ingests the pointer of any type and returns the
// zero value if the pointer is nil. Otherwise, returns the dereferenced value.
func Safe[T any](t *T) T {
	var zero T
	if t == nil {
		return zero
	}
	return *t
}
