package wire

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/edwces/gobt/wire/handshake"
	"github.com/edwces/gobt/wire/message"
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
    _, err := handshake.Write(c.conn, handshake.New(infoHash, peerId))
    if err != nil {
        return err
    }

	hs, err := handshake.Read(c.conn)
	if err != nil {
        return err
	}

	if hs.InfoHash != infoHash {
		return errors.New("invalid field: InfoHash")
	}

	return nil
}

func (c *Conn) Send(msg *message.Message) (int, error) {
    fmt.Printf("SEND: Message{KeepAlive: %t, ID: %d}\n", msg.KeepAlive, msg.ID)
    return message.Write(c.conn, msg) 
}

func (c *Conn) Recv() (*message.Message, error) {
    msg, err := message.Read(c.conn)
    fmt.Printf("RECV: Message{KeepAlive: %t, ID: %d}\n", msg.KeepAlive, msg.ID)
    return msg, err
}

func (c *Conn) SendInterested() (int, error) {
    return message.Write(c.conn, &message.Message{ID: message.IDInterested})
}

func (c *Conn) SendRequest(req message.Request) (int, error) {
    return message.Write(c.conn, &message.Message{ID: message.IDRequest, Payload: message.NewRequestPayload(req)}) 
}

func (c *Conn) Close() error {
    return c.conn.Close()
}
