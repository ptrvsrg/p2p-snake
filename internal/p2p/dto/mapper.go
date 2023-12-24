package dto

import "p2p-snake/internal/p2p/protocol"

func toConfigDto(config *protocol.GameConfig) ConfigDto {
	return NewConfigDto(
		config.GetWidth(),
		config.GetHeight(),
		config.GetFoodStatic(),
		config.GetStateDelayMs(),
	)
}

func toSnakeDto(snake *protocol.GameState_Snake) SnakeDto {
	return NewSnakeDto(
		snake.GetPlayerId(),
		toDirection(snake.GetHeadDirection()),
		toCoordDtos(snake.GetPoints()),
	)
}

func toDirection(direction protocol.Direction) Direction {
	switch direction {
	case protocol.Direction_UP:
		return UP
	case protocol.Direction_DOWN:
		return DOWN
	case protocol.Direction_LEFT:
		return LEFT
	case protocol.Direction_RIGHT:
		return RIGHT
	}
	return 0
}

func toRole(role protocol.NodeRole) NodeRole {
	switch role {
	case protocol.NodeRole_MASTER:
		return MASTER
	case protocol.NodeRole_NORMAL:
		return NORMAL
	case protocol.NodeRole_DEPUTY:
		return DEPUTY
	case protocol.NodeRole_VIEWER:
		return VIEWER
	}
	return 0
}

func toCoordDtos(coords []*protocol.GameState_Coord) []CoordDto {
	coordDtos := make([]CoordDto, len(coords))
	for i, coord := range coords {
		coordDtos[i] = NewCoordDto(coord)
	}
	return coordDtos
}

func toSnakeDtos(snakes []*protocol.GameState_Snake) []SnakeDto {
	snakeDtos := make([]SnakeDto, len(snakes))
	for i, snake := range snakes {
		snakeDtos[i] = toSnakeDto(snake)
	}
	return snakeDtos
}

func toPlayerDto(player *protocol.GamePlayer) PlayerDto {
	return NewPlayerDto(
		player.GetName(),
		player.GetId(),
		player.GetScore(),
		toRole(player.GetRole()),
	)
}

func toPlayerDtos(players *protocol.GamePlayers) []PlayerDto {
	playerDtos := make([]PlayerDto, len(players.GetPlayers()))
	for i, player := range players.GetPlayers() {
		playerDtos[i] = toPlayerDto(player)
	}
	return playerDtos
}

func ToGameStateDto(
	stateOrder int32,
	config *protocol.GameConfig,
	snakes []*protocol.GameState_Snake,
	foods []*protocol.GameState_Coord,
	players *protocol.GamePlayers) GameStateDto {
	return NewGameStateDto(
		stateOrder,
		toConfigDto(config),
		toSnakeDtos(snakes),
		toCoordDtos(foods),
		toPlayerDtos(players),
	)
}
