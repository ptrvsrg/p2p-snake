package protocol

import (
	"google.golang.org/protobuf/proto"
)

func NewAckMsg(msgSeq int64, senderId int32, receiverId int32) *GameMessage {
	return &GameMessage{
		MsgSeq:     proto.Int64(msgSeq),
		SenderId:   proto.Int32(senderId),
		ReceiverId: proto.Int32(receiverId),
		Type: &GameMessage_Ack{
			Ack: &GameMessage_AckMsg{},
		},
	}
}

func NewAnnouncementMsg(msgSeq int64, gameName string, width int32, height int32, foodStatic int32,
	stateDelay int32, players *GamePlayers) *GameMessage {
	return &GameMessage{
		MsgSeq: proto.Int64(msgSeq),
		Type: &GameMessage_Announcement{
			Announcement: &GameMessage_AnnouncementMsg{
				Games: []*GameAnnouncement{
					{
						GameName: proto.String(gameName),
						Config: &GameConfig{
							Width:        proto.Int32(width),
							Height:       proto.Int32(height),
							FoodStatic:   proto.Int32(foodStatic),
							StateDelayMs: proto.Int32(stateDelay),
						},
						CanJoin: proto.Bool(true),
						Players: players,
					},
				},
			},
		},
	}
}

func NewDiscoverMsg(msgSeq int64) *GameMessage {
	return &GameMessage{
		MsgSeq: proto.Int64(msgSeq),
		Type: &GameMessage_Discover{
			Discover: &GameMessage_DiscoverMsg{},
		},
	}
}

func NewErrorMsg(msgSeq int64, senderId int32, receiverId int32, error string) *GameMessage {
	return &GameMessage{
		MsgSeq:     proto.Int64(msgSeq),
		SenderId:   proto.Int32(senderId),
		ReceiverId: proto.Int32(receiverId),
		Type: &GameMessage_Error{
			Error: &GameMessage_ErrorMsg{
				ErrorMessage: proto.String(error),
			},
		},
	}
}

func NewJoinMsg(msgSeq int64, gameName string, playerName string, role NodeRole) *GameMessage {
	return &GameMessage{
		MsgSeq: proto.Int64(msgSeq),
		Type: &GameMessage_Join{
			Join: &GameMessage_JoinMsg{
				GameName:      proto.String(gameName),
				PlayerName:    proto.String(playerName),
				PlayerType:    (*PlayerType)(proto.Int32((int32)(Default_GamePlayer_Type))),
				RequestedRole: (*NodeRole)(proto.Int32((int32)(role))),
			},
		},
	}
}

func NewPingMsg(msgSeq int64, senderId int32, receiverId int32) *GameMessage {
	return &GameMessage{
		MsgSeq:     proto.Int64(msgSeq),
		SenderId:   proto.Int32(senderId),
		ReceiverId: proto.Int32(receiverId),
		Type: &GameMessage_Ping{
			Ping: &GameMessage_PingMsg{},
		},
	}
}

func NewRoleChangeMsg(msgSeq int64, senderId int32, receiverId int32, senderRole *NodeRole, receiverRole *NodeRole) *GameMessage {
	return &GameMessage{
		MsgSeq:     proto.Int64(msgSeq),
		SenderId:   proto.Int32(senderId),
		ReceiverId: proto.Int32(receiverId),
		Type: &GameMessage_RoleChange{
			RoleChange: &GameMessage_RoleChangeMsg{
				SenderRole:   senderRole,
				ReceiverRole: receiverRole,
			},
		},
	}
}

func NewStateMsg(msgSeq int64, senderId int32, receiverId int32, stateOrder int32,
	snakes []*GameState_Snake, foods []*GameState_Coord, players *GamePlayers) *GameMessage {
	return &GameMessage{
		MsgSeq:     proto.Int64(msgSeq),
		SenderId:   proto.Int32(senderId),
		ReceiverId: proto.Int32(receiverId),
		Type: &GameMessage_State{
			State: &GameMessage_StateMsg{
				State: &GameState{
					StateOrder: proto.Int32(stateOrder),
					Snakes:     snakes,
					Foods:      foods,
					Players:    players,
				},
			},
		},
	}
}

func NewSteerMsg(msgSeq int64, senderId int32, receiverId int32, direction Direction) *GameMessage {
	return &GameMessage{
		MsgSeq:     proto.Int64(msgSeq),
		SenderId:   proto.Int32(senderId),
		ReceiverId: proto.Int32(receiverId),
		Type: &GameMessage_Steer{
			Steer: &GameMessage_SteerMsg{
				Direction: (*Direction)(proto.Int32((int32)(direction))),
			},
		},
	}
}
