package dto

import (
	"p2p-snake/internal/p2p/protocol"
)

//////// Game info DTO ////////

type GameInfoDto struct {
	Name       string
	Width      int32
	Height     int32
	StateDelay int32
}

func NewGameInfoDto(name string, width int32, height int32, stateDelay int32) GameInfoDto {
	return GameInfoDto{
		Name:       name,
		Width:      width,
		Height:     height,
		StateDelay: stateDelay,
	}
}

//////// Config DTO ////////

type ConfigDto struct {
	Width      int32
	Height     int32
	FoodStatic int32
	StateDelay int32
}

func NewConfigDto(width int32, height int32, foodStatic int32, stateDelay int32) ConfigDto {
	return ConfigDto{
		Width:      width,
		Height:     height,
		FoodStatic: foodStatic,
		StateDelay: stateDelay,
	}
}

//////// Coordinate DTO ////////

type CoordDto struct {
	X int32
	Y int32
}

func NewCoordDto(coord *protocol.GameState_Coord) CoordDto {
	return CoordDto{
		X: coord.GetX(),
		Y: coord.GetY(),
	}
}

//////// Snake DTO ////////

type Direction int32

const (
	UP    Direction = 1
	DOWN  Direction = 2
	LEFT  Direction = 3
	RIGHT Direction = 4
)

type SnakeDto struct {
	PlayerId      int32
	HeadDirection Direction
	Points        []CoordDto
}

func NewSnakeDto(playerId int32, headDirection Direction, points []CoordDto) SnakeDto {
	return SnakeDto{
		PlayerId:      playerId,
		HeadDirection: headDirection,
		Points:        points,
	}
}

//////// Player DTO ////////

type NodeRole int32

const (
	MASTER NodeRole = 1
	NORMAL NodeRole = 2
	DEPUTY NodeRole = 3
	VIEWER NodeRole = 4
)

type PlayerDto struct {
	Name  string
	Id    int32
	Score int32
	Role  NodeRole
}

func NewPlayerDto(name string, id int32, score int32, role NodeRole) PlayerDto {
	return PlayerDto{
		Name:  name,
		Id:    id,
		Score: score,
		Role:  role,
	}
}

//////// Game state DTO ////////

type GameStateDto struct {
	StateOrder int32
	Config     ConfigDto
	Snakes     []SnakeDto
	Foods      []CoordDto
	Players    []PlayerDto
}

func NewGameStateDto(stateOrder int32, config ConfigDto, snakes []SnakeDto, foods []CoordDto, players []PlayerDto) GameStateDto {
	return GameStateDto{
		StateOrder: stateOrder,
		Config:     config,
		Snakes:     snakes,
		Foods:      foods,
		Players:    players,
	}
}
