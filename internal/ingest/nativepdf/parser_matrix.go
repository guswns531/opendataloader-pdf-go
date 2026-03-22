package nativepdf

import "github.com/guswns531/opendataloader-pdf-go/internal/model"

type matrix2D struct {
	a float64
	b float64
	c float64
	d float64
	e float64
	f float64
}

func identityMatrix() matrix2D {
	return matrix2D{a: 1, d: 1}
}

func (m matrix2D) multiply(other matrix2D) matrix2D {
	return matrix2D{
		a: m.a*other.a + m.c*other.b,
		b: m.b*other.a + m.d*other.b,
		c: m.a*other.c + m.c*other.d,
		d: m.b*other.c + m.d*other.d,
		e: m.a*other.e + m.c*other.f + m.e,
		f: m.b*other.e + m.d*other.f + m.f,
	}
}

func (m matrix2D) transformPoint(p model.Point) model.Point {
	return model.Point{
		X: m.a*p.X + m.c*p.Y + m.e,
		Y: m.b*p.X + m.d*p.Y + m.f,
	}
}

func (m matrix2D) transformVector(p model.Point) model.Point {
	return model.Point{
		X: m.a*p.X + m.c*p.Y,
		Y: m.b*p.X + m.d*p.Y,
	}
}

func (m matrix2D) transformBox(box model.Box) model.Box {
	box = box.Normalize()
	points := []model.Point{
		{X: box.Left, Y: box.Bottom},
		{X: box.Left, Y: box.Top},
		{X: box.Right, Y: box.Bottom},
		{X: box.Right, Y: box.Top},
	}
	return boxFromPoints([]model.Point{
		m.transformPoint(points[0]),
		m.transformPoint(points[1]),
		m.transformPoint(points[2]),
		m.transformPoint(points[3]),
	})
}

func boxFromPoints(points []model.Point) model.Box {
	if len(points) == 0 {
		return model.Box{}
	}
	bounds := model.Box{
		Left:   points[0].X,
		Right:  points[0].X,
		Bottom: points[0].Y,
		Top:    points[0].Y,
	}
	for _, p := range points[1:] {
		if p.X < bounds.Left {
			bounds.Left = p.X
		}
		if p.X > bounds.Right {
			bounds.Right = p.X
		}
		if p.Y < bounds.Bottom {
			bounds.Bottom = p.Y
		}
		if p.Y > bounds.Top {
			bounds.Top = p.Y
		}
	}
	return bounds
}

func rectangleFromSize(width, height float64) model.Box {
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}
	return model.Box{Left: 0, Bottom: 0, Right: width, Top: height}
}
