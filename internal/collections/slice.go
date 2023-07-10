package collections

func SliceRemoveElement[T any](s []T, index int) []T {
	if index >= len(s) || index < 0 {
		return s
	}

	newSlice := make([]T, 0, len(s)-1)
	newSlice = append(newSlice, s[:index]...)
	newSlice = append(newSlice, s[index+1:]...)

	return newSlice
}
