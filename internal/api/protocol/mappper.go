package protocol

import (
	"google.golang.org/protobuf/proto"

	"p2p-snake/internal/p2p/dto"
	"p2p-snake/internal/p2p/protocol"
)

func NewError(error string) *APIResponse {
	return &APIResponse{
		Type: &APIResponse_Error{
			Error: &APIResponse_ErrorMsg{
				ErrorMessage: &error,
			},
		},
	}
}

func NewSuccessConnect(token string, timeout int32) *APIResponse {
	return &APIResponse{
		Type: &APIResponse_SuccessConnect{
			SuccessConnect: &APIResponse_SuccessConnectMsg{
				Token:   proto.String(token),
				Timeout: proto.Int32(timeout),
			},
		},
	}
}

func NewAck() *APIResponse {
	return &APIResponse{
		Type: &APIResponse_Ack{
			Ack: &APIResponse_AckMsg{},
		},
	}
}

func NewGameState(stateDto dto.GameStateDto) *APIResponse {
	return &APIResponse{
		Type: &APIResponse_GameState{
			GameState: &APIResponse_GameStateMsg{
				Snakes:  mapToP2PSnakes(stateDto.Snakes),
				Foods:   mapToCoords(stateDto.Foods),
				Players: mapToPlayers(stateDto.Players),
			},
		},
	}
}

func NewGameList(games []dto.GameInfoDto) *APIResponse {
	return &APIResponse{
		Type: &APIResponse_GameList{
			GameList: &APIResponse_GameListMsg{
				Games: mapToGameInfos(games),
			},
		},
	}
}

func mapToP2PSnakes(snakeDtos []dto.SnakeDto) []*APIResponse_GameStateMsg_Snake {
	snakes := make([]*APIResponse_GameStateMsg_Snake, len(snakeDtos))
	for i, snakeDto := range snakeDtos {
		snakes[i] = &APIResponse_GameStateMsg_Snake{
			PlayerId:      proto.Int32(snakeDto.PlayerId),
			HeadDirection: (*Direction)(proto.Int32((int32)(MapToP2PDirection(Direction(snakeDto.HeadDirection))))),
			Points:        mapToCoords(snakeDto.Points),
		}
	}
	return snakes
}

func mapToDirection(p2pDirection protocol.Direction) Direction {
	switch p2pDirection {
	case protocol.Direction_UP:
		return Direction_UP
	case protocol.Direction_DOWN:
		return Direction_DOWN
	case protocol.Direction_LEFT:
		return Direction_LEFT
	case protocol.Direction_RIGHT:
		return Direction_RIGHT
	}
	return 0
}

func MapToP2PDirection(direction Direction) protocol.Direction {
	switch direction {
	case Direction_UP:
		return protocol.Direction_UP
	case Direction_DOWN:
		return protocol.Direction_DOWN
	case Direction_LEFT:
		return protocol.Direction_LEFT
	case Direction_RIGHT:
		return protocol.Direction_RIGHT
	}
	return 0
}

func mapToCoords(coordDtos []dto.CoordDto) []*APIResponse_GameStateMsg_Coord {
	coords := make([]*APIResponse_GameStateMsg_Coord, len(coordDtos))
	for i, coordDto := range coordDtos {
		coords[i] = &APIResponse_GameStateMsg_Coord{
			X: proto.Int32(coordDto.X),
			Y: proto.Int32(coordDto.Y),
		}
	}
	return coords
}

func mapToNodeRole(roleDto dto.NodeRole) *APIResponse_GameStateMsg_Role {
	switch roleDto {
	case dto.MASTER:
		return APIResponse_GameStateMsg_MASTER.Enum()
	case dto.NORMAL:
		return APIResponse_GameStateMsg_NORMAL.Enum()
	case dto.DEPUTY:
		return APIResponse_GameStateMsg_DEPUTY.Enum()
	case dto.VIEWER:
		return APIResponse_GameStateMsg_VIEWER.Enum()
	}
	return nil
}

func mapToPlayers(playerDtos []dto.PlayerDto) []*APIResponse_GameStateMsg_Player {
	players := make([]*APIResponse_GameStateMsg_Player, len(playerDtos))
	for i, playerDto := range playerDtos {
		players[i] = &APIResponse_GameStateMsg_Player{
			Name:  proto.String(playerDto.Name),
			Id:    proto.Int32(playerDto.Id),
			Score: proto.Int32(playerDto.Score),
			Role:  mapToNodeRole(playerDto.Role),
		}
	}
	return players
}

func mapToGameInfos(gameInfoDtos []dto.GameInfoDto) []*APIResponse_GameListMsg_GameInfo {
	games := make([]*APIResponse_GameListMsg_GameInfo, len(gameInfoDtos))
	for i, gameInfoDto := range gameInfoDtos {
		games[i] = &APIResponse_GameListMsg_GameInfo{
			GameName:   proto.String(gameInfoDto.Name),
			Width:      proto.Int32(gameInfoDto.Width),
			Height:     proto.Int32(gameInfoDto.Height),
			StateDelay: proto.Int32(gameInfoDto.StateDelay),
		}
	}
	return games
}
