package rectf

import (
	"image"

	"github.com/nooby-gamedev/spritepacker/pkg/transformation/pointf"
)

// RectF contains only real pixels, while image.Rectangle doesn't.
//
// A 10x10 image would be:
//
// RectF: 0,0 -> 9,9
// image.Rectangle: 0,0 -> 10,10
//
// but the 10,10 pixel are NOT usable.
//
// RectF uses pixels only for calculations.
type RectF struct {
	Min pointf.PointF
	Max pointf.PointF
}

func New(min, max pointf.PointF) RectF {
	return RectF{
		Min: min,
		Max: max,
	}
}

func NewRectFromPointF(pts [4]pointf.PointF) RectF {
	var minX, minY, maxX, maxY float64
	for i, pt := range pts {
		if i == 0 {
			minX = pt.X
			minY = pt.Y
			maxX = pt.X
			maxY = pt.Y
		}

		minX = min(minX, pt.X)
		minY = min(minY, pt.Y)

		maxX = max(maxX, pt.X)
		maxY = max(maxY, pt.Y)
	}

	return New(pointf.New(minX, minY), pointf.New(maxX, maxY))
}

// RectF contains only real pixels, while image.Rectangle cdoesn'y.
// A 10x10 image would be:
// RectF: 0,0 -> 9,9
// image.Rectangle: 0,0 -> 10,10
// but the 10,10 pixel are NOT usable.
//
// RectF uses pixels only for calculations.
func NewFromRect(r image.Rectangle) RectF {
	return New(pointf.NewFromPoint(r.Min), pointf.NewFromPoint(r.Max.Sub(image.Pt(1, 1))))
}

func (r RectF) TopLeft() pointf.PointF {
	return pointf.New(r.Min.X, r.Min.Y)
}
func (r RectF) TopRight() pointf.PointF {
	return pointf.New(r.Max.X, r.Min.Y)
}
func (r RectF) BottomLeft() pointf.PointF {
	return pointf.New(r.Min.X, r.Max.Y)
}
func (r RectF) BottomRight() pointf.PointF {
	return pointf.New(r.Max.X, r.Max.Y)
}
