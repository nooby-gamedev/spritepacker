package transformation2doptions

import (
	"math"

	"github.com/nooby-gamedev/spritepacker/pkg/transformation/pointf"
)

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

type Translation2D struct {
	MoveBy pointf.PointF
}
type Rotation2D struct {
	Degrees float64
}

type Transformation2DOptions struct {
	OriginPointType OriginPointType // Default: OriginCenterImage
	OriginPoint     pointf.PointF   // Custom only

	Rotation  Rotation2D // Expressed in degrees (0° to 360°)
	FlipImage FlipImage
}

// Returns the difference from t and t2.
// For example, if the current Rotation is 90° and t2 has a Rotation of 135°,
// it returns 45°.
func (t Transformation2DOptions) Difference(t2 Transformation2DOptions) Transformation2DOptions {
	t.Rotation.Degrees = math.Mod(t.Rotation.Degrees, 360)
	t2.Rotation.Degrees = math.Mod(t2.Rotation.Degrees, 360)

	rotationDiff := t.Rotation.Degrees - t2.Rotation.Degrees
	flip := t.FlipImage.RemoveFlag(t2.FlipImage)

	return Transformation2DOptions{
		OriginPointType: t2.OriginPointType,
		OriginPoint:     t2.OriginPoint,
		Rotation:        Rotation2D{Degrees: rotationDiff},
		FlipImage:       flip,
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
