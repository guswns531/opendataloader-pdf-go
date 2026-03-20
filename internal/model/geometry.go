package model

import "math"

// Point represents a 2D coordinate in PDF points.
type Point struct {
	X float64
	Y float64
}

// Size represents a width/height pair in PDF points.
type Size struct {
	Width  float64
	Height float64
}

// Box represents a bounding box in PDF coordinates: left, bottom, right, top.
type Box struct {
	Left   float64
	Bottom float64
	Right  float64
	Top    float64
}

// NewBox constructs a box from corners.
func NewBox(left, bottom, right, top float64) Box {
	return Box{Left: left, Bottom: bottom, Right: right, Top: top}
}

// IsZero reports whether the box is empty.
func (b Box) IsZero() bool {
	return b.Left == 0 && b.Bottom == 0 && b.Right == 0 && b.Top == 0
}

// Normalize returns a box whose corners are ordered consistently.
func (b Box) Normalize() Box {
	if b.Left > b.Right {
		b.Left, b.Right = b.Right, b.Left
	}
	if b.Bottom > b.Top {
		b.Bottom, b.Top = b.Top, b.Bottom
	}
	return b
}

// Width returns the box width.
func (b Box) Width() float64 {
	b = b.Normalize()
	return math.Max(0, b.Right-b.Left)
}

// Height returns the box height.
func (b Box) Height() float64 {
	b = b.Normalize()
	return math.Max(0, b.Top-b.Bottom)
}

// Area returns the box area.
func (b Box) Area() float64 {
	return b.Width() * b.Height()
}

// Contains reports whether the point lies inside the box.
func (b Box) Contains(p Point) bool {
	b = b.Normalize()
	return p.X >= b.Left && p.X <= b.Right && p.Y >= b.Bottom && p.Y <= b.Top
}

// Intersects reports whether two boxes overlap.
func (b Box) Intersects(other Box) bool {
	a := b.Normalize()
	c := other.Normalize()
	return a.Left <= c.Right && a.Right >= c.Left && a.Bottom <= c.Top && a.Top >= c.Bottom
}

// Intersection returns the overlapping region of two boxes.
func (b Box) Intersection(other Box) (Box, bool) {
	if !b.Intersects(other) {
		return Box{}, false
	}
	a := b.Normalize()
	c := other.Normalize()
	return Box{
		Left:   math.Max(a.Left, c.Left),
		Bottom: math.Max(a.Bottom, c.Bottom),
		Right:  math.Min(a.Right, c.Right),
		Top:    math.Min(a.Top, c.Top),
	}, true
}

// Union returns the smallest box covering both boxes.
func (b Box) Union(other Box) Box {
	a := b.Normalize()
	c := other.Normalize()
	if a.IsZero() {
		return c
	}
	if c.IsZero() {
		return a
	}
	return Box{
		Left:   math.Min(a.Left, c.Left),
		Bottom: math.Min(a.Bottom, c.Bottom),
		Right:  math.Max(a.Right, c.Right),
		Top:    math.Max(a.Top, c.Top),
	}
}

// MultiBox stores a collection of bounding boxes.
type MultiBox []Box

// Bounds returns the union of all boxes in the collection.
func (m MultiBox) Bounds() (Box, bool) {
	if len(m) == 0 {
		return Box{}, false
	}
	bounds := m[0]
	for _, box := range m[1:] {
		bounds = bounds.Union(box)
	}
	return bounds, true
}

// PageSize describes the nominal page dimensions.
type PageSize struct {
	Width  float64
	Height float64
}
