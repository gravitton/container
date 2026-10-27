package container

import (
	"encoding/json"
	"fmt"
	"iter"
	"math"
	"slices"

	geom "github.com/gravitton/geometry"
)

// Array is a table of values of type T, one for every index of a size in columns and rows,
// stored row by row. Its sides and indices are of the integer type S, which bounds each side by
// what S holds; only the number of values is an int. It is a handle: a copy shares its values,
// and [Array.Clone] gives storage of its own. The zero value is the empty array.
type Array[T any, S geom.Integer] struct {
	size geom.Size[S]
	data []T
}

// Arr returns an array of the given size, every value zero. It panics for a negative width or
// height, and for a size of more values than an int counts.
func Arr[T any, S geom.Integer](size geom.Size[S]) Array[T, S] {
	if err := validateSize(size); err != nil {
		panic(err)
	}

	return Array[T, S]{size, make([]T, int(size.Width)*int(size.Height))}
}

// Size returns the number of columns and rows of the array.
func (a Array[T, S]) Size() geom.Size[S] {
	return a.size
}

// Width returns the number of columns of the array.
func (a Array[T, S]) Width() S {
	return a.size.Width
}

// Height returns the number of rows of the array.
func (a Array[T, S]) Height() S {
	return a.size.Height
}

// Len returns the number of values of the array (width·height), an int whatever S is.
func (a Array[T, S]) Len() int {
	return len(a.data)
}

// Get returns a pointer into the array to the value at the index: a write through it changes the
// array. It panics for an index outside the array.
func (a Array[T, S]) Get(index geom.Point[S]) *T {
	if !a.Contains(index) {
		panic(arrayIndexError[S]{index, a.size})
	}

	return &a.data[int(index.X)+int(index.Y)*int(a.size.Width)]
}

// Set stores value at the index. It panics for an index outside the array.
func (a Array[T, S]) Set(index geom.Point[S], value T) {
	*a.Get(index) = value
}

// Data returns the values of the array row by row as a slice into the array: a write through it
// changes the array, and an append to it never does. It is nil for the zero array.
func (a Array[T, S]) Data() []T {
	return a.data[:len(a.data):len(a.data)]
}

// Fill sets every value of the array to value.
func (a Array[T, S]) Fill(value T) {
	for i := range a.data {
		a.data[i] = value
	}
}

// Clear resets every value of the array to the zero value.
func (a Array[T, S]) Clear() {
	clear(a.data)
}

// Clone returns a copy of the array with storage of its own, its values copied as an
// assignment copies them.
func (a Array[T, S]) Clone() Array[T, S] {
	return Array[T, S]{a.size, slices.Clone(a.data)}
}

// Keys returns an iterator over every index of the array, row by row.
func (a Array[T, S]) Keys() iter.Seq[geom.Point[S]] {
	return func(yield func(geom.Point[S]) bool) {
		for y := S(0); y < a.size.Height; y++ {
			for x := S(0); x < a.size.Width; x++ {
				if !yield(geom.Pt(x, y)) {
					return
				}
			}
		}
	}
}

// Values returns an iterator over pointers into the array to its values, in the order of Keys.
func (a Array[T, S]) Values() iter.Seq[*T] {
	return func(yield func(*T) bool) {
		for i := range a.data {
			if !yield(&a.data[i]) {
				return
			}
		}
	}
}

// All returns an iterator over every index with a pointer into the array to its value, in the
// order of Keys.
func (a Array[T, S]) All() iter.Seq2[geom.Point[S], *T] {
	return func(yield func(geom.Point[S], *T) bool) {
		width := int(a.size.Width)
		for y := S(0); y < a.size.Height; y++ {
			row := a.data[int(y)*width : (int(y)+1)*width]
			for x := range row {
				if !yield(geom.Pt(S(x), y), &row[x]) {
					return
				}
			}
		}
	}
}

// Contains reports whether the index is that of a value of the array.
func (a Array[T, S]) Contains(index geom.Point[S]) bool {
	return index.X >= 0 && index.X < a.size.Width && index.Y >= 0 && index.Y < a.size.Height
}

// MarshalJSON implements json.Marshaler with the size of the array, flat as geom.Size writes
// it, and its values row by row, null for the zero array.
func (a Array[T, S]) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Size   geom.Size[S] `json:",embed"`
		Values []T          `json:"d"`
	}{a.size, a.data})
}

// UnmarshalJSON implements json.Unmarshaler, reading null as the zero array and null values as
// none. It returns an error for a size Arr panics for or S cannot hold, and for a number of
// values other than the width times the height, and leaves the array as it was on an error.
func (a *Array[T, S]) UnmarshalJSON(data []byte) error {
	var decoded struct {
		Size   geom.Size[S] `json:",embed"`
		Values []T          `json:"d"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	size := decoded.Size
	if err := validateSize(size); err != nil {
		return err
	}

	if len(decoded.Values) != int(size.Width)*int(size.Height) {
		return fmt.Errorf("container: array of size %s with %d values", size, len(decoded.Values))
	}

	*a = Array[T, S]{size, decoded.Values}

	return nil
}

// validateSize returns the error of a size no array has: a negative width or height, or a
// product past the largest int, counted in uint64 so that a side wider than an int cannot wrap.
func validateSize[S geom.Integer](size geom.Size[S]) error {
	if size.Width < 0 || size.Height < 0 {
		return fmt.Errorf("container: array of negative size %s", size)
	}

	width, height := uint64(size.Width), uint64(size.Height)
	if height != 0 && width > math.MaxInt/height {
		return fmt.Errorf("container: array of size %s is too large", size)
	}

	return nil
}

// arrayIndexError is the panic of [Array.Get] and [Array.Set] for an index outside the array. It
// formats its message only when printed, which keeps both small enough to inline.
type arrayIndexError[S geom.Integer] struct {
	index geom.Point[S]
	size  geom.Size[S]
}

// Error returns the message of the panic.
func (e arrayIndexError[S]) Error() string {
	return fmt.Sprintf("container: index %s outside array of size %s", e.index, e.size)
}
