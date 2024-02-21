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

func (pm *PeersManager) Disconnect() {
	pm.peers.Range(func(key, value any) bool {
		value.(*Conn).Close()
		return true
	})
}

func (pm *PeersManager) WriteHave(have int) {
	pm.peers.Range(func(key, value any) bool {
		peer := value.(*Conn)

		_, err := peer.WriteHave(have)
		if err != nil {
			peer.Close()
		}

		return true
	})
}
