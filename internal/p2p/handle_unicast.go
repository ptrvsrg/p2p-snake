package p2p

import (
	"context"
	"net"

	"p2p-snake/internal/log"
	"p2p-snake/internal/p2p/protocol"
)

const (
	gameNameNotMatchError    = "game name does not match"
	receiverIsNotInGameError = "receiver is not in any game"
	receiverIsMasterError    = "receiver is master for this game"
	receiverIsNotMasterError = "receiver is not master for this game"
	senderIsNotInGameError   = "sender is not in this game"
	senderIsViewerError      = "sender is viewer"
	duplicatePlayerNameError = "player with such name already exists"
	duplicatePlayerAddrError = "player with such address already exists"
)

func (p *Peer) listenUnicast(ctx context.Context) {
	defer p.wg.Done()

	log.Logger.Debug("listenUnicast goroutine is running")
	for {
		select {
		case <-ctx.Done():
			log.Logger.Debug("listenUnicast goroutine has completed")
			return
		default:
			gameMsg := &protocol.GameMessage{}
			addr, ok := p.receiveUnicastProto(gameMsg)
			if ok {
				go p.handleUnicastMsg(gameMsg, addr)
			}
		}
	}
}

func (p *Peer) handleUnicastMsg(gameMsg *protocol.GameMessage, addr *net.UDPAddr) {
	switch gameMsg.GetType().(type) {
	case *protocol.GameMessage_Ack:
		p.handleAckMsg(gameMsg, addr)
	case *protocol.GameMessage_Error:
		p.handleErrorMsg(gameMsg, addr)
	case *protocol.GameMessage_Join:
		p.handleJoinMsg(gameMsg, addr)
	case *protocol.GameMessage_Ping:
		p.handlePingMsg(gameMsg, addr)
	case *protocol.GameMessage_RoleChange:
		p.handleRoleChangeMsg(gameMsg, addr)
	case *protocol.GameMessage_State:
		p.handleStateMsg(gameMsg, addr)
	case *protocol.GameMessage_Steer:
		p.handleSteerMsg(gameMsg, addr)
	}
}

func (p *Peer) handleAckMsg(msg *protocol.GameMessage, addr *net.UDPAddr) {
	if p.gameInfo != nil && p.gameInfo.CurrentNode().PlayerId() != msg.GetReceiverId() {
		return
	}
	if response, ok := p.notAckMsg[msg.GetMsgSeq()]; ok {
		response <- msg
	}
	if p.gameInfo != nil {
		if node, ok := p.gameInfo.Node(msg.GetSenderId()); ok {
			node.UpdateTimeAsNow()
		}
	}
}

func (p *Peer) handleErrorMsg(msg *protocol.GameMessage, addr *net.UDPAddr) {
	if p.gameInfo != nil && p.gameInfo.CurrentNode().PlayerId() != msg.GetReceiverId() {
		return
	}
	if response, ok := p.notAckMsg[msg.GetMsgSeq()]; ok {
		response <- msg
	}
	if p.gameInfo != nil {
		if node, ok := p.gameInfo.Node(msg.GetSenderId()); ok {
			node.UpdateTimeAsNow()
		}
	}
}

func (p *Peer) handleJoinMsg(msg *protocol.GameMessage, addr *net.UDPAddr) {
	if p.gameInfo == nil {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), receiverIsNotInGameError, addr)
		return
	}
	if !p.gameInfo.CurrentNode().IsMasterNode() {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), receiverIsNotMasterError, addr)
		return
	}
	if msg.GetJoin().GetGameName() != p.gameInfo.GameName() {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), gameNameNotMatchError, addr)
		return
	}
	if p.gameInfo.ExistsPlayerByName(msg.GetJoin().GetPlayerName()) {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), duplicatePlayerNameError, addr)
		return
	}
	if p.gameInfo.ExistsPlayerByAddr(addr) {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), duplicatePlayerAddrError, addr)
		return
	}

	node, err := p.gameInfo.AddPlayer(msg.GetJoin().GetPlayerName(), msg.GetJoin().GetRequestedRole(), addr)
	if err != nil {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), err.Error(), addr)
		return
	}

	p.sendAckMsg(
		msg.GetMsgSeq(),
		p.gameInfo.CurrentNode().PlayerId(),
		node.PlayerId(),
		addr,
	)
}

func (p *Peer) handlePingMsg(msg *protocol.GameMessage, addr *net.UDPAddr) {
	if p.gameInfo == nil {
		return
	}
	if p.gameInfo.CurrentNode().PlayerId() != msg.GetReceiverId() {
		return
	}

	if node, ok := p.gameInfo.Node(msg.GetSenderId()); ok {
		node.UpdateTimeAsNow()
	}
}

func (p *Peer) handleRoleChangeMsg(msg *protocol.GameMessage, addr *net.UDPAddr) {
	if p.gameInfo == nil {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), receiverIsNotInGameError, addr)
		return
	}
	if p.gameInfo.CurrentNode().PlayerId() != msg.GetReceiverId() {
		return
	}
	if msg.GetRoleChange().SenderRole != nil {
		if node, ok := p.gameInfo.Node(msg.GetSenderId()); ok && msg.GetRoleChange().GetSenderRole() != protocol.NodeRole_MASTER {
			node.SetRole(msg.GetRoleChange().GetSenderRole())
		}
	}
	if msg.GetRoleChange().ReceiverRole != nil {
		p.gameInfo.CurrentNode().SetRole(msg.GetRoleChange().GetSenderRole())
	}
	p.sendAckMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), addr)

	if node, ok := p.gameInfo.Node(msg.GetSenderId()); ok {
		node.UpdateTimeAsNow()
	}
}

func (p *Peer) handleStateMsg(msg *protocol.GameMessage, addr *net.UDPAddr) {
	if p.gameInfo == nil {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), receiverIsNotInGameError, addr)
		return
	}
	if p.gameInfo.CurrentNode().PlayerId() != msg.GetReceiverId() {
		return
	}
	if p.gameInfo.CurrentNode().IsMasterNode() {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), receiverIsMasterError, addr)
		return
	}
	if p.gameInfo.StateOrder() >= msg.GetState().GetState().GetStateOrder() {
		return
	}

	p.sendAckMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), addr)
	p.gameInfo.SetState(msg.GetReceiverId(), msg.GetState().GetState(), addr)
	if node, ok := p.gameInfo.Node(msg.GetSenderId()); ok {
		node.UpdateTimeAsNow()
	}
}

func (p *Peer) handleSteerMsg(msg *protocol.GameMessage, addr *net.UDPAddr) {
	if p.gameInfo == nil {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), receiverIsNotInGameError, addr)
		return
	}
	if p.gameInfo.CurrentNode().PlayerId() != msg.GetReceiverId() {
		return
	}
	if !p.gameInfo.CurrentNode().IsMasterNode() {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), receiverIsNotMasterError, addr)
		return
	}

	node, ok := p.gameInfo.Node(msg.GetSenderId())
	if !ok {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), senderIsNotInGameError, addr)
		return
	}
	if node.Role() == protocol.NodeRole_VIEWER {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), senderIsViewerError, addr)
		return
	}

	if err := p.gameInfo.AddMove(msg.GetSenderId(), msg.GetSteer().GetDirection()); err != nil {
		p.sendErrorMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), err.Error(), addr)
		return
	}
	p.sendAckMsg(msg.GetMsgSeq(), msg.GetReceiverId(), msg.GetSenderId(), addr)
	if node, ok := p.gameInfo.Node(msg.GetSenderId()); ok {
		node.UpdateTimeAsNow()
	}
}
