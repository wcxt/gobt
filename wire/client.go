package wire

import (
	"errors"
	"net"
	"time"
)

const (
    DefaultTimeout = time.Second * 5
)

type Client struct {
    PeerId [20]byte
}

func (c *Client) NewConnection(addr string) (net.Conn, error) {
    conn, err := net.DialTimeout("tcp", addr, DefaultTimeout)
    if err != nil {
        return nil, err
    }

    return conn, nil
}

func (c *Client) Handshake(conn net.Conn, infoHash [20]byte, peerId [20]byte) error {
    hs := NewHandshake(infoHash, peerId)
    conn.Write(MarshalHandshake(hs))

    hs, err := UnmarshalHandshake(conn)
    if err != nil {
        return err
    }
    
    if hs.InfoHash != infoHash {
        return errors.New("invalid field: InfoHash")
    }

    return nil
}
