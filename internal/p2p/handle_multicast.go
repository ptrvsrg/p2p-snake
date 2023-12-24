package p2p

import (
	"context"
	"net"

	"p2p-snake/internal/log"
	"p2p-snake/internal/p2p/announcements"
	"p2p-snake/internal/p2p/protocol"
)

func (p *Peer) listenMulticast(ctx context.Context) {
	defer p.wg.Done()

	log.Logger.Debug("listenMulticast goroutine is running")
	for {
		select {
		case <-ctx.Done():
			log.Logger.Debug("listenMulticast goroutine has completed")
			return
		default:
			gameMsg := &protocol.GameMessage{}
			addr, ok := p.receiveMulticastProto(gameMsg)
			if ok {
				go p.handleMulticastMsg(gameMsg, addr)
			}
		}
	}
}

func (p *Peer) handleMulticastMsg(gameMsg *protocol.GameMessage, addr *net.UDPAddr) {
	switch gameMsg.GetType().(type) {
	case *protocol.GameMessage_Announcement:
		p.handleAnnouncementMsg(gameMsg.GetAnnouncement(), addr)
	case *protocol.GameMessage_Discover:
		p.handleDiscoverMsg()
	}
}

func (p *Peer) handleAnnouncementMsg(announcementMsg *protocol.GameMessage_AnnouncementMsg, addr *net.UDPAddr) {
	if p.gameInfo != nil && announcementMsg.GetGames()[0].GetGameName() == p.gameInfo.GameName() {
		return
	}
	p.announcementCollector.AddAnnouncement(
		announcements.NewAnnouncement(
			addr,
			announcementMsg.GetGames()[0].GetGameName(),
			announcementMsg.GetGames()[0].GetConfig().GetWidth(),
			announcementMsg.GetGames()[0].GetConfig().GetHeight(),
			announcementMsg.GetGames()[0].GetConfig().GetFoodStatic(),
			announcementMsg.GetGames()[0].GetConfig().GetStateDelayMs(),
		),
	)
}

func (p *Peer) handleDiscoverMsg() {
	if p.gameInfo != nil && p.gameInfo.CurrentNode().IsMasterNode() {
		p.sendAnnouncementMsg(p.gameInfo, p.multicastAddr)
	}
}
