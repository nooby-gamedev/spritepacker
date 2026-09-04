package spritepackreader

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/nooby-gamedev/spritepacker/pkg/extensions"
	"github.com/nooby-gamedev/spritepacker/pkg/spritepack"
	"github.com/rs/zerolog/log"
)

type SpriteName string
type DrawableImage interface {
	Set(x int, y int, clr color.Color)
	Bounds() image.Rectangle
}

type SpritePackReader struct {
	spritePack      *spritepack.SpritePack
	spritePackImage image.Image
}

func NewEmpty() *SpritePackReader {
	return &SpritePackReader{
		spritePack: spritepack.New(),
	}
}

func NewBuf(spritePackImage, spritePackJson []byte, spritePackImageExtension extensions.SupportedExtension) (*SpritePackReader, error) {
	s := &SpritePackReader{
		spritePack: spritepack.New(),
	}
	if err := s.LoadPackJsonBuf(spritePackJson); err != nil {
		return nil, err
	}
	if err := s.LoadPackImageBuf(spritePackImage, spritePackImageExtension); err != nil {
		return nil, err
	}
	return s, nil
}

// Creates a new instance of SpritePackReader and automatically set JSON and Image files
func New(spritePackImage, spritePackJson string) (*SpritePackReader, error) {
	s := &SpritePackReader{
		spritePack: spritepack.New(),
	}

	if err := s.LoadPackJson(spritePackJson); err != nil {
		return nil, err
	}

	if err := s.LoadPackImage(spritePackImage); err != nil {
		return nil, err
	}

	return s, nil
}

// Draw the sprite in dst at the given dst coordinates.
// It only draws the coordinates within the boundaries.
//
// If the sprite was not found, it returns an error.
//
// If the sprite sheet has not been loaded, it panics.
func (s *SpritePackReader) DrawSprite(name SpriteName, dst DrawableImage, dstX, dstY int) error {
	if s.spritePackImage == nil {
		log.Fatal().Msg("fatal error: the sprite sheet has not been loaded yet")
	}

	sprite := s.spritePack.Sprite(string(name))
	if sprite == nil {
		return nil
	}

	for x := range sprite.Width {
		newDstX := x + dstX
		if newDstX >= dst.Bounds().Dx() {
			break
		}
		for y := range sprite.Height {
			newDstY := y + dstY
			if newDstY >= dst.Bounds().Dy() {
				break
			}
			spriteX := sprite.X + x
			spriteY := sprite.Y + y
			clr := s.spritePackImage.At(spriteX, spriteY)
			dst.Set(newDstX, newDstY, clr)
		}
	}
	return nil
}

// Reload the sprite pack json
func (s *SpritePackReader) LoadPackJson(spritePackJson string) error {
	err := s.spritePack.Import(spritePackJson)
	if err != nil {
		return err
	}
	return nil
}
func (s *SpritePackReader) LoadPackJsonBuf(buf []byte) error {
	return s.spritePack.ImportBuf(buf)
}

// Reload the sprite pack image
func (s *SpritePackReader) LoadPackImage(spritePackImage string) error {
	return s.loadPackImage(spritePackImage)
}
func (s *SpritePackReader) LoadPackImageBuf(buf []byte, ext extensions.SupportedExtension) error {
	reader := bytes.NewReader(buf)
	switch ext {
	case extensions.Png:
		img, err := png.Decode(reader)
		if err != nil {
			return err
		}
		s.spritePackImage = img
		return nil
	default:
		return fmt.Errorf("unsupported file extension (%s)", ext)
	}
}
func (s *SpritePackReader) loadPackImage(spritePackImage string) error {
	handle, err := os.Open(spritePackImage)
	if err != nil {
		return err
	}
	defer handle.Close()

	ext := strings.ToLower(filepath.Ext(spritePackImage))
	switch extensions.SupportedExtension(ext) {
	case extensions.Png:
		img, err := png.Decode(handle)
		if err != nil {
			return err
		}
		s.spritePackImage = img
	default:
		return fmt.Errorf("unsupported file extension (%s)", ext)
	}

	return nil
}
