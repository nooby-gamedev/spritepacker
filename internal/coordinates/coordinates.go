package coordinates

type Coordinates struct {
	X, Y int
}

func New(x, y int) Coordinates {
	return Coordinates{X: x, Y: y}
}
