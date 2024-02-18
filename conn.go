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

	writeKeepAlivePeriod time.Duration
	writeKeepAliveTicker *time.Ticker
}

func DialTimeout(address string) (*Conn, error) {
	conn, err := net.DialTimeout("tcp", address, DefaultConnTimeout)
	if err != nil {
		return nil, err
	}

	return &Conn{conn: conn}, nil
}

func (c *Conn) Handshake(hash, clientID [20]byte) error {
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

func (c *Conn) SetReadKeepAlive(period time.Duration) {
	c.conn.SetReadDeadline(time.Now().Add(period))
}

func (c *Conn) SetWriteKeepAlive(period time.Duration) {
	c.writeKeepAlivePeriod = period
	c.writeKeepAliveTicker = time.NewTicker(c.writeKeepAlivePeriod)

	go func() {
		for range c.writeKeepAliveTicker.C {
			_, err := c.WriteKeepAlive()
			if err != nil {
				c.Close()
			}
		}
	}()
}

func (c *Conn) ReadMsg() (*message.Message, error) {
	msg, err := message.Read(c.conn)
	if err != nil {
		return nil, err
	}

	c.conn.SetReadDeadline(time.Time{})

	fmt.Printf("%s READ: %s\n", c.conn.RemoteAddr().String(), msg.String())

	return msg, nil
}

func (c *Conn) WriteMsg(id message.ID, payload message.Payload) (int, error) {
	nmsg := &message.Message{ID: id, Payload: payload}
	wb, err := message.Write(c.conn, nmsg)
	if err != nil {
		return wb, err
	}
	c.writeKeepAliveTicker.Reset(c.writeKeepAlivePeriod)

	if nmsg.ID != message.IDRequest {
		fmt.Printf("%s WRITE: %s\n", c.conn.RemoteAddr().String(), nmsg.String())
	}

	return wb, nil
}

func (c *Conn) WriteKeepAlive() (int, error) {
	nmsg := &message.Message{KeepAlive: true}
	return message.Write(c.conn, nmsg)
}

func (c *Conn) WriteUnchoke() (int, error) {
	return c.WriteMsg(message.IDUnchoke, nil)
}

func (c *Conn) WriteInterested() (int, error) {
	return c.WriteMsg(message.IDInterested, nil)
}

func (c *Conn) WriteNotInterested() (int, error) {
	return c.WriteMsg(message.IDNotInterested, nil)
}

func (c *Conn) WriteRequest(index, offset, length int) (int, error) {
	req := message.Request{Index: uint32(index), Offset: uint32(offset), Length: uint32(length)}
	fmt.Printf("%s WRITE REQUEST: %d %d %d\n", c.conn.RemoteAddr().String(), index, offset, length)
	payload := message.NewRequestPayload(req)
	return c.WriteMsg(message.IDRequest, payload)
}

func (c *Conn) WriteHave(index int) (int, error) {
	payload := message.NewHavePayload(uint32(index))
	return c.WriteMsg(message.IDHave, payload)
}

func (c *Conn) String() string {
	return c.conn.RemoteAddr().String()
}

func (c *Conn) Close() error {
	if c.writeKeepAliveTicker != nil {
		c.writeKeepAliveTicker.Stop()
	}
	return c.conn.Close()
}
