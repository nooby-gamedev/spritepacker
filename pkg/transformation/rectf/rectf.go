package rectf

import (
	"image"

	"github.com/nooby-gamedev/spritepacker/pkg/transformation/pointf"
)

const (
	// An image 10x10 will have a rect with Min (0,0) to Max (9,9).
	// Both Min an Max will contain valid pixels.
	CreateRectWithValidPixelsOnly bool = true
	// An image 10x10 will have a rect with Min (0,0) to Max (10,10) (reflect image.Rectangle behavior).
	// Min will contain valid pixels, while Max will NOT (last valid pixel: Max(9,9))
	CreateRectWithExclusiveBorders bool = false
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

	// if Max contains the coordinates of valid pixels, must be true.
	// For example: image 10x10.
	// Max(9,9) -> TRUE.
	// Max(10,10) -> FALSE.
	//
	// See consts: CreateRectWithValidPixelsOnly amd CreateRectWithExclusiveBorders
	bordersContainValidPixels bool
}

// if Max contains the coordinates of valid pixels, bordersContainValidPixels must be true.
// For example: image 10x10.
// Max(9,9) -> TRUE.
// Max(10,10) -> FALSE.
func New(min, max pointf.PointF, bordersContainValidPixels bool) RectF {
	return RectF{
		Min:                       min,
		Max:                       max,
		bordersContainValidPixels: bordersContainValidPixels,
	}
}

func NewRectFromPointF(pts [4]pointf.PointF, bordersContainValidPixels bool) RectF {
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

	return New(pointf.New(minX, minY), pointf.New(maxX, maxY), bordersContainValidPixels)
}

// RectF contains only real pixels, while image.Rectangle cdoesn'y.
// A 10x10 image would be:
// RectF: 0,0 -> 9,9
// image.Rectangle: 0,0 -> 10,10
// but the 10,10 pixel are NOT usable.
//
// RectF uses pixels only for calculations.
func NewFromRect(r image.Rectangle) RectF {
	return New(pointf.NewFromPoint(r.Min), pointf.NewFromPoint(r.Max.Sub(image.Pt(1, 1))), CreateRectWithValidPixelsOnly)
}

// If the RectF has borders with valid pixels (bordersContainValidPixels=true),
// an image.Pt(1,1) will be added to the returned Max().
//
// For example, for an image 10x10:
//
// RectF Min(0,0) Max(9,9) (bordersContainValidPixels=true): return image.Rect(0,0,10,10)
//
// RectF Min(0,0) Max(10,10) (bordersContainValidPixels=false): return image.Rect(0,0,10,10)
func (r RectF) ToRect() image.Rectangle {
	min := r.Min.ToPointRound()
	max := r.Max.ToPointRound()

	if r.bordersContainValidPixels {
		max = max.Add(image.Pt(1, 1))
	}
	return image.Rect(min.X, min.Y, max.X, max.Y)
}

// Return Min.X, Min.Y
func (r RectF) TopLeft() pointf.PointF {
	return pointf.New(r.Min.X, r.Min.Y)
}

// Return Max.X, Min.Y
func (r RectF) TopRight() pointf.PointF {
	return pointf.New(r.Max.X, r.Min.Y)
}

// Return Min.X, Max.Y
func (r RectF) BottomLeft() pointf.PointF {
	return pointf.New(r.Min.X, r.Max.Y)
}

// Return Max.X, Max.Y
func (r RectF) BottomRight() pointf.PointF {
	return pointf.New(r.Max.X, r.Max.Y)
}
