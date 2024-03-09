package gobt

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/edwces/gobt/handshake"
	"github.com/edwces/gobt/message"
)

const (
	MaxRequestCountPerPeer = 5
	MaxHashFails           = 15
)

// NOTES: This module will be probably be one of the higher level and will
// be responsible for handling most of the logic for messages. Thus he will need
// to have access torrent size, piece size etc.; piece state; and overral most of the modules
// REFACTOR: Messy for now, Leaking most of the information about peer
type Peer struct {
	conn net.Conn

	IsInteresting bool
	IsChoking     bool

	Requests  [][]int
	HashFails int

	writeKeepAlivePeriod time.Duration
	writeKeepAliveTicker *time.Ticker
}

func NewPeer(conn net.Conn) *Peer {
	return &Peer{conn: conn, IsInteresting: false, IsChoking: true, Requests: [][]int{}, HashFails: 0}
}

func (p *Peer) Handshake(hash, clientID [20]byte) error {
	hs := handshake.New(hash, clientID)
	handshake.Write(p.conn, hs)

	hs, err := handshake.Read(p.conn)
	if err != nil {
		return err
	}

	if hs.InfoHash != hash {
		return fmt.Errorf("InfoHash unexpected value: %s", hs.InfoHash)
	}

	return nil
}

func (p *Peer) SetReadKeepAlive(period time.Duration) {
	p.conn.SetReadDeadline(time.Now().Add(period))
}

func (p *Peer) SetWriteKeepAlive(period time.Duration) {
	p.writeKeepAlivePeriod = period
	p.writeKeepAliveTicker = time.NewTicker(p.writeKeepAlivePeriod)

	go func() {
		for range p.writeKeepAliveTicker.C {
			_, err := p.WriteKeepAlive()
			if err != nil {
				p.Close()
			}
		}
	}()
}

func (p *Peer) SendInterested() error {
	_, err := p.WriteMsg(message.IDInterested, nil)
	if err != nil {
		return nil
	}

	p.IsInteresting = true
	return nil
}

func (p *Peer) SendNotInterested() error {
	_, err := p.WriteMsg(message.IDNotInterested, nil)
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

	p.Requests = p.Requests[1:]
	return nil
}

func (p *Peer) SendRequest(index, offset, length int) error {
	req := message.Request{Index: uint32(index), Offset: uint32(offset), Length: uint32(length)}
	fmt.Printf("%s WRITE REQUEST: %d %d %d\n", p.conn.RemoteAddr().String(), index, offset, length)
	payload := message.NewRequestPayload(req)

	p.Requests = append(p.Requests, []int{index, offset / MaxBlockLength, length})
	_, err := p.WriteMsg(message.IDRequest, payload)
	if err != nil {
		return err
	}

	return nil
}

func (p *Peer) SendCancel(index, offset, length int) error {
	req := message.Request{Index: uint32(index), Offset: uint32(offset), Length: uint32(length)}
	fmt.Printf("%s WRITE CANCEL: %d %d %d\n", p.conn.RemoteAddr().String(), index, offset, length)
	payload := message.NewRequestPayload(req)

	_, err := p.WriteMsg(message.IDCancel, payload)
	if err != nil {
		return err
	}

	for i, req := range p.Requests {
		if req[0] != index {
			continue
		}
		if req[1]*MaxBlockLength != offset {
			continue
		}
		if req[2] != length {
			continue
		}

		p.Requests = append(p.Requests[:i], p.Requests[i+1:]...)
		break
	}

	return nil
}

func (p *Peer) IsRequestable() bool {
	return len(p.Requests) < MaxRequestCountPerPeer && p.IsInteresting
}

func (p *Peer) ReadMsg() (*message.Message, error) {
	msg, err := message.Read(p.conn)
	if err != nil {
		return nil, err
	}

	p.conn.SetReadDeadline(time.Time{})

	fmt.Printf("%s READ: %s\n", p.conn.RemoteAddr().String(), msg.String())

	return msg, nil
}

func (p *Peer) WriteMsg(id message.ID, payload message.Payload) (int, error) {
	nmsg := &message.Message{ID: id, Payload: payload}
	wb, err := message.Write(p.conn, nmsg)
	if err != nil {
		return wb, err
	}
	p.writeKeepAliveTicker.Reset(p.writeKeepAlivePeriod)

	if nmsg.ID != message.IDRequest {
		fmt.Printf("%s WRITE: %s\n", p.conn.RemoteAddr().String(), nmsg.String())
	}

	return wb, nil
}

func (p *Peer) WriteKeepAlive() (int, error) {
	nmsg := &message.Message{KeepAlive: true}
	return message.Write(p.conn, nmsg)
}

func (p *Peer) WriteUnchoke() (int, error) {
	return p.WriteMsg(message.IDUnchoke, nil)
}

func (p *Peer) WriteHave(index int) (int, error) {
	payload := message.NewHavePayload(uint32(index))
	return p.WriteMsg(message.IDHave, payload)
}

func (p *Peer) String() string {
	return p.conn.RemoteAddr().String()
}

func (p *Peer) Close() error {
	if p.writeKeepAliveTicker != nil {
		p.writeKeepAliveTicker.Stop()
	}
	return p.conn.Close()
}
