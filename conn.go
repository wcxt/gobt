package gobt

import (
	"fmt"
	"net"
	"time"

	"github.com/edwces/gobt/handshake"
	"github.com/edwces/gobt/message"
)

const DefaultConnTimeout = 3 * time.Second

type Conn struct {
    conn net.Conn
}

func DialTimeout(address string) (*Conn, error) {
    conn, err := net.DialTimeout("tcp", address, DefaultConnTimeout)
    if err != nil {
        return nil, err
    }

    return &Conn{conn: conn}, nil
}

func (c *Conn) Handshake(hash [20]byte, clientID [20]byte) error {
    hs := handshake.New(hash, clientID)
    handshake.Write(c.conn, hs)

    hs, err := handshake.Read(c.conn)
    if err != nil {
        return err
    }

    if hs.InfoHash != hash {
        return fmt.Errorf("InfoHash unexpected value: %s", hs.InfoHash) 
    }
    
    return nil

}

func (c *Conn) ReadMsg() (*message.Message, error) {
    msg, err := message.Read(c.conn)
    if err != nil {
        return nil, err
    }

    fmt.Printf("MSG READ: %d\n", msg.ID)
    return msg, nil
}

func (c *Conn) WriteMsg(id message.ID, payload message.Payload) (int, error) {
    nmsg := &message.Message{ID: id, Payload: payload}
    wb, err := message.Write(c.conn, nmsg)
    if err != nil {
        return wb, err
    }

    fmt.Printf("MSG WRITE: %d\n", nmsg.ID)
    return wb, nil
}

func (c *Conn) KeepAlive() (int, error) {
    nmsg := &message.Message{KeepAlive: true}
    return message.Write(c.conn, nmsg)
}

func (c *Conn) WriteInterested() (int, error) {
    return c.WriteMsg(message.IDInterested, nil)
}

func (c *Conn) WriteNotInterested() (int, error) {
    return c.WriteMsg(message.IDNotInterested, nil)
}

func (c *Conn) WriteRequest(index, offset, length int) (int, error) {
    req := message.Request{Index: uint32(index), Offset: uint32(offset), Length: uint32(length)}
    payload := message.NewRequestPayload(req)
    return c.WriteMsg(message.IDRequest, payload)
}

func (c *Conn) Close() error {
    return c.conn.Close()
}
