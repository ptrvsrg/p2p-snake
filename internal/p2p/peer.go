package p2p

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"p2p-snake/internal/log"
	"p2p-snake/internal/p2p/announcements"
	"p2p-snake/internal/p2p/dto"
	"p2p-snake/internal/p2p/game"
	"p2p-snake/internal/p2p/protocol"
)

var (
	playerAlreadyInGameError   = fmt.Errorf("player is already in game")
	gameAlreadyExistsError     = fmt.Errorf("game with same name already exists")
	gameNotFoundError          = fmt.Errorf("game with same name not found")
	masterIsNotRespondingError = fmt.Errorf("master node is not responding")
	unexpectedResponseError    = fmt.Errorf("unexpected response")
	notParticipateInGameError  = fmt.Errorf("node does not participate in game")
)

type Peer struct {
	// Network
	multicastAddr *net.UDPAddr
	multicast     *net.UDPConn
	unicast       *net.UDPConn
	notAckMsg     map[int64]chan *protocol.GameMessage

	// Game
	msgSeq     *atomic.Int64
	gameInfo   *game.GameInfo
	cancelGame context.CancelFunc

	// Announcements
	announcementCollector *announcements.AnnouncementCollector

	// Closing
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func NewPeer(multicastAddr *net.UDPAddr) *Peer {
	return &Peer{
		multicastAddr: multicastAddr,
		notAckMsg:     make(map[int64]chan *protocol.GameMessage),

		msgSeq:     &atomic.Int64{},
		cancelGame: func() {},

		announcementCollector: announcements.NewAnnouncementCollector(),

		cancel: func() {},
		wg:     &sync.WaitGroup{},
	}
}

//////////// START NODE ////////////

func (p *Peer) Start() error {
	// Create multicast socket
	var err error
	p.multicast, err = net.ListenMulticastUDP("udp", nil, p.multicastAddr)
	if err != nil {
		return err
	}
	log.Logger.Infof("P2P node is listening on multicast %v", p.multicastAddr.String())

	// Create unicast socket
	p.unicast, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		p.multicast.Close()
		return err
	}
	log.Logger.Infof("P2P node is listening on unicast %v", p.unicast.LocalAddr().String())

	// Collect announcements
	p.announcementCollector.Start()

	// Start listening on sockets
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())
	p.wg.Add(2)
	go p.listenMulticast(ctx)
	go p.listenUnicast(ctx)

	return nil
}

//////////// CLOSE NODE ////////////

func (p *Peer) Close() error {
	p.cancelGame()
	p.announcementCollector.Close()
	p.cancel()
	p.wg.Wait()

	err1 := p.multicast.Close()
	err2 := p.unicast.Close()
	if err1 != nil || err2 != nil {
		if err1 == nil {
			return err2
		}
		if err2 == nil {
			return err1
		}
		return fmt.Errorf("%v, %v", err1, err2)
	}
	return nil
}

//////////// CREATE GAME ////////////

func (p *Peer) CreateGame(gameName string, width int32, height int32, foodStatic int32, stateDelay int32, playerName string) error {
	if p.gameInfo != nil {
		return playerAlreadyInGameError
	}

	// Check if game with same name exists
	if p.announcementCollector.ExistsAnnouncementByGameName(gameName) {
		return gameAlreadyExistsError
	}

	// Init game
	var err error
	p.gameInfo = game.NewGameInfo()
	if err = p.gameInfo.CreateNewGame(gameName, width, height, foodStatic, stateDelay); err != nil {
		return err
	}

	// Add MASTER
	player, err := p.gameInfo.AddPlayer(playerName, protocol.NodeRole_MASTER, nil)
	if err != nil {
		return err
	}
	p.gameInfo.SetCurrentNode(player)

	// Start sending messages with announcement and game state, receiving messages from players and
	// deleting “dead” nodes
	var ctx context.Context
	ctx, p.cancelGame = context.WithCancel(context.Background())
	p.wg.Add(4)
	go p.announceGame(ctx)
	go p.publishState(ctx)
	go p.pingNode(ctx)
	go p.deleteExpiredNode(ctx)

	log.Logger.Infof("create new game \"%s\" (%dx%d, %dms)", gameName, width, height, stateDelay)
	return nil
}

func (p *Peer) announceGame(ctx context.Context) {
	defer p.wg.Done()

	log.Logger.Debug("announceGame goroutine is running")
	for {
		if p.gameInfo == nil {
			log.Logger.Debug("announceGame goroutine has completed")
			return
		}
		select {
		case <-ctx.Done():
			log.Logger.Debug("announceGame goroutine has completed")
			return
		case <-time.After(time.Second):
			p.sendAnnouncementMsg(
				p.gameInfo,
				p.multicastAddr,
			)
		}
	}
}

func (p *Peer) publishState(ctx context.Context) {
	defer p.wg.Done()

	log.Logger.Debug("publishState goroutine is running")
	for {
		if p.gameInfo == nil {
			log.Logger.Debug("publishState goroutine has completed")
			return
		}
		select {
		case <-ctx.Done():
			log.Logger.Debug("publishState goroutine has completed")
			return
		case <-time.After(p.gameInfo.StateDelay()):
			deadSnakes, err := p.gameInfo.GenerateNextState()
			if err != nil {
				log.Logger.Errorf("P2P node error: %v", err)
				continue
			}

			for playerId, node := range p.gameInfo.Nodes() {
				if node != p.gameInfo.CurrentNode() {
					p.sendStateMsg(
						p.gameInfo.CurrentNode().PlayerId(),
						playerId,
						p.gameInfo,
						node.Addr(),
					)
				}
			}

			for _, playerId := range deadSnakes {
				if node, ok := p.gameInfo.Node(playerId); ok && node.Addr() != nil {
					p.sendRoleChangeMsg(
						p.gameInfo.CurrentNode().PlayerId(),
						playerId,
						nil,
						protocol.NodeRole_VIEWER.Enum(),
						node.Addr(),
					)
					node.SetRole(protocol.NodeRole_VIEWER)
				}
				if p.gameInfo.CurrentNode().PlayerId() == playerId {
					_ = p.ExitGame()
				}
			}

			if !p.gameInfo.ExistsDeputyNode() {
				// Appoint new DEPUTY
				for _, normal := range p.gameInfo.NormalNodes() {
					if p.appointDeputy(normal) {
						break
					}
				}
			}
		}
	}
}

func (p *Peer) pingNode(ctx context.Context) {
	defer p.wg.Done()

	log.Logger.Debug("pingNode goroutine is running")
	for {
		if p.gameInfo == nil {
			log.Logger.Debug("pingNode goroutine has completed")
			return
		}
		select {
		case <-ctx.Done():
			log.Logger.Debug("pingNode goroutine has completed")
			return
		case <-time.After(p.gameInfo.StateDelay() / 5):
			for _, node := range p.gameInfo.Nodes() {
				if !node.IsMasterNode() {
					p.sendPingMsg(p.gameInfo.CurrentNode().PlayerId(), node.PlayerId(), node.Addr())
				}
			}
		}
	}
}

func (p *Peer) deleteExpiredNode(ctx context.Context) {
	defer p.wg.Done()

	log.Logger.Debug("deleteExpiredNode goroutine is running")
	for {
		if p.gameInfo == nil {
			log.Logger.Debug("deleteExpiredNode goroutine has completed")
			return
		}
		select {
		case <-ctx.Done():
			log.Logger.Debug("deleteExpiredNode goroutine has completed")
			return
		case <-time.After(p.gameInfo.StateDelay() / 2):
			for _, node := range p.gameInfo.Nodes() {
				if time.Since(node.LastUpdateTime()) > p.gameInfo.StateDelay()*8/10 && !node.IsMasterNode() {
					_ = p.gameInfo.DeletePlayer(node.PlayerId())
				}
			}
		}
	}
}

//////////// DISCOVER GAMES ////////////

func (p *Peer) DiscoverGames() []dto.GameInfoDto {
	p.sendDiscoverMsg(p.multicastAddr)
	return p.announcementCollector.GetGameInfoDtos()
}

//////////// JOIN GAME ////////////

func (p *Peer) JoinGame(gameName string, playerName string, isPlayer bool) error {
	if p.gameInfo != nil {
		return playerAlreadyInGameError
	}

	announcement, ok := p.announcementCollector.FindByGameName(gameName)
	if !ok {
		return gameNotFoundError
	}

	var role protocol.NodeRole
	if isPlayer {
		role = protocol.NodeRole_NORMAL
	} else {
		role = protocol.NodeRole_VIEWER
	}

	_, res := p.sendJoinMsg(
		gameName,
		playerName,
		role,
		announcement.Addr(),
	)
	if res == nil {
		return masterIsNotRespondingError
	}
	if _, ok := res.GetType().(*protocol.GameMessage_Ack); ok {
		p.gameInfo = game.NewGameInfo()
		p.gameInfo.SetCurrentNode(game.NewNodeInfo(res.GetReceiverId(), role, nil))
		_ = p.gameInfo.CreateNewGame(
			announcement.GameName(),
			announcement.Width(),
			announcement.Height(),
			announcement.FoodStatic(),
			announcement.StateDelay(),
		)

		var ctx context.Context
		ctx, p.cancelGame = context.WithCancel(context.Background())
		p.wg.Add(2)
		go p.pingMaster(ctx)
		go p.deleteExpiredMaster(ctx)

		log.Logger.Infof("join to game %s (%dx%d, %d)", announcement.GameName(),
			announcement.Width(), announcement.Height(), announcement.StateDelay())
		return nil
	}
	if _, ok := res.GetType().(*protocol.GameMessage_Error); ok {
		return fmt.Errorf(res.GetError().GetErrorMessage())
	}
	return unexpectedResponseError
}

func (p *Peer) pingMaster(ctx context.Context) {
	defer p.wg.Done()

	log.Logger.Debug("pingMaster goroutine is running")
	for {
		if p.gameInfo == nil {
			log.Logger.Debug("pingMaster goroutine has completed")
			return
		}
		select {
		case <-ctx.Done():
			log.Logger.Debug("pingMaster goroutine has completed")
			return
		case <-time.After(p.gameInfo.StateDelay() / 5):
			if master := p.gameInfo.MasterNode(); master != nil {
				p.sendPingMsg(p.gameInfo.CurrentNode().PlayerId(), master.PlayerId(), master.Addr())
			}
		}
	}
}

func (p *Peer) deleteExpiredMaster(ctx context.Context) {
	defer p.wg.Done()

	log.Logger.Debug("deleteExpiredMaster goroutine is running")
	for {
		if p.gameInfo == nil {
			log.Logger.Debug("deleteExpiredMaster goroutine has completed")
			return
		}
		select {
		case <-ctx.Done():
			log.Logger.Debug("deleteExpiredMaster goroutine has completed")
			return
		case <-time.After(p.gameInfo.StateDelay() / 2):
			master := p.gameInfo.MasterNode()
			if master != nil && time.Since(master.LastUpdateTime()) > p.gameInfo.StateDelay()*8/10 {
				// Delete expired master
				_ = p.gameInfo.DeletePlayer(master.PlayerId())

				deputy := p.gameInfo.DeputyNode()
				if deputy == nil {
					_ = p.ExitGame()
					continue
				}

				deputy.SetRole(protocol.NodeRole_MASTER)
				if !p.gameInfo.CurrentNode().IsMasterNode() {
					deputy.UpdateTime(time.Now().Add(2 * p.gameInfo.StateDelay()))
					continue
				}

				p.cancelGame()

				// Broadcast message about new MASTER
				for playerId, node := range p.gameInfo.Nodes() {
					if !node.IsMasterNode() {
						go p.sendRoleChangeMsg(
							p.gameInfo.CurrentNode().PlayerId(),
							playerId,
							protocol.NodeRole_MASTER.Enum(),
							nil,
							node.Addr(),
						)
						node.UpdateTime(time.Now().Add(2 * p.gameInfo.StateDelay()))
					}
				}

				// Start sending messages with announcement and game state, receiving messages from
				// players and deleting “dead” nodes
				var ctx context.Context
				ctx, p.cancelGame = context.WithCancel(context.Background())
				p.wg.Add(4)
				go p.announceGame(ctx)
				go p.publishState(ctx)
				go p.pingNode(ctx)
				go p.deleteExpiredNode(ctx)
			}
		}
	}
}

//////////// GET GAME STATE ////////////

func (p *Peer) GetState() (dto.GameStateDto, error) {
	if p.gameInfo == nil {
		return dto.GameStateDto{}, notParticipateInGameError
	}
	return dto.ToGameStateDto(
		p.gameInfo.StateOrder(),
		p.gameInfo.Config(),
		p.gameInfo.Snakes(),
		p.gameInfo.Foods(),
		p.gameInfo.Players(),
	), nil
}

//////////// ADD MOVE ////////////

func (p *Peer) AddMove(direction protocol.Direction) error {
	if p.gameInfo == nil {
		return notParticipateInGameError
	}
	if p.gameInfo.CurrentNode().IsMasterNode() {
		return p.gameInfo.AddMove(p.gameInfo.CurrentNode().PlayerId(), direction)
	}

	if master := p.gameInfo.MasterNode(); master != nil {
		_, res := p.sendSteerMsg(
			p.gameInfo.CurrentNode().PlayerId(),
			master.PlayerId(),
			direction,
			master.Addr(),
		)
		if res == nil {
			return masterIsNotRespondingError
		} else if _, ok := res.GetType().(*protocol.GameMessage_Ack); ok {
			return nil
		} else if _, ok := res.GetType().(*protocol.GameMessage_Error); ok {
			return fmt.Errorf(res.GetError().GetErrorMessage())
		} else {
			return unexpectedResponseError
		}
	}
	return nil
}

//////////// EXIT GAME ////////////

func (p *Peer) ExitGame() error {
	if p.gameInfo == nil {
		return notParticipateInGameError
	}
	p.cancelGame()
	p.gameInfo = nil
	return nil
}

func (p *Peer) appointDeputy(player *game.NodeInfo) bool {
	if !p.gameInfo.CurrentNode().IsMasterNode() {
		return false
	}

	_, res := p.sendRoleChangeMsg(
		p.gameInfo.CurrentNode().PlayerId(),
		player.PlayerId(),
		nil,
		protocol.NodeRole_DEPUTY.Enum(),
		player.Addr(),
	)
	if res != nil {
		if _, ok := res.GetType().(*protocol.GameMessage_Ack); ok {
			player.SetRole(protocol.NodeRole_DEPUTY)
			return true
		}
	}
	return false
}
