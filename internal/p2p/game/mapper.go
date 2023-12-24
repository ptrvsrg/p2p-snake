package game

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/protobuf/proto"

	"p2p-snake/internal/engine"
	"p2p-snake/internal/p2p/protocol"
)

//////// ENGINE -> P2P ////////

func toConfig(width int32, height int32, foodStatic int32, stateDelay time.Duration) *protocol.GameConfig {
	return &protocol.GameConfig{
		Width:        proto.Int32(width),
		Height:       proto.Int32(height),
		FoodStatic:   proto.Int32(foodStatic),
		StateDelayMs: proto.Int32(int32(stateDelay / time.Millisecond)),
	}
}

func toCoord(coord engine.Coord) *protocol.GameState_Coord {
	return &protocol.GameState_Coord{
		X: proto.Int32(coord.X()),
		Y: proto.Int32(coord.Y()),
	}
}

func toCoords(engineCoords []engine.Coord) []*protocol.GameState_Coord {
	coords := make([]*protocol.GameState_Coord, len(engineCoords))
	for i, engineCoord := range engineCoords {
		coords[i] = toCoord(engineCoord)
	}
	return coords
}

func toSnakeState(isZombie bool) protocol.GameState_Snake_SnakeState {
	if isZombie {
		return protocol.GameState_Snake_ZOMBIE
	}
	return protocol.GameState_Snake_ALIVE
}

func toDirection(engineDirection engine.Direction) protocol.Direction {
	switch engineDirection {
	case engine.UP:
		return protocol.Direction_UP
	case engine.DOWN:
		return protocol.Direction_DOWN
	case engine.LEFT:
		return protocol.Direction_LEFT
	case engine.RIGHT:
		return protocol.Direction_RIGHT
	}
	return -1
}

func toSnake(snake *engine.Snake) *protocol.GameState_Snake {
	coords := toCoords(snake.Points)
	state := toSnakeState(snake.IsZombie)
	headDirection := toDirection(snake.HeadDirection)

	return &protocol.GameState_Snake{
		PlayerId:      proto.Int32(snake.PlayerId),
		State:         state.Enum(),
		HeadDirection: headDirection.Enum(),
		Points:        coords,
	}
}

func toSnakes(engineSnakes map[int32]*engine.Snake) []*protocol.GameState_Snake {
	snakes := make([]*protocol.GameState_Snake, 0)
	for _, engineSnake := range engineSnakes {
		snakes = append(snakes, toSnake(engineSnake))
	}
	return snakes
}

func toPlayer(enginePlayer *engine.Player, nodeInfo *NodeInfo) *protocol.GamePlayer {
	var ip *string = nil
	var port *int32 = nil
	if nodeInfo.addr != nil {
		ip = proto.String(nodeInfo.addr.IP.String())
		port = proto.Int32(int32(nodeInfo.addr.Port))
	}

	return &protocol.GamePlayer{
		Name:      proto.String(enginePlayer.Name),
		Id:        proto.Int32(enginePlayer.Id),
		IpAddress: ip,
		Port:      port,
		Role:      nodeInfo.role.Enum(),
		Type:      protocol.Default_GamePlayer_Type.Enum(),
		Score:     proto.Int32(enginePlayer.Score),
	}
}

func toPlayers(enginePlayers map[int32]*engine.Player, nodeInfos map[int32]*NodeInfo) *protocol.GamePlayers {
	gamePlayers := make([]*protocol.GamePlayer, 0)
	for playerId, enginePlayer := range enginePlayers {
		gamePlayers = append(gamePlayers, toPlayer(enginePlayer, nodeInfos[playerId]))
	}
	return &protocol.GamePlayers{
		Players: gamePlayers,
	}
}

//////// P2P -> ENGINE ////////

func toEngineCoord(coord *protocol.GameState_Coord) engine.Coord {
	return engine.NewCoord(coord.GetX(), coord.GetY())
}

func toEngineCoords(coords []*protocol.GameState_Coord) []engine.Coord {
	engineCoords := make([]engine.Coord, len(coords))
	for i, coord := range coords {
		engineCoords[i] = toEngineCoord(coord)
	}
	return engineCoords
}

func toEngineSnakeState(state protocol.GameState_Snake_SnakeState) bool {
	return state == protocol.GameState_Snake_ZOMBIE
}

func toEngineDirection(direction protocol.Direction) engine.Direction {
	switch direction {
	case protocol.Direction_UP:
		return engine.UP
	case protocol.Direction_DOWN:
		return engine.DOWN
	case protocol.Direction_LEFT:
		return engine.LEFT
	case protocol.Direction_RIGHT:
		return engine.RIGHT
	}
	return -1
}

func toEngineSnake(snake *protocol.GameState_Snake) *engine.Snake {
	playerId := snake.GetPlayerId()
	coords := toEngineCoords(snake.Points)
	state := toEngineSnakeState(snake.GetState())
	headDirection := toEngineDirection(snake.GetHeadDirection())

	return engine.NewSnake(
		playerId,
		coords,
		state,
		headDirection,
		false,
	)
}

func toEngineSnakes(snakes []*protocol.GameState_Snake) map[int32]*engine.Snake {
	engineSnakes := make(map[int32]*engine.Snake)
	for _, snake := range snakes {
		engineSnakes[snake.GetPlayerId()] = toEngineSnake(snake)
	}
	return engineSnakes
}

func toEnginePlayer(gamePlayer *protocol.GamePlayer) *engine.Player {
	return engine.NewPlayer(
		gamePlayer.GetId(),
		gamePlayer.GetName(),
		gamePlayer.GetScore(),
	)
}

func toEnginePlayers(gamePlayers *protocol.GamePlayers) map[int32]*engine.Player {
	players := make(map[int32]*engine.Player)
	for _, gamePlayer := range gamePlayers.GetPlayers() {
		players[gamePlayer.GetId()] = toEnginePlayer(gamePlayer)
	}
	return players
}

func toNodeInfo(gamePlayer *protocol.GamePlayer) *NodeInfo {
	var addr *net.UDPAddr
	if gamePlayer.IpAddress == nil && gamePlayer.Port == nil {
		addr = nil
	} else {
		addr, _ = net.ResolveUDPAddr("udp",
			fmt.Sprintf("%s:%d", gamePlayer.GetIpAddress(), gamePlayer.GetPort()))
	}

	return NewNodeInfo(
		gamePlayer.GetId(),
		gamePlayer.GetRole(),
		addr,
	)
}

func toNodeInfos(gamePlayers *protocol.GamePlayers) map[int32]*NodeInfo {
	nodeInfos := make(map[int32]*NodeInfo)
	for _, gamePlayer := range gamePlayers.GetPlayers() {
		nodeInfos[gamePlayer.GetId()] = toNodeInfo(gamePlayer)
	}
	return nodeInfos
}
