package pack

import (
	"fmt"
	"image"

	"github.com/nooby-gamedev/spritepacker/internal/coordinates"
	"github.com/nooby-gamedev/spritepacker/internal/imageinfo"
	"github.com/nooby-gamedev/spritepacker/internal/size"

	"github.com/rs/zerolog/log"
)

type BoxType string

const (
	MainBox   BoxType = "main"   // The very first box created
	RightBox  BoxType = "right"  // Right free box
	BottomBox BoxType = "bottom" // Bottom free box
)

// A Pack is composed of boxes.
type Box struct {
	// Can be used to recognize different boxes.
	// By setting id to the packCoordinates, it becomes easier to troubleshoot.
	id string
	// Defines the packCoordinates where the box is placed
	packCoordinates coordinates.Coordinates
	size            size.Size

	imgInfo *imageinfo.ImageInfo

	// For debugging purpose
	boxType BoxType
}

func newBox(size size.Size, packCoordinates coordinates.Coordinates, boxType BoxType) *Box {

	id := fmt.Sprintf("%dx%d", packCoordinates.X, packCoordinates.Y)
	return &Box{
		id:              id, //uuid.NewString(),
		size:            size,
		packCoordinates: packCoordinates,
		boxType:         boxType,
	}
}

// The box size MUST be able to contain the imgInfo.
// If it doesn't, the application panics.
// If imgInfo is nil, the application panics.
//
// This method sets automatically the size of box to match exactly the imgInfo size.
//
// After setting the image, the method returns an array of all free slots created (min 0, max 2)
func (b *Box) setImageInfo(imgInfo *imageinfo.ImageInfo) (freeBoxes []*Box) {
	if imgInfo == nil {
		log.Fatal().Msg("fatal error: imgInfo cannot be nil on calling setImageInfo")
	}

	if canFit, _ := b.canFit(imgInfo); !canFit {
		log.Fatal().
			Int("img_width", imgInfo.Width()).
			Int("img_height", imgInfo.Height()).
			Int("box_width", b.width()).
			Int("box_height", b.height()).
			Str("full_path", imgInfo.Path).
			Msg("fatal error: imgInfo size cannot be wider or higher than the box")
	}

	freeBoxes = make([]*Box, 0)

	// Below the logic of how free boxes are calculated
	/*
		// Free box on the right (free r)
		// and free box on the bottom (free b)
		 _ _ _ _ _ _ _ _ _ _
		|	 img	| free r|
		| _ _ _ _ _ | _ _ _	|
		|		free b		|
		| _ _ _ _ _ _ _ _ _ |

		// Free box on the bottom
		 _ _ _ _ _ _ _ _ _ _
		|	 	 img		|
		| _ _ _ _ _ _ _ _ _	|
		|		free b		|
		| _ _ _ _ _ _ _ _ _ |

		// Free box on the right (free r)
		 _ _ _ _ _ _ _ _ _ _
		|	 img	| free r|
		| 			| 		|
		|			|		|
		| _ _ _ _ _ | _ _ _ |
	*/

	/* Calculate the free box on the right */

	calculateFreeBoxRight := func() {
		rX := imgInfo.Width()
		rWidth := b.width() - rX

		rY := 0
		rHeight := imgInfo.Height()

		if rWidth > 0 {
			box := newBox(
				size.New(rWidth, rHeight),
				coordinates.New(rX+b.left(), rY+b.top()),
				RightBox,
			)
			freeBoxes = append(freeBoxes, box)
		}
	}

	/* Calculate the free box on the bottom */
	calculateFreeBoxBottom := func() {
		bX := 0
		bWidth := b.width()

		bY := imgInfo.Height()
		bHeight := b.height() - bY

		if bHeight > 0 {
			box := newBox(
				size.New(bWidth, bHeight),
				coordinates.New(bX+b.left(), bY+b.top()),
				BottomBox,
			)
			freeBoxes = append(freeBoxes, box)
		}
	}

	calculateFreeBoxRight()
	calculateFreeBoxBottom()

	b.imgInfo = imgInfo
	b.size.Width = imgInfo.Width()
	b.size.Height = imgInfo.Height()

	return freeBoxes
}

// Check if imgInfo can fit in the current box.
//
// Panics if imgInfo is nil.
func (b *Box) canFit(imgInfo *imageinfo.ImageInfo) (canFit bool, totCoveredArea float32) {
	if imgInfo == nil {
		log.Fatal().Msg("fatal error: imgInfo cannot be nil on calling canFit")
	}

	canFit = b.width() >= imgInfo.Width() && b.height() >= imgInfo.Height()
	if !canFit {
		return false, 0
	}

	boxArea := float32(b.area())
	imgArea := float32(imgInfo.Area())

	boxWidth := float32(b.width())
	boxHeight := float32(b.height())

	imgWidth := float32(imgInfo.Width())
	imgHeight := float32(imgInfo.Height())

	coveredWidth := (100 / boxWidth) * imgWidth
	coveredHeight := (100 / boxHeight) * imgHeight
	coveredArea := (100 / boxArea) * imgArea

	return true, (coveredWidth + coveredHeight + coveredArea) / 3
}

func (b *Box) imageHasBeenSet() bool {
	return b.imgInfo != nil
}

func (b *Box) area() int {
	return b.size.Area()
}

func (b *Box) isInRange(x, y int) bool {
	return x >= 0 && x < b.width() && y >= 0 && y < b.height()
}

func (b *Box) absoluteCoordinates(x, y int) coordinates.Coordinates {
	return coordinates.New(x+b.packCoordinates.X, y+b.packCoordinates.Y)
}

// Return true if the RELATIVE coordinates are the borders of the box
func (b *Box) isBorder(x, y int) bool {
	borderX := x == 0 || x == b.width()-1
	borderY := y == 0 || y == b.height()-1
	return borderX || borderY
}

// Returns b.packCoordinates.X
func (b *Box) left() int {
	return b.packCoordinates.X
}

// Returns b.packCoordinates.Y
func (b *Box) top() int {
	return b.packCoordinates.Y
}

// Returns Width
func (b *Box) width() int {
	return b.size.Width
}

// Returns Height
func (b *Box) height() int {
	return b.size.Height
}

// Returns b.packCoordinates.X + b.size.Width
func (b *Box) right() int {
	return b.left() + b.width()
}

// Returns b.packCoordinates.Y + b.size.Height
func (b *Box) bottom() int {
	return b.top() + b.height()
}

// Returns the image name
// Panics if image has not been set.
func (b *Box) imageName() string {
	if b.imgInfo == nil {
		log.Fatal().Msg("fatal error: imgInfo cannot be nil on calling imageName")
	}
	return b.imgInfo.Name
}

// Returns the image animation group
// Panics if image has not been set.
func (b *Box) animationGroup() string {
	if b.imgInfo == nil {
		log.Fatal().Msg("fatal error: imgInfo cannot be nil on calling animationGroup")
	}
	return b.imgInfo.AnimationGroup
}

// Releases all the resources
func (b *Box) close() {
	if b.imgInfo == nil {
		return
	}
	b.imgInfo.Close()
	b.imgInfo = nil
}

func (b *Box) image() image.Image {
	return b.imgInfo.Image()
}
