# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased](https://github.com/gravitton/container/compare/v1.0.0...main)

## [v1.0.0](https://github.com/gravitton/container/commits/v1.0.0) (2026-10-27)
### Added
- `Array`, a table of values per index of a size in columns and rows, moved from `grid` with `Arr`, `Data`, its values row by row as a slice into the array, and its JSON, which refuses a size past what an int counts on any platform and reads `null` as the zero array
- `Set`, an immutable sorted set with its set operations and JSON, moved from `grid`, with the variadic `SetOf`; elements that compare the same are one, and a result may hold that of either set; the empty set is `null` in JSON, as the zero array's values are, and `Set.At` panics with a `container:` error for an index outside the set
- `SetFrom`, the set of a slice it takes as its own, at no allocation
