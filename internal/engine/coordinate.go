package engine

type Coord struct {
	x int32
	y int32
}

func NewCoord(x int32, y int32) Coord {
	return Coord{
		x: x,
		y: y,
	}
}

func (c Coord) X() int32 {
	return c.x
}

func (c Coord) Y() int32 {
	return c.y
}
