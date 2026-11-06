package main

import (
	"slices"
	"strconv"

	"github.com/gravitton/container"
	geom "github.com/gravitton/geometry"
)

type region = container.Set[geom.Point[int]]

var figures = []figure{
	{"array", array},
	{"operations", operations},
}

func array(t theme) []*panel {
	a := container.Arr[int](geom.Sz(4, 3))
	offset := 0
	for _, value := range a.All() {
		*value = offset
		offset++
	}
	rows := []string{t.accentSoft, t.coolSoft, t.leafSoft}

	keys := newPanel(t, 44, "a.Keys()")
	for index, value := range a.All() {
		keys.cell(index, rows[index.Y])
		keys.text(index, -6, font{size: 12}, pair(t, index)...)
		keys.text(index, 11, font{size: 10}, span{strconv.Itoa(*value), t.muted})
	}

	data := newPanel(t, 36, "a.Data()")
	indices := slices.Collect(a.Keys())
	for i, value := range a.Data() {
		at := geom.Pt(i, 0)
		data.cell(at, rows[indices[i].Y])
		data.text(at, -5, font{size: 12}, span{strconv.Itoa(value), t.ink})
		data.text(at, 9, font{size: 9}, pair(t, indices[i])...)
	}

	return []*panel{keys, data}
}

func pair(t theme, index geom.Point[int]) []span {
	return []span{{strconv.Itoa(index.X), t.ink}, {",", t.muted}, {strconv.Itoa(index.Y), t.ink}}
}

func operations(t theme) []*panel {
	a := container.SetOf(geom.Pt(1, 1), geom.Pt(2, 1), geom.Pt(1, 2), geom.Pt(2, 2))
	b := container.SetOf(geom.Pt(2, 2), geom.Pt(3, 2), geom.Pt(2, 3), geom.Pt(3, 3))
	results := []struct {
		caption string
		set     region
	}{
		{"a.Union(b)", a.Union(b)},
		{"a.Intersection(b)", a.Intersection(b)},
		{"a.Difference(b)", a.Difference(b)},
		{"a.SymmetricDifference(b)", a.SymmetricDifference(b)},
	}
	field := container.Arr[struct{}](geom.Sz(5, 5))

	panels := make([]*panel, 0, len(results))
	for _, result := range results {
		p := newPanel(t, 26, result.caption)
		for index := range field.Keys() {
			p.cell(index, membership(t, a, b, result.set, index))
		}
		bound(p, a, t.accent)
		bound(p, b, t.cool)
		first, _ := a.First()
		last, _ := b.Last()
		p.text(first, 0, font{size: 12}, span{"a", t.ink})
		p.text(last, 0, font{size: 12}, span{"b", t.ink})
		panels = append(panels, p)
	}

	return panels
}

func membership(t theme, a, b, result region, index geom.Point[int]) string {
	switch {
	case !result.Contains(index):
		return t.cell
	case a.Contains(index) && b.Contains(index):
		return t.leafSoft
	case a.Contains(index):
		return t.accentSoft
	default:
		return t.coolSoft
	}
}

func bound(p *panel, s region, color string) {
	first, _ := s.First()
	last, _ := s.Last()
	p.outline(first, last, color)
}
