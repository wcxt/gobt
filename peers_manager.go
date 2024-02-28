package gobt

import "sync"

type PeersManager struct {
	peers sync.Map
}

func NewPeersManager() *PeersManager {
	return &PeersManager{peers: sync.Map{}}
}

func (pm *PeersManager) Add(peer *Peer) {
	pm.peers.Store(peer.String(), peer)
}

func (pm *PeersManager) Remove(peer *Peer) {
	pm.peers.Delete(peer.String())
}

func (pm *PeersManager) Disconnect() {
	pm.peers.Range(func(key, value any) bool {
		value.(*Peer).Close()
		return true
	})
}

func (pm *PeersManager) WriteHave(have int, peerID string) {
	pm.peers.Range(func(key, value any) bool {
		peer := value.(*Peer)

		if peer.String() == peerID {
			return true
		}

		_, err := peer.WriteHave(have)
		if err != nil {
			peer.Close()
		}

		return true
	})
}

func (pm *PeersManager) WriteCancel(index int, offset int, length int, peerID string) {
	pm.peers.Range(func(key, value any) bool {
		peer := value.(*Peer)

		if peer.String() == peerID {
			return true
		}

		err := peer.SendCancel(index, offset, length)
		if err != nil {
			peer.Close()
			return true
		}

		return true
	})
}
