package spritepackreader

import (
	"image/draw"
	"time"

	"github.com/nooby-gamedev/spritepacker/pkg/spritepack"
	"github.com/nooby-gamedev/spritepacker/pkg/transformation/transformation2doptions"
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
// If id is empty, it returns ErrAnimationIdCannotBeEmpty.
// If id already exists, it returns ErrAnimationIdAlreadyExists.
func (a *AnimationGroup) newAnimation(id string) (*Animation, error) {
	if id == "" {
		return nil, ErrAnimationIdCannotBeEmpty
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
// If id doesn't exist, it automatically creates it.
// If id is empty, it returns ErrAnimationIdCannotBeEmpty.
//
// The ID is used to keep track of the current sprite for a given animation.
// For example, if we have two players, we can use Animation("player1") and Animation("player2") to
// track both separately.
func (a *AnimationGroup) Animation(id string) (*Animation, error) {
	if id == "" {
		return nil, ErrAnimationIdCannotBeEmpty
	}
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
func (a *Animation) Draw(dst draw.Image, opts transformation2doptions.Transformation2DOptions, useCache bool) error {
	dt, now := a.dt()
	threshold := a.tickThreshold()
	if dt >= threshold {
		a.setNextSprite()
		a.tick = &now
	}
	return a.animationGroup.
		spritePackReader.
		DrawSprite(SpriteName(a.sprite.NormalizedName), dst, opts, useCache)
}

func (a *AnimationGroup) SetTargetFPS(fps int) {
	a.targetFPS = fps
}
