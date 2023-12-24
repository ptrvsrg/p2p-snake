package engine

import (
	"p2p-snake/internal/log"
)

type Direction int

const (
	UP    Direction = 1
	LEFT  Direction = 2
	DOWN  Direction = 3
	RIGHT Direction = 4
)

type Snake struct {
	PlayerId      int32
	Points        []Coord
	IsZombie      bool
	HeadDirection Direction
	IsEating      bool
}

func NewSnake(playerId int32, points []Coord, IsZombie bool, headDirection Direction, isEating bool) *Snake {
	return &Snake{
		PlayerId:      playerId,
		Points:        points,
		IsZombie:      IsZombie,
		HeadDirection: headDirection,
		IsEating:      isEating,
	}
}

func (s *Snake) Move(direction Direction, width int32, height int32) {
	if direction < 1 || direction > 4 {
		log.Logger.Errorf("Not valid direction")
		return
	}

	head := s.Points[0]
	if s.HeadDirection == direction || // The snake moves in the original direction
		s.HeadDirection-direction == 2 || // Received the opposite direction (Snake moves in the original direction)
		s.HeadDirection-direction == -2 {
		switch s.HeadDirection {
		case UP:
			s.Points[0].y = (head.y - 1 + height) % height
			s.Points[1].y = s.Points[1].y + 1
		case DOWN:
			s.Points[0].y = (head.y + 1) % height
			s.Points[1].y = s.Points[1].y - 1
		case LEFT:
			s.Points[0].x = (head.x - 1 + width) % width
			s.Points[1].x = s.Points[1].x + 1
		case RIGHT:
			s.Points[0].x = (head.x + 1) % width
			s.Points[1].x = s.Points[1].x - 1
		}
	} else { // Direction has changed
		switch direction {
		case UP:
			s.Points = append([]Coord{{0, 1}}, s.Points[1:]...)
			s.Points = append([]Coord{{head.x, (head.y - 1 + height) % height}}, s.Points...)
		case DOWN:
			s.Points = append([]Coord{{0, -1}}, s.Points[1:]...)
			s.Points = append([]Coord{{head.x, (head.y + 1) % height}}, s.Points...)
		case LEFT:
			s.Points = append([]Coord{{1, 0}}, s.Points[1:]...)
			s.Points = append([]Coord{{(head.x - 1 + width) % width, head.y}}, s.Points...)
		case RIGHT:
			s.Points = append([]Coord{{-1, 0}}, s.Points[1:]...)
			s.Points = append([]Coord{{(head.x + 1) % width, head.y}}, s.Points...)
		}
		s.HeadDirection = direction
	}

	if !s.IsEating { // The snake didn’t eat last turn and should shrink
		tail := s.Points[len(s.Points)-1]
		if tail.x != 0 {
			if tail.x > 0 {
				tail.x = tail.x - 1
			} else {
				tail.x = tail.x + 1
			}

			if tail.x == 0 { // The tail occupied one cell
				s.Points = s.Points[:len(s.Points)-1]
			} else {
				s.Points[len(s.Points)-1] = tail
			}
		} else {
			if tail.y > 0 {
				tail.y = tail.y - 1
			} else {
				tail.y = tail.y + 1
			}

			if tail.y == 0 { // The tail occupied one cell
				s.Points = s.Points[:len(s.Points)-1]
			} else {
				s.Points[len(s.Points)-1] = tail
			}
		}
	} else {
		s.IsEating = false
	}
}

func (s *Snake) convertToPoints(width int32, height int32) []Coord {
	points := make([]Coord, 0)

	x := s.Points[0].x
	y := s.Points[0].y
	points = append(points, NewCoord(x, y))

	for idx, coord := range s.Points {
		if idx == 0 {
			continue
		}

		if coord.y != 0 {
			offset := coord.y
			for offset != 0 {
				if offset > 0 {
					y = (y + 1) % height
					offset--
				} else {
					y = (y - 1 + height) % height
					offset++
				}
				points = append(points, NewCoord(x, y))
			}
		} else {
			offset := coord.x
			for offset != 0 {
				if offset > 0 {
					x = (x + 1) % width
					offset--
				} else {
					x = (x - 1 + width) % width
					offset++
				}
				points = append(points, NewCoord(x, y))
			}
		}
	}

	return points
}
