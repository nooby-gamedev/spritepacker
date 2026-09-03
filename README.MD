# Sprite Packer

Usage example

```Go
import (
	"github.com/nooby-gamedev/spritepacker/internal/loader"
	"github.com/nooby-gamedev/spritepacker/internal/pack"

	"github.com/rs/zerolog"
)

func main() {
	imgLoader := loader.New(loader.LoadAllImagesInMemory, zerolog.DebugLevel)
	// Load sprites under /tmp/sprites and save the sprite pack as "output.png".
    // An "output.png.json" that contains the sprite pack info will also be created.
    imgLoader.
		LoadImages("/tmp/sprites").
		CreateSpritePack("output.png", pack.DontDrawBoxBorders)
}
```

Sprite Pack JSON sample
```json
{
	"sprites": {
		"0.png": {
			"name": "0.png",
			"x": 872,
			"y": 267,
			"width": 24,
			"height": 36
		},
		"1.png": {
			"name": "1.png",
			"x": 896,
			"y": 267,
			"width": 16,
			"height": 36
		},
		"2.png": {
			"name": "2.png",
			"x": 576,
			"y": 432,
			"width": 24,
			"height": 36
		},
	}
}
```