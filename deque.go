package container

import (
	"fmt"
	"iter"
	"sync"
)

// minRing is the length of the first ring a push allocates, a power of two as every ring is.
const minRing = 4

// Deque is a double-ended queue of elements of type E, held in order from its front to its back
// on a ring that grows by doubling. Unlike the other containers it changes in place and is used
// through a pointer: a copy shares the ring but not its ends, so a deque is never copied once
// pushed to, and go vet reports a copy. The zero value is the empty deque.
type Deque[E any] struct {
	_     [0]sync.Mutex
	ring  []E
	front int
	len   int
}

// Len returns the number of elements in the deque.
func (d *Deque[E]) Len() int {
	return d.len
}

// At returns the element at the index, counted from the front of the deque. It panics for an
// index outside the deque.
func (d *Deque[E]) At(i int) E {
	if i < 0 || i >= d.len {
		panic(dequeIndexError{i, d.len})
	}

	return d.ring[d.slot(i)]
}

// Front returns the element at the front of the deque and true, or false for the empty deque.
func (d *Deque[E]) Front() (E, bool) {
	if d.IsEmpty() {
		var zero E

		return zero, false
	}

	return d.ring[d.front], true
}

// Back returns the element at the back of the deque and true, or false for the empty deque.
func (d *Deque[E]) Back() (E, bool) {
	if d.IsEmpty() {
		var zero E

		return zero, false
	}

	return d.ring[d.slot(d.len-1)], true
}

// slot returns the position in the ring of the element at the index from the front.
func (d *Deque[E]) slot(i int) int {
	return (d.front + i) & (len(d.ring) - 1)
}

// PushFront puts the element at the front of the deque. It allocates only where the ring is
// full, a ring of twice the length.
func (d *Deque[E]) PushFront(element E) {
	if d.len == len(d.ring) {
		d.grow()
	}

	d.front = (d.front - 1) & (len(d.ring) - 1)
	d.ring[d.front] = element
	d.len++
}

// PushBack puts the element at the back of the deque. It allocates only where the ring is full,
// a ring of twice the length.
func (d *Deque[E]) PushBack(element E) {
	if d.len == len(d.ring) {
		d.grow()
	}

	d.ring[(d.front+d.len)&(len(d.ring)-1)] = element
	d.len++
}

// PopFront removes the element at the front of the deque and returns it and true, or false for
// the empty deque.
func (d *Deque[E]) PopFront() (E, bool) {
	var zero E
	if d.len == 0 {
		return zero, false
	}

	element := d.ring[d.front]
	d.ring[d.front] = zero
	d.front = d.slot(1)
	d.len--

	return element, true
}

// PopBack removes the element at the back of the deque and returns it and true, or false for
// the empty deque.
func (d *Deque[E]) PopBack() (E, bool) {
	var zero E
	if d.len == 0 {
		return zero, false
	}

	d.len--
	back := d.slot(d.len)
	element := d.ring[back]
	d.ring[back] = zero

	return element, true
}

// Clear removes every element of the deque, keeping its ring for the elements pushed next.
func (d *Deque[E]) Clear() {
	end := min(d.front+d.len, len(d.ring))
	clear(d.ring[d.front:end])
	clear(d.ring[:d.front+d.len-end])
	d.front, d.len = 0, 0
}

// grow moves the elements into a ring of twice the length, its front first.
func (d *Deque[E]) grow() {
	ring := make([]E, max(minRing, 2*len(d.ring)))
	moved := copy(ring, d.ring[d.front:])
	copy(ring[moved:], d.ring[:d.front])
	d.ring, d.front = ring, 0
}

// All returns an iterator over the index and the element of every element of the deque, from its
// front to its back. It reads the deque as it goes, so a push or a pop while it runs changes what
// it yields.
func (d *Deque[E]) All() iter.Seq2[int, E] {
	return func(yield func(int, E) bool) {
		for i := range d.len {
			if !yield(i, d.ring[d.slot(i)]) {
				return
			}
		}
	}
}

// IsEmpty reports whether the deque holds no element.
func (d *Deque[E]) IsEmpty() bool {
	return d.len == 0
}

// dequeIndexError is the panic of [Deque.At] for an index outside the deque. It formats its
// message only when printed, which keeps At small enough to inline.
type dequeIndexError struct {
	index int
	len   int
}

// Error returns the message of the panic.
func (e dequeIndexError) Error() string {
	return fmt.Sprintf("container: index %d outside deque of %d elements", e.index, e.len)
}
