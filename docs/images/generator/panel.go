package main

import (
	geom "github.com/gravitton/geometry"
)

type panel struct {
	canvas
	cellSize float64
}

func newPanel(t theme, cellSize float64, caption string) *panel {
	return &panel{canvas: newCanvas(t, caption), cellSize: cellSize}
}

func (p *panel) corner(index geom.Point[int]) point {
	return index.Float().Multiply(p.cellSize)
}

func (p *panel) center(index geom.Point[int]) point {
	return p.corner(index).AddXY(p.cellSize/2, p.cellSize/2)
}

func (p *panel) cell(index geom.Point[int], fill string) {
	p.polygon(p.box(index, index).Polygon().Points, style{fill: fill, stroke: p.theme.edge, width: 1})
}

func (p *panel) outline(from, to geom.Point[int], color string) {
	p.polygon(p.box(from, to).Polygon().Points, style{stroke: color, width: 2, dashed: true})
}

func (p *panel) box(from, to geom.Point[int]) geom.Box[float64] {
	return geom.Bx(p.corner(from), p.corner(to).AddXY(p.cellSize, p.cellSize))
}

func (p *panel) text(index geom.Point[int], dy float64, f font, spans ...span) {
	p.canvas.text(p.center(index).AddXY(0, dy), f, spans...)
}
