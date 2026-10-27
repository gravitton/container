package container

import (
	"encoding/json"
	"fmt"
	"iter"
	"slices"
)

// Set is an immutable set of elements, held once each in the order of their Compare method,
// which alone decides that two elements are the same and must be a total order. A set is safe to
// copy and share. The zero value is the empty set.
type Set[E interface{ Compare(E) int }] struct {
	elements []E
}

// SetOf returns the set of the given elements, in any order and with any repeats.
func SetOf[E interface{ Compare(E) int }](elements ...E) Set[E] {
	return SetFrom(slices.Clone(elements))
}

// SetFrom returns the set of the given elements, in any order and with any repeats, taking the
// slice as its own: it rearranges what the slice holds, and a later write to it changes the set.
func SetFrom[E interface{ Compare(E) int }](elements []E) Set[E] {
	if len(elements) == 0 {
		return Set[E]{}
	}

	slices.SortFunc(elements, E.Compare)

	return Set[E]{slices.CompactFunc(elements, same)}
}

// Len returns the number of elements in the set.
func (s Set[E]) Len() int {
	return len(s.elements)
}

// At returns the element at the index in the order of the set. It panics for an index outside
// the set.
func (s Set[E]) At(i int) E {
	if i < 0 || i >= len(s.elements) {
		panic(setIndexError{i, len(s.elements)})
	}

	return s.elements[i]
}

// Index returns the index of the element in the order of the set, the one At reads it back
// from, and true, or false where the set does not hold it.
func (s Set[E]) Index(element E) (int, bool) {
	i, found := s.search(element)
	if !found {
		return 0, false
	}

	return i, true
}

// First returns the first element in the order of the set and true, or false for the empty set.
func (s Set[E]) First() (E, bool) {
	if s.IsEmpty() {
		var zero E

		return zero, false
	}

	return s.elements[0], true
}

// Last returns the last element in the order of the set and true, or false for the empty set.
func (s Set[E]) Last() (E, bool) {
	if s.IsEmpty() {
		var zero E

		return zero, false
	}

	return s.elements[len(s.elements)-1], true
}

// search returns the index of the element in the order of the set and true, or the index it
// would be inserted at and false.
func (s Set[E]) search(element E) (int, bool) {
	return slices.BinarySearchFunc(s.elements, element, E.Compare)
}

// Elements returns the elements of the set in order, and nil for the empty set.
func (s Set[E]) Elements() []E {
	if s.IsEmpty() {
		return nil
	}

	return s.AppendElements(make([]E, 0, len(s.elements)))
}

// AppendElements appends the elements of Elements to dst and returns the extended slice.
func (s Set[E]) AppendElements(dst []E) []E {
	return append(dst, s.elements...)
}

// All returns an iterator over the elements of the set, in order.
func (s Set[E]) All() iter.Seq[E] {
	return slices.Values(s.elements)
}

// Backward returns an iterator over the elements of the set, in reverse order.
func (s Set[E]) Backward() iter.Seq[E] {
	return func(yield func(E) bool) {
		for _, element := range slices.Backward(s.elements) {
			if !yield(element) {
				return
			}
		}
	}
}

// Union returns the set of the elements in s, in other or in both.
func (s Set[E]) Union(other Set[E]) Set[E] {
	return s.merge(other, func(int) bool {
		return true
	})
}

// Intersection returns the set of the elements in both s and other.
func (s Set[E]) Intersection(other Set[E]) Set[E] {
	return s.merge(other, func(side int) bool {
		return side == 0
	})
}

// Difference returns the set of the elements in s that are not in other.
func (s Set[E]) Difference(other Set[E]) Set[E] {
	return s.merge(other, func(side int) bool {
		return side < 0
	})
}

// SymmetricDifference returns the set of the elements in s or in other, and not in both.
func (s Set[E]) SymmetricDifference(other Set[E]) Set[E] {
	return s.merge(other, func(side int) bool {
		return side != 0
	})
}

// merge returns the set of the elements of s and other on a side to keep, walked once to count
// them and once to place them. Where the result holds the elements of one set, it is that set
// itself, which immutability makes safe to share.
func (s Set[E]) merge(other Set[E], keep func(side int) bool) Set[E] {
	switch {
	case other.IsEmpty() && keep(-1):
		return s
	case s.IsEmpty() && keep(1):
		return other
	}

	size, onlyOwn, onlyOther := 0, 0, 0
	s.walk(other, func(_ E, side int) {
		if !keep(side) {
			return
		}

		size++
		switch {
		case side < 0:
			onlyOwn++
		case side > 0:
			onlyOther++
		}
	})

	switch {
	case size == 0:
		return Set[E]{}
	case onlyOther == 0 && size == s.Len():
		return s
	case onlyOwn == 0 && size == other.Len():
		return other
	}

	elements := make([]E, 0, size)
	s.walk(other, func(element E, side int) {
		if keep(side) {
			elements = append(elements, element)
		}
	})

	return Set[E]{elements}
}

// walk visits every element of s and other once, in order, with the side it lies on: negative
// in s alone, zero in both, positive in other alone.
func (s Set[E]) walk(other Set[E], visit func(element E, side int)) {
	i, j := 0, 0
	for i < len(s.elements) && j < len(other.elements) {
		switch side := s.elements[i].Compare(other.elements[j]); {
		case side < 0:
			visit(s.elements[i], side)
			i++
		case side > 0:
			visit(other.elements[j], side)
			j++
		default:
			visit(s.elements[i], side)
			i++
			j++
		}
	}

	for _, element := range s.elements[i:] {
		visit(element, -1)
	}

	for _, element := range other.elements[j:] {
		visit(element, 1)
	}
}

// Contains reports whether the element is in the set.
func (s Set[E]) Contains(element E) bool {
	_, found := s.search(element)

	return found
}

// Encloses reports whether s holds every element of other. Every set encloses the empty one.
func (s Set[E]) Encloses(other Set[E]) bool {
	if other.Len() > s.Len() {
		return false
	}

	for i, j := 0, 0; j < len(other.elements); {
		if i == len(s.elements) {
			return false
		}

		switch side := s.elements[i].Compare(other.elements[j]); {
		case side < 0:
			i++
		case side > 0:
			return false
		default:
			i++
			j++
		}
	}

	return true
}

// Intersects reports whether s and other hold an element in common.
func (s Set[E]) Intersects(other Set[E]) bool {
	for i, j := 0, 0; i < len(s.elements) && j < len(other.elements); {
		switch side := s.elements[i].Compare(other.elements[j]); {
		case side < 0:
			i++
		case side > 0:
			j++
		default:
			return true
		}
	}

	return false
}

// Equal reports whether s and other hold the same elements.
func (s Set[E]) Equal(other Set[E]) bool {
	return slices.EqualFunc(s.elements, other.elements, same)
}

// IsEmpty reports whether the set holds no element.
func (s Set[E]) IsEmpty() bool {
	return len(s.elements) == 0
}

// MarshalJSON implements json.Marshaler with the elements as an array in order, null for the
// empty set.
func (s Set[E]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.elements)
}

// UnmarshalJSON implements json.Unmarshaler, reading an array of elements in any order and
// with any repeats, and null or an empty array as the empty set. It leaves the set as it was on
// an error.
func (s *Set[E]) UnmarshalJSON(data []byte) error {
	var elements []E
	if err := json.Unmarshal(data, &elements); err != nil {
		return err
	}

	*s = SetFrom(elements)

	return nil
}

// same reports whether two elements are the same one, by their order.
func same[E interface{ Compare(E) int }](a, b E) bool {
	return a.Compare(b) == 0
}

// setIndexError is the panic of [Set.At] for an index outside the set. It formats its message
// only when printed, which keeps At small enough to inline.
type setIndexError struct {
	index int
	len   int
}

// Error returns the message of the panic.
func (e setIndexError) Error() string {
	return fmt.Sprintf("container: index %d outside set of %d elements", e.index, e.len)
}
