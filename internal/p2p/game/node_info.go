package game

import (
	"net"
	"sync"
	"time"

	"p2p-snake/internal/p2p/protocol"
)

type NodeInfo struct {
	playerId int32
	role     protocol.NodeRole
	addr     *net.UDPAddr

	lastUpdate time.Time

	lock *sync.RWMutex
}

func NewNodeInfo(playerId int32, role protocol.NodeRole, addr *net.UDPAddr) *NodeInfo {
	return &NodeInfo{
		playerId: playerId,
		role:     role,
		addr:     addr,

		lastUpdate: time.Now(),

		lock: &sync.RWMutex{},
	}
}

func (n *NodeInfo) PlayerId() int32 {
	return n.playerId
}

func (n *NodeInfo) Role() protocol.NodeRole {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.role
}

func (n *NodeInfo) SetRole(role protocol.NodeRole) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.role = role
}

func (n *NodeInfo) Addr() *net.UDPAddr {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.addr
}

func (n *NodeInfo) SetAddr(addr *net.UDPAddr) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.addr = addr
}

func (n *NodeInfo) LastUpdateTime() time.Time {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.lastUpdate
}

func (n *NodeInfo) UpdateTime(time time.Time) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.lastUpdate = time
}

func (n *NodeInfo) UpdateTimeAsNow() {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.lastUpdate = time.Now()
}

func (n *NodeInfo) IsMasterNode() bool {
	return n.Role() == protocol.NodeRole_MASTER
}

func (n *NodeInfo) IsDeputyNode() bool {
	return n.Role() == protocol.NodeRole_DEPUTY
}

func (n *NodeInfo) IsNormalNode() bool {
	return n.Role() == protocol.NodeRole_NORMAL
}

func (n *NodeInfo) IsViewerNode() bool {
	return n.Role() == protocol.NodeRole_VIEWER
}
