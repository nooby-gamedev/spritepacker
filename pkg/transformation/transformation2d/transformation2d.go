package transformation2d

import (
	"image"
	"image/draw"
	"math"

	"github.com/nooby-gamedev/spritepacker/pkg/transformation/pointf"
	"github.com/nooby-gamedev/spritepacker/pkg/transformation/rectf"
	"github.com/nooby-gamedev/spritepacker/pkg/transformation/transformation2doptions"
	"github.com/rs/zerolog/log"
)

type Transformation2D struct {
	originalImage    image.Image
	transformedImage image.Image
	origin           pointf.PointF

	// It carries the current transform options.
	// It's used to determine whether transformations have to be made or not.
	//
	// For example, if it contains a Rotation of 90° and the Transform method
	// receives a Rotation of 180°, the image would be rotated by 90°.
	currentOpts transformation2doptions.Transformation2DOptions
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
	return pointF.ToPointRound()
}
func (t *Transformation2D) rotateRectF(r rectf.RectF, radians float64) rectf.RectF {
	pts := [4]pointf.PointF{
		t.rotatePointF(r.TopLeft(), radians),
		t.rotatePointF(r.TopRight(), radians),
		t.rotatePointF(r.BottomLeft(), radians),
		t.rotatePointF(r.BottomRight(), radians)}
	return rectf.NewRectFromPointF(pts, rectf.CreateRectWithValidPixelsOnly)
}
func (t *Transformation2D) rotateRect(r image.Rectangle, radians float64) image.Rectangle {
	return t.rotateRectF(rectf.NewFromRect(r), radians).ToRect()
}

// If transformedImage is NOT nil, a transformation already happened.
// If so, the originalImage will become the transformed image, so that new transformations
// will happens from that point.
//
// If transformedImage is NIL, nothing happens.
func (t *Transformation2D) setOriginalImageIfNecessary() {
	if t.transformedImage != nil {
		t.originalImage = image.NewRGBA(t.transformedImage.Bounds())
		draw.Draw(t.originalImage.(draw.Image), t.originalImage.Bounds(), t.transformedImage, image.Pt(0, 0), draw.Over)
	}
}

func (t *Transformation2D) FlipImage(flipHorizontally, flipVertically bool) {
	t.setOriginalImageIfNecessary()

	if t.transformedImage == nil {
		t.transformedImage = image.NewRGBA(t.originalImage.Bounds())
	}

	centerImage := t.centerImage(t.transformedImage).ToPointRound()

	if flipHorizontally {
		left := t.transformedImage.Bounds().Min.X
		top := t.transformedImage.Bounds().Min.Y
		bottom := t.transformedImage.Bounds().Max.Y - 1
		right := t.transformedImage.Bounds().Max.X - 1

		counter := 0
		for x := left; x < centerImage.X; x++ {
			dstRect := image.Rect(x, top, x+1, bottom+1)
			srcPt := image.Pt(right-counter, top)
			draw.Draw(t.transformedImage.(draw.Image), dstRect, t.originalImage, srcPt, draw.Over)

			dstRect2 := image.Rect(right-counter-1, top, (right - counter), bottom+1)
			srcPt2 := image.Pt(x, top)
			if !dstRect2.In(t.transformedImage.Bounds()) {
				log.Warn().Msg("no buono")
			}
			draw.Draw(t.transformedImage.(draw.Image), dstRect2, t.originalImage, srcPt2, draw.Over)
			counter++
		}
	}

	if flipVertically {
		left := t.transformedImage.Bounds().Min.X
		top := t.transformedImage.Bounds().Min.Y
		bottom := t.transformedImage.Bounds().Max.Y - 1
		right := t.transformedImage.Bounds().Max.X - 1

		counter := 0
		for y := top; y < centerImage.Y; y++ {
			dstRect := image.Rect(left, y, right+1, y+1)
			srcPt := image.Pt(left, bottom-counter)
			draw.Draw(t.transformedImage.(draw.Image), dstRect, t.originalImage, srcPt, draw.Over)

			dstRect2 := image.Rect(left, bottom-counter-1, right+1, bottom-counter)
			srcPt2 := image.Pt(left, y)
			draw.Draw(t.transformedImage.(draw.Image), dstRect2, t.originalImage, srcPt2, draw.Over)
			counter++
		}
	}
}

// Rotate the original image and returns the result.
// Rotate accepts degrees (e.g., 45°, 90°, 180° etc...)
func (t *Transformation2D) Rotate(degrees float64) {
	t.setOriginalImageIfNecessary()

	radians := t.Radians(degrees)
	log.Trace().Float64("degrees", degrees).Float64("radians", radians).Msg("rotating image")
	rect := t.rotateRect(t.originalImage.Bounds(), radians)

	left := rect.Min.X
	right := rect.Max.X
	top := rect.Min.Y
	bottom := rect.Max.Y

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
			t.transformedImage.(draw.Image).Set(dst.X, dst.Y, clr)
		}
	}
}

func (t *Transformation2D) Transform(opts transformation2doptions.Transformation2DOptions) (image.Image, error) {

	optsDiff := opts.Difference(t.currentOpts)
	t.currentOpts = opts

	switch optsDiff.OriginPointType {
	case transformation2doptions.OriginCenterImage:
		t.SetOriginToCenterImage()
	case transformation2doptions.OriginTopLeft:
		t.SetOriginToTopLeft()
	case transformation2doptions.OriginTopRight:
		t.SetOriginToTopRight()
	case transformation2doptions.OriginBottomLeft:
		t.SetOriginToBottomLeft()
	case transformation2doptions.OriginBottomRight:
		t.SetOriginToBottomRight()
	case transformation2doptions.OriginCustom:
		t.SetOrigin(t.Origin())
	default:
		return nil, ErrInvalidOriginPointType
	}

	if optsDiff.Rotation.Degrees > 0 {
		t.Rotate(opts.Rotation.Degrees)
	}

	if optsDiff.FlipImage != transformation2doptions.DontFlipImage {
		flipH := optsDiff.FlipImage.HasFlag(transformation2doptions.FlipImageHorizontally)
		flipV := optsDiff.FlipImage.HasFlag(transformation2doptions.FlipImageVertically)
		t.FlipImage(flipH, flipV)
	}

	// No transformations happened.
	// Return the original image.
	if t.transformedImage == nil {
		return t.originalImage, nil
	}

	return t.transformedImage, nil
}

/* Set Origin */

// Set the origin to the center of image.
// If the image is a sub image, it automatically recognizes the coordinates
// using img.Bounds().
func (t *Transformation2D) SetOriginToCenterImage() {
	t.origin = t.centerImage(t.originalImage)
}
func (t *Transformation2D) centerImage(img image.Image) pointf.PointF {
	left := float64(img.Bounds().Min.X)
	width := float64(img.Bounds().Dx() - 1)

	top := float64(img.Bounds().Min.Y)
	height := float64(img.Bounds().Dy() - 1)

	return pointf.New(left+(width/2), top+(height/2))
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
