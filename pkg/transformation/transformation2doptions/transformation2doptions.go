package transformation2doptions

import (
	"fmt"
	"math"

	"github.com/nooby-gamedev/spritepacker/pkg/transformation/pointf"
)

var originalImageCacheKey = NewEmpty().CacheKey()

const DefaultScalingFactor float64 = 100

type OriginPointType byte
type FlipImage byte

const (
	OriginCenterImage OriginPointType = 0
	OriginTopLeft     OriginPointType = 1
	OriginTopRight    OriginPointType = 2
	OriginBottomLeft  OriginPointType = 3
	OriginBottomRight OriginPointType = 4
	OriginCustom      OriginPointType = 5

	DontFlipImage         FlipImage = 0
	FlipImageHorizontally FlipImage = 1
	FlipImageVertically   FlipImage = 2
)

type Position struct {
	Coordinates pointf.PointF
}
type Rotation2D struct {
	Degrees float64
}

// Scale an image by a percentage.
//
// The default values of the scaling MUST be 100 (defaultScalingValue) for both Horizontal and Vertical.
// This means the image is not scaled at all.
type Scaling2D struct {
	Horizontal float64 // Default: 100 (defaultScalingValue)
	Vertical   float64 // Default: 100 (defaultScalingValue)
}

// Always use New() to create a new Transformation2DOptions() instance.
//
// You should NEVER instance this by doing opts := Transformation2DOptions{}.
// Initializing without New() may cause calculation errors.
type Transformation2DOptions struct {
	/*
		Caching System
		Properties that transform the original image by either rotating or flipping,
		are cached.

		Position, speed and scalingFactor are NOT being cached.
	*/

	OriginPointType OriginPointType // Default: OriginCenterImage
	OriginPoint     pointf.PointF   // Custom only
	Rotation        Rotation2D      // Expressed in degrees (0° to 360°)
	FlipImage       FlipImage
	Scaling         Scaling2D

	Position      Position
	speed         float64
	scalingFactor float64 // Default value: DefaultScalingValue (100)

	// used to tell whether the Transformation2DOptions{} has been initialized correctly
	initialized bool
}

func OriginalImageCacheKey() string {
	return originalImageCacheKey
}

// Returns the difference from t and t2.
// For example, if the current Rotation is 90° and t2 has a Rotation of 135°,
// it returns 45°.
func (t *Transformation2DOptions) Difference(t2 Transformation2DOptions) Transformation2DOptions {
	if !t2.initialized {
		return *t
	}
	t.Rotation.Degrees = math.Mod(t.Rotation.Degrees, 360)
	t2.Rotation.Degrees = math.Mod(t2.Rotation.Degrees, 360)

	rotationDiff := t.Rotation.Degrees - t2.Rotation.Degrees
	flip := t.FlipImage.RemoveFlag(t2.FlipImage)

	speedDiff := t.speed - t2.speed
	posDiff := t.Position.Coordinates.Sub(t2.Position.Coordinates)

	scalingHorDiff := (t.Scaling.Horizontal - t2.Scaling.Horizontal) + DefaultScalingFactor
	scalingVerDiff := (t.Scaling.Vertical - t2.Scaling.Vertical) + DefaultScalingFactor

	return Transformation2DOptions{
		OriginPointType: t2.OriginPointType,
		OriginPoint:     t2.OriginPoint,
		Rotation:        Rotation2D{Degrees: rotationDiff},
		FlipImage:       flip,
		Scaling:         Scaling2D{Horizontal: scalingHorDiff, Vertical: scalingVerDiff},
		speed:           speedDiff,
		Position:        Position{Coordinates: posDiff},
		initialized:     true,
	}
}

func (f FlipImage) HasFlag(f2 FlipImage) bool {
	return f&f2 == f2
}
func (f FlipImage) RemoveFlag(f2 FlipImage) FlipImage {
	return f & (^f2)
}
func (f FlipImage) Swap(f2 FlipImage) FlipImage {
	if f.HasFlag(f2) {
		f = f.RemoveFlag(f2)
	} else {
		f |= f2
	}
	return f
}

// Returns true if either s.Horizontal or s.Vertical are NOT 100 (defaultScalingValue).
func (s Scaling2D) Scaled() bool {
	return s.Horizontal != DefaultScalingFactor || s.Vertical != DefaultScalingFactor
}

func DefaultScaling2D() Scaling2D {
	return Scaling2D{
		Horizontal: DefaultScalingFactor,
		Vertical:   DefaultScalingFactor,
	}
}
func New(speed float64, initialPosition pointf.PointF) *Transformation2DOptions {
	return &Transformation2DOptions{
		speed: speed,
		Position: Position{
			Coordinates: initialPosition,
		},
		Scaling:       DefaultScaling2D(),
		scalingFactor: DefaultScalingFactor,
		initialized:   true,
	}
}
func NewEmpty() *Transformation2DOptions {
	return New(0, pointf.New(0, 0))
}

// Returns a cache key that identifies a specific Transformation2DOptions state.
//
//	WARNING: Position, speed and scalingFactor are ALWAYS IGNORED in CacheKey.
func (t *Transformation2DOptions) CacheKey() string {
	originPointTypeStr := fmt.Sprintf("optt.%d", t.OriginPointType)

	originPt := pointf.New(0, 0)
	if t.OriginPointType == OriginCustom {
		originPt = t.OriginPoint
	}

	originPtStr := fmt.Sprintf("opt.%fx%f", originPt.X, originPt.Y)
	flipImgStr := fmt.Sprintf("flp.%d", t.FlipImage)

	rotationStr := fmt.Sprintf("rot.%f", t.Rotation.Degrees)
	scalingStr := fmt.Sprintf("scale.%f.%f", t.Scaling.Horizontal, t.Scaling.Vertical)

	key := fmt.Sprintf("%s.%s.%s.%s.%s", originPointTypeStr, originPtStr, flipImgStr, rotationStr, scalingStr)
	return key
}

// Set the Speed.
// If speed is less than 0, it's automatically multiplied by -1 to obtain a positive number.
func (t *Transformation2DOptions) SetSpeed(speed float64) *Transformation2DOptions {
	if speed < 0 {
		speed *= -1
	}
	t.speed = speed
	return t
}

func (t *Transformation2DOptions) Speed() float64 {
	return t.speed
}

func (t *Transformation2DOptions) Move(x, y float64) *Transformation2DOptions {
	t.Position.Coordinates.X += x
	t.Position.Coordinates.Y += y
	return t
}

// Move t.Position.Y by -t.speed * dt, where dt is the Delta Time between each Update call.
func (t *Transformation2DOptions) MoveUp(dt float64) *Transformation2DOptions {
	t.Move(0, -t.speed*dt)
	return t
}

// Move t.Position.Y by t.speed * dt, where dt is the Delta Time between each Update call.
func (t *Transformation2DOptions) MoveDown(dt float64) *Transformation2DOptions {
	t.Move(0, t.speed*dt)
	return t
}

// Move t.Position.X by -t.speed * dt, where dt is the Delta Time between each Update call.
func (t *Transformation2DOptions) MoveLeft(dt float64) *Transformation2DOptions {
	t.Move(-t.speed*dt, 0)
	return t
}

// Move t.Position.X by t.speed * dt, where dt is the Delta Time between each Update call.
func (t *Transformation2DOptions) MoveRight(dt float64) *Transformation2DOptions {
	t.Move(t.speed*dt, 0)
	return t
}

// Upscale horizontally and vertically by the scalingFactor (use t.SetScalingFactor()).
//
// For example, if the current scale if 100% horizontally and vertically (original image),
// amd we upscale by 100%, it sets scaling to 200% horizontally and vertically (image doubled).
func (t *Transformation2DOptions) Upscale() *Transformation2DOptions {
	t.Scaling.Horizontal += t.scalingFactor
	t.Scaling.Vertical += t.scalingFactor

	t.Scaling.Horizontal = math.Max(t.Scaling.Horizontal, 0)
	t.Scaling.Vertical = math.Max(t.Scaling.Vertical, 0)
	return t
}

// Downscale horizontally and vertically by the scalingFactor (use t.SetScalingFactor()).
//
// For example, if the current scale if 100% horizontally and vertically (original image),
// amd we downscale by 50%, it sets scaling to 50% horizontally and vertically (image halved).
func (t *Transformation2DOptions) Downscale() *Transformation2DOptions {
	t.Scaling.Horizontal -= t.scalingFactor
	t.Scaling.Vertical -= t.scalingFactor

	t.Scaling.Horizontal = math.Max(t.Scaling.Horizontal, 0)
	t.Scaling.Vertical = math.Max(t.Scaling.Vertical, 0)
	return t
}

func (t *Transformation2DOptions) ScalingFactor() float64 {
	return t.scalingFactor
}

// When upscaling or downscaling, it uses the set scaling factor.
// Default value: DefaultScalingValue (100)
func (t *Transformation2DOptions) SetScalingFactor(value float64) *Transformation2DOptions {
	if value < 0 {
		value *= -1
	}
	t.scalingFactor = value
	return t
}
