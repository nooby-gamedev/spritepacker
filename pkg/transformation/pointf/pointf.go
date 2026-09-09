package pointf

import (
	"image"
	"math"
)

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

// It uses math.Round() to calculate Min and Max
func (p PointF) ToPointRound() image.Point {
	return image.Pt(int(math.Round(p.X)), int(math.Round(p.Y)))
}

// It uses math.Floor() to calculate Min, and math.Ceil() to calculate Max
func (p PointF) ToPointFloorCeil() image.Point {
	return image.Pt(int(math.Floor(p.X)), int(math.Ceil(p.Y)))
}
