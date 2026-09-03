package size

import "fmt"

type Size struct {
	Width, Height int
}

func New(width, height int) Size {
	return Size{
		Width:  width,
		Height: height,
	}
}

func (s Size) Area() int {
	return s.Width * s.Height
}

func (s Size) ToString() string {
	return fmt.Sprintf("%dx%d", s.Width, s.Height)
}
