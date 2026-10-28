<div align="center" width="100%">

<a href="https://github.com/gravitton">
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/gravitton/container/refs/heads/main/docs/images/logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/gravitton/container/refs/heads/main/docs/images/logo-light.svg">
  <img alt="Gravitton container" src="https://raw.githubusercontent.com/gravitton/container/refs/heads/main/docs/images/logo-light.svg" width="300">
</picture>
</a>

[![Latest Stable Version][ico-release]][link-release]
[![Build Status][ico-workflow]][link-workflow]
[![Coverage Status][ico-coverage]][link-coverage]
[![Go Dev Reference][ico-go-dev-reference]][link-go-dev-reference]
[![Software License][ico-license]][link-licence]

Generic containers for Go

<hr>

</div>


## Features

- **Arrays**: a value per `geom.Point` of a size, in one slice, with sides of any integer type.
- **Sets**: immutable sorted sets of any `Compare` element, with union, intersection and difference.

The rules behind them are in the [package documentation][link-go-dev-reference].

## Installation

```shell
go get github.com/gravitton/container
```

```go
import "github.com/gravitton/container"
```

## Usage

### Arrays

An `Array[T, S]` holds one value for every index of a size in columns and rows. `S` is the
integer type of its sides and indices, taken from the size it is made with. An array is a handle,
as a slice is: a copy shares its values, and `Clone` gives storage of its own.

```go
a := container.Arr[int](geom.Sz(4, 3))                  // an Array[int, int], every value zero
small := container.Arr[uint8](geom.Sz[int16](320, 200)) // an Array[uint8, int16], indexed by geom.Point[int16]

a.Set(geom.Pt(2, 1), 7)
*a.Get(geom.Pt(2, 1)) += 1 // a pointer into the array
a.Contains(geom.Pt(4, 0))  // false, the check before Get and Set, which panic outside
a.Size()                   // 4x3; also Width, Height and Len
snapshot := a.Clone()      // storage of its own
```

### Iteration

The values are stored row by row, and every traversal follows that order: `Keys` walks the
indices, `Data` is the storage itself.

<p align="center">
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/gravitton/container/refs/heads/main/docs/images/array-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/gravitton/container/refs/heads/main/docs/images/array-light.svg">
  <img alt="An array of four columns and three rows, each cell labelled with its index and its offset, beside the same twelve values as one slice, row after row" src="https://raw.githubusercontent.com/gravitton/container/refs/heads/main/docs/images/array-light.svg">
</picture>
</p>

```go
for index, value := range a.All() { // also Keys and Values
	*value += index.X
}

values := a.Data() // the values row by row as a slice into the array, the fastest loop
for i := range values {
	values[i] = min(values[i], 9)
}

data, err := json.Marshal(a)   // {"w":4,"h":3,"d":[…]}, the values row by row
err = json.Unmarshal(data, &a) // and back; an error where the values do not fill the size
```

### Sets

A `Set[E]` holds any element with a `Compare` method, once each and in that order. It never
changes: every operation returns a set, so one is safe to copy and share.

```go
a := container.SetOf(geom.Pt(1, 1), geom.Pt(2, 1), geom.Pt(1, 2), geom.Pt(2, 2))
from := container.SetFrom(points) // the same from a slice it takes as its own, without a copy

a.Contains(geom.Pt(2, 1)) // true, a binary search
a.Len()                   // 4; also IsEmpty
a.First()                 // (1,1), true; also Last
a.Index(geom.Pt(2, 1))    // 2, true, the index At reads it back from
```

### Set operations

Each operation walks both sets once and allocates its result once, or not at all where it equals
one of them.

<p align="center">
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/gravitton/container/refs/heads/main/docs/images/operations-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/gravitton/container/refs/heads/main/docs/images/operations-light.svg">
  <img alt="Two overlapping squares of points, a and b, and the cells of their union, intersection, difference and symmetric difference" src="https://raw.githubusercontent.com/gravitton/container/refs/heads/main/docs/images/operations-light.svg">
</picture>
</p>

```go
b := container.SetOf(geom.Pt(2, 2), geom.Pt(3, 2), geom.Pt(2, 3), geom.Pt(3, 3))

a.Union(b)               // (1,1) (1,2) (2,1) (2,2) (2,3) (3,2) (3,3)
a.Intersection(b)        // (2,2)
a.Difference(b)          // (1,1) (1,2) (2,1)
a.SymmetricDifference(b) // (1,1) (1,2) (2,1) (2,3) (3,2) (3,3)

a.Intersects(b) // true, without building the intersection
a.Encloses(b)   // false, whether a holds every element of b
a.Equal(b)      // false
```

A set is walked and stored in its order:

```go
for p := range a.All() { // also Backward
	draw(p)
}

a.Elements()    // a fresh slice; also AppendElements
json.Marshal(a) // [{"x":1,"y":1},…], null for the empty set
```

Full reference: [pkg.go.dev][link-go-dev-reference].

## Credits

- [Tomáš Novotný](https://github.com/tomas-novotny)
- [All Contributors][link-contributors]


## License

The MIT License (MIT). Please see [License File][link-licence] for more information.


[ico-license]:              https://img.shields.io/github/license/gravitton/container.svg?style=flat-square&color=blue
[ico-workflow]:             https://img.shields.io/github/actions/workflow/status/gravitton/container/main.yml?branch=main&style=flat-square
[ico-release]:              https://img.shields.io/github/v/release/gravitton/container?style=flat-square&color=blue
[ico-go-dev-reference]:     https://img.shields.io/badge/go.dev-reference-blue?style=flat-square
[ico-coverage]:             https://img.shields.io/coverallsCoverage/github/gravitton/container?style=flat-square

[link-release]:             https://github.com/gravitton/container/releases
[link-contributors]:        https://github.com/gravitton/container/contributors
[link-licence]:             ./LICENSE.md
[link-workflow]:            https://github.com/gravitton/container/actions
[link-go-dev-reference]:    https://pkg.go.dev/github.com/gravitton/container
[link-coverage]:            https://coveralls.io/github/gravitton/container
