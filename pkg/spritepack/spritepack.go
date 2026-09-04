package spritepack

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
)

const emptyAnimationGroup string = ""

type SpritePack struct {
	Sprites         map[string]Sprite   `json:"sprites"`
	AnimationGroups map[string][]string `json:"animation_groups"`
}

type Sprite struct {
	AnimationGroup string `json:"animation_group"`
	NormalizedName string `json:"normalized_name"`
	Name           string `json:"name"`
	X              int    `json:"x"`
	Y              int    `json:"y"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
}

func New() *SpritePack {
	return &SpritePack{
		Sprites:         make(map[string]Sprite, 0),
		AnimationGroups: make(map[string][]string, 0),
	}
}

func newSprite(animationGroup, normalizedName, name string, x, y, width, height int) Sprite {
	return Sprite{
		AnimationGroup: animationGroup,
		NormalizedName: normalizedName,
		Name:           name,
		X:              x,
		Y:              y,
		Width:          width,
		Height:         height,
	}
}

var normalizeRegexp = regexp.MustCompile(`(?i)[^a-z_0-9]`)
var findMultipleUnderscoreRegexp = regexp.MustCompile(`[_]{2,}`)
var capitalizeLetterRegexp = regexp.MustCompile(`(?i)[_]([a-z0-9]{1})`)

// Normalize the sprite name.
// The normalized name can be used also as variable name in Go.
func (s *SpritePack) normalizeSpriteName(name string, isAnimationGroup bool) string {
	// Set "Sprite" as prefix for Sprites.
	// Add an extra underscore so that the next code will recognize it and
	// set the next letter as uppercase, if necessary.
	if !isAnimationGroup && !strings.HasPrefix(strings.ToLower(name), "sprite") {
		name = fmt.Sprintf("Sprite_%s", name)
	}

	// Set "Animation" as prefix for Animation Groups.
	// Add an extra underscore so that the next code will recognize it and
	// set the next letter as uppercase, if necessary.
	if isAnimationGroup && !strings.HasPrefix(strings.ToLower(name), "animation") {
		name = fmt.Sprintf("Animation_%s", name)
	}

	// Remove the extension (only if is NOT an animation group, as
	// animation groups are folders)
	if !isAnimationGroup {
		name, _ = strings.CutSuffix(name, filepath.Ext(name))
	}
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

func (s *SpritePack) NewSprite(animationGroup, name string, x, y, width, height int) {
	isAnimationGroup := animationGroup != emptyAnimationGroup
	normalizedSpriteName := s.normalizeSpriteName(name, false)

	normalizedAnimationGroupName := animationGroup
	if isAnimationGroup {
		normalizedAnimationGroupName = s.normalizeSpriteName(animationGroup, true)
	}

	nameIsValid := false
	counter := 0
	for !nameIsValid {
		nameWithCounter := normalizedSpriteName
		if counter > 0 {
			nameWithCounter = fmt.Sprintf("%s%d", nameWithCounter, counter)
		}
		_, ok := s.Sprites[nameWithCounter]
		if ok {
			counter++
			continue
		}
		normalizedSpriteName = nameWithCounter
		nameIsValid = true
	}

	s.Sprites[normalizedSpriteName] = newSprite(animationGroup, normalizedSpriteName, name, x, y, width, height)

	// If the sprite is part of an animation group, add the
	// sprite to the group.
	if isAnimationGroup {
		existing, ok := s.AnimationGroups[normalizedAnimationGroupName]
		if !ok {
			existing = make([]string, 0)
		}
		existing = append(existing, normalizedSpriteName)
		s.AnimationGroups[normalizedAnimationGroupName] = existing
	}

}

func (s *SpritePack) ImportBuf(buf []byte) error {
	return json.Unmarshal(buf, s)
}
func (s *SpritePack) Import(importPath string) error {
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

func (s *SpritePack) ExportJson(savePath string) error {
	extJson := ".json"
	if !strings.HasSuffix(savePath, extJson) {
		savePath += extJson
	}

	log.Info().
		Str("full_path", savePath).
		Msg("exporting sprite pack as JSON")

	handleJson, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer handleJson.Close()

	jEncoder := json.NewEncoder(handleJson)
	jEncoder.SetIndent("", "	")
	if err := jEncoder.Encode(s); err != nil {
		return err
	}

	return nil
}

func (s *SpritePack) ExportGoConstants(savePath string) error {
	extGo := ".go"

	if !strings.HasSuffix(savePath, extGo) {
		savePath += extGo
	}

	log.Info().
		Str("full_path", savePath).
		Msg("exporting sprite pack as Go constants")

	template := s.createGoTemplate()
	handle, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer handle.Close()

	_, err = handle.WriteString(template)
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

	sortedSpriteNames := make([]string, 0)
	for _, sprite := range s.Sprites {
		sortedSpriteNames = append(sortedSpriteNames, sprite.NormalizedName)
	}
	slices.Sort(sortedSpriteNames)

	if len(sortedSpriteNames) > 0 {
		fmt.Fprintf(&builder, "/* Sprites */\n\n")
	}
	for _, spriteName := range sortedSpriteNames {
		sprite, _ := s.Sprites[spriteName]
		fmt.Fprintf(&builder, "const %s spritepackreader.SpriteName = \"%s\"\n", sprite.NormalizedName, sprite.NormalizedName)
	}

	sortedAnimationGroupNames := make([]string, 0)
	for groupName := range s.AnimationGroups {
		sortedAnimationGroupNames = append(sortedAnimationGroupNames, groupName)
	}
	slices.Sort(sortedAnimationGroupNames)

	if len(sortedAnimationGroupNames) > 0 {
		fmt.Fprintf(&builder, "\n\n/* Sprite Animation Groups */\n\n")
	}

	for _, groupName := range sortedAnimationGroupNames {
		fmt.Fprintf(&builder, "const %s spritepackreader.SpriteAnimationGroupName = \"%s\"\n", groupName, groupName)
	}
	return builder.String()
}

func (s *SpritePack) SearchSpritesByName(names ...string) []*Sprite {
	sprites := make([]*Sprite, 0)
	for _, name := range names {
		sprite, ok := s.Sprites[name]
		if ok {
			sprites = append(sprites, &sprite)
		}
	}
	return sprites
}
func (s *SpritePack) Sprite(spriteNormalizedName string) *Sprite {
	existing, ok := s.Sprites[spriteNormalizedName]
	if !ok {
		return nil
	}
	return &existing
}

func (s Sprite) Rect() image.Rectangle {
	return image.Rect(s.X, s.Y, s.Right(), s.Bottom())
}
func (s Sprite) Point() image.Point {
	return image.Point{
		X: s.X,
		Y: s.Y,
	}
}

// Returns X + Width
func (s Sprite) Right() int {
	return s.X + s.Width
}

// Returns Y + Height
func (s Sprite) Bottom() int {
	return s.Y + s.Height
}
