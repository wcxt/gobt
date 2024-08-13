package gobt

import (
	"errors"
	"net"
	"time"

	"github.com/edwces/gobt/protocol"
)

const (
	MaxRequestCountPerPeer = 5
	MaxHashFails           = 15
)

type Peer struct {
	conn net.Conn

	IsInteresting bool
	IsChoking     bool

	Requests  [][]int
	Cancelled [][]int
	HashFails int

	keepAlivePeriod time.Duration
	keepAliveTicker *time.Ticker
}

func NewPeer(conn net.Conn) *Peer {
	return &Peer{conn: conn, IsInteresting: false, IsChoking: true, Requests: [][]int{}, Cancelled: [][]int{}, HashFails: 0}
}

func (p *Peer) SetReadDeadline(period time.Duration) {
	p.conn.SetReadDeadline(time.Now().Add(period))
}

func (p *Peer) KeepAlive(period time.Duration) {
	p.keepAlivePeriod = period
	p.keepAliveTicker = time.NewTicker(p.keepAlivePeriod)

	go func() {
		for range p.keepAliveTicker.C {
			_, err := p.WriteKeepAlive()
			if err != nil {
				p.Close()
			}
		}
	}()
}

func (p *Peer) SendInterested() error {
	_, err := p.WriteMsg(protocol.IDInterested, nil)
	if err != nil {
		return nil
	}

	p.IsInteresting = true
	return nil
}

func (p *Peer) SendNotInterested() error {
	_, err := p.WriteMsg(protocol.IDNotInterested, nil)
	if err != nil {
		return nil
	}

	p.IsInteresting = false
	return nil
}

func (p *Peer) RecvRequest(index, offset, length int) error {
	req := p.Requests[0]

	if req[0] != index {
		return errors.New("invalid index")
	}
	if req[1]*MaxBlockLength != offset {
		return errors.New("invalid offset")
	}
	if req[2] != length {
		return errors.New("invalid length")
	}

	// TEMP: Fix for cancelled requests, BUT we still process them in our client
	if len(p.Cancelled) != 0 {
		if p.Cancelled[0][0] != req[0] && p.Cancelled[0][1] != req[1] && p.Cancelled[0][2] != req[2] {
			p.Requests = p.Requests[1:]
		}
		p.Cancelled = p.Cancelled[1:]
	}

	p.Requests = p.Requests[1:]
	return nil
}

func (p *Peer) SendRequest(index, offset, length int) error {
	req := protocol.Request{Index: uint32(index), Offset: uint32(offset), Length: uint32(length)}

	p.Requests = append(p.Requests, []int{index, offset / MaxBlockLength, length})
	_, err := p.WriteMsg(protocol.IDRequest, req.Marshal())
	if err != nil {
		return err
	}

	return nil
}

func (p *Peer) SendCancel(index, offset, length int) error {
	req := protocol.Request{Index: uint32(index), Offset: uint32(offset), Length: uint32(length)}

	_, err := p.WriteMsg(protocol.IDCancel, req.Marshal())
	if err != nil {
		return err
	}

	for _, req := range p.Requests {
		if req[0] != index {
			continue
		}
		if req[1]*MaxBlockLength != offset {
			continue
		}
		if req[2] != length {
			continue
		}

		p.Cancelled = append(p.Cancelled, req)
		break
	}

	return nil
}

func (p *Peer) IsRequestable() bool {
	return len(p.Requests) < MaxRequestCountPerPeer && p.IsInteresting
}

func (p *Peer) ReadMsg() (*protocol.Message, error) {
	msg, err := protocol.UnmarshalMessage(p.conn)
	if err != nil {
		return nil, err
	}

	p.conn.SetReadDeadline(time.Time{})

	// fmt.Printf("%s READ: %s\n", p.conn.RemoteAddr().String(), msg.String())

	return msg, nil
}

func (p *Peer) WriteMsg(id protocol.MessageID, payload []byte) (int, error) {
	nmsg := &protocol.Message{ID: id, Payload: payload}
	wb, err := p.conn.Write(nmsg.Marshal())
	if err != nil {
		return wb, err
	}
	p.keepAliveTicker.Reset(p.keepAlivePeriod)

	// fmt.Printf("%s WRITE: %s\n", p.conn.RemoteAddr().String(), nmsg.String())

	return wb, nil
}

func (p *Peer) WriteKeepAlive() (int, error) {
	nmsg := &protocol.Message{KeepAlive: true}
	return p.conn.Write(nmsg.Marshal())
}

func (p *Peer) WriteUnchoke() (int, error) {
	return p.WriteMsg(protocol.IDUnchoke, nil)
}

func (p *Peer) WriteHave(index int) (int, error) {
	payload := protocol.Have(index).Marshal()
	return p.WriteMsg(protocol.IDHave, payload)
}

func (p *Peer) String() string {
	return p.conn.RemoteAddr().String()
}

func (p *Peer) Close() error {
	if p.keepAliveTicker != nil {
		p.keepAliveTicker.Stop()
	}
	return p.conn.Close()
}
