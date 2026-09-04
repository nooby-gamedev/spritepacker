# Sprite Packer

## What is Sprite Packer?
**Sprite Packer** is a tool that can be used to create a single PNG stylesheet, using multiple images as source.  

## Why Sprite Packer?
I'm aware that there are tools out there that works way better than mine, but I just wanted to create something new as a challenge.  
And, as a challenge, I've decided to not use AI, and to not search for known algorithms on the Internet.  

## Creating a Sprite Pack
Creating a Sprite Pack is fairly easy. 
First, we need to create a folder where we put all our images (e.g., **/tmp/sprites**). 
Then, we have to decide where we want to save our files and the basename.  
For example, if we want to save our new files under **/tmp** using **spritesheet** as basename, we can just specify **/tmp/spritesheet**.  

```Go
import (
	"github.com/nooby-gamedev/spritepacker/internal/loader"
	"github.com/nooby-gamedev/spritepacker/internal/pack"

	"github.com/rs/zerolog"
)

func main() {
	imgLoader := loader.New(loader.LoadAllImagesInMemory, zerolog.DebugLevel)
	// Load sprites under /tmp/sprites and save the sprite pack as /tmp/spritesheet.png.
    // A "spritesheet.json" and "spritesheet.go" that contain the sprite pack info will also be created.
    imgLoader.
		LoadImages("/tmp/sprites").
		CreateSpritePack("/tmp/spritesheet", pack.DontDrawBoxBorders)
}
```

## Sprite Packer Output
As already mentioned, **Sprite Packer** create a PNG stylesheet using several images as source.  
However, the PNG is not the only output file: in fact, two more files will be created:
1. **spritesheet.json**: a JSON file that contains all the coordinates and sizes of the single sprites.
2. **spritesheet.go**: a simple Go file that contains all the constants that can be used to retrieve specific sprites, using **pkg/spritepackreader**

**spritesheet.json** sample:
```json
{
	"sprites": {
		"BackgroundDay": {
			"normalized_name": "BackgroundDay",
			"name": "background-day.png",
			"x": 0,
			"y": 0,
			"width": 288,
			"height": 512
		},
		"BackgroundNight": {
			"normalized_name": "BackgroundNight",
			"name": "background-night.png",
			"x": 288,
			"y": 0,
			"width": 288,
			"height": 512
		},
		"Base": {
			"normalized_name": "Base",
			"name": "base.png",
			"x": 576,
			"y": 320,
			"width": 336,
			"height": 112
		},
	}
}
```

**spritesheet.go** sample:
```Go
package images

import "github.com/nooby-gamedev/spritepacker/pkg/spritepackreader"

const BackgroundDay spritepackreader.SpriteName = "BackgroundDay"
const BackgroundNight spritepackreader.SpriteName = "BackgroundNight"
const Base spritepackreader.SpriteName = "Base"
```

## Reading sprites
Reading a sprite is even simpler.  
Let's say we're using **Ebitengine** to create our game.  
In the **Draw** method we can just:

```Go
var sprites *spritepackreader.SpritePackReader

// Note: I know, this is an awful implementation of the Draw method. 
// I just wanted to show you how easy it is to use.
func (g *Game) Draw(screen *ebiten.Image) {
	if sprites == nil {
		var err error
		sprites, err = spritepackreader.New("/tmp/spritesheet.png", "/tmp/spritesheet.json")
		if err != nil {
			log.Fatal().Err(err).Msg("fatal error")
		}
	}
    // dstX and dstY are relative to the dst image (screen)
    // If coordinates are out of boundaries, the DrawSprite will draw only
    // visible coordinates.
	dstX := 0
	dstY := 0
	if err := sprites.DrawSprite(images.BackgroundDay, screen, dstX, dstY); err != nil {
		log.Fatal().Err(err).Msg("unable to draw background")
	}
}
```