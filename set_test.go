package container_test

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"

	. "github.com/gravitton/container"
)

// Tile is an element ordered by x and then by y.
type Tile struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (t Tile) Compare(other Tile) int {
	if order := cmp.Compare(t.X, other.X); order != 0 {
		return order
	}

	return cmp.Compare(t.Y, other.Y)
}

func (t Tile) String() string {
	return fmt.Sprintf("(%d,%d)", t.X, t.Y)
}

var (
	testSet      = SetOf(square(Tile{0, 0}, 2)...)
	testSetOther = SetOf(square(Tile{3, -1}, 2)...)
	testSetEmpty = Set[Tile]{}

	sinkSet   Set[Tile]
	sinkTiles []Tile
)

// square returns the tiles within n columns and rows of the center, ordered by x and then by y.
func square(center Tile, n int) []Tile {
	tiles := make([]Tile, 0, (2*n+1)*(2*n+1))
	for dx := -n; dx <= n; dx++ {
		for dy := -n; dy <= n; dy++ {
			tiles = append(tiles, Tile{center.X + dx, center.Y + dy})
		}
	}

	return tiles
}

// referenceSet returns the elements a map-based set keeps of the two sets, sorted.
func referenceSet(a, b Set[Tile], keep func(inA, inB bool) bool) []Tile {
	inA, inB := referenceMap(a), referenceMap(b)

	var kept []Tile
	for element := range inA {
		if keep(true, inB[element]) {
			kept = append(kept, element)
		}
	}
	for element := range inB {
		if !inA[element] && keep(false, true) {
			kept = append(kept, element)
		}
	}
	slices.SortFunc(kept, Tile.Compare)

	return kept
}

// referenceMap returns the elements of the set as the keys of a map.
func referenceMap(s Set[Tile]) map[Tile]bool {
	elements := map[Tile]bool{}
	for element := range s.All() {
		elements[element] = true
	}

	return elements
}

// testSets returns the sets the properties of a single set are checked on.
func testSets() []Set[Tile] {
	return []Set[Tile]{testSet, testSetOther, testSetEmpty, SetOf(Tile{9, 9}), SetOf(square(Tile{0, 0}, 1)...)}
}

// testSetPairs returns the pairs of sets every operation is checked on.
func testSetPairs() [][2]Set[Tile] {
	return [][2]Set[Tile]{
		{testSet, testSetOther},
		{testSetOther, testSet},
		{testSet, testSet},
		{testSet, testSetEmpty},
		{testSetEmpty, testSet},
		{testSetEmpty, testSetEmpty},
		{testSet, SetOf(Tile{9, 9})},
		{SetOf(Tile{9, 9}), testSet},
		{testSet, SetOf(square(Tile{0, 0}, 1)...)},
		{SetOf(square(Tile{0, 0}, 1)...), testSet},
	}
}

// testSetOperation runs the sections every set operation shares, checking it against the
// elements a map keeps of the two sets.
func testSetOperation(t *testing.T, operation func(a, b Set[Tile]) Set[Tile], keep func(inA, inB bool) bool) {
	t.Run("keeps what a map keeps", func(t *testing.T) {
		for _, pair := range testSetPairs() {
			assert.Equal(t, slices.Collect(operation(pair[0], pair[1]).All()), referenceSet(pair[0], pair[1], keep), fmt.Sprint(pair[0].Elements(), " and ", pair[1].Elements(), ": "))
		}
	})
	t.Run("leaves both sets alone", func(t *testing.T) {
		before, beforeOther := testSet.Elements(), testSetOther.Elements()
		sinkSet = operation(testSet, testSetOther)

		assert.Equal(t, testSet.Elements(), before)
		assert.Equal(t, testSetOther.Elements(), beforeOther)
	})
	t.Run("allocates once, and nothing where the result is empty or equals a side", func(t *testing.T) {
		for _, pair := range testSetPairs() {
			allocs := 1.0
			if result := operation(pair[0], pair[1]); result.IsEmpty() || result.Equal(pair[0]) || result.Equal(pair[1]) {
				allocs = 0
			}

			assert.Equal(t, testing.AllocsPerRun(10, func() {
				sinkSet = operation(pair[0], pair[1])
			}), allocs, fmt.Sprint(pair[0].Elements(), " and ", pair[1].Elements(), ": "))
		}
	})
}

// fuzzSetOperation checks a set operation on two sets of fuzzed tiles against the elements a
// map keeps of them.
func fuzzSetOperation(f *testing.F, operation func(a, b Set[Tile]) Set[Tile], keep func(inA, inB bool) bool) {
	f.Add([]byte{1, 2, 3, 4, 5, 6}, []byte{3, 4, 5, 6, 7, 0})
	f.Add([]byte{}, []byte{0, 0})
	f.Add([]byte{7, 7, 7, 7}, []byte{})

	f.Fuzz(func(t *testing.T, a, b []byte) {
		first, second := fuzzSet(a), fuzzSet(b)

		assert.Equal(t, slices.Collect(operation(first, second).All()), referenceSet(first, second, keep), fmt.Sprint(first.Elements(), " and ", second.Elements(), ": "))
	})
}

// fuzzSetRelation checks a relation of two sets of fuzzed tiles against a map: it holds where
// the map finds an element to keep, or where it finds none when found is false.
func fuzzSetRelation(f *testing.F, relation func(a, b Set[Tile]) bool, keep func(inA, inB bool) bool, found bool) {
	f.Add([]byte{1, 2, 3, 4, 5, 6}, []byte{3, 4, 5, 6, 7, 0})
	f.Add([]byte{}, []byte{0, 0})
	f.Add([]byte{7, 7, 7, 7}, []byte{7, 7})

	f.Fuzz(func(t *testing.T, a, b []byte) {
		first, second := fuzzSet(a), fuzzSet(b)

		assert.Equal(t, relation(first, second), (len(referenceSet(first, second, keep)) > 0) == found, fmt.Sprint(first.Elements(), " and ", second.Elements(), ": "))
	})
}

// fuzzSet returns the set of the tiles the bytes give in pairs, within four of the origin, so
// two fuzzed sets often share some.
func fuzzSet(data []byte) Set[Tile] {
	tiles := make([]Tile, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		tiles = append(tiles, Tile{int(data[i]%8) - 4, int(data[i+1]%8) - 4})
	}

	return SetFrom(tiles)
}

func TestSetOf(t *testing.T) {
	t.Run("orders the elements and drops repeats", func(t *testing.T) {
		assert.Equal(t, SetOf(Tile{1, 0}, Tile{0, 1}, Tile{1, 0}, Tile{0, 0}).Elements(), []Tile{{0, 0}, {0, 1}, {1, 0}})
	})
	t.Run("does not change with the caller's slice", func(t *testing.T) {
		tiles := []Tile{{1, 0}, {0, 0}}
		s := SetOf(tiles...)
		tiles[0] = Tile{9, 9}

		assert.Equal(t, tiles, []Tile{{9, 9}, {0, 0}})
		assert.Equal(t, s.Elements(), []Tile{{0, 0}, {1, 0}})
	})
	t.Run("no elements is the empty set", func(t *testing.T) {
		assert.True(t, SetOf[Tile]().Equal(testSetEmpty))
		assert.True(t, SetOf([]Tile{}...).IsEmpty())
	})
	t.Run("allocates once", func(t *testing.T) {
		tiles := square(Tile{0, 0}, 2)

		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkSet = SetOf(tiles...)
		}), 1.0)
	})
}

func ExampleSetOf() {
	a := SetOf(geom.Pt(1, 1), geom.Pt(2, 1), geom.Pt(1, 2), geom.Pt(2, 2))
	points := []geom.Point[int]{geom.Pt(2, 2), geom.Pt(1, 1), geom.Pt(2, 1), geom.Pt(1, 2), geom.Pt(1, 1)}
	from := SetFrom(points)

	first, ok := a.First()
	index, found := a.Index(geom.Pt(2, 1))

	fmt.Println(from.Equal(a), a.Contains(geom.Pt(2, 1)), a.Len())
	fmt.Println(first, ok, index, found)
	fmt.Println(slices.Collect(a.All()), slices.Collect(a.Backward()))
	fmt.Println(a.Elements())

	data, _ := json.Marshal(a)
	fmt.Println(string(data))
	// Output:
	// true true 4
	// (1,1) true 2 true
	// [(1,1) (1,2) (2,1) (2,2)] [(2,2) (2,1) (1,2) (1,1)]
	// [(1,1) (1,2) (2,1) (2,2)]
	// [{"x":1,"y":1},{"x":1,"y":2},{"x":2,"y":1},{"x":2,"y":2}]
}

func TestSetFrom(t *testing.T) {
	t.Run("agrees with SetOf", func(t *testing.T) {
		tiles := []Tile{{1, 0}, {0, 1}, {1, 0}, {0, 0}}

		assert.True(t, SetFrom(slices.Clone(tiles)).Equal(SetOf(tiles...)))
	})
	t.Run("rearranges the slice", func(t *testing.T) {
		tiles := []Tile{{1, 0}, {0, 0}}
		sinkSet = SetFrom(tiles)

		assert.Equal(t, tiles, []Tile{{0, 0}, {1, 0}})
	})
	t.Run("a later write to the slice changes the set", func(t *testing.T) {
		tiles := []Tile{{1, 0}, {0, 0}}
		s := SetFrom(tiles)
		tiles[1] = Tile{2, 0}

		assert.Equal(t, s.Elements(), []Tile{{0, 0}, {2, 0}})
	})
	t.Run("no elements is the empty set", func(t *testing.T) {
		assert.True(t, SetFrom([]Tile{}).Equal(testSetEmpty))
		assert.True(t, SetFrom[Tile](nil).IsEmpty())
	})
	t.Run("allocates nothing", func(t *testing.T) {
		tiles := square(Tile{0, 0}, 2)

		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkSet = SetFrom(tiles)
		}), 0.0)
	})
}

func TestSet_Len(t *testing.T) {
	t.Run("counts each element once", func(t *testing.T) {
		assert.Equal(t, testSet.Len(), 25)
		assert.Equal(t, SetOf(Tile{1, 1}, Tile{1, 1}).Len(), 1)
	})
	t.Run("the empty set has none", func(t *testing.T) {
		assert.Equal(t, testSetEmpty.Len(), 0)
	})
	t.Run("agrees with the elements", func(t *testing.T) {
		for _, s := range testSets() {
			assert.Length(t, s.Elements(), s.Len(), fmt.Sprint(s.Elements(), ": "))
		}
	})
}

func TestSet_At(t *testing.T) {
	t.Run("is the element at the index All yields it", func(t *testing.T) {
		i := 0
		for element := range testSet.All() {
			assert.Equal(t, testSet.At(i), element)
			i++
		}
	})
	t.Run("panics past the last element", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			testSet.At(testSet.Len())
		}, "container: index 25 outside set of 25 elements")
	})
	t.Run("panics before the first element", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			testSet.At(-1)
		}, "container: index -1 outside set of 25 elements")
	})
	t.Run("panics for any index of the empty set", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			testSetEmpty.At(0)
		}, "container: index 0 outside set of 0 elements")
	})
}

func TestSet_Index(t *testing.T) {
	t.Run("finds every element at its index", func(t *testing.T) {
		for i := range testSet.Len() {
			index, found := testSet.Index(testSet.At(i))

			assert.True(t, found)
			assert.Equal(t, index, i)
		}
	})
	t.Run("an element that is not there has none", func(t *testing.T) {
		for _, element := range []Tile{{9, 9}, {-9, -9}} {
			_, found := testSet.Index(element)

			assert.False(t, found, element.String()+": ")
		}
	})
	t.Run("the empty set holds no index", func(t *testing.T) {
		_, found := testSetEmpty.Index(Tile{0, 0})

		assert.False(t, found)
	})
}

func TestSet_First(t *testing.T) {
	t.Run("is the first element", func(t *testing.T) {
		element, found := testSet.First()

		assert.True(t, found)
		assert.Equal(t, element, Tile{-2, -2})
	})
	t.Run("the empty set has none", func(t *testing.T) {
		_, found := testSetEmpty.First()

		assert.False(t, found)
	})
	t.Run("agrees with At", func(t *testing.T) {
		for _, s := range testSets() {
			element, found := s.First()

			assert.Equal(t, found, !s.IsEmpty(), fmt.Sprint(s.Elements(), ": "))
			if found {
				assert.Equal(t, element, s.At(0), fmt.Sprint(s.Elements(), ": "))
			}
		}
	})
}

func TestSet_Last(t *testing.T) {
	t.Run("is the last element", func(t *testing.T) {
		element, found := testSet.Last()

		assert.True(t, found)
		assert.Equal(t, element, Tile{2, 2})
	})
	t.Run("the empty set has none", func(t *testing.T) {
		_, found := testSetEmpty.Last()

		assert.False(t, found)
	})
	t.Run("agrees with At", func(t *testing.T) {
		for _, s := range testSets() {
			element, found := s.Last()

			assert.Equal(t, found, !s.IsEmpty(), fmt.Sprint(s.Elements(), ": "))
			if found {
				assert.Equal(t, element, s.At(s.Len()-1), fmt.Sprint(s.Elements(), ": "))
			}
		}
	})
}

func TestSet_Elements(t *testing.T) {
	t.Run("is the elements All yields, in order", func(t *testing.T) {
		assert.Equal(t, testSet.Elements(), slices.Collect(testSet.All()))
	})
	t.Run("is a fresh slice", func(t *testing.T) {
		elements := testSet.Elements()
		elements[0] = Tile{9, 9}

		assert.False(t, testSet.Contains(Tile{9, 9}))
	})
	t.Run("the empty set has none", func(t *testing.T) {
		assert.Equal(t, testSetEmpty.Elements(), nil)
	})
	t.Run("allocates once at its length", func(t *testing.T) {
		elements := testSet.Elements()

		assert.Equal(t, cap(elements), len(elements))
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkTiles = testSet.Elements()
		}), 1.0)
	})
}

func TestSet_AppendElements(t *testing.T) {
	t.Run("appends what Elements returns", func(t *testing.T) {
		assert.Equal(t, testSet.AppendElements([]Tile{{9, 9}}), append([]Tile{{9, 9}}, testSet.Elements()...))
	})
	t.Run("the empty set appends nothing", func(t *testing.T) {
		assert.Equal(t, testSetEmpty.AppendElements([]Tile{{9, 9}}), []Tile{{9, 9}})
	})
	t.Run("allocates nothing with room", func(t *testing.T) {
		buffer := make([]Tile, 0, testSet.Len())

		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkTiles = testSet.AppendElements(buffer[:0])
		}), 0.0)
	})
}

func TestSet_All(t *testing.T) {
	t.Run("yields the elements in order", func(t *testing.T) {
		assert.True(t, slices.IsSortedFunc(slices.Collect(testSet.All()), Tile.Compare))
		assert.Length(t, slices.Collect(testSet.All()), testSet.Len())
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range testSet.All() {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
	t.Run("the empty set yields nothing", func(t *testing.T) {
		assert.Length(t, slices.Collect(testSetEmpty.All()), 0)
	})
}

func TestSet_Backward(t *testing.T) {
	t.Run("yields the elements in reverse order", func(t *testing.T) {
		elements := slices.Collect(testSet.All())
		slices.Reverse(elements)

		assert.Equal(t, slices.Collect(testSet.Backward()), elements)
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range testSet.Backward() {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
	t.Run("the empty set yields nothing", func(t *testing.T) {
		assert.Length(t, slices.Collect(testSetEmpty.Backward()), 0)
	})
}

func TestSet_Union(t *testing.T) {
	testSetOperation(t, Set[Tile].Union, func(inA, inB bool) bool {
		return inA || inB
	})
}

func ExampleSet_Union() {
	a := SetOf(geom.Pt(1, 1), geom.Pt(2, 1), geom.Pt(1, 2), geom.Pt(2, 2))
	b := SetOf(geom.Pt(2, 2), geom.Pt(3, 2), geom.Pt(2, 3), geom.Pt(3, 3))

	fmt.Println(a.Union(b).Elements())
	fmt.Println(a.Intersection(b).Elements())
	fmt.Println(a.Difference(b).Elements())
	fmt.Println(a.SymmetricDifference(b).Elements())
	fmt.Println(a.Intersects(b), a.Encloses(b), a.Equal(b))
	// Output:
	// [(1,1) (1,2) (2,1) (2,2) (2,3) (3,2) (3,3)]
	// [(2,2)]
	// [(1,1) (1,2) (2,1)]
	// [(1,1) (1,2) (2,1) (2,3) (3,2) (3,3)]
	// true false false
}

func BenchmarkSet_Union(b *testing.B) {
	for b.Loop() {
		sinkSet = testSet.Union(testSetOther)
	}
}

func FuzzSet_Union(f *testing.F) {
	fuzzSetOperation(f, Set[Tile].Union, func(inA, inB bool) bool {
		return inA || inB
	})
}

func TestSet_Intersection(t *testing.T) {
	testSetOperation(t, Set[Tile].Intersection, func(inA, inB bool) bool {
		return inA && inB
	})
}

func BenchmarkSet_Intersection(b *testing.B) {
	for b.Loop() {
		sinkSet = testSet.Intersection(testSetOther)
	}
}

func FuzzSet_Intersection(f *testing.F) {
	fuzzSetOperation(f, Set[Tile].Intersection, func(inA, inB bool) bool {
		return inA && inB
	})
}

func TestSet_Difference(t *testing.T) {
	testSetOperation(t, Set[Tile].Difference, func(inA, inB bool) bool {
		return inA && !inB
	})
}

func FuzzSet_Difference(f *testing.F) {
	fuzzSetOperation(f, Set[Tile].Difference, func(inA, inB bool) bool {
		return inA && !inB
	})
}

func TestSet_SymmetricDifference(t *testing.T) {
	testSetOperation(t, Set[Tile].SymmetricDifference, func(inA, inB bool) bool {
		return inA != inB
	})
}

func FuzzSet_SymmetricDifference(f *testing.F) {
	fuzzSetOperation(f, Set[Tile].SymmetricDifference, func(inA, inB bool) bool {
		return inA != inB
	})
}

func TestSet_Contains(t *testing.T) {
	t.Run("holds its elements", func(t *testing.T) {
		for element := range testSet.All() {
			assert.True(t, testSet.Contains(element), element.String()+": ")
		}
	})
	t.Run("holds nothing else", func(t *testing.T) {
		assert.False(t, testSet.Contains(Tile{3, 0}))
		assert.False(t, testSetEmpty.Contains(Tile{0, 0}))
	})
	t.Run("agrees with Index", func(t *testing.T) {
		for _, pair := range testSetPairs() {
			for element := range pair[1].All() {
				_, found := pair[0].Index(element)

				assert.Equal(t, pair[0].Contains(element), found, fmt.Sprint(pair[0].Elements(), " at ", element, ": "))
			}
		}
	})
}

func BenchmarkSet_Contains(b *testing.B) {
	for b.Loop() {
		sinkBool = testSet.Contains(Tile{1, 1})
	}
}

func TestSet_Encloses(t *testing.T) {
	t.Run("a set encloses a set inside it", func(t *testing.T) {
		assert.True(t, testSet.Encloses(SetOf(square(Tile{0, 0}, 1)...)))
		assert.True(t, testSet.Encloses(testSet))
	})
	t.Run("a set does not enclose one reaching past it", func(t *testing.T) {
		assert.False(t, testSet.Encloses(testSetOther))
		assert.False(t, SetOf(square(Tile{0, 0}, 1)...).Encloses(testSet))
	})
	t.Run("every set encloses the empty one, and the empty one no other", func(t *testing.T) {
		assert.True(t, testSet.Encloses(testSetEmpty))
		assert.True(t, testSetEmpty.Encloses(testSetEmpty))
		assert.False(t, testSetEmpty.Encloses(testSet))
	})
	t.Run("agrees with the difference being empty", func(t *testing.T) {
		for _, pair := range testSetPairs() {
			assert.Equal(t, pair[0].Encloses(pair[1]), pair[1].Difference(pair[0]).IsEmpty(), fmt.Sprint(pair[0].Elements(), " and ", pair[1].Elements(), ": "))
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkBool = testSet.Encloses(testSetOther)
		}), 0.0)
	})
}

func FuzzSet_Encloses(f *testing.F) {
	fuzzSetRelation(f, Set[Tile].Encloses, func(inA, inB bool) bool {
		return !inA && inB
	}, false)
}

func TestSet_Intersects(t *testing.T) {
	t.Run("sets sharing an element intersect", func(t *testing.T) {
		assert.True(t, testSet.Intersects(testSetOther))
		assert.True(t, testSet.Intersects(SetOf(Tile{2, 2}, Tile{9, 9})))
	})
	t.Run("sets sharing none do not", func(t *testing.T) {
		assert.False(t, testSet.Intersects(SetOf(Tile{9, 9})))
	})
	t.Run("the empty set intersects no set", func(t *testing.T) {
		assert.False(t, testSet.Intersects(testSetEmpty))
		assert.False(t, testSetEmpty.Intersects(testSetEmpty))
	})
	t.Run("agrees with the intersection being empty", func(t *testing.T) {
		for _, pair := range testSetPairs() {
			assert.Equal(t, pair[0].Intersects(pair[1]), !pair[0].Intersection(pair[1]).IsEmpty(), fmt.Sprint(pair[0].Elements(), " and ", pair[1].Elements(), ": "))
		}
	})
	t.Run("allocates nothing", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkBool = testSet.Intersects(testSetOther)
		}), 0.0)
	})
}

func FuzzSet_Intersects(f *testing.F) {
	fuzzSetRelation(f, Set[Tile].Intersects, func(inA, inB bool) bool {
		return inA && inB
	}, true)
}

func TestSet_Equal(t *testing.T) {
	t.Run("the same elements in any order given", func(t *testing.T) {
		elements := square(Tile{0, 0}, 2)
		slices.Reverse(elements)

		assert.True(t, testSet.Equal(SetOf(elements...)))
	})
	t.Run("different elements", func(t *testing.T) {
		assert.False(t, testSet.Equal(testSetOther))
		assert.False(t, testSet.Equal(testSet.Difference(SetOf(Tile{0, 0}))))
	})
	t.Run("the empty set equals the zero value", func(t *testing.T) {
		assert.True(t, testSet.Difference(testSet).Equal(testSetEmpty))
	})
	t.Run("agrees with enclosing both ways", func(t *testing.T) {
		for _, pair := range testSetPairs() {
			assert.Equal(t, pair[0].Equal(pair[1]), pair[0].Encloses(pair[1]) && pair[1].Encloses(pair[0]), fmt.Sprint(pair[0].Elements(), " and ", pair[1].Elements(), ": "))
		}
	})
}

func TestSet_IsEmpty(t *testing.T) {
	t.Run("the zero value and a set of no elements are", func(t *testing.T) {
		assert.True(t, testSetEmpty.IsEmpty())
		assert.True(t, testSet.Difference(testSet).IsEmpty())
	})
	t.Run("a set with an element is not", func(t *testing.T) {
		assert.False(t, SetOf(Tile{0, 0}).IsEmpty())
	})
	t.Run("agrees with Len", func(t *testing.T) {
		for _, s := range testSets() {
			assert.Equal(t, s.IsEmpty(), s.Len() == 0, fmt.Sprint(s.Elements(), ": "))
		}
	})
}

func TestSet_JSON(t *testing.T) {
	t.Run("round-trips as an array of elements", func(t *testing.T) {
		data, err := json.Marshal(SetOf(Tile{1, 0}, Tile{0, 0}))
		assert.NoError(t, err)
		assert.Equal(t, string(data), `[{"x":0,"y":0},{"x":1,"y":0}]`)

		var decoded Set[Tile]
		assert.NoError(t, json.Unmarshal(data, &decoded))
		assert.True(t, decoded.Equal(SetOf(Tile{0, 0}, Tile{1, 0})))
	})
	t.Run("the empty set is null and reads back empty", func(t *testing.T) {
		data, err := json.Marshal(testSetEmpty)
		assert.NoError(t, err)
		assert.Equal(t, string(data), "null")

		decoded := testSet
		assert.NoError(t, json.Unmarshal(data, &decoded))
		assert.True(t, decoded.IsEmpty())
	})
	t.Run("an empty array is the empty set", func(t *testing.T) {
		decoded := testSet
		assert.NoError(t, json.Unmarshal([]byte("[]"), &decoded))
		assert.True(t, decoded.Equal(testSetEmpty))
	})
	t.Run("orders and drops repeats on the way in", func(t *testing.T) {
		var decoded Set[Tile]
		assert.NoError(t, json.Unmarshal([]byte(`[{"x":1,"y":0},{"x":0,"y":0},{"x":1,"y":0}]`), &decoded))
		assert.Equal(t, slices.Collect(decoded.All()), []Tile{{0, 0}, {1, 0}})
	})
	t.Run("anything but an array of elements fails", func(t *testing.T) {
		var decoded Set[Tile]
		assert.Error(t, json.Unmarshal([]byte(`{"x":1}`), &decoded))
	})
	t.Run("a failed read leaves the set as it was", func(t *testing.T) {
		decoded := testSet

		assert.Error(t, json.Unmarshal([]byte(`[{"x":"a"}]`), &decoded))
		assert.True(t, decoded.Equal(testSet))
	})
}

func FuzzSet_JSON(f *testing.F) {
	f.Add([]byte(`[{"x":1,"y":0},{"x":0,"y":0},{"x":1,"y":0}]`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var decoded Set[Tile]
		if json.Unmarshal(data, &decoded) != nil {
			return
		}

		elements := decoded.Elements()
		assert.True(t, slices.IsSortedFunc(elements, Tile.Compare))
		assert.Length(t, slices.CompactFunc(slices.Clone(elements), func(a, b Tile) bool {
			return a == b
		}), len(elements))

		encoded, err := json.Marshal(decoded)
		assert.NoError(t, err)

		var again Set[Tile]
		assert.NoError(t, json.Unmarshal(encoded, &again))
		assert.True(t, again.Equal(decoded))
	})
}
