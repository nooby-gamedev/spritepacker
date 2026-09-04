package imageinfo

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/nooby-gamedev/spritepacker/internal/size"
	"github.com/nooby-gamedev/spritepacker/pkg/extensions"

	"github.com/rs/zerolog/log"
)

type ImageInfo struct {
	Name string
	Path string
	Size size.Size
	img  image.Image
	ext  extensions.SupportedExtension
}

func New(name string, path string, size size.Size, img image.Image, ext extensions.SupportedExtension) *ImageInfo {
	return &ImageInfo{
		Name: name,
		Path: path,
		Size: size,
		img:  img,
		ext:  ext,
	}
}

// Load the images if not already loaded
func (i *ImageInfo) loadImage() {
	if i.img != nil {
		log.Trace().Str("full_path", i.Path).Msg("image already loaded")
		return
	}

	log.Debug().Str("full_path", i.Path).Msg("reloading the image")
	handle, err := os.Open(i.Path)
	if err != nil {
		log.Fatal().Err(err).Str("full_path", i.Path).Msg("fatal error: the image cannot be reloaded due to an unexpected error")
	}

	switch i.ext {
	case extensions.Png:
		i.img, err = png.Decode(handle)
		if err != nil {
			log.Fatal().
				Str("full_path", i.Path).
				Msg("fatal error: the image cannot be reloaded due to an unexpected error")
		}
	default:
		log.Fatal().
			Str("full_path", i.Path).
			Str("ext", string(i.ext)).
			Msg("fatal error: the image extension is either invalid or not supported yet. Unable to reload the image")
	}
}
func (i *ImageInfo) ColorAt(x, y int) color.Color {
	i.loadImage()
	return i.img.At(x, y)
}
func (i *ImageInfo) Area() int {
	return i.Size.Area()
}
func (i *ImageInfo) Width() int {
	return i.Size.Width
}

func (i *ImageInfo) Height() int {
	return i.Size.Height
}

func (i *ImageInfo) Close() {
	if i.img == nil {
		return
	}
	i.img = nil
}
