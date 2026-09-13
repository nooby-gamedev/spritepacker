package transformation2d

import (
	"image"
	"image/draw"
	"math"
	"sync"

	"github.com/nooby-gamedev/spritepacker/pkg/performancemonitor"
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
	performance *performancemonitor.PerformanceMonitor
	mu          sync.RWMutex
}

// Returns a new instance of Transformation2D.
//
// This function automatically sets the origin to the center of the image.
func New(img image.Image) *Transformation2D {
	t := &Transformation2D{
		originalImage: img,
		performance:   performancemonitor.Monitor(),
	}
	t.setOriginToCenterImage()
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
func (t *Transformation2D) GetCurrentOptions() transformation2doptions.Transformation2DOptions {
	return t.currentOpts
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

func (t *Transformation2D) getSourceImage() image.Image {
	if t.transformedImage != nil {
		return t.transformedImage
	}

	return t.originalImage
}

func (t *Transformation2D) flipImage(flipHorizontally, flipVertically bool) {
	if flipHorizontally {
		dst := image.NewRGBA(t.getSourceImage().Bounds())
		centerImage := t.centerImage(dst).ToPointRound()

		left := dst.Bounds().Min.X
		top := dst.Bounds().Min.Y
		bottom := dst.Bounds().Max.Y - 1
		right := dst.Bounds().Max.X - 1

		xCounter := 0
		for x := left; x < centerImage.X; x++ {
			// let's say we have a 3x3 image where Min (0,0) and Max (3,3)
			// If we want to flip the image horizontally, we have to swap the first column
			// with the last (0 <-> 3) and the second with the second-last (1 <-> 2).
			//
			// First, we create the destination rect for the first column:
			// image.Rect(0, 0, 1, 4)
			// The reason we're not using an image.Rect(0, 0, 0, 3) is that image.Rect().Max
			// does NOT contain valid pixels.
			//
			// Original Image
			// 0 1 2 3 |
			// 1 1 2 3 |
			// 2 1 2 3 |
			// 3 1 2 3 |
			// - - - - 4,4 (Max)
			//
			// Flipped Image
			//				// dstRect | srcRect (3,0 is the srcPt)
			// 3 2 1 0 |	// 0,0:0,3 | 3,0:3,3
			// 3 2 1 1 |	// 1,0:1,3 | 2,0:2,3
			// 3 2 1 2 |
			// 3 2 1 3 |
			// - - - - 4,4 (Max)
			dstRect := image.Rect(x, top, x+1, bottom+1)
			srcPt := image.Pt(right-xCounter, top)
			draw.Draw(dst, dstRect, t.getSourceImage(), srcPt, draw.Over)

			dstRect2 := image.Rect(right-xCounter, top, right-xCounter+1, bottom+1)
			srcPt2 := image.Pt(x, top)

			draw.Draw(dst, dstRect2, t.getSourceImage(), srcPt2, draw.Over)
			xCounter++
		}
		t.transformedImage = dst
	}

	if flipVertically {
		dst := image.NewRGBA(t.getSourceImage().Bounds())
		centerImage := t.centerImage(dst).ToPointRound()

		left := dst.Bounds().Min.X
		top := dst.Bounds().Min.Y
		bottom := dst.Bounds().Max.Y - 1
		right := dst.Bounds().Max.X - 1

		yCounter := 0
		for y := top; y < centerImage.Y; y++ {
			// Vertical flip:

			// 3 1 2 3 |	// 0,0:3,0 | 0,3:3,3
			// 2 1 2 3 |	// 0,1:3,1 | 2,2,3,2
			// 1 1 2 3 |
			// 0 1 2 3 |
			// - - - - 4,4 (Max)
			dstRect := image.Rect(left, y, right+1, y+1)
			srcPt := image.Pt(left, bottom-yCounter)
			draw.Draw(dst, dstRect, t.getSourceImage(), srcPt, draw.Over)

			dstRect2 := image.Rect(left, bottom-yCounter, right+1, bottom-yCounter+1)
			srcPt2 := image.Pt(left, y)
			draw.Draw(dst, dstRect2, t.getSourceImage(), srcPt2, draw.Over)
			yCounter++
		}
		t.transformedImage = dst
	}
}

// rotate the original image and returns the result.
// rotate accepts degrees (e.g., 45°, 90°, 180° etc...)
func (t *Transformation2D) rotate(degrees float64) {
	radians := t.Radians(degrees)
	log.Trace().Float64("degrees", degrees).Float64("radians", radians).Msg("rotating image")
	rect := t.rotateRect(t.originalImage.Bounds(), radians)

	left := rect.Min.X
	right := rect.Max.X
	top := rect.Min.Y
	bottom := rect.Max.Y

	dst := image.NewRGBA(rect)

	for x := left; x < right; x++ {
		for y := top; y < bottom; y++ {
			dstPt := image.Pt(x, y)
			src := t.rotatePoint(image.Pt(x, y), radians*-1)
			if !src.In(t.getSourceImage().Bounds()) {
				continue
			}
			clr := t.getSourceImage().At(src.X, src.Y)
			dst.Set(dstPt.X, dstPt.Y, clr)
		}
	}
	t.transformedImage = dst
}

func (t *Transformation2D) Transform(opts transformation2doptions.Transformation2DOptions) (image.Image, error) {
	t.performance.StartMeasureAverageDeltaTime("transform.transforming")
	defer t.performance.StopMeasureAverageDeltaTime("transform.transforming")

	t.mu.Lock()
	optsDiff := opts.Difference(t.currentOpts)
	t.currentOpts = opts
	t.mu.Unlock()

	switch optsDiff.OriginPointType {
	case transformation2doptions.OriginCenterImage:
		t.setOriginToCenterImage()
	case transformation2doptions.OriginTopLeft:
		t.setOriginToTopLeft()
	case transformation2doptions.OriginTopRight:
		t.setOriginToTopRight()
	case transformation2doptions.OriginBottomLeft:
		t.setOriginToBottomLeft()
	case transformation2doptions.OriginBottomRight:
		t.setOriginToBottomRight()
	case transformation2doptions.OriginCustom:
		t.setOrigin(t.Origin())
	default:
		return nil, ErrInvalidOriginPointType
	}

	if optsDiff.Rotation.Degrees > 0 {
		t.performance.StartMeasureAverageDeltaTime("transform.rotate")
		t.rotate(opts.Rotation.Degrees)
		t.performance.StopMeasureAverageDeltaTime("transform.rotate")
	}

	if optsDiff.FlipImage != transformation2doptions.DontFlipImage {
		t.performance.StartMeasureAverageDeltaTime("transform.flip")
		flipH := optsDiff.FlipImage.HasFlag(transformation2doptions.FlipImageHorizontally)
		flipV := optsDiff.FlipImage.HasFlag(transformation2doptions.FlipImageVertically)
		t.flipImage(flipH, flipV)
		t.performance.StopMeasureAverageDeltaTime("transform.flip")
	}

	// No transformations happened.
	// TransformedImage is equal to OriginalImage
	if t.transformedImage == nil {
		t.transformedImage = t.originalImage
	}

	return t.transformedImage, nil
}

/* Set Origin */

// Set the origin to the center of image.
// If the image is a sub image, it automatically recognizes the coordinates
// using img.Bounds().
func (t *Transformation2D) setOriginToCenterImage() {
	t.origin = t.centerImage(t.originalImage)
}
func (t *Transformation2D) centerImage(img image.Image) pointf.PointF {
	left := float64(img.Bounds().Min.X)
	width := float64(img.Bounds().Dx() - 1)

	top := float64(img.Bounds().Min.Y)
	height := float64(img.Bounds().Dy() - 1)

	return pointf.New(left+(width/2), top+(height/2))
}
func (t *Transformation2D) setOriginToTopLeft() {
	t.origin.X = float64(t.originalImage.Bounds().Min.X)
	t.origin.Y = float64(t.originalImage.Bounds().Min.Y)
}
func (t *Transformation2D) setOriginToTopRight() {
	t.origin.X = float64(t.originalImage.Bounds().Max.X - 1)
	t.origin.Y = float64(t.originalImage.Bounds().Min.Y)
}
func (t *Transformation2D) setOriginToBottomLeft() {
	t.origin.X = float64(t.originalImage.Bounds().Min.X)
	t.origin.Y = float64(t.originalImage.Bounds().Max.Y - 1)
}
func (t *Transformation2D) setOriginToBottomRight() {
	t.origin.X = float64(t.originalImage.Bounds().Max.X - 1)
	t.origin.Y = float64(t.originalImage.Bounds().Max.Y - 1)
}
func (t *Transformation2D) setOrigin(origin pointf.PointF) {
	t.origin.X = origin.X
	t.origin.Y = origin.Y
}
