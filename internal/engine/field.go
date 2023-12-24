package engine

import (
	"errors"
	"math/rand"
)

type cellState int

const (
	cellState_EMPTY cellState = 0
	cellState_FOOD  cellState = 1
	cellState_SNAKE cellState = 2
)

var (
	headPlaceNotFoundError = errors.New("place for snake head not found")
	tailPlaceNotFoundError = errors.New("place for snake tail not found")
)

func (g *Game) createField() [][]cellState {
	// Init
	field := make([][]cellState, g.Height)
	for i := int32(0); i < g.Height; i++ {
		field[i] = make([]cellState, g.Width)
	}

	// Add snakes
	for _, snake := range g.Snakes {
		points := snake.convertToPoints(g.Width, g.Height)
		for _, point := range points {
			field[point.y][point.x] = cellState_SNAKE
		}
	}

	// Add foods
	for _, coord := range g.Foods {
		field[coord.y][coord.x] = cellState_FOOD
	}

	return field
}

func createSnake(field [][]cellState) ([]Coord, error) {
	headCell, err := findHeadCell(field)
	if err != nil {
		return nil, err
	}

	tailCell, err := findTailCell(field, headCell)
	if err != nil {
		return nil, err
	}

	return []Coord{headCell, tailCell}, nil
}

func createFoods(field [][]cellState, count int32) []Coord {
	if count <= 0 {
		return []Coord{}
	}

	emptyCells := getAllEmptyCells(field)
	if int32(len(emptyCells)) <= count {
		return emptyCells
	}

	foodCells := make([]Coord, count)
	for i := int32(0); i < count; i++ {
		cellIdx := rand.Intn(len(emptyCells))
		foodCells[i] = emptyCells[cellIdx]
		emptyCells = append(emptyCells[:cellIdx], emptyCells[cellIdx+1:]...)
	}

	return foodCells
}

func findHeadCell(field [][]cellState) (Coord, error) {
	height := int32(len(field))
	width := int32(len(field[0]))
	emptyCells := getAllEmptyCells(field)

	// Checking that the 5*5 area centered at the selected point does not contain other snakes
	for _, cell := range emptyCells {
		check := true
		for k := int32(-2); k <= 2 && check; k++ {
			neighbourY := (cell.y + k + height) % height
			for l := int32(-2); l <= 2 && check; l++ {
				neighbourX := (cell.x + l + width) % width
				if field[neighbourY][neighbourX] == cellState_SNAKE {
					check = false
				}
			}
		}

		if check {
			return cell, nil
		}
	}

	return Coord{}, headPlaceNotFoundError
}

func findTailCell(field [][]cellState, headCell Coord) (Coord, error) {
	height := int32(len(field))
	width := int32(len(field[0]))

	// Check that the snake's tail will occupy a cell free of food
	availableTailCoords := []Coord{
		{0, 1}, {1, 0}, {0, -1}, {-1, 0},
	}
	for _, availableTailCoord := range availableTailCoords {
		tailCellY := (headCell.y + availableTailCoord.y + height) % height
		tailCellX := (headCell.x + availableTailCoord.x + width) % width
		if field[tailCellY][tailCellX] != cellState_FOOD {
			return availableTailCoord, nil
		}
	}
	return Coord{}, tailPlaceNotFoundError
}

func getAllEmptyCells(field [][]cellState) []Coord {
	emptyCells := make([]Coord, 0)
	for i := int32(0); i < int32(len(field)); i++ {
		for j := int32(0); j < int32(len(field[i])); j++ {
			if field[i][j] == cellState_EMPTY {
				emptyCells = append(emptyCells, Coord{
					x: j,
					y: i,
				})
			}
		}
	}

	return emptyCells
}
