package api

import (
	"net"
	"strings"
	"time"

	"p2p-snake/internal/api/protocol"
	"p2p-snake/internal/log"
	"p2p-snake/internal/p2p/dto"
	"p2p-snake/internal/util"
)

func (server *Server) receiveProto(request *protocol.APIRequest) (*net.UDPAddr, bool) {
	addr, err := util.ReceiveProto(request, server.conn)
	if err != nil {
		// Error due to timeout
		if !strings.Contains(err.Error(), "i/o timeout") {
			log.Logger.Errorf("API server error: %v", err)
		}
		return nil, false
	}
	return addr, true
}

func (server *Server) sendProto(response *protocol.APIResponse, addr *net.UDPAddr) {
	if err := util.SendProto(response, server.conn, addr); err != nil {
		log.Logger.Errorf("API server error: %v", err)
	}
}

func (server *Server) sendError(error string, addr *net.UDPAddr) {
	server.sendProto(protocol.NewError(error), addr)
}

func (server *Server) sendSuccessConnect(addr *net.UDPAddr) {
	server.sendProto(protocol.NewSuccessConnect(server.token, int32(server.timeout/time.Millisecond)), addr)
}

func (server *Server) sendAck(addr *net.UDPAddr) {
	server.sendProto(protocol.NewAck(), addr)
}

func (server *Server) sendGameState(stateDto dto.GameStateDto, addr *net.UDPAddr) {
	server.sendProto(protocol.NewGameState(stateDto), addr)
}

func (server *Server) sendGameList(games []dto.GameInfoDto, addr *net.UDPAddr) {
	server.sendProto(protocol.NewGameList(games), addr)
}
