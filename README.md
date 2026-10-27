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

- **Arrays** – a table of values per index of a size, stored row by row in one slice.
- **Sets** – immutable sorted sets of any element with a `Compare` method, with their set operations.

The rules behind them are in the [package documentation][link-go-dev-reference].

## Installation

```shell
go get github.com/gravitton/container
```


## Usage

### Arrays

`Array[T, S]` is a table of values, one for every index of a size in columns and rows. `S` is
the integer type of its sides and indices, taken from the size it is made with:

```go
heat := container.Arr[float64](geom.Sz(64, 64))        // an Array[float64, int]
small := container.Arr[uint8](geom.Sz[int16](320, 200)) // an Array[uint8, int16], indexed by geom.Point[int16]

heat.Set(geom.Pt(1, 2), 0.5)
*heat.Get(geom.Pt(1, 2)) += 0.25

for _, value := range heat.All() { // also Keys and Values
	*value *= 0.9
}

values := heat.Data() // every value row by row as a slice into the array, the fastest loop
for i := range values {
	values[i] = min(values[i], 1)
}

data, err := json.Marshal(heat)   // {"w":64,"h":64,"d":[…]}, the values row by row
err = json.Unmarshal(data, &heat) // and back; an error where the values do not fill the size
```

### Sets

`Set[E]` is an immutable set of any element with a `Compare` method, held once each and sorted:

```go
seen := container.SetOf(points...) // in any order, with repeats; SetFrom takes a slice as its own
hit := container.SetOf(geom.Pt(2, 0), geom.Pt(3, 0))

seen.Intersection(hit)       // also Union, Difference and SymmetricDifference, at most one allocation
seen.Intersects(hit)         // whether they share an element, without building it; also Encloses
seen.Contains(geom.Pt(1, 0)) // a binary search
seen.Len()                   // also IsEmpty and Equal

for p := range seen.All() { // in order; also Backward
	draw(p)
}

seen.Elements()    // the elements as a fresh slice; also AppendElements
json.Marshal(seen) // an array of the elements in order, null for the empty set
```


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
