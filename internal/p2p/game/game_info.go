package game

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"p2p-snake/internal/engine"
	"p2p-snake/internal/p2p/protocol"
)

var (
	nextPlayerId int32 = 1

	notValidWidthError        = fmt.Errorf("width should be from 10 to 100")
	notValidHeightError       = fmt.Errorf("height should be from 10 to 100")
	notValidFoodStaticError   = fmt.Errorf("initial amount of foods should be from 0 to 100")
	notValidStateDelayError   = fmt.Errorf("state delay should be from 100 to 3000")
	gameIsNotInitializedError = fmt.Errorf("game is not initialized")
)

type GameInfo struct {
	// State
	currentNode *NodeInfo
	stateOrder  *atomic.Int32
	stateDelay  time.Duration
	game        *engine.Game
	nodes       map[int32]*NodeInfo

	// Player moves
	moves map[int32]engine.Direction

	// Sync
	lock *sync.RWMutex
}

func NewGameInfo() *GameInfo {
	return &GameInfo{
		currentNode: nil,
		stateOrder:  &atomic.Int32{},
		stateDelay:  -1,
		game:        nil,
		nodes:       make(map[int32]*NodeInfo),

		moves: make(map[int32]engine.Direction),

		lock: &sync.RWMutex{},
	}
}

func (i *GameInfo) CurrentNode() *NodeInfo {
	return i.currentNode
}

func (i *GameInfo) SetCurrentNode(currentNode *NodeInfo) {
	i.currentNode = currentNode
}

func (i *GameInfo) StateOrder() int32 {
	return i.stateOrder.Load()
}

func (i *GameInfo) SetStateOrder(stateOrder int32) {
	i.stateOrder.Store(stateOrder)
}

func (i *GameInfo) StateDelay() time.Duration {
	return i.stateDelay
}

func (i *GameInfo) SetStateDelay(stateDelay time.Duration) {
	i.stateDelay = stateDelay
}

func (i *GameInfo) GameName() string {
	return i.game.Name
}

func (i *GameInfo) SetGameName(gameName string) {
	i.game.Name = gameName
}

func (i *GameInfo) Width() int32 {
	return i.game.Width
}

func (i *GameInfo) SetWidth(width int32) {
	i.game.Width = width
}

func (i *GameInfo) Height() int32 {
	return i.game.Height
}

func (i *GameInfo) SetHeight(height int32) {
	i.game.Height = height
}

func (i *GameInfo) FoodStatic() int32 {
	return i.game.FoodStatic
}

func (i *GameInfo) SetFoodStatic(foodStatic int32) {
	i.game.FoodStatic = foodStatic
}

func (i *GameInfo) Config() *protocol.GameConfig {
	return toConfig(
		i.Width(),
		i.Height(),
		i.FoodStatic(),
		i.StateDelay(),
	)
}

func (i *GameInfo) SetConfig(config *protocol.GameConfig) {
	i.SetWidth(config.GetWidth())
	i.SetHeight(config.GetHeight())
	i.SetFoodStatic(config.GetFoodStatic())
	i.SetStateDelay(time.Duration(config.GetStateDelayMs()) * time.Millisecond)
}

func (i *GameInfo) Snakes() []*protocol.GameState_Snake {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return toSnakes(i.game.Snakes)
}

func (i *GameInfo) SetSnakes(snakes []*protocol.GameState_Snake) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.game.Snakes = toEngineSnakes(snakes)
}

func (i *GameInfo) Foods() []*protocol.GameState_Coord {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return toCoords(i.game.Foods)
}

func (i *GameInfo) SetFoods(foods []*protocol.GameState_Coord) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.game.Foods = toEngineCoords(foods)
}

func (i *GameInfo) Players() *protocol.GamePlayers {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return toPlayers(i.game.Players, i.nodes)
}

func (i *GameInfo) SetPlayers(players *protocol.GamePlayers) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.game.Players = toEnginePlayers(players)
}

func (i *GameInfo) Nodes() map[int32]*NodeInfo {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.nodes
}

func (i *GameInfo) SetNodes(players *protocol.GamePlayers) {
	for playerId, node := range toNodeInfos(players) {
		i.SetNode(playerId, node)
	}
}

func (i *GameInfo) Node(playerId int32) (*NodeInfo, bool) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	node, ok := i.nodes[playerId]
	return node, ok
}

func (i *GameInfo) SetNode(playerId int32, node *NodeInfo) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.nodes[playerId] = node
}

func (i *GameInfo) NormalNodes() []*NodeInfo {
	normals := make([]*NodeInfo, 0)
	for _, node := range i.Nodes() {
		if node.Role() == protocol.NodeRole_NORMAL {
			normals = append(normals, node)
		}
	}
	return normals
}

func (i *GameInfo) MasterNode() *NodeInfo {
	for _, node := range i.Nodes() {
		if node.Role() == protocol.NodeRole_MASTER {
			return node
		}
	}
	return nil
}

func (i *GameInfo) DeputyNode() *NodeInfo {
	for _, node := range i.Nodes() {
		if node.Role() == protocol.NodeRole_DEPUTY {
			return node
		}
	}
	return nil
}

func (i *GameInfo) ExistsDeputyNode() bool {
	for _, node := range i.Nodes() {
		if node.Role() == protocol.NodeRole_DEPUTY {
			return true
		}
	}
	return false
}

func (i *GameInfo) ExistsPlayerByName(name string) bool {
	for _, player := range i.Players().GetPlayers() {
		if player.GetName() == name {
			return true
		}
	}
	return false
}

func (i *GameInfo) ExistsPlayerByAddr(addr *net.UDPAddr) bool {
	for _, player := range i.Players().GetPlayers() {
		if player.GetIpAddress() == addr.IP.String() && player.GetPort() == int32(addr.Port) {
			return true
		}
	}
	return false
}

func (i *GameInfo) CreateNewGame(gameName string, width int32, height int32, foodStatic int32, stateDelay int32) error {
	// Check
	if width < 10 || width > 100 {
		return notValidWidthError
	}
	if height < 10 || height > 100 {
		return notValidHeightError
	}
	if foodStatic < 0 || foodStatic > 100 {
		return notValidFoodStaticError
	}
	if stateDelay < 100 || stateDelay > 3000 {
		return notValidStateDelayError
	}

	i.lock.Lock()
	defer i.lock.Unlock()

	i.game = engine.NewGame(gameName, width, height, foodStatic)
	i.SetStateDelay(time.Duration(stateDelay) * time.Millisecond)
	return nil
}

func (i *GameInfo) AddPlayer(playerName string, role protocol.NodeRole, addr *net.UDPAddr) (*NodeInfo, error) {
	if i.game == nil {
		return nil, gameIsNotInitializedError
	}

	i.lock.Lock()
	defer i.lock.Unlock()

	// Add player
	err := i.game.AddPlayer(nextPlayerId, playerName, role != protocol.NodeRole_VIEWER)
	if err != nil {
		return nil, err
	}

	// Create node
	node := NewNodeInfo(nextPlayerId, role, addr)
	i.nodes[nextPlayerId] = node
	nextPlayerId = nextPlayerId + 1

	return node, nil
}

func (i *GameInfo) DeletePlayer(playerId int32) error {
	if i.game == nil {
		return gameIsNotInitializedError
	}

	i.lock.Lock()
	defer i.lock.Unlock()

	delete(i.nodes, playerId)
	delete(i.game.Players, playerId)
	return nil
}

func (i *GameInfo) AddMove(playerId int32, direction protocol.Direction) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.moves[playerId] = toEngineDirection(direction)
	return nil
}

func (i *GameInfo) GenerateNextState() ([]int32, error) {
	if i.game == nil {
		return []int32{}, gameIsNotInitializedError
	}

	// Generate next state
	deadSnakes := i.game.NextState(i.moves)
	i.stateOrder.Add(1)
	i.moves = make(map[int32]engine.Direction)

	return deadSnakes, nil
}

func (i *GameInfo) SetState(currentPlayerId int32, state *protocol.GameState, addr *net.UDPAddr) {
	i.SetStateOrder(state.GetStateOrder())
	i.SetNodes(state.GetPlayers())
	i.SetPlayers(state.GetPlayers())
	i.SetSnakes(state.GetSnakes())
	i.SetFoods(state.GetFoods())

	for _, player := range state.GetPlayers().GetPlayers() {
		if player.GetId() >= nextPlayerId {
			nextPlayerId = player.GetId() + 1
		}
	}

	if master := i.MasterNode(); master == nil {
		if deputy := i.DeputyNode(); deputy != nil {
			deputy.SetRole(protocol.NodeRole_MASTER)
		}
	}

	currentNode, _ := i.Node(currentPlayerId)
	i.SetCurrentNode(currentNode)

	i.MasterNode().SetAddr(addr)
}
