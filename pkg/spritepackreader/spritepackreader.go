package spritepackreader

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/nooby-gamedev/spritepacker/pkg/extensions"
	"github.com/nooby-gamedev/spritepacker/pkg/spritepack"
	"github.com/nooby-gamedev/spritepacker/pkg/transformation/transformation2d"
	"github.com/rs/zerolog/log"
)

type SpriteName string
type SpriteAnimationGroupName string

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

// Draws.
//
// If the sprite was not found, it returns ErrSpriteNotFound.
//
// If the sprite sheet has not been loaded, it returns ErrSpriteSheetNodLoaded.
func (s *SpritePackReader) DrawSprite(spriteNormalizedName SpriteName, dst draw.Image, dstX, dstY int, rotation float64) error {
	if s.spritePackImage == nil {
		return ErrSpriteSheetNodLoaded
	}

	sprite := s.spritePack.Sprite(string(spriteNormalizedName))
	if sprite == nil {
		return ErrSpriteNotFound
	}

	subImg, ok := s.spritePackImage.(interface {
		SubImage(r image.Rectangle) image.Image
	})
	if !ok {
		return ErrSpritesheetNotValidPng
	}

	transform2d := transformation2d.New(subImg.SubImage(sprite.Rect()))
	rotatedImg := transform2d.Rotate(rotation)

	dstRect := image.Rect(dstX, dstY, (dstX + rotatedImg.Bounds().Dx()), (dstY + rotatedImg.Bounds().Dy()))
	srcPoint := image.Point{X: rotatedImg.Bounds().Min.X, Y: rotatedImg.Bounds().Min.Y}

	draw.Draw(dst, dstRect, rotatedImg, srcPoint, 0)

	// dstRect := image.Rect(dstX, dstY, (dstX + sprite.Width), (dstY + sprite.Height))
	// srcPoint := sprite.Point()

	// draw.Draw(dst, dstRect, s.spritePackImage, srcPoint, 0)
	return nil
}

// Returns a new animation group.
//
// If targetFPS is <= 0, it panicsl
func (s *SpritePackReader) AnimationGroup(animationGroupName SpriteAnimationGroupName, targetFPS int) (*AnimationGroup, error) {
	if targetFPS <= 0 {
		log.Fatal().Msg("fatal error: targetFPS must be greater than 0")
	}
	sprites, ok := s.spritePack.AnimationGroups[string(animationGroupName)]
	if !ok {
		return nil, ErrAnimationGroupNotFound
	}
	group := &AnimationGroup{
		name:             animationGroupName,
		sprites:          s.spritePack.SearchSpritesByName(sprites...),
		animations:       make(map[string]*Animation, 0),
		spritePackReader: s,
		targetFPS:        targetFPS,
	}
	return group, nil
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
