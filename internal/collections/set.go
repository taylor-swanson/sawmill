package collections

// Set is a collection data structure that retains single values of a given type.
type Set[T comparable] struct {
	m map[T]struct{}
}

// Add adds one or more values to the Set.
func (s Set[T]) Add(values ...T) {
	for _, v := range values {
		s.m[v] = struct{}{}
	}
}

// Has returns true if value is in the Set.
func (s Set[T]) Has(value T) bool {
	_, exists := s.m[value]

	return exists
}

// Values returns the values in this set as a slice of T. The elements will not be sorted.
func (s Set[T]) Values() []T {
	if len(s.m) == 0 {
		return nil
	}

	values := make([]T, 0, len(s.m))
	for k := range s.m {
		values = append(values, k)
	}

	return values
}

// Remove removes an element from the Set. If the element is not present in the
// Set, the operation is a no-op.
func (s Set[T]) Remove(value T) {
	delete(s.m, value)
}

func (s Set[T]) Len() int {
	return len(s.m)
}

func (s Set[T]) Clear() {
	for k := range s.m {
		delete(s.m, k)
	}
}

func (s Set[T]) IsDisjoint(other Set[T]) bool {
	// TODO: Implement
	panic("Not yet implemented")
}

func (s Set[T]) IsSubset(other Set[T]) bool {
	// TODO: Implement
	panic("Not yet implemented")
}

func (s Set[T]) IsSuperset(other Set[T]) bool {
	// TODO: Implement
	panic("Not yet implemented")
}

func (s Set[T]) Union(other Set[T]) Set[T] {
	// TODO: Implement
	panic("Not yet implemented")
}

func (s Set[T]) Intersection(other Set[T]) Set[T] {
	// TODO: Implement
	panic("Not yet implemented")
}

func (s Set[T]) Difference(other Set[T]) Set[T] {
	// TODO: Implement
	panic("Not yet implemented")
}

func (s Set[T]) SymmetricDifference(other Set[T]) Set[T] {
	// TODO: Implement
	panic("Not yet implemented")
}

// NewSet creates a new set and optionally adds one or more values to the Set.
func NewSet[T comparable](values ...T) Set[T] {
	s := Set[T]{m: map[T]struct{}{}}
	s.Add(values...)

	return s
}
