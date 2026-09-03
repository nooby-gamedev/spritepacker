package spritepack

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

type SpritePack struct {
	Sprites map[string]Sprite `json:"sprites"`
}

type Sprite struct {
	Name   string `json:"name"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func New() *SpritePack {
	return &SpritePack{
		Sprites: make(map[string]Sprite, 0),
	}
}

func newSprite(name string, x, y, width, height int) Sprite {
	return Sprite{
		Name:   name,
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

func (s *SpritePack) NewSprite(name string, x, y, width, height int) {
	s.Sprites[name] = newSprite(name, x, y, width, height)
}

func (s *SpritePack) Import(importPath string) error {
	log.Info().
		Str("full_path", importPath).
		Msg("importing sprite pack as JSON")

	handle, err := os.Open(importPath)
	closeHandle := func() {
		if handle == nil {
			return
		}
		if err := handle.Close(); err != nil {
			log.Fatal().Err(err).Msg("fatal error while importing SpritePack as JSON")
		}
	}
	defer closeHandle()
	if err != nil {
		return err
	}

	jDecoder := json.NewDecoder(handle)
	if err := jDecoder.Decode(s); err != nil {
		return err
	}

	return nil
}

// Saves the SpritePack as JSON
func (s *SpritePack) Export(savePath string) error {
	ext := ".json"
	if !strings.HasSuffix(savePath, ext) {
		savePath += ext
	}

	log.Info().
		Str("full_path", savePath).
		Msg("exporting sprite pack as JSON")

	handle, err := os.Create(savePath)
	closeHandle := func() {
		if handle == nil {
			return
		}
		if err := handle.Close(); err != nil {
			log.Fatal().Err(err).Msg("fatal error while exporting SpritePack as JSON")
		}
	}
	defer closeHandle()
	if err != nil {
		return err
	}

	jEncoder := json.NewEncoder(handle)
	jEncoder.SetIndent("", "	")
	if err := jEncoder.Encode(s); err != nil {
		return err
	}
	return nil
}
