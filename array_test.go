package container_test

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/gravitton/assert"
	geom "github.com/gravitton/geometry"

	. "github.com/gravitton/container"
)

var (
	sinkArray Array[int, int]
	sinkPoint geom.Point[int]
	sinkBool  bool
	sinkInt   int
	sinkInts  []int
)

// testArraySizes returns the sizes the properties of an array are checked on.
func testArraySizes() []geom.Size[int] {
	return []geom.Size[int]{geom.Sz(4, 3), geom.Sz(1, 1), geom.Sz(1, 5), geom.Sz(0, 0), geom.Sz(4, 0), geom.Sz(0, 3)}
}

func TestArray_Constructor(t *testing.T) {
	a := Arr[int](geom.Sz(4, 3))

	t.Run("has the given size", func(t *testing.T) {
		assert.Equal(t, a.Size(), geom.Sz(4, 3))
	})
	t.Run("holds a zero value per index", func(t *testing.T) {
		assert.Equal(t, a.Len(), 12)

		for index, value := range a.All() {
			assert.Equal(t, *value, 0, index.String()+": ")
		}
	})
	t.Run("a size without columns or rows is empty", func(t *testing.T) {
		for _, size := range []geom.Size[int]{geom.Sz(0, 0), geom.Sz(4, 0), geom.Sz(0, 3)} {
			empty := Arr[int](size)

			assert.Equal(t, empty.Size(), size, size.String()+": ")
			assert.Equal(t, empty.Len(), 0, size.String()+": ")
			assert.False(t, empty.Contains(geom.Pt(0, 0)), size.String()+": ")
		}
	})
	t.Run("panics for a negative width or height", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			Arr[int](geom.Sz(-1, 5))
		}, "container: array of negative size -1x5")
		assert.PanicsWith(t, func() {
			Arr[int](geom.Sz(4, -3))
		}, "container: array of negative size 4x-3")
	})
	t.Run("panics for two negatives, whose area is positive", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			Arr[int](geom.Sz(-2, -3))
		}, "container: array of negative size -2x-3")
	})
	t.Run("a narrow size type counts and places past what it holds", func(t *testing.T) {
		narrow := Arr[int](geom.Sz[int8](127, 3))
		narrow.Set(geom.Pt[int8](126, 2), 7)

		assert.Equal(t, narrow.Size(), geom.Sz[int8](127, 3))
		assert.Equal(t, narrow.Len(), 381)
		assert.Equal(t, *narrow.Get(geom.Pt[int8](126, 2)), 7)
		assert.Equal(t, narrow.Row(2)[126], 7)
		assert.False(t, narrow.Contains(geom.Pt[int8](-1, 0)))
	})
	t.Run("panics for a size of more values than an int counts", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			Arr[int](geom.Sz[int64](1<<32, 1<<32))
		}, "container: array of size 4294967296x4294967296 is too large")
	})
	t.Run("the zero value is the empty array", func(t *testing.T) {
		var zero Array[int, int]

		assert.Equal(t, zero.Len(), 0)
		assert.False(t, zero.Contains(geom.Pt(0, 0)))

		for range zero.All() {
			assert.Fail(t, "the zero array yielded an index")
		}
	})
	t.Run("allocates its values once", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkArray = Arr[int](geom.Sz(4, 3))
		}), 1.0)
	})
}

func ExampleArr() {
	heat := Arr[float64](geom.Sz(3, 2))
	heat.Set(geom.Pt(1, 1), 0.5)
	*heat.Get(geom.Pt(1, 1)) += 0.25

	for _, value := range heat.All() {
		*value *= 2
	}

	row := heat.Row(1)
	for x := range row {
		row[x] /= 2
	}

	data, _ := json.Marshal(heat)
	small := Arr[uint8](geom.Sz[int16](320, 200))

	fmt.Println(heat.Size(), heat.Len(), *heat.Get(geom.Pt(1, 1)))
	fmt.Println(string(data))
	fmt.Println(small.Size(), small.Len())
	// Output:
	// 3x2 6 0.75
	// {"w":3,"h":2,"d":[0,0,0,0,0.75,0]}
	// 320x200 64000
}

func TestArray_Size(t *testing.T) {
	t.Run("is the size the array is made with", func(t *testing.T) {
		for _, size := range testArraySizes() {
			assert.Equal(t, Arr[int](size).Size(), size, size.String()+": ")
		}
	})
	t.Run("the zero value has no columns or rows", func(t *testing.T) {
		assert.Equal(t, Array[int, int]{}.Size(), geom.Sz(0, 0))
	})
}

func TestArray_Width(t *testing.T) {
	t.Run("is the number of columns", func(t *testing.T) {
		assert.Equal(t, Arr[int](geom.Sz(4, 3)).Width(), 4)
	})
	t.Run("agrees with the size", func(t *testing.T) {
		for _, size := range testArraySizes() {
			assert.Equal(t, Arr[int](size).Width(), size.Width, size.String()+": ")
		}
	})
}

func TestArray_Height(t *testing.T) {
	t.Run("is the number of rows", func(t *testing.T) {
		assert.Equal(t, Arr[int](geom.Sz(4, 3)).Height(), 3)
	})
	t.Run("agrees with the size", func(t *testing.T) {
		for _, size := range testArraySizes() {
			assert.Equal(t, Arr[int](size).Height(), size.Height, size.String()+": ")
		}
	})
}

func TestArray_Len(t *testing.T) {
	t.Run("is the width times the height", func(t *testing.T) {
		for _, size := range testArraySizes() {
			assert.Equal(t, Arr[int](size).Len(), size.Width*size.Height, size.String()+": ")
		}
	})
	t.Run("agrees with the number of indices", func(t *testing.T) {
		for _, size := range testArraySizes() {
			a := Arr[int](size)

			assert.Length(t, slices.Collect(a.Keys()), a.Len(), size.String()+": ")
		}
	})
}

func TestArray_Get(t *testing.T) {
	a := Arr[int](geom.Sz(4, 3))

	t.Run("returns a pointer to the stored value", func(t *testing.T) {
		a.Set(geom.Pt(1, 2), 42)
		*a.Get(geom.Pt(1, 2))++

		assert.Equal(t, *a.Get(geom.Pt(1, 2)), 43)
	})
	t.Run("panics past the width instead of reading the next row", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			a.Get(geom.Pt(4, 0))
		}, "container: index (4,0) outside array of size 4x3")
	})
	t.Run("panics before the first column", func(t *testing.T) {
		assert.Panics(t, func() {
			a.Get(geom.Pt(-1, 1))
		})
	})
	t.Run("panics outside the rows", func(t *testing.T) {
		assert.Panics(t, func() {
			a.Get(geom.Pt(0, 3))
		})
	})
}

func BenchmarkArray_Get(b *testing.B) {
	b.Run("int", func(b *testing.B) {
		benchmarkArrayGet(b, Arr[int](geom.Sz(64, 64)))
	})
	b.Run("int32", func(b *testing.B) {
		benchmarkArrayGet(b, Arr[int](geom.Sz[int32](64, 64)))
	})
}

// benchmarkArrayGet reads every value of a 64 by 64 array through Get.
func benchmarkArrayGet[S geom.Integer](b *testing.B, a Array[int, S]) {
	for b.Loop() {
		for y := range 64 {
			for x := range 64 {
				sinkInt = *a.Get(geom.Pt(S(x), S(y)))
			}
		}
	}
}

func TestArray_Set(t *testing.T) {
	a := Arr[int](geom.Sz(4, 3))

	t.Run("stores the value at the index alone", func(t *testing.T) {
		a.Set(geom.Pt(1, 2), 42)
		a.Set(geom.Pt(0, 0), 7)

		assert.Equal(t, *a.Get(geom.Pt(1, 2)), 42)
		assert.Equal(t, *a.Get(geom.Pt(0, 0)), 7)
	})
	t.Run("panics past the width instead of writing the next row", func(t *testing.T) {
		assert.Panics(t, func() {
			a.Set(geom.Pt(4, 0), 7)
		})
		assert.Equal(t, *a.Get(geom.Pt(0, 1)), 0)
	})
}

func BenchmarkArray_Set(b *testing.B) {
	b.Run("int", func(b *testing.B) {
		benchmarkArraySet(b, Arr[int](geom.Sz(64, 64)))
	})
	b.Run("int32", func(b *testing.B) {
		benchmarkArraySet(b, Arr[int](geom.Sz[int32](64, 64)))
	})
}

// benchmarkArraySet writes every value of a 64 by 64 array through Set.
func benchmarkArraySet[S geom.Integer](b *testing.B, a Array[int, S]) {
	for b.Loop() {
		for y := range 64 {
			for x := range 64 {
				a.Set(geom.Pt(S(x), S(y)), x)
			}
		}
	}
}

func FuzzArray_Set(f *testing.F) {
	f.Add(uint8(4), uint8(3), []byte{0, 0, 1, 3, 2, 9, 4, 0, 5, 1, 1, 7, 1, 1, 8})
	f.Add(uint8(0), uint8(5), []byte{0, 0, 1})
	f.Add(uint8(1), uint8(1), []byte{})

	f.Fuzz(func(t *testing.T, width, height uint8, writes []byte) {
		size := geom.Sz(int(width%16), int(height%16))
		a := Arr[int](size)
		rows := make([][]int, size.Height)
		for y := range rows {
			rows[y] = make([]int, size.Width)
		}

		for i := 0; i+2 < len(writes); i += 3 {
			index, value := geom.Pt(int(writes[i]%20)-2, int(writes[i+1]%20)-2), int(writes[i+2])
			inside := index.X >= 0 && index.X < size.Width && index.Y >= 0 && index.Y < size.Height

			assert.Equal(t, a.Contains(index), inside, index.String()+": ")

			if !inside {
				assert.Panics(t, func() {
					a.Set(index, value)
				}, index.String()+": ")

				continue
			}

			a.Set(index, value)
			rows[index.Y][index.X] = value
		}

		for y, row := range rows {
			assert.Equal(t, a.Row(y), row, fmt.Sprint("row ", y, ": "))
		}

		count := 0
		for index, value := range a.All() {
			count++

			assert.Equal(t, *value, rows[index.Y][index.X], index.String()+": ")
		}

		assert.Equal(t, count, size.Width*size.Height)
		assert.Equal(t, a.Len(), size.Width*size.Height)
	})
}

func TestArray_Row(t *testing.T) {
	a := Arr[int](geom.Sz(4, 3))
	for index, value := range a.All() {
		*value = index.X + 10*index.Y
	}

	t.Run("holds the values of the row in column order", func(t *testing.T) {
		for y := range a.Height() {
			row := a.Row(y)

			assert.Length(t, row, a.Width())
			for x, value := range row {
				assert.Equal(t, value, *a.Get(geom.Pt(x, y)), geom.Pt(x, y).String()+": ")
			}
		}
	})
	t.Run("a write through it changes the array", func(t *testing.T) {
		a.Row(1)[2] = 99

		assert.Equal(t, *a.Get(geom.Pt(2, 1)), 99)
	})
	t.Run("an append past its end leaves the next row alone", func(t *testing.T) {
		sinkInts = append(a.Row(0), 77)

		assert.Equal(t, *a.Get(geom.Pt(0, 1)), 10)
	})
	t.Run("a row of no columns is empty", func(t *testing.T) {
		assert.Length(t, Arr[int](geom.Sz(0, 3)).Row(1), 0)
	})
	t.Run("panics outside the rows", func(t *testing.T) {
		assert.PanicsWith(t, func() {
			a.Row(3)
		}, "container: row 3 outside array of size 4x3")
		assert.PanicsWith(t, func() {
			a.Row(-1)
		}, "container: row -1 outside array of size 4x3")
	})
	t.Run("allocates nothing", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkInts = a.Row(2)
		}), 0.0)
	})
}

func BenchmarkArray_Row(b *testing.B) {
	a := Arr[int](geom.Sz(64, 64))

	for b.Loop() {
		for y := range 64 {
			row := a.Row(y)
			for x := range row {
				row[x] = x
			}
		}
	}
}

func TestArray_Fill(t *testing.T) {
	a := Arr[int](geom.Sz(3, 3))
	a.Fill(9)

	for index, value := range a.All() {
		assert.Equal(t, *value, 9, index.String()+": ")
	}
}

func TestArray_Clear(t *testing.T) {
	a := Arr[int](geom.Sz(3, 3))
	a.Fill(9)
	a.Clear()

	for index, value := range a.All() {
		assert.Equal(t, *value, 0, index.String()+": ")
	}
}

func TestArray_Clone(t *testing.T) {
	original := Arr[int](geom.Sz(3, 3))
	original.Fill(1)

	clone := original.Clone()
	clone.Set(geom.Pt(0, 0), 99)

	t.Run("holds the same values", func(t *testing.T) {
		assert.Equal(t, clone.Size(), original.Size())
		assert.Equal(t, *clone.Get(geom.Pt(2, 2)), 1)
	})
	t.Run("a write to the clone leaves the original alone", func(t *testing.T) {
		assert.Equal(t, *original.Get(geom.Pt(0, 0)), 1)
		assert.Equal(t, *clone.Get(geom.Pt(0, 0)), 99)
	})
	t.Run("a plain copy shares the values", func(t *testing.T) {
		shared := original
		shared.Set(geom.Pt(1, 1), 7)

		assert.Equal(t, *original.Get(geom.Pt(1, 1)), 7)
	})
	t.Run("the zero value clones to the zero value", func(t *testing.T) {
		assert.Equal(t, Array[int, int]{}.Clone(), Array[int, int]{})
	})
	t.Run("an empty array clones to an empty array of its size", func(t *testing.T) {
		clone := Arr[int](geom.Sz(4, 0)).Clone()

		assert.Equal(t, clone.Size(), geom.Sz(4, 0))
		assert.Equal(t, clone.Len(), 0)
	})
	t.Run("allocates its values once", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			sinkArray = original.Clone()
		}), 1.0)
	})
}

func TestArray_Keys(t *testing.T) {
	t.Run("yields every index row by row", func(t *testing.T) {
		var indices []geom.Point[int]
		for index := range Arr[int](geom.Sz(3, 2)).Keys() {
			indices = append(indices, index)
		}

		assert.Equal(t, indices, []geom.Point[int]{geom.Pt(0, 0), geom.Pt(1, 0), geom.Pt(2, 0), geom.Pt(0, 1), geom.Pt(1, 1), geom.Pt(2, 1)})
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range Arr[int](geom.Sz(5, 5)).Keys() {
			count++
			if count == 3 {
				break
			}
		}

		assert.Equal(t, count, 3)
	})
	t.Run("allocates nothing", func(t *testing.T) {
		a := Arr[int](geom.Sz(8, 8))

		assert.Equal(t, testing.AllocsPerRun(10, func() {
			for index := range a.Keys() {
				sinkPoint = index
			}
		}), 0.0)
	})
}

func TestArray_Values(t *testing.T) {
	a := Arr[int](geom.Sz(2, 2))

	t.Run("yields a pointer into the array", func(t *testing.T) {
		for value := range a.Values() {
			*value = 5
		}

		assert.Equal(t, *a.Get(geom.Pt(1, 1)), 5)
	})
	t.Run("follows the order of Keys", func(t *testing.T) {
		var values, expected []*int
		for value := range a.Values() {
			values = append(values, value)
		}
		for index := range a.Keys() {
			expected = append(expected, a.Get(index))
		}

		assert.Equal(t, values, expected)
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range a.Values() {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
	t.Run("allocates nothing", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			for value := range a.Values() {
				sinkBool = *value == 0
			}
		}), 0.0)
	})
}

func TestArray_All(t *testing.T) {
	a := Arr[int](geom.Sz(2, 2))
	a.Set(geom.Pt(1, 0), 42)

	t.Run("yields every index with its value", func(t *testing.T) {
		count := 0
		for index, value := range a.All() {
			count++
			assert.Equal(t, value, a.Get(index), index.String()+": ")
		}

		assert.Equal(t, count, 4)
	})
	t.Run("follows the order of Keys", func(t *testing.T) {
		var indices []geom.Point[int]
		for index := range a.All() {
			indices = append(indices, index)
		}

		assert.Equal(t, indices, slices.Collect(a.Keys()))
	})
	t.Run("stops at the break", func(t *testing.T) {
		count := 0
		for range a.All() {
			count++

			break
		}

		assert.Equal(t, count, 1)
	})
	t.Run("allocates nothing", func(t *testing.T) {
		assert.Equal(t, testing.AllocsPerRun(10, func() {
			for index, value := range a.All() {
				sinkPoint, sinkBool = index, *value == 0
			}
		}), 0.0)
	})
}

func TestArray_Contains(t *testing.T) {
	a := Arr[int](geom.Sz(4, 3))

	t.Run("the corners are inside", func(t *testing.T) {
		assert.True(t, a.Contains(geom.Pt(0, 0)))
		assert.True(t, a.Contains(geom.Pt(3, 2)))
	})
	t.Run("a negative coordinate is outside", func(t *testing.T) {
		assert.False(t, a.Contains(geom.Pt(-1, 0)))
		assert.False(t, a.Contains(geom.Pt(0, -1)))
	})
	t.Run("the size itself is outside", func(t *testing.T) {
		assert.False(t, a.Contains(geom.Pt(4, 0)))
		assert.False(t, a.Contains(geom.Pt(0, 3)))
	})
}

func TestArray_JSON(t *testing.T) {
	t.Run("round-trips as its size and its values row by row", func(t *testing.T) {
		a := Arr[int](geom.Sz(3, 2))
		a.Set(geom.Pt(1, 0), 7)
		a.Set(geom.Pt(2, 1), 9)

		data, err := json.Marshal(a)
		assert.NoError(t, err)
		assert.Equal(t, string(data), `{"w":3,"h":2,"d":[0,7,0,0,0,9]}`)

		var decoded Array[int, int]
		assert.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, decoded.Size(), a.Size())
		assert.Equal(t, slices.Collect(decoded.Keys()), slices.Collect(a.Keys()))

		for index, value := range decoded.All() {
			assert.Equal(t, *value, *a.Get(index), index.String()+": ")
		}
	})
	t.Run("the zero value holds no values and reads back empty", func(t *testing.T) {
		data, err := json.Marshal(Array[int, int]{})
		assert.NoError(t, err)
		assert.Equal(t, string(data), `{"w":0,"h":0,"d":null}`)

		var decoded Array[int, int]
		assert.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, decoded.Len(), 0)
	})
	t.Run("a size without rows keeps its width", func(t *testing.T) {
		data, err := json.Marshal(Arr[int](geom.Sz(4, 0)))
		assert.NoError(t, err)
		assert.Equal(t, string(data), `{"w":4,"h":0,"d":[]}`)

		var decoded Array[int, int]
		assert.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, decoded.Size(), geom.Sz(4, 0))
	})
	t.Run("null values are none", func(t *testing.T) {
		var decoded Array[int, int]

		assert.NoError(t, json.Unmarshal([]byte(`{"w":0,"h":3,"d":null}`), &decoded))
		assert.Equal(t, decoded.Size(), geom.Sz(0, 3))
		assert.Equal(t, decoded.Len(), 0)
	})
	t.Run("rejects a number of values that is not the size", func(t *testing.T) {
		var decoded Array[int, int]

		assert.ErrorContains(t, json.Unmarshal([]byte(`{"w":2,"h":2,"d":[1,2,3]}`), &decoded), "container: array of size 2x2 with 3 values")
		assert.ErrorContains(t, json.Unmarshal([]byte(`{"w":2,"h":2}`), &decoded), "container: array of size 2x2 with 0 values")
	})
	t.Run("rejects a negative size, of two negatives too", func(t *testing.T) {
		var decoded Array[int, int]

		assert.ErrorContains(t, json.Unmarshal([]byte(`{"w":-1,"h":2,"d":[]}`), &decoded), "container: array of negative size -1x2")
		assert.ErrorContains(t, json.Unmarshal([]byte(`{"w":-2,"h":-3,"d":[1,2,3,4,5,6]}`), &decoded), "container: array of negative size -2x-3")
	})
	t.Run("rejects a side the size type cannot hold", func(t *testing.T) {
		var decoded Array[int, int8]

		assert.Error(t, json.Unmarshal([]byte(`{"w":200,"h":1,"d":[]}`), &decoded))
	})
	t.Run("rejects a size of more values than an int counts", func(t *testing.T) {
		var decoded Array[int, int]

		assert.ErrorContains(t, json.Unmarshal([]byte(`{"w":4294967296,"h":4294967296,"d":[]}`), &decoded), "container: array of size 4294967296x4294967296 is too large")
	})
	t.Run("rejects what is not an array", func(t *testing.T) {
		var decoded Array[int, int]

		assert.Error(t, json.Unmarshal([]byte(`[1,2,3]`), &decoded))
	})
	t.Run("a failed read leaves the array as it was", func(t *testing.T) {
		kept := Arr[int](geom.Sz(2, 1))
		kept.Set(geom.Pt(1, 0), 5)

		assert.Error(t, json.Unmarshal([]byte(`{"w":3,"h":3,"d":[1]}`), &kept))
		assert.Equal(t, kept.Size(), geom.Sz(2, 1))
		assert.Equal(t, *kept.Get(geom.Pt(1, 0)), 5)
	})
}

func FuzzArray_JSON(f *testing.F) {
	f.Add([]byte(`{"w":3,"h":2,"d":[0,7,0,0,0,9]}`))
	f.Add([]byte(`{"w":0,"h":0,"d":null}`))
	f.Add([]byte(`{"w":2,"h":2,"d":[1,2,3]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var decoded Array[int, int8]
		if json.Unmarshal(data, &decoded) != nil {
			return
		}

		encoded, err := json.Marshal(decoded)
		assert.NoError(t, err)

		var again Array[int, int8]
		assert.NoError(t, json.Unmarshal(encoded, &again))

		reencoded, err := json.Marshal(again)
		assert.NoError(t, err)
		assert.Equal(t, string(reencoded), string(encoded))
	})
}
