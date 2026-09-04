package pack

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"sync"

	"github.com/nooby-gamedev/spritepacker/internal/coordinates"
	"github.com/nooby-gamedev/spritepacker/internal/imageinfo"
	"github.com/nooby-gamedev/spritepacker/internal/size"
	"github.com/nooby-gamedev/spritepacker/pkg/extensions"
	"github.com/nooby-gamedev/spritepacker/pkg/spritepack"

	"github.com/rs/zerolog/log"
)

type PackOrientation string

const (
	Horizontal PackOrientation = "horizontal"
	Vertical   PackOrientation = "vertical"

	DrawBoxBorders     bool = true
	DontDrawBoxBorders bool = false
)

type Pack struct {
	// Defines the packCoordinates of the pack (usually 0,0)
	packCoordinates coordinates.Coordinates

	// Defines the initial pack size (calculated automatically)
	packSize size.Size

	// Defines the final pack size (calculated automatically).
	//
	// This value gets updated after each image is inserted,
	// and contains the rightmost X coordinate consumed by boxes, and the
	// bottommost Y coordinate consumed by boxes.
	//
	// Starts from Width = 0, Height = 0
	consumedPackSize size.Size

	// Width represents the size of the widest image detected by the Loader.
	// Height represents the size of the highest image detected by the loader.
	//
	// When packOrientation is vertical, the maxImageSize.width is used as
	// the REAL width of the pack, while height is ignored.
	//
	// When packOrientation is horizontal, the maxImageSize.height is used instead.
	maxImageSize size.Size

	// Width represents the sum of width of all images, while
	// Height represents the sum of height of all images.
	//
	// When packOrientation is vertical, the totImageSize.height is used as
	// the temporary height of the pack, while width is ignored.
	//
	// When packOrientation is horizontal, the totImageSize.width is used instead.
	totImageSize size.Size

	// Defines the orientation of the pack.
	//
	// When the maxImageSize.width is >= than maxImageSize.height, the orientation
	// will be Vertical, as we already know the max width necessary.
	//
	// Otherwise, the orientation will be Horizontal.
	packOrientation PackOrientation

	// Contains all the free boxes (updated automatically)
	freeBoxes map[string]*Box

	// Contains all the boxes with the image set
	consumedBoxes map[string]*Box

	// If true, the final image will have box borders (dev only)
	drawBoxBorders bool
	mu             sync.Mutex
}

func New(maxImageSize, totImageSize size.Size, packCoordinates coordinates.Coordinates, drawBoxBorders bool) *Pack {
	var packSize size.Size
	var orientation PackOrientation

	if maxImageSize.Width >= maxImageSize.Height {
		orientation = Vertical
		packSize = size.New(maxImageSize.Width, totImageSize.Height)
	} else {
		orientation = Horizontal
		packSize = size.New(totImageSize.Width, maxImageSize.Height)
	}

	p := &Pack{
		maxImageSize:     maxImageSize,
		totImageSize:     totImageSize,
		packCoordinates:  packCoordinates,
		packOrientation:  orientation,
		packSize:         packSize,
		consumedPackSize: size.New(0, 0),
		freeBoxes:        make(map[string]*Box, 0),
		consumedBoxes:    make(map[string]*Box, 0),
		drawBoxBorders:   drawBoxBorders,
	}

	log.Debug().
		Str("maxImgSize", p.maxImageSize.ToString()).
		Str("totImageSize", p.totImageSize.ToString()).
		Str("packOrientation", string(p.packOrientation)).
		Msg("new pack initialized")
	p.setInitialFreeBox()

	return p
}

// Insert the image in the pack.
//
// If the image cannot be inserted, it panics.
func (p *Pack) InsertImage(imgInfo *imageinfo.ImageInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.freeBoxes) == 0 {
		log.Fatal().Msg("fatal error: cannot insert image as there are no free boxes left")
	}

	var selectedBox *Box
	var boxCoveredArea float32
	var wouldUpdateConsumedPackSize bool

	for _, freeBox := range p.freeBoxes {
		canFit, coveredArea := freeBox.canFit(imgInfo)
		if !canFit {
			log.Debug().
				Int("img_width", imgInfo.Width()).
				Int("img_height", imgInfo.Height()).
				Int("box_x", freeBox.x()).
				Int("box_y", freeBox.y()).
				Int("box_width", freeBox.width()).
				Int("box_height", freeBox.height()).
				Str("box_type", string(freeBox.boxType)).
				Msg("box cannot fit image")
			continue
		}
		if selectedBox == nil {
			boxCoveredArea = coveredArea
			selectedBox = freeBox
			wouldUpdateConsumedPackSize = p.wouldUpdateConsumedPackSize(freeBox, imgInfo)
			continue
		}

		// If the previous box would've updated the "consumedPackSize", but the
		// current box doesn't, then use the current box
		if wouldUpdateConsumedPackSize && !p.wouldUpdateConsumedPackSize(freeBox, imgInfo) {
			boxCoveredArea = coveredArea
			selectedBox = freeBox
			wouldUpdateConsumedPackSize = false
			continue
		}

		if coveredArea > boxCoveredArea {
			boxCoveredArea = coveredArea
			selectedBox = freeBox
			wouldUpdateConsumedPackSize = p.wouldUpdateConsumedPackSize(freeBox, imgInfo)
		}
	}

	if selectedBox == nil {
		log.Fatal().
			Int("img_width", imgInfo.Width()).
			Int("img_height", imgInfo.Height()).
			Int("free_box_count", len(p.freeBoxes)).
			Int("consumed_box_count", len(p.consumedBoxes)).
			Str("full_path", imgInfo.Path).
			Msg("fatal error: the image cannot fit into any box")
	}

	log.Debug().
		Int("box_x", selectedBox.x()).
		Int("box_y", selectedBox.y()).
		Int("box_width", selectedBox.width()).
		Int("box_height", selectedBox.height()).
		Str("box_type", string(selectedBox.boxType)).
		Str("image_full_path", imgInfo.Path).
		Int("img_width", imgInfo.Width()).
		Int("img_height", imgInfo.Height()).
		Float32("box_covered_area", boxCoveredArea).
		Msg("a free box has been selected for the image")

	newFreeBoxes := selectedBox.setImageInfo(imgInfo)

	p.consumeFreeBox(selectedBox)
	p.addFreeBoxes(newFreeBoxes...)
	p.updateConsumedPackSize(selectedBox)
}

// Creates the final packed image as PNG.
// If the .png extension is not present, it's automatically added.
//
// In case of any error, it panics.
func (p *Pack) Finalize(savePath string) {

	spritePack := spritepack.New()
	var savePathPng string

	savePathPng = savePath

	if !strings.HasSuffix(savePathPng, string(extensions.Png)) {
		savePathPng += string(extensions.Png)
	}

	if p.drawBoxBorders {
		log.Warn().Msg("warning: drawBoxBorders is true. The final image will have box borders.")
	}

	handle, err := os.Create(savePathPng)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("fatal error: unable to finalize")
	}
	defer handle.Close()

	log.Debug().
		Str("pack_size", p.consumedPackSize.ToString()).
		Msg("finalizing pack")

	imgSize := image.Rect(0, 0, p.width(), p.height())
	img := image.NewNRGBA(imgSize)

	for _, box := range p.consumedBoxes {
		for x := range box.width() {
			for y := range box.height() {
				clr, absoluteCoordinates := box.colorAt(x, y)

				if p.drawBoxBorders {
					if box.isBorder(x, y) {
						clr = color.RGBA{R: 0, G: 255, B: 255, A: 255}
					}
				}

				img.Set(absoluteCoordinates.X, absoluteCoordinates.Y, clr)
			}
		}
		spritePack.NewSprite(box.imageName(), box.x(), box.y(), box.width(), box.height())
		box.close()
	}

	if p.drawBoxBorders {
		if len(p.freeBoxes) > 0 {
			log.Debug().Msg("drawing box borders of free boxes")
		}
		for _, box := range p.freeBoxes {
			for x := range box.width() {
				for y := range box.height() {
					if !box.isInRange(x, y) {
						continue
					}

					var clr color.Color
					absoluteCoordinates := box.absoluteCoordinates(x, y)

					if box.isBorder(x, y) {
						clr = color.RGBA{R: 255, G: 0, B: 255, A: 255}
					} else {
						continue
					}

					imgBounds := img.Bounds()
					isInRange :=
						absoluteCoordinates.X >= 0 && absoluteCoordinates.X < imgBounds.Dx() &&
							absoluteCoordinates.Y >= 0 && absoluteCoordinates.Y < imgBounds.Dy()

					if !isInRange {
						continue
					}
					img.Set(absoluteCoordinates.X, absoluteCoordinates.Y, clr)
				}
			}
			box.close()
		}
	}

	if err := png.Encode(handle, img); err != nil {
		log.Fatal().
			Err(err).
			Str("save_path", savePathPng).
			Msg("fatal error: unable to save the image as an unexpected error occurred")
	}

	log.Info().Str("save_path", savePathPng).Msg("Pack created successfully!")

	if err := spritePack.Export(savePath); err != nil {
		log.Error().Err(err).Msg("unexpected error occurred")
	}
}

// Set the initial Free Box (same dimension as the pack).
//
// If freeBoxes already contains elements, it panics.
func (p *Pack) setInitialFreeBox() {
	if len(p.freeBoxes) > 0 {
		log.Fatal().Int("num_of_free_boxes", len(p.freeBoxes)).Msg("fatal error: cannot set initial free box, as the list already contains elements")
	}

	box := newBox(p.packSize, p.packCoordinates, MainBox)
	p.addFreeBoxes(box)
}

// Update the consumedPackSize.
// This method must be called after the image has been set in box.
// If the image has not been set yet, it panics.
//
// If the box dx (x + width) is wider than the consumedPackSize width, it gets updated.
//
// If the box dy (y + height) is higher than the consumedPackSize height, it gets updated.
func (p *Pack) updateConsumedPackSize(box *Box) {
	if !box.imageHasBeenSet() {
		log.Fatal().Str("box_id", box.id).Msg("fatal error: cannot call updateConsumedPackSize because the box has no image set")
	}

	if box.dx() > p.consumedPackSize.Width {
		p.consumedPackSize.Width = box.dx()
	}

	if box.dy() > p.consumedPackSize.Height {
		p.consumedPackSize.Height = box.dy()
	}
}

// Returns true if the box would update the consumedPackSize.
// This can be useful to use or discard a specific box.
func (p *Pack) wouldUpdateConsumedPackSize(box *Box, img *imageinfo.ImageInfo) bool {
	dx := box.x() + img.Width()
	dy := box.y() + img.Height()
	return dx > p.consumedPackSize.Width || dy > p.consumedPackSize.Height
}

// Add the boxes to p.freeBoxes.
//
// If a box has already been added, it panics.
func (p *Pack) addFreeBoxes(boxes ...*Box) {
	for _, box := range boxes {
		_, ok := p.freeBoxes[box.id]
		if ok {
			log.Fatal().Str("box_id", box.id).Msg("fatal error: the randomly generated box id has already been generated")
		}
		log.Debug().
			Int("box_x", box.x()).
			Int("box_y", box.y()).
			Int("box_width", box.width()).
			Int("box_height", box.height()).
			Str("box_type", string(box.boxType)).
			Msg("new free box added")
		p.freeBoxes[box.id] = box
	}
}

// Remove a box from p.freeBoxes.
//
// If a box doesn't exist, it panics.
// If a box is already consumed, it panics.
func (p *Pack) consumeFreeBox(box *Box) {
	_, ok := p.freeBoxes[box.id]
	if !ok {
		log.Fatal().
			Str("box_id", box.id).
			Msg("fatal error: cannot consume free box as the ID doesn't exist")
	}
	_, ok = p.consumedBoxes[box.id]
	if ok {
		log.Fatal().
			Str("box_id", box.id).
			Msg("fatal error: cannot consume free box as the box has already been consumed")
	}
	p.consumedBoxes[box.id] = box
	delete(p.freeBoxes, box.id)
}

// Returns p.consumedPackSize.width
//
// Is the width used to create the final image.
func (p *Pack) width() int {
	return p.consumedPackSize.Width
}

// Returns p.consumedPackSize.height
//
// Is the height used to create the final image.
func (p *Pack) height() int {
	return p.consumedPackSize.Height
}
