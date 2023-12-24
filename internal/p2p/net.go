package p2p

import (
	"net"
	"strings"
	"time"

	"google.golang.org/protobuf/runtime/protoimpl"

	"p2p-snake/internal/log"
	"p2p-snake/internal/p2p/game"
	"p2p-snake/internal/p2p/protocol"
	"p2p-snake/internal/util"
)

func (p *Peer) receiveUnicastProto(msg *protocol.GameMessage) (*net.UDPAddr, bool) {
	addr, err := util.ReceiveProto(msg, p.unicast)
	if err != nil {
		// Error due to timeout
		if !strings.Contains(err.Error(), "i/o timeout") {
			log.Logger.Errorf("P2P node error: %v", err)
		}
		return nil, false
	}
	return addr, true
}

func (p *Peer) receiveMulticastProto(msg *protocol.GameMessage) (*net.UDPAddr, bool) {
	addr, err := util.ReceiveProto(msg, p.multicast)
	if err != nil {
		// Error due to timeout
		if !strings.Contains(err.Error(), "i/o timeout") {
			log.Logger.Errorf("P2P node error: %v", err)
		}
		return nil, false
	}
	return addr, true
}

func (p *Peer) sendProto(msg *protocol.GameMessage, addr *net.UDPAddr) *protocol.GameMessage {
	err := util.SendProto(msg, p.unicast, addr)
	if err != nil {
		log.Logger.Debugf("P2P node error: %v, Message: %v", err, protoimpl.X.MessageStringOf(msg))
		return nil
	}
	return msg
}

func (p *Peer) sendProtoWithResponse(msg *protocol.GameMessage, timeout time.Duration, addr *net.UDPAddr) (*protocol.GameMessage, *protocol.GameMessage) {
	respCh := make(chan *protocol.GameMessage)
	p.notAckMsg[msg.GetMsgSeq()] = respCh

	p.sendProto(msg, addr)

	timeoutCh := time.After(timeout)
	select {
	case response := <-respCh:
		close(p.notAckMsg[msg.GetMsgSeq()])
		delete(p.notAckMsg, msg.GetMsgSeq())
		return msg, response
	case <-timeoutCh:
		close(p.notAckMsg[msg.GetMsgSeq()])
		delete(p.notAckMsg, msg.GetMsgSeq())
		return msg, nil
	}
}

func (p *Peer) sendAckMsg(msgSeq int64, senderId int32, receiverId int32, addr *net.UDPAddr) *protocol.GameMessage {
	return p.sendProto(
		protocol.NewAckMsg(msgSeq, senderId, receiverId),
		addr,
	)
}

func (p *Peer) sendAnnouncementMsg(gameInfo *game.GameInfo, addr *net.UDPAddr) *protocol.GameMessage {
	curMsgSeq := p.msgSeq.Load()
	p.msgSeq.Add(1)
	return p.sendProto(
		protocol.NewAnnouncementMsg(
			curMsgSeq,
			gameInfo.GameName(),
			gameInfo.Width(),
			gameInfo.Height(),
			gameInfo.FoodStatic(),
			int32(gameInfo.StateDelay()/time.Millisecond),
			gameInfo.Players(),
		),
		addr,
	)
}

func (p *Peer) sendErrorMsg(msgSeq int64, senderId int32, receiverId int32, error string, addr *net.UDPAddr) *protocol.GameMessage {
	return p.sendProto(
		protocol.NewErrorMsg(msgSeq, senderId, receiverId, error),
		addr,
	)
}

func (p *Peer) sendDiscoverMsg(addr *net.UDPAddr) *protocol.GameMessage {
	curMsgSeq := p.msgSeq.Load()
	p.msgSeq.Add(1)
	return p.sendProto(
		protocol.NewDiscoverMsg(curMsgSeq),
		addr,
	)
}

func (p *Peer) sendJoinMsg(gameName string, playerName string, role protocol.NodeRole, addr *net.UDPAddr) (*protocol.GameMessage, *protocol.GameMessage) {
	curMsgSeq := p.msgSeq.Load()
	p.msgSeq.Add(1)
	return p.sendProtoWithResponse(
		protocol.NewJoinMsg(curMsgSeq, gameName, playerName, role),
		time.Second,
		addr,
	)
}

func (p *Peer) sendPingMsg(senderId int32, receiverId int32, addr *net.UDPAddr) *protocol.GameMessage {
	curMsgSeq := p.msgSeq.Load()
	p.msgSeq.Add(1)
	return p.sendProto(
		protocol.NewPingMsg(curMsgSeq, senderId, receiverId),
		addr,
	)
}

func (p *Peer) sendRoleChangeMsg(senderId int32, receiverId int32, senderRole *protocol.NodeRole, receiverRole *protocol.NodeRole, addr *net.UDPAddr) (*protocol.GameMessage, *protocol.GameMessage) {
	curMsgSeq := p.msgSeq.Load()
	p.msgSeq.Add(1)
	return p.sendProtoWithResponse(
		protocol.NewRoleChangeMsg(curMsgSeq, senderId, receiverId, senderRole, receiverRole),
		p.gameInfo.StateDelay()*8/10,
		addr,
	)
}

func (p *Peer) sendStateMsg(senderId int32, receiverId int32, gameInfo *game.GameInfo, addr *net.UDPAddr) (*protocol.GameMessage, *protocol.GameMessage) {
	curMsgSeq := p.msgSeq.Load()
	p.msgSeq.Add(1)
	return p.sendProtoWithResponse(
		protocol.NewStateMsg(
			curMsgSeq,
			senderId,
			receiverId,
			gameInfo.StateOrder(),
			gameInfo.Snakes(),
			gameInfo.Foods(),
			gameInfo.Players(),
		),
		p.gameInfo.StateDelay()*8/10,
		addr,
	)
}

func (p *Peer) sendSteerMsg(senderId int32, receiverId int32, direction protocol.Direction, addr *net.UDPAddr) (*protocol.GameMessage, *protocol.GameMessage) {
	curMsgSeq := p.msgSeq.Load()
	p.msgSeq.Add(1)
	return p.sendProtoWithResponse(
		protocol.NewSteerMsg(curMsgSeq, senderId, receiverId, direction),
		p.gameInfo.StateDelay()*8/10,
		addr,
	)
}
