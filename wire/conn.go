package wire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

const DefaultTimeout = 5 * time.Second

type Block struct {
    Index uint32
    Offset uint32
    Bytes []byte
}

type Request struct {
    Index uint32
    Offset uint32
    Length uint32
}

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
	c.conn.Write(hs.Marshal())

	hs, err := ReadHandshake(c.conn)
	if err != nil {
		return err
	}

	if hs.InfoHash != infoHash {
		return errors.New("invalid field: InfoHash")
	}

	return nil
}

func (c *Conn) Send(msg *Message) (int, error) {
    fmt.Printf("SEND: Message{KeepAlive: %t, ID: %d}\n", msg.KeepAlive, msg.ID)
    return c.conn.Write(msg.Marshal())
}

func (c *Conn) Recv() (*Message, error) {
    msg, err := ReadMessage(c.conn)
    fmt.Printf("RECV: Message{KeepAlive: %t, ID: %d}\n", msg.KeepAlive, msg.ID)
    return msg, err
}

func (c *Conn) SendInterested() (int, error) {
    return c.Send(&Message{ID: MessageInterested})
}

func (c *Conn) SendRequest(req *Request) (int, error) {
    var buf bytes.Buffer

    binary.Write(&buf, binary.BigEndian, req.Index)
    binary.Write(&buf, binary.BigEndian, req.Offset)
    binary.Write(&buf, binary.BigEndian, req.Length)

    return c.Send(&Message{ID: MessageRequest, Payload: buf.Bytes()}) 
}

func (c *Conn) Close() error {
    return c.conn.Close()
}
