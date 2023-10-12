package wire

import (
	"errors"
	"net"

	"github.com/edwces/gobt/wire/handshake"
	"github.com/edwces/gobt/wire/message"
)

const MaxPendingRequests = 5

func Dial(addr string, infoHash [20]byte, peerId [20]byte) (*Peer, error) {
	conn, err := net.DialTimeout("tcp", addr, ConnDialTimeout)
	if err != nil {
		return nil, err
	}

	peer := NewPeer(conn)
	err = peer.handshake(infoHash, peerId)
	if err != nil {
		return nil, err
	}

	return peer, nil
}

type Peer struct {
	bitfield message.Bitfield
	conn     net.Conn

	Choking     bool
	Interested  bool
	Choked      bool
	Interesting bool

    pending chan int
}

func NewPeer(conn net.Conn) *Peer {
	return &Peer{
		bitfield: message.Bitfield{},
		conn:     conn,

		Choking:     true,
		Interested:  false,
		Choked:      true,
		Interesting: false,
        pending: make(chan int, MaxPendingRequests),
	}
}

func (p *Peer) Has(index int) bool {
    if len(p.bitfield) <= index {
        return false
    }
	return p.bitfield.Get(index)
}

func (p *Peer) Recv() (*message.Message, error) {
    return message.Read(p.conn)
}

func (p *Peer) Handle(msg *message.Message) {
	if msg.KeepAlive {
		return
	}

	switch msg.ID {
	case message.IDChoke:
		p.Choking = true
	case message.IDUnchoke:
		p.Choking = false
	case message.IDInterested:
		p.Interested = true
	case message.IDNotInterested:
		p.Interested = false
	case message.IDHave:
		p.bitfield.Set(int(msg.Payload.Have()))
	case message.IDBitfield:
		p.bitfield = msg.Payload.Bitfield()
	case message.IDPiece:
        <-p.pending
	}
}

func (p *Peer) Uninterest() error {
	msg := &message.Message{ID: message.IDNotInterested}
	_, err := message.Write(p.conn, msg)
	if err != nil {
		return err
	}

	p.Interesting = false
	return nil

}

func (p *Peer) Interest() error {
	msg := &message.Message{ID: message.IDInterested}
	_, err := message.Write(p.conn, msg)
	if err != nil {
		return err
	}

	p.Interesting = true
	return nil
}

func (p *Peer) Choke() error {
	msg := &message.Message{ID: message.IDChoke}
	_, err := message.Write(p.conn, msg)
	if err != nil {
		return err
	}

	p.Choked = true
	return nil
}

func (p *Peer) Unchoke() error {
	msg := &message.Message{ID: message.IDUnchoke}
	_, err := message.Write(p.conn, msg)
	if err != nil {
		return err
	}

	p.Choked = false
	return nil
}

func (p *Peer) Request(index, offset, length int) error {
    p.pending<-1
    req := message.Request{Index: uint32(index), Offset: uint32(offset), Length: uint32(length)}
    msg := &message.Message{ID: message.IDRequest, Payload: message.NewRequestPayload(req)}
    _, err := message.Write(p.conn, msg)
    if err != nil {
        <-p.pending
        return err
    }
    
    return nil 
}

func (p *Peer) handshake(infoHash, peerId [20]byte) error {
	_, err := handshake.Write(p.conn, handshake.New(infoHash, peerId))
	if err != nil {
		return err
	}

	hs, err := handshake.Read(p.conn)
	if err != nil {
		return err
	}

	if hs.InfoHash != infoHash {
		return errors.New("invalid field: InfoHash")
	}

	return nil
}

func (p *Peer) Close() error {
	return p.conn.Close()
}
