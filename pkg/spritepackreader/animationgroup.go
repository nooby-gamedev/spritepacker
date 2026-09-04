package spritepackreader

import (
	"image/draw"

	"github.com/nooby-gamedev/spritepacker/pkg/spritepack"
)

// Represents an animation group.
type AnimationGroup struct {
	name          SpriteAnimationGroupName
	sprites       []*spritepack.Sprite
	targetFPS     int
	ticks         int
	currentSprite int
	s             *SpritePackReader
}

func (a *AnimationGroup) DrawSprite(dst draw.Image, dstX, dstY int) error {
	a.ticks++
	spriteCount := len(a.sprites)
	// Change animation after tickThreshold
	tickThreshold := a.targetFPS / spriteCount

	if a.ticks >= tickThreshold {
		a.currentSprite++
		a.ticks = 0
		if a.currentSprite >= len(a.sprites) {
			a.currentSprite = 0
		}
	}

	sprite := a.sprites[a.currentSprite]
	return a.s.DrawSprite(SpriteName(sprite.NormalizedName), dst, dstX, dstY)
}

func (a *AnimationGroup) SetTargetFPS(targetFPS int) {
	a.targetFPS = targetFPS
}
