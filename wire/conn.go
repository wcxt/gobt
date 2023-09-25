package wire

import (
	"errors"
	"net"
	"time"
)

const DefaultTimeout = 5 * time.Second

type Conn struct {
    conn net.Conn

	ClientChoking    bool
	PeerChoking      bool
	ClientInterested bool
	PeerInterested   bool
}

func Dial(addr string) (*Conn, error) {
    conn, err := net.DialTimeout("tcp", addr, DefaultTimeout)
    if err != nil {
        return nil, err
    }

    return &Conn{
        conn: conn,
        ClientChoking: true,
        ClientInterested: false,
        PeerChoking: true,
        PeerInterested: false,
    }, nil
}

func (c *Conn) Handshake(infoHash, peerId [20]byte) error {
	hs := NewHandshake(infoHash, peerId)
	c.conn.Write(MarshalHandshake(hs))

	hs, err := UnmarshalHandshake(c.conn)
	if err != nil {
		return err
	}

	if hs.InfoHash != infoHash {
		return errors.New("invalid field: InfoHash")
	}

	return nil
}

func (c *Conn) Send(msg *Message) (int, error) {
     return c.conn.Write(MarshalMessage(msg))
}

func (c *Conn) Recv() (*Message, error) {
     return UnmarshalMessage(c.conn)
}

func (c *Conn) RecvBitfield() (Bitfield, error) {
    msg, err := c.Recv()
    if err != nil {
        return nil, err
    }
    if msg.ID != MessageBitfield {
        return nil, errors.New("expected bitfield message")
    }
    return Bitfield(msg.Payload), nil
}

func (c *Conn) Close() error {
    return c.conn.Close()
}
