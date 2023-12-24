package engine

type Game struct {
	// Config
	Name       string
	Width      int32
	Height     int32
	FoodStatic int32

	// State
	Snakes  map[int32]*Snake
	Players map[int32]*Player
	Foods   []Coord
}

func NewGame(gameName string, width int32, height int32, foodStatic int32) *Game {
	return &Game{
		Name:       gameName,
		Width:      width,
		Height:     height,
		FoodStatic: foodStatic,

		Snakes:  make(map[int32]*Snake),
		Players: make(map[int32]*Player),
		Foods:   make([]Coord, 0),
	}
}

func (g *Game) AddPlayer(playerId int32, playerName string, withSnake bool) error {
	// Create player
	g.Players[playerId] = NewPlayer(playerId, playerName, 0)

	if withSnake {
		// Create snake
		field := g.createField()
		snakeCoords, err := createSnake(field)
		if err != nil {
			return err
		}

		var headDirection Direction
		switch {
		case snakeCoords[1].y > 0:
			headDirection = UP
		case snakeCoords[0].y < 0:
			headDirection = DOWN
		case snakeCoords[0].x > 0:
			headDirection = LEFT
		case snakeCoords[0].x < 0:
			headDirection = RIGHT
		}

		g.Snakes[playerId] = NewSnake(playerId, snakeCoords, false, headDirection, false)
	}

	return nil
}

func (g *Game) DeletePlayer(playerId int32) {
	delete(g.Players, playerId)
	g.Snakes[playerId].IsZombie = true
}

func (g *Game) addFood(field [][]cellState) {
	newFoodCoords := createFoods(field, g.FoodStatic+int32(len(g.Players))-int32(len(g.Foods)))
	g.Foods = append(g.Foods, newFoodCoords...)
}

func (g *Game) NextState(directionChanges map[int32]Direction) []int32 {
	for playerId, snake := range g.Snakes {
		if snake.IsZombie {
			continue
		}

		// Move the snake 1 cell
		if newDirection, ok := directionChanges[playerId]; ok {
			snake.Move(newDirection, g.Width, g.Height)
		} else {
			snake.Move(snake.HeadDirection, g.Width, g.Height)
		}

		// Check if the snake has eaten the food
		for idx, foodCoord := range g.Foods {
			headCoord := snake.Points[0]
			if foodCoord.x == headCoord.x && foodCoord.y == headCoord.y {
				g.Foods = append(g.Foods[:idx], g.Foods[idx+1:]...)
				snake.IsEating = true
				if player, ok := g.Players[playerId]; ok {
					player.Score++
				}
				break
			}
		}
	}

	// Check all pairs of snake for collision
	snakePoints := make(map[int32][]Coord)
	for playerId, snake := range g.Snakes {
		snakePoints[playerId] = snake.convertToPoints(g.Width, g.Height)
	}

	deadSnakes := make(map[int32]bool)
	for playerId1, snakePoints1 := range snakePoints {
		for playerId2, snakePoints2 := range snakePoints {
			if playerId1 == playerId2 {
				if containsPoint(snakePoints1[1:], snakePoints1[0]) {
					deadSnakes[playerId1] = true
				}
				continue
			}
			if containsPoint(snakePoints1, snakePoints2[0]) {
				deadSnakes[playerId2] = true
			}
			if containsPoint(snakePoints2, snakePoints1[0]) {
				deadSnakes[playerId1] = true
			}
		}
	}
	deadPlayerIds := make([]int32, 0)
	for playerId := range deadSnakes {
		delete(g.Snakes, playerId)
		deadPlayerIds = append(deadPlayerIds, playerId)
	}

	field := g.createField()
	g.addFood(field)

	return deadPlayerIds
}

func containsPoint(points []Coord, searchPoint Coord) bool {
	for _, point := range points {
		if searchPoint.x == point.x && searchPoint.y == point.y {
			return true
		}
	}
	return false
}
