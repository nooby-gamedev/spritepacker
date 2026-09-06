package pointf

import "image"

type PointF struct {
	X, Y float64
}

func (p PointF) Add(p2 PointF) PointF {
	return New(p.X+p2.X, p.Y+p2.X)
}
func (p PointF) Sub(p2 PointF) PointF {
	return New(p.X-p2.X, p.Y-p2.X)
}

func New(x, y float64) PointF {
	return PointF{
		X: x,
		Y: y,
	}
}
func NewFromPoint(pt image.Point) PointF {
	return PointF{
		X: float64(pt.X),
		Y: float64(pt.Y),
	}
}
