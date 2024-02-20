package gobt

import "sync"

type PeersManager struct {
	peers sync.Map
}

func NewPeersManager() *PeersManager {
	return &PeersManager{peers: sync.Map{}}
}

func (pm *PeersManager) Add(peer *Conn) {
	pm.peers.Store(peer.String(), peer)
}

func (pm *PeersManager) Remove(peer *Conn) {
	pm.peers.Delete(peer.String())
}

func (pm *PeersManager) GetPeers() *sync.Map {
	return &pm.peers
}
