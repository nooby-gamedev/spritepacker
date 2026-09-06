package spritepackreader

import (
	"image/draw"
	"time"

	"github.com/google/uuid"
	"github.com/nooby-gamedev/spritepacker/pkg/spritepack"
	"github.com/rs/zerolog/log"
)

// Represents an animation group.
type AnimationGroup struct {
	name             SpriteAnimationGroupName
	sprites          []*spritepack.Sprite
	animations       map[string]*Animation
	spritePackReader *SpritePackReader
	targetFPS        int
}

type Animation struct {
	id             string
	sprite         *spritepack.Sprite
	spriteIndex    int
	animationGroup *AnimationGroup

	tick *time.Time
}

// Create a new animation.
//
// If id is empty, it creates a random id (uuid).
// If id already exists, it returns ErrAnimationIdAlreadyExists.
func (a *AnimationGroup) newAnimation(id string) (*Animation, error) {
	if id == "" {
		id = uuid.NewString()
	}

	_, ok := a.animations[id]
	if ok {
		return nil, ErrAnimationIdAlreadyExists
	}

	animation := &Animation{
		id:             id,
		sprite:         a.sprites[0],
		spriteIndex:    0,
		animationGroup: a,
	}

	a.animations[id] = animation
	return animation, nil
}

// Returns the animation by id.
// If it doesn't exist, it automatically creates it.
func (a *AnimationGroup) Animation(id string) (*Animation, error) {
	animation, ok := a.animations[id]
	if ok {
		return animation, nil
	}
	return a.newAnimation(id)
}

// Returns delta time and "now".
//
// This method DOES NOT update a.tick (the only exception is when a.tick is nil).
func (a *Animation) dt() (dt float64, now time.Time) {
	now = time.Now()
	if a.tick == nil {
		a.tick = &now
	}

	dt = now.Sub(*a.tick).Seconds()
	return dt, now
}
func (a *Animation) setNextSprite() {
	a.spriteIndex++
	if a.spriteIndex >= len(a.animationGroup.sprites) {
		a.spriteIndex = 0
	}
	a.sprite = a.animationGroup.sprites[a.spriteIndex]
}
func (a *Animation) tickThreshold() float64 {
	if a.animationGroup.targetFPS <= 0 {
		log.Fatal().Msg("fatal error: targetFPS must be greater than 0")
	}
	return float64(len(a.animationGroup.sprites)) / float64(a.animationGroup.targetFPS)
}
func (a *Animation) Draw(dst draw.Image, dstX, dstY int, rotation float64) error {
	dt, now := a.dt()
	threshold := a.tickThreshold()
	if dt >= threshold {
		a.setNextSprite()
		a.tick = &now
	}
	return a.animationGroup.
		spritePackReader.
		DrawSprite(SpriteName(a.sprite.NormalizedName), dst, dstX, dstY, rotation)
}

func (a *AnimationGroup) SetTargetFPS(fps int) {
	a.targetFPS = fps
}
