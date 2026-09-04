package loader

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/nooby-gamedev/spritepacker/internal/coordinates"
	"github.com/nooby-gamedev/spritepacker/internal/imageinfo"
	"github.com/nooby-gamedev/spritepacker/internal/pack"
	"github.com/nooby-gamedev/spritepacker/internal/size"
	"github.com/nooby-gamedev/spritepacker/pkg/extensions"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// Faster but consumes more resources
	LoadAllImagesInMemory bool = true
	// Slower but consumes less resources
	LoadOneImageAtTime bool = false
)

type SizeType string

const (
	SizeTypeWidth  SizeType = "width"
	SizeTypeHeight SizeType = "height"
)

type Loader struct {
	// Contains all the images
	images map[string]*imageinfo.ImageInfo
	// Containes the paths of the images in descending order,
	// ordered by either width or height (depending on maxSizeType)
	//
	// WARNING: be sure the order is DESCENDING, otherwise the algorithm will fail.
	descOrderedImages []string
	// Contains all the widths detected
	widths map[int][]string
	// Contains all the heights detected
	heights map[int][]string
	// If true, load all images in memory at once (faster but consumes more resources).
	// If false, load one image at time.
	loadAllImagesInMemory bool
	// Max width detected. Useful for creating the BaseBox.
	maxWidth int
	// Max height detected. Useful for creating the BaseBox.
	maxHeight int
	// Sum of all widths
	totWidth int
	// Sum of all heights
	totHeight int
	// Defines the max size detected
	maxSize int
	// Defines the type of max size detected (width or height)
	maxSizeType SizeType
	// Set by LoadImages(dir string).
	// If false, calling CreateSpritePack(out string) will make
	// the application panic
	imagesLoaded bool
}

// If loadAllImagesInMemory is true, load all images in memory at once (faster but consumes more resources).
// If false, load one image at time.
//
// Padding defines the distance between each image in the pack, and
// the distance from the borders.
func New(loadAllImagesInMemory bool, logLevel zerolog.Level) *Loader {
	log.Logger = zerolog.New(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		},
	).With().Timestamp().Logger().Level(logLevel)

	return &Loader{
		images:                make(map[string]*imageinfo.ImageInfo, 0),
		widths:                make(map[int][]string, 0),
		heights:               make(map[int][]string, 0),
		loadAllImagesInMemory: loadAllImagesInMemory,
	}
}
func (l *Loader) isExtensionSupported(ext string) bool {
	e := extensions.SupportedExtension(ext)
	return e.IsValid()
}

// Load all images and order them by either width or height
// depending on the max value detected.
func (l *Loader) LoadImages(dir string) *Loader {
	entries, err := os.ReadDir(dir)

	if err != nil {
		log.Error().Err(err).Str("directory", dir).Msg("an error occurred while reading directoru")
	}

	if len(entries) == 0 {
		log.Warn().Str("directory", dir).Msg("no files found")
		return l
	}

	for _, entry := range entries {
		fullPath := filepath.Clean(dir + "/" + entry.Name())

		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(fullPath))

		if !l.isExtensionSupported(ext) {
			log.Warn().Str("file", fullPath).Str("ext", ext).Msg("the file won't be loaded as the extension is not supported")
			continue
		}

		log.Info().Str("file", fullPath).Str("ext", ext).Msg("loading file with a supported extension")
		handle, err := os.Open(fullPath)
		if err != nil {
			log.Error().Err(err).Str("file", fullPath).Msg("an error occurred while opening file")
			continue
		}
		defer handle.Close()

		var img image.Image
		switch extensions.SupportedExtension(ext) {
		case extensions.Png:
			img, err = png.Decode(handle)
			if err != nil {
				log.Error().Err(err).Str("file", fullPath).Msg("an error occurred while decoding the image")
				continue
			}
		default:
			log.Fatal().Err(err).Str("file", fullPath).Str("ext", ext).Msg("fatal error: the extension is supported but \"decoder switch\" doesn't handle it")
			continue
		}

		rect := img.Bounds()
		width := rect.Dx()
		height := rect.Dy()

		l.totWidth += width
		l.totHeight += height

		if width > l.maxSize {
			l.maxSize = width
			l.maxSizeType = SizeTypeWidth
		}

		if height > l.maxSize {
			l.maxSize = height
			l.maxSizeType = SizeTypeHeight
		}

		if width > l.maxWidth {
			l.maxWidth = width
		}
		if height > l.maxHeight {
			l.maxHeight = height
		}

		var imgHandle image.Image
		if l.loadAllImagesInMemory {
			imgHandle = img
		}
		imgInfo := imageinfo.New(entry.Name(), fullPath, size.New(width, height), imgHandle, extensions.SupportedExtension(ext))
		l.images[fullPath] = imgInfo

		existingWidth, ok := l.widths[width]
		if !ok {
			existingWidth = make([]string, 0)
		}
		existingWidth = append(existingWidth, fullPath)
		l.widths[width] = existingWidth

		existingHeight, ok := l.heights[height]
		if !ok {
			existingHeight = make([]string, 0)
		}
		existingHeight = append(existingHeight, fullPath)
		l.heights[height] = existingHeight

		log.Debug().Int("width", width).Int("height", height).Str("full_path", fullPath).Msg("image loaded")
	}

	log.Info().Int("max_size", l.maxSize).Str("max_size_type", string(l.maxSizeType)).Msg("max size detected")
	log.Info().Str("order_by", string(l.maxSizeType)).Int("loaded_images", len(l.images)).Msg("ordering images by size (descending order)")

	l.descOrderedImages = make([]string, 0, len(l.images))

	switch l.maxSizeType {
	case SizeTypeWidth:
		widths := make([]int, 0, len(l.widths))
		for width := range l.widths {
			widths = append(widths, width)
		}
		sort.Ints(widths)
		slices.Reverse(widths)

		for _, width := range widths {
			paths, ok := l.widths[width]
			if !ok {
				log.Fatal().Msg("fatal error: unable to retrieve the list of paths from l.widths")
			}
			for _, path := range paths {
				l.descOrderedImages = append(l.descOrderedImages, path)
			}
		}
	case SizeTypeHeight:
		heights := make([]int, 0, len(l.heights))
		for height := range l.heights {
			heights = append(heights, height)
		}
		sort.Ints(heights)
		slices.Reverse(heights)

		for _, height := range heights {
			paths, ok := l.heights[height]
			if !ok {
				log.Fatal().Msg("fatal error: unable to retrieve the list of paths from l.heights")
			}
			for _, path := range paths {
				l.descOrderedImages = append(l.descOrderedImages, path)
			}
		}
	default:
		log.Fatal().Str("max_size_type", string(l.maxSizeType)).Msg("fatal error: unable to order images due to invalid maxSizeType")
	}

	l.imagesLoaded = true
	return l
}

// CreateSpritePack must be called AFTER LoadImages(dir string),
// otherwise it panics.
func (l *Loader) CreateSpritePack(savePath string, drawBoxBorders bool) {
	if !l.imagesLoaded {
		log.Fatal().Msg("fatal error: LoadImages must be called before CreateSpritePack")
	}
	if len(l.descOrderedImages) == 0 {
		log.Warn().Msg("cannot create sprite pack: no images loaded")
	}

	p := pack.New(
		size.New(l.maxWidth, l.maxHeight),
		size.New(l.totWidth, l.totHeight),
		coordinates.New(0, 0),
		drawBoxBorders,
	)

	for _, imgName := range l.descOrderedImages {
		img, ok := l.images[imgName]
		if !ok {
			log.Fatal().
				Msg("fatal error: the image cannot be found in l.images (CreateSpritePack)")
		}
		log.Info().Str("full_path", img.Path).Msg("inserting image")
		p.InsertImage(img)
		log.Info().Str("full_path", img.Path).Msg("image inserted successfully")
	}

	log.Info().Str("save_path", savePath).Msg("finalizing pack")
	p.Finalize(savePath)
}
