package container_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"

	. "github.com/gravitton/container"
)

func TestDeque_Len(t *testing.T) {
	t.Run("the zero value has none", func(t *testing.T) {
		assert.Equal(t, (&Deque[int]{}).Len(), 0)
	})
	t.Run("agrees with the elements", func(t *testing.T) {
		for _, test := range testDeques() {
			assert.Equal(t, test.deque.Len(), len(test.elements), fmt.Sprint(test.elements, ": "))
		}
	})
}

func TestDeque_At(t *testing.T) {
	t.Run("is the element at the index from the front", func(t *testing.T) {
		for _, test := range testDeques() {
			for i, element := range test.elements {
				assert.Equal(t, test.deque.At(i), element, fmt.Sprint(test.elements, " at ", i, ": "))
			}
		}
	})
	t.Run("panics past the back", func(t *testing.T) {
		d, _ := testDeque(5)

		assert.PanicsWith(t, func() {
			d.At(5)
		}, "container: index 5 outside deque of 5 elements")
	})
	t.Run("panics before the front", func(t *testing.T) {
		d, _ := testDeque(5)

		assert.PanicsWith(t, func() {
			d.At(-1)
		}, "container: index -1 outside deque of 5 elements")
	})
	t.Run("panics for any index of the empty deque", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			(&Deque[int]{}).At(0)
		}, "container: index 0 outside deque of 0 elements")
	})
}

func BenchmarkDeque_At(b *testing.B) {
	d, _ := testDeque(17)
	for b.Loop() {
		sinkInt = d.At(9)
	}
}

func TestDeque_Front(t *testing.T) {
	t.Run("is the element pushed to the front last", func(t *testing.T) {
		d, _ := testDeque(4)
		element, found := d.Front()

		assert.True(t, found)
		assert.Equal(t, element, 3)
	})
	t.Run("the empty deque has none", func(t *testing.T) {
		_, found := (&Deque[int]{}).Front()

		assert.False(t, found)
	})
	t.Run("agrees with At", func(t *testing.T) {
		for _, test := range testDeques() {
			element, found := test.deque.Front()

			assert.Equal(t, found, !test.deque.IsEmpty(), fmt.Sprint(test.elements, ": "))
			if found {
				assert.Equal(t, element, test.deque.At(0), fmt.Sprint(test.elements, ": "))
			}
		}
	})
}

func TestDeque_Back(t *testing.T) {
	t.Run("is the element pushed to the back last", func(t *testing.T) {
		d, _ := testDeque(4)
		element, found := d.Back()

		assert.True(t, found)
		assert.Equal(t, element, 2)
	})
	t.Run("the empty deque has none", func(t *testing.T) {
		_, found := (&Deque[int]{}).Back()

		assert.False(t, found)
	})
	t.Run("agrees with At", func(t *testing.T) {
		for _, test := range testDeques() {
			element, found := test.deque.Back()

			assert.Equal(t, found, !test.deque.IsEmpty(), fmt.Sprint(test.elements, ": "))
			if found {
				assert.Equal(t, element, test.deque.At(test.deque.Len()-1), fmt.Sprint(test.elements, ": "))
			}
		}
	})
}

func TestDeque_PushFront(t *testing.T) {
	t.Run("puts each element before the others", func(t *testing.T) {
		d := &Deque[int]{}
		for i := range 9 {
			d.PushFront(i)
		}

		assert.Equal(t, dequeElements(d), []int{8, 7, 6, 5, 4, 3, 2, 1, 0})
	})
	t.Run("keeps the order across a growth of a wrapped ring", func(t *testing.T) {
		for _, test := range testDeques() {
			test.deque.PushFront(-1)

			assert.Equal(t, dequeElements(test.deque), append([]int{-1}, test.elements...), fmt.Sprint(test.elements, ": "))
		}
	})
	t.Run("allocates only where the ring is full", func(t *testing.T) {
		d := &Deque[int]{}
		d.PushFront(0)

		assert.Allocates(t, func() {
			d.PushFront(1)
			d.PopBack()
		}, 0)
		assert.AllocatesRuns(t, func() {
			for range d.Len() + 4 {
				d.PushFront(1)
			}
		}, 1, 1)
	})
	t.Run("with a PopBack moves every element back by one and keeps the length", func(t *testing.T) {
		for _, test := range testDeques()[1:] {
			test.deque.PushFront(-1)
			test.deque.PopBack()

			assert.Equal(t, dequeElements(test.deque), append([]int{-1}, test.elements[:len(test.elements)-1]...), fmt.Sprint(test.elements, ": "))
		}
	})
}

func BenchmarkDeque_PushFront(b *testing.B) {
	for _, length := range []int{16, 1024} {
		b.Run(fmt.Sprint("deque/", length), func(b *testing.B) {
			d, _ := testDeque(length)
			for b.Loop() {
				d.PushFront(1)
				sinkInt, _ = d.PopBack()
			}
		})
		b.Run(fmt.Sprint("slice/", length), func(b *testing.B) {
			_, s := testDeque(length)
			for b.Loop() {
				s = slices.Insert(s, 0, 1)
				sinkInt, s = s[len(s)-1], s[:len(s)-1]
			}
		})
	}
}

func TestDeque_PushBack(t *testing.T) {
	t.Run("puts each element after the others", func(t *testing.T) {
		d := &Deque[int]{}
		for i := range 9 {
			d.PushBack(i)
		}

		assert.Equal(t, dequeElements(d), []int{0, 1, 2, 3, 4, 5, 6, 7, 8})
	})
	t.Run("keeps the order across a growth of a wrapped ring", func(t *testing.T) {
		for _, test := range testDeques() {
			test.deque.PushBack(-1)

			assert.Equal(t, dequeElements(test.deque), append(slices.Clone(test.elements), -1), fmt.Sprint(test.elements, ": "))
		}
	})
	t.Run("with a PopFront for each element hands out a ring before what it pushes", func(t *testing.T) {
		d, _ := testDeque(5)
		var order []int
		for range 3 {
			for range d.Len() {
				element, _ := d.PopFront()
				order = append(order, element)
				d.PushBack(element + 10)
			}
		}

		assert.Equal(t, order, []int{3, 1, 0, 2, 4, 13, 11, 10, 12, 14, 23, 21, 20, 22, 24})
	})
	t.Run("allocates only where the ring is full", func(t *testing.T) {
		d := &Deque[int]{}
		d.PushBack(0)

		assert.Allocates(t, func() {
			d.PushBack(1)
			d.PopFront()
		}, 0)
		assert.AllocatesRuns(t, func() {
			for range d.Len() + 4 {
				d.PushBack(1)
			}
		}, 1, 1)
	})
}

func BenchmarkDeque_PushBack(b *testing.B) {
	const pushes = 1 << 12

	for _, width := range []int{16, 1024} {
		b.Run(fmt.Sprint("deque/", width), func(b *testing.B) {
			var d Deque[Tile]
			for b.Loop() {
				d.Clear()
				for i := range width {
					d.PushBack(Tile{i, i})
				}
				for pushed := width; !d.IsEmpty(); {
					tile, _ := d.PopFront()
					sinkInt += tile.X
					if pushed < pushes {
						d.PushBack(Tile{tile.Y, tile.X})
						pushed++
					}
				}
			}
		})
		b.Run(fmt.Sprint("slice/", width), func(b *testing.B) {
			var queue []Tile
			for b.Loop() {
				queue = queue[:0]
				for i := range width {
					queue = append(queue, Tile{i, i})
				}
				for head := 0; head < len(queue); head++ {
					tile := queue[head]
					sinkInt += tile.X
					if len(queue) < pushes {
						queue = append(queue, Tile{tile.Y, tile.X})
					}
				}
			}
		})
	}
}

func FuzzDeque_PushBack(f *testing.F) {
	f.Add([]byte{0, 1, 0, 2, 1, 3, 2, 3, 1, 4, 0, 5, 5, 3})
	f.Add([]byte{3, 2, 4, 0, 0, 1, 1})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, operations []byte) {
		d := &Deque[int]{}
		var elements []int
		for i, operation := range operations {
			switch operation % 6 {
			case 0:
				d.PushFront(i)
				elements = slices.Insert(elements, 0, i)
			case 1:
				d.PushBack(i)
				elements = append(elements, i)
			case 2:
				element, found := d.PopFront()

				assert.Equal(t, found, len(elements) > 0, fmt.Sprint(i, ": "))
				if found {
					assert.Equal(t, element, elements[0], fmt.Sprint(i, ": "))
					elements = elements[1:]
				}
			case 3:
				element, found := d.PopBack()

				assert.Equal(t, found, len(elements) > 0, fmt.Sprint(i, ": "))
				if found {
					assert.Equal(t, element, elements[len(elements)-1], fmt.Sprint(i, ": "))
					elements = elements[:len(elements)-1]
				}
			case 4:
				d.Clear()
				elements = nil
			case 5:
				for j, element := range elements {
					assert.Equal(t, d.At(j), element, fmt.Sprint(i, " at ", j, ": "))
				}

				front, found := d.Front()
				assert.Equal(t, found, len(elements) > 0, fmt.Sprint(i, ": "))
				if found {
					back, _ := d.Back()
					assert.Equal(t, front, elements[0], fmt.Sprint(i, ": "))
					assert.Equal(t, back, elements[len(elements)-1], fmt.Sprint(i, ": "))
				}
			}

			assert.Equal(t, d.Len(), len(elements), fmt.Sprint(i, ": "))
		}

		assert.Equal(t, dequeElements(d), append([]int{}, elements...))
	})
}

func TestDeque_PopFront(t *testing.T) {
	t.Run("removes the elements from the front, in order", func(t *testing.T) {
		for _, test := range testDeques() {
			for _, want := range test.elements {
				element, found := test.deque.PopFront()

				assert.True(t, found, fmt.Sprint(test.elements, ": "))
				assert.Equal(t, element, want, fmt.Sprint(test.elements, ": "))
			}

			assert.True(t, test.deque.IsEmpty(), fmt.Sprint(test.elements, ": "))
		}
	})
	t.Run("the empty deque has none", func(t *testing.T) {
		d := &Deque[int]{}
		_, found := d.PopFront()

		assert.False(t, found)
		assert.Equal(t, d.Len(), 0)
	})
	t.Run("undoes PushFront", func(t *testing.T) {
		for _, test := range testDeques() {
			test.deque.PushFront(-1)
			element, _ := test.deque.PopFront()

			assert.Equal(t, element, -1, fmt.Sprint(test.elements, ": "))
			assert.Equal(t, dequeElements(test.deque), test.elements, fmt.Sprint(test.elements, ": "))
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		d, _ := testDeque(17)

		assert.Allocates(t, func() {
			sinkInt, _ = d.PopFront()
		}, 0)
	})
}

func TestDeque_PopBack(t *testing.T) {
	t.Run("removes the elements from the back, in reverse order", func(t *testing.T) {
		for _, test := range testDeques() {
			for _, want := range slices.Backward(test.elements) {
				element, found := test.deque.PopBack()

				assert.True(t, found, fmt.Sprint(test.elements, ": "))
				assert.Equal(t, element, want, fmt.Sprint(test.elements, ": "))
			}

			assert.True(t, test.deque.IsEmpty(), fmt.Sprint(test.elements, ": "))
		}
	})
	t.Run("the empty deque has none", func(t *testing.T) {
		d := &Deque[int]{}
		_, found := d.PopBack()

		assert.False(t, found)
		assert.Equal(t, d.Len(), 0)
	})
	t.Run("undoes PushBack", func(t *testing.T) {
		for _, test := range testDeques() {
			test.deque.PushBack(-1)
			element, _ := test.deque.PopBack()

			assert.Equal(t, element, -1, fmt.Sprint(test.elements, ": "))
			assert.Equal(t, dequeElements(test.deque), test.elements, fmt.Sprint(test.elements, ": "))
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		d, _ := testDeque(17)

		assert.Allocates(t, func() {
			sinkInt, _ = d.PopBack()
		}, 0)
	})
}

func TestDeque_Clear(t *testing.T) {
	t.Run("removes every element", func(t *testing.T) {
		for _, test := range testDeques() {
			test.deque.Clear()

			assert.True(t, test.deque.IsEmpty(), fmt.Sprint(test.elements, ": "))
			assert.Equal(t, dequeElements(test.deque), []int{}, fmt.Sprint(test.elements, ": "))
		}
	})
	t.Run("keeps the ring for the elements pushed next", func(t *testing.T) {
		d, elements := testDeque(17)

		assert.Allocates(t, func() {
			d.Clear()
			for _, element := range elements {
				d.PushBack(element)
			}
		}, 0)
		assert.Equal(t, dequeElements(d), elements)
	})
}

func TestDeque_All(t *testing.T) {
	t.Run("yields the index and the element from the front to the back", func(t *testing.T) {
		for _, test := range testDeques() {
			i := 0
			for index, element := range test.deque.All() {
				assert.Equal(t, index, i, fmt.Sprint(test.elements, ": "))
				assert.Equal(t, element, test.elements[i], fmt.Sprint(test.elements, ": "))
				i++
			}

			assert.Equal(t, i, len(test.elements), fmt.Sprint(test.elements, ": "))
		}
	})
	t.Run("stops at the break", func(t *testing.T) {
		d, _ := testDeque(5)
		count := 0
		for range d.All() {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
	t.Run("allocates nothing", func(t *testing.T) {
		d, _ := testDeque(17)

		assert.Allocates(t, func() {
			for _, element := range d.All() {
				sinkInt += element
			}
		}, 0)
	})
}

func BenchmarkDeque_All(b *testing.B) {
	d, _ := testDeque(1024)
	for b.Loop() {
		for _, element := range d.All() {
			sinkInt += element
		}
	}
}

func TestDeque_IsEmpty(t *testing.T) {
	t.Run("the zero value and a deque popped empty are", func(t *testing.T) {
		d, _ := testDeque(1)
		d.PopBack()

		assert.True(t, (&Deque[int]{}).IsEmpty())
		assert.True(t, d.IsEmpty())
	})
	t.Run("a deque with an element is not", func(t *testing.T) {
		d, _ := testDeque(1)

		assert.False(t, d.IsEmpty())
	})
	t.Run("agrees with Len", func(t *testing.T) {
		for _, test := range testDeques() {
			assert.Equal(t, test.deque.IsEmpty(), test.deque.Len() == 0, fmt.Sprint(test.elements, ": "))
		}
	})
}

func ExampleDeque() {
	var body Deque[geom.Point[int]]
	body.PushBack(geom.Pt(2, 1))
	body.PushBack(geom.Pt(1, 1))

	body.PushFront(geom.Pt(3, 1))
	tail, _ := body.PopBack()
	head, _ := body.Front()

	fmt.Println(head, tail, body.At(1), body.Len())
	for i, cell := range body.All() {
		fmt.Println(i, cell)
	}
	// Output:
	// (3,1) (1,1) (2,1) 2
	// 0 (3,1)
	// 1 (2,1)
}

// dequeElements returns the elements of the deque in order from its front, collected by All.
func dequeElements(d *Deque[int]) []int {
	elements := []int{}
	for _, element := range d.All() {
		elements = append(elements, element)
	}

	return elements
}

// dequeCase is a deque with its elements in order from the front, as a slice holds them.
type dequeCase struct {
	deque    *Deque[int]
	elements []int
}

// testDeques returns the deques the properties of a deque are checked on: empty, wrapped, full
// and grown.
func testDeques() []dequeCase {
	var cases []dequeCase
	for _, n := range []int{0, 1, 2, 4, 5, 17} {
		d, elements := testDeque(n)
		cases = append(cases, dequeCase{d, elements})
	}

	return cases
}

// testDeque returns a deque of the elements, pushed to its front and its back in turn so its
// ring wraps, and the elements in order from its front.
func testDeque(n int) (*Deque[int], []int) {
	d := &Deque[int]{}
	elements := []int{}
	for i := range n {
		if i%2 == 0 {
			d.PushBack(i)
			elements = append(elements, i)

			continue
		}

		d.PushFront(i)
		elements = slices.Insert(elements, 0, i)
	}

	return d, elements
}
