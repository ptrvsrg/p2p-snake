package api

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"

	"p2p-snake/internal/api/protocol"
	"p2p-snake/internal/log"
	"p2p-snake/internal/p2p"
)

const (
	unrecognizedRequestError = "unrecognized API request"
	nodeIsBusyError          = "node is busy"
	apiTokenCreationError    = "failed to create API token."
	notValidTokenError       = "not valid token (maybe, client has changed)"
)

type Server struct {
	// Network
	port    int
	timeout time.Duration
	conn    *net.UDPConn

	// Peer
	node *p2p.Peer

	// Connect state
	isFree          bool
	token           string
	lastRequestTime time.Time

	// Close
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func NewServer(port int, timeout time.Duration, node *p2p.Peer) *Server {
	return &Server{
		port:    port,
		timeout: timeout,

		node: node,

		isFree: true,

		cancel: func() {},
		wg:     &sync.WaitGroup{},
	}
}

func (server *Server) IsFree() bool {
	return server.isFree
}

func (server *Server) Start() error {
	server.cancel()
	server.wg.Wait()

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", server.port))
	if err != nil {
		return err
	}

	server.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	log.Logger.Infof("API server is listening on %v", addr.String())

	// Start p2p node
	if err := server.node.Start(); err != nil {
		return err
	}

	var ctx context.Context
	ctx, server.cancel = context.WithCancel(context.Background())

	server.wg.Add(2)
	go server.checkTimeout(ctx)
	go server.listen(ctx)

	return nil
}

func (server *Server) checkTimeout(ctx context.Context) {
	defer server.wg.Done()

	log.Logger.Debug("checkTimeout goroutine is running")
	for {
		select {
		case <-ctx.Done():
			log.Logger.Debug("checkTimeout goroutine has completed")
			return
		case <-time.After(server.timeout):
			if !server.IsFree() && time.Since(server.lastRequestTime) > server.timeout {
				server.token = ""
				server.isFree = true
				_ = server.node.ExitGame()
				log.Logger.Infof("client connection with the client is lost")
			}
		}
	}
}

func (server *Server) listen(ctx context.Context) {
	defer server.wg.Done()

	log.Logger.Debug("listen goroutine is running")
	for {
		select {
		case <-ctx.Done():
			log.Logger.Debug("listen goroutine has completed")
			return
		default:
			request := &protocol.APIRequest{}
			addr, ok := server.receiveProto(request)
			if ok {
				go server.handleMessage(request, addr)
			}
		}
	}
}

func (server *Server) Close() error {
	server.cancel()
	server.wg.Wait()

	return server.conn.Close()
}

func (server *Server) handleMessage(request *protocol.APIRequest, addr *net.UDPAddr) {
	switch request.GetType().(type) {
	case *protocol.APIRequest_Connect:
		server.handleConnect(addr)
	case *protocol.APIRequest_Ping:
		server.handlePing(request.GetPing(), addr)
	case *protocol.APIRequest_CreateGame:
		server.handleCreateGame(request.GetCreateGame(), addr)
	case *protocol.APIRequest_DiscoverGames:
		server.handleDiscoverGame(request.GetDiscoverGames(), addr)
	case *protocol.APIRequest_JoinGame:
		server.handleJoinGame(request.GetJoinGame(), addr)
	case *protocol.APIRequest_SteerSnake:
		server.handleSteerSnake(request.GetSteerSnake(), addr)
	case *protocol.APIRequest_GetGameState:
		server.handleGetGameState(request.GetGetGameState(), addr)
	case *protocol.APIRequest_ExitGame:
		server.handleExitGame(request.GetExitGame(), addr)
	case *protocol.APIRequest_Disconnect:
		server.handleDisconnect(request.GetDisconnect(), addr)
	default:
		server.sendError(unrecognizedRequestError, addr)
	}
}

func (server *Server) handleConnect(addr *net.UDPAddr) {
	if !server.IsFree() {
		server.sendError(nodeIsBusyError, addr)
		return
	}
	server.lastRequestTime = time.Now()

	token, err := uuid.NewUUID()
	if err != nil {
		server.sendError(apiTokenCreationError, addr)
		return
	}
	server.token = token.String()
	server.isFree = false

	server.sendSuccessConnect(addr)
}

func (server *Server) handlePing(request *protocol.APIRequest_PingMsg, addr *net.UDPAddr) {
	if request.GetToken() != server.token {
		server.sendError(notValidTokenError, addr)
		return
	}
	server.lastRequestTime = time.Now()
}

func (server *Server) handleCreateGame(request *protocol.APIRequest_CreateGameMsg, addr *net.UDPAddr) {
	if request.GetToken() != server.token {
		server.sendError(notValidTokenError, addr)
		return
	}
	server.lastRequestTime = time.Now()

	err := server.node.CreateGame(
		request.GetGameName(),
		request.GetWidth(),
		request.GetHeight(),
		request.GetFoodStatic(),
		request.GetStateDelayMs(),
		request.GetPlayerName(),
	)
	if err == nil {
		server.sendAck(addr)
	} else {
		server.sendError(err.Error(), addr)
	}
}

func (server *Server) handleDiscoverGame(request *protocol.APIRequest_DiscoverGamesMsg, addr *net.UDPAddr) {
	if request.GetToken() != server.token {
		server.sendError(notValidTokenError, addr)
		return
	}
	server.lastRequestTime = time.Now()

	gameInfoDtos := server.node.DiscoverGames()
	server.sendGameList(gameInfoDtos, addr)
}

func (server *Server) handleJoinGame(request *protocol.APIRequest_JoinGameMsg, addr *net.UDPAddr) {
	if request.GetToken() != server.token {
		server.sendError(notValidTokenError, addr)
		return
	}
	server.lastRequestTime = time.Now()

	err := server.node.JoinGame(request.GetGameName(), request.GetPlayerName(), request.GetIsPlayer())
	if err == nil {
		server.sendAck(addr)
	} else {
		server.sendError(err.Error(), addr)
	}
}

func (server *Server) handleSteerSnake(request *protocol.APIRequest_SteerSnakeMsg, addr *net.UDPAddr) {
	if request.GetToken() != server.token {
		server.sendError(notValidTokenError, addr)
		return
	}
	server.lastRequestTime = time.Now()

	err := server.node.AddMove(protocol.MapToP2PDirection(*request.Direction))
	if err == nil {
		server.sendAck(addr)
	} else {
		server.sendError(err.Error(), addr)
	}
}

func (server *Server) handleGetGameState(request *protocol.APIRequest_GetGameStateMsg, addr *net.UDPAddr) {
	if request.GetToken() != server.token {
		server.sendError(notValidTokenError, addr)
		return
	}
	server.lastRequestTime = time.Now()

	stateDto, err := server.node.GetState()
	if err != nil {
		server.sendError(err.Error(), addr)
		return
	}

	server.sendGameState(stateDto, addr)
}

func (server *Server) handleExitGame(request *protocol.APIRequest_ExitGameMsg, addr *net.UDPAddr) {
	if request.GetToken() != server.token {
		server.sendError(notValidTokenError, addr)
		return
	}
	server.lastRequestTime = time.Now()

	err := server.node.ExitGame()
	if err == nil {
		server.sendAck(addr)
	} else {
		server.sendError(err.Error(), addr)
	}
}

func (server *Server) handleDisconnect(request *protocol.APIRequest_DisconnectMsg, addr *net.UDPAddr) {
	if request.GetToken() != server.token {
		server.sendError(notValidTokenError, addr)
		return
	}

	server.token = ""
	server.isFree = true
	err := server.node.ExitGame()
	if err == nil {
		server.sendAck(addr)
	} else {
		server.sendError(err.Error(), addr)
	}
}
