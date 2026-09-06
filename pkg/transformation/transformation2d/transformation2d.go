package transformation2d

import (
	"image"
	"image/color"
	"math"

	"github.com/nooby-gamedev/spritepacker/pkg/transformation/pointf"
	"github.com/nooby-gamedev/spritepacker/pkg/transformation/rectf"
	"github.com/rs/zerolog/log"
)

type Transformation2D struct {
	originalImage    image.Image
	transformedImage image.Image
	origin           pointf.PointF
}

// Returns a new instance of Transformation2D.
//
// This function automatically sets the origin to the center of the image.
func New(img image.Image) *Transformation2D {
	t := &Transformation2D{
		originalImage: img,
	}
	t.SetOriginToCenterImage()
	return t
}

func (t *Transformation2D) Origin() pointf.PointF {
	return t.origin
}
func (t *Transformation2D) OriginalImage() image.Image {
	return t.originalImage
}

func (t *Transformation2D) nearestNeighborPointF(pt pointf.PointF) image.Point {
	return image.Pt(int(math.Round(pt.X)), int(math.Round(pt.Y)))
}
func (t *Transformation2D) nearestNeighborRectF(r rectf.RectF) image.Rectangle {
	min := t.nearestNeighborPointF(r.Min)
	max := t.nearestNeighborPointF(r.Max)
	return image.Rect(min.X, min.Y, max.X, max.Y)
}

// Returns the transformed image
func (t *Transformation2D) Image() image.Image {
	return t.transformedImage
}

func (t *Transformation2D) Radians(degrees float64) float64 {
	return (degrees * math.Pi) / 180
}
func (t *Transformation2D) Degrees(radians float64) float64 {
	return (radians * 180) / math.Pi
}

// The value of rotation passed to rotate() must be expressed in radians.
//
// Use Radians(degrees float64) to obtain the value from degrees.
func (t *Transformation2D) rotatePointF(pt pointf.PointF, radians float64) pointf.PointF {
	pt = pt.Sub(t.origin)
	cos := math.Cos(radians)
	sin := math.Sin(radians)
	x := (pt.X * cos) - (pt.Y * sin)
	y := (pt.X * sin) + (pt.Y * cos)
	return pointf.New(x, y).Add(t.origin)
}
func (t *Transformation2D) rotatePoint(pt image.Point, radians float64) image.Point {
	pointF := t.rotatePointF(pointf.NewFromPoint(pt), radians)
	return t.nearestNeighborPointF(pointF)
}
func (t *Transformation2D) rotateRectF(r rectf.RectF, radians float64) rectf.RectF {
	pts := [4]pointf.PointF{
		t.rotatePointF(r.TopLeft(), radians),
		t.rotatePointF(r.TopRight(), radians),
		t.rotatePointF(r.BottomLeft(), radians),
		t.rotatePointF(r.BottomRight(), radians)}
	return rectf.NewRectFromPointF(pts)
}
func (t *Transformation2D) rotateRect(r image.Rectangle, radians float64) image.Rectangle {
	rectF := t.rotateRectF(rectf.NewFromRect(r), radians)
	r2 := t.nearestNeighborRectF(rectF)
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
	r2.Max = r2.Max.Add(image.Pt(1, 1))
	return r2
}

// Rotate the original image and returns the result.
// Rotate accepts degrees (e.g., 45°, 90°, 180° etc...)
func (t *Transformation2D) Rotate(degrees float64) image.Image {
	radians := t.Radians(degrees)
	log.Trace().Float64("degrees", degrees).Float64("radians", radians).Msg("rotating image")
	rect := t.rotateRect(t.originalImage.Bounds(), radians)

	left := rect.Min.X
	right := rect.Max.X
	top := rect.Min.Y
	bottom := rect.Max.Y

	// temp
	if t.transformedImage == nil {
		t.transformedImage = image.NewRGBA(rect)
	}

	for x := left; x < right; x++ {
		for y := top; y < bottom; y++ {
			dst := image.Pt(x, y)
			src := t.rotatePoint(image.Pt(x, y), radians*-1)
			if !src.In(t.originalImage.Bounds()) {
				continue
			}
			clr := t.originalImage.At(src.X, src.Y)
			t.transformedImage.(interface{ Set(int, int, color.Color) }).Set(dst.X, dst.Y, clr)
		}
	}

	return t.transformedImage
}

/* Set Origin */

// Set the origin to the center of image.
// If the image is a sub image, it automatically recognizes the coordinates
// using img.Bounds().
func (t *Transformation2D) SetOriginToCenterImage() {
	left := float64(t.originalImage.Bounds().Min.X)
	width := float64(t.originalImage.Bounds().Dx() - 1)

	top := float64(t.originalImage.Bounds().Min.Y)
	height := float64(t.originalImage.Bounds().Dy() - 1)

	t.origin.X = left + (width / 2)
	t.origin.Y = top + (height / 2)
}
func (t *Transformation2D) SetOriginToTopLeft() {
	t.origin.X = float64(t.originalImage.Bounds().Min.X)
	t.origin.Y = float64(t.originalImage.Bounds().Min.Y)
}
func (t *Transformation2D) SetOriginToTopRight() {
	t.origin.X = float64(t.originalImage.Bounds().Max.X - 1)
	t.origin.Y = float64(t.originalImage.Bounds().Min.Y)
}
func (t *Transformation2D) SetOriginToBottomLeft() {
	t.origin.X = float64(t.originalImage.Bounds().Min.X)
	t.origin.Y = float64(t.originalImage.Bounds().Max.Y - 1)
}
func (t *Transformation2D) SetOriginToBottomRight() {
	t.origin.X = float64(t.originalImage.Bounds().Max.X - 1)
	t.origin.Y = float64(t.originalImage.Bounds().Max.Y - 1)
}
func (t *Transformation2D) SetOrigin(origin pointf.PointF) {
	t.origin.X = origin.X
	t.origin.Y = origin.Y
}
