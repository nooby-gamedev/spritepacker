package spritepack

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

type SpritePack struct {
	Sprites map[string]Sprite `json:"sprites"`
	mu      sync.Mutex
}

type Sprite struct {
	NormalizedName string `json:"normalized_name"`
	Name           string `json:"name"`
	X              int    `json:"x"`
	Y              int    `json:"y"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
}

func New() *SpritePack {
	return &SpritePack{
		Sprites: make(map[string]Sprite, 0),
	}
}

func newSprite(normalizedName, name string, x, y, width, height int) Sprite {
	return Sprite{
		NormalizedName: normalizedName,
		Name:           name,
		X:              x,
		Y:              y,
		Width:          width,
		Height:         height,
	}
}

var normalizeRegexp = regexp.MustCompile(`(?i)[^a-z_0-9]`)
var startsWithoutLetterRegexp = regexp.MustCompile(`(?i)^[^a-z]`)
var findMultipleUnderscoreRegexp = regexp.MustCompile(`[_]{2,}`)
var capitalizeLetterRegexp = regexp.MustCompile(`(?i)[_]([a-z]{1})`)

// Normalize the sprite name.
// The normalized name can be used also as variable name in Go.
func (s *SpritePack) normalizeSpriteName(name string) string {
	// Set "Sprite" as prefix for sprites with a name that
	// starts without a letter
	if startsWithoutLetterRegexp.MatchString(name) {
		name = fmt.Sprintf("Sprite%s", name)
	}

	// Set an underscore as a prefix.
	// This will be used by "capitalizeLetterRegexp" to set the first letter ToUpper,
	// if necessary (the underscore will be removed)
	name = fmt.Sprintf("_%s", name)
	// Remove the extension
	name, _ = strings.CutSuffix(name, filepath.Ext(name))
	// Replace invalid characters with "_"
	name = normalizeRegexp.ReplaceAllString(name, "_")
	// Replace multiple underscores with a single underscore
	name = findMultipleUnderscoreRegexp.ReplaceAllString(name, "_")
	// Capitalize the letters after each underscore (_) and remove the underscore
	name = capitalizeLetterRegexp.ReplaceAllStringFunc(name, func(s string) string {
		s, _ = strings.CutPrefix(s, "_")
		return strings.ToUpper(s)
	})
	return name
}

func (s *SpritePack) NewSprite(name string, x, y, width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	normalizedName := s.normalizeSpriteName(name)

	nameIsValid := false
	counter := 0
	for !nameIsValid {
		nameWithCounter := normalizedName
		if counter > 0 {
			nameWithCounter = fmt.Sprintf("%s%d", nameWithCounter, counter)
		}
		_, ok := s.Sprites[nameWithCounter]
		if ok {
			counter++
			continue
		}
		normalizedName = nameWithCounter
		nameIsValid = true
	}

	s.Sprites[normalizedName] = newSprite(normalizedName, name, x, y, width, height)
}

func (s *SpritePack) ImportBuf(buf []byte) error {
	return json.Unmarshal(buf, s)
}
func (s *SpritePack) Import(importPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Info().
		Str("full_path", importPath).
		Msg("importing sprite pack as JSON")

	s.Sprites = make(map[string]Sprite, 0)

	handle, err := os.Open(importPath)
	if err != nil {
		return err
	}
	defer handle.Close()

	jDecoder := json.NewDecoder(handle)
	if err := jDecoder.Decode(s); err != nil {
		return err
	}

	return nil
}

// Saves the SpritePack as JSON.
// It will also create a Go file with the sprite constants.
func (s *SpritePack) Export(savePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	extJson := ".json"
	extGo := ".go"

	var savePathJson, savePathGo string
	savePathJson = savePath
	savePathGo = savePath

	if !strings.HasSuffix(savePathJson, extJson) {
		savePathJson += extJson
	}

	if !strings.HasSuffix(savePathGo, extGo) {
		savePathGo += extGo
	}

	log.Info().
		Str("full_path", savePathJson).
		Msg("exporting sprite pack as JSON")

	handleJson, err := os.Create(savePathJson)
	if err != nil {
		return err
	}
	defer handleJson.Close()

	jEncoder := json.NewEncoder(handleJson)
	jEncoder.SetIndent("", "	")
	if err := jEncoder.Encode(s); err != nil {
		return err
	}

	template := s.createGoTemplate()
	handleGo, err := os.Create(savePathGo)
	if err != nil {
		return err
	}
	defer handleGo.Close()

	_, err = handleGo.WriteString(template)
	if err != nil {
		return err
	}
	return nil
}

//go:embed template.go.tpl
var goTemplate []byte

func (s *SpritePack) createGoTemplate() string {
	template := string(goTemplate)

	builder := strings.Builder{}
	builder.WriteString(template)

	sortedNames := make([]string, 0)
	for _, sprite := range s.Sprites {
		sortedNames = append(sortedNames, sprite.NormalizedName)
	}
	slices.Sort(sortedNames)

	for _, spriteName := range sortedNames {
		sprite, _ := s.Sprites[spriteName]
		fmt.Fprintf(&builder, "const %s spritepackreader.SpriteName = \"%s\"\n", sprite.NormalizedName, sprite.NormalizedName)
	}
	return builder.String()
}

func (s *SpritePack) Sprite(spriteNormalizedName string) *Sprite {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.Sprites[spriteNormalizedName]
	if !ok {
		return nil
	}
	return &existing
}
