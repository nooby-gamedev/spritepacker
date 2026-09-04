package spritepackreader

import "errors"

var ErrSpriteSheetNodLoaded = errors.New("the sprite sheet has not been loaded yet")
var ErrSpriteNotFound = errors.New("sprite not found")
var ErrSpritesheetNotValidPng = errors.New("spritesheets doesn't seem to be a valid PNG")
