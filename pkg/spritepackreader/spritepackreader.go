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
	"sync"

	"github.com/nooby-gamedev/spritepacker/pkg/extensions"
	"github.com/nooby-gamedev/spritepacker/pkg/performancemonitor"
	"github.com/nooby-gamedev/spritepacker/pkg/spritepack"
	"github.com/nooby-gamedev/spritepacker/pkg/transformation/transformation2d"
	"github.com/nooby-gamedev/spritepacker/pkg/transformation/transformation2doptions"
	"github.com/rs/zerolog/log"
)

type SpriteName string
type SpriteAnimationGroupName string
type SpriteCacheKey string

type DrawableImage interface {
	Set(x int, y int, clr color.Color)
	Bounds() image.Rectangle
}

type SpritePackReader struct {
	spritePack      *spritepack.SpritePack
	spritePackImage image.Image
	caches          map[SpriteCacheKey]image.Image
	mu              sync.RWMutex
	performance     *performancemonitor.PerformanceMonitor
}

func NewEmpty() *SpritePackReader {
	return &SpritePackReader{
		spritePack:  spritepack.New(),
		caches:      make(map[SpriteCacheKey]image.Image),
		performance: performancemonitor.Monitor(),
	}
}

func NewBuf(spritePackImage, spritePackJson []byte, spritePackImageExtension extensions.SupportedExtension) (*SpritePackReader, error) {
	s := &SpritePackReader{
		spritePack:  spritepack.New(),
		caches:      make(map[SpriteCacheKey]image.Image),
		performance: performancemonitor.Monitor(),
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
		spritePack:  spritepack.New(),
		caches:      make(map[SpriteCacheKey]image.Image),
		performance: performancemonitor.Monitor(),
	}

	if err := s.LoadPackJson(spritePackJson); err != nil {
		return nil, err
	}

	if err := s.LoadPackImage(spritePackImage); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *SpritePackReader) getSpriteCacheKey(spriteNormalizedName SpriteName, optsCacheKey string) SpriteCacheKey {
	return SpriteCacheKey(fmt.Sprintf("%s.%s", spriteNormalizedName, optsCacheKey))
}

func (s *SpritePackReader) SetSpriteCache(spriteNormalizedName SpriteName, optsCacheKey string, img image.Image) {
	s.mu.Lock()
	defer s.mu.Unlock()
	spriteCacheKey := s.getSpriteCacheKey(spriteNormalizedName, optsCacheKey)
	s.caches[spriteCacheKey] = img
}
func (s *SpritePackReader) GetSpriteCache(spriteNormalizedName SpriteName, optsCacheKey string) (image.Image, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	spriteCacheKey := s.getSpriteCacheKey(spriteNormalizedName, optsCacheKey)
	spriteCache, ok := s.caches[spriteCacheKey]

	if !ok {
		return nil, false
	}
	return spriteCache, true
}

// Draws the sprite.
// If cacheCustomKey is NOT empty, it uses caches.
// If the sprite was not found, it returns ErrSpriteNotFound.
//
// If the sprite sheet has not been loaded, it returns ErrSpriteSheetNodLoaded.
func (s *SpritePackReader) DrawSprite(spriteNormalizedName SpriteName, dst draw.Image, opts transformation2doptions.Transformation2DOptions, useCache bool) error {
	s.performance.StartMeasureAverageDeltaTime("spritepackreader.draw_sprite")
	defer s.performance.StopMeasureAverageDeltaTime("spritepackreader.draw_sprite")

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

	optsCacheKey := opts.CacheKey()

	// If optsCacheKey is equal to the default options cache key, it means
	// no transformations will happen.
	//
	// By setting useCache to FALSE, it will return the original image directly.
	// We want to avoid caching the original image.
	if optsCacheKey == transformation2doptions.OriginalImageCacheKey() {
		useCache = false
	}
	var transform2d *transformation2d.Transformation2D

	// When using caches, we try to get the cached sprite.
	// If it exists, we use that image for a new Transform2D instance, with empty options.
	//
	// By using empty options, the transform2d.Transform(opts) will just return the image itself.
	if useCache {
		s.performance.StartMeasureAverageDeltaTime("spritepackreader.get_cache")
		img, ok := s.GetSpriteCache(spriteNormalizedName, optsCacheKey)
		if ok {
			transform2d = transformation2d.New(img)
			pos := opts.Position
			speed := opts.Speed()
			scalingFactor := opts.ScalingFactor()
			opts = *transformation2doptions.New(speed, pos.Coordinates)
			opts.SetScalingFactor(scalingFactor)
		}
		s.performance.StopMeasureAverageDeltaTime("spritepackreader.get_cache")
	}

	if transform2d == nil {
		transform2d = transformation2d.New(subImg.SubImage(sprite.Rect()))
	}

	transformedImg, err := transform2d.Transform(opts)

	if err != nil {
		return err
	}

	dstPos := opts.Position.Coordinates.ToPointRound()
	dstX := dstPos.X
	dstY := dstPos.Y
	dstRect := image.Rect(dstX, dstY, (dstX + transformedImg.Bounds().Dx()), (dstY + transformedImg.Bounds().Dy()))
	srcPoint := image.Point{X: transformedImg.Bounds().Min.X, Y: transformedImg.Bounds().Min.Y}

	draw.Draw(dst, dstRect, transformedImg, srcPoint, draw.Over)

	if useCache {
		s.performance.StartMeasureAverageDeltaTime("spritepackreader.set_cache")
		s.SetSpriteCache(spriteNormalizedName, optsCacheKey, transformedImg)
		s.performance.StopMeasureAverageDeltaTime("spritepackreader.set_cache")
	}

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
