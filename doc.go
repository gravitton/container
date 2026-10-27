// Package container provides generic containers, each used as a value, that know no layout. It
// builds on github.com/gravitton/geometry for the points and sizes an array speaks.
//
// # Types
//
//   - [Array] is a table of values, one for every index of a size in columns and rows, stored
//     row by row. It is a handle: a copy shares its values.
//   - [Set] is an immutable set held in the order of its elements' Compare method, which alone
//     decides that two elements are the same. [SetOf] copies the elements it is given; [SetFrom]
//     takes the slice as its own.
//
// # Rules
//
// A [Set] never changes: every operation returns a set, and one that would equal a side is that
// side. The one way to change a set is a write to the slice given to [SetFrom].
//
// Compare must be a total order, and the same each time it is asked; on any other the answers of
// a set are undefined.
//
// The zero value of every container is the empty one, ready to use.
//
// # Allocations
//
// A slice or set result is allocated once, at its final size, and each slice result has an
// Append form that allocates nothing once the caller's slice has room. [SetOf] allocates once,
// for its copy of the elements, and [SetFrom] nothing. A read, a write and a test allocate
// nothing.
//
// # Panics
//
// A degenerate value gets an answer; an argument with no meaning gets a panic, whose value is an
// error:
//
//   - [Arr] for a negative width or height, or a size of more values than an int counts.
//   - [Array.Get] and [Array.Set] for an index outside the array, [Array.Row] for a row outside
//     it, and [Set.At] for an index outside the set.
package container
