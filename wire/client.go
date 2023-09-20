package wire

import (
	"bytes"
	"net"
	"time"
)

const (
    DefaultTimeout = time.Second * 5
    Protocolstr = "BitTorrent protocol"
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

func (c *Client) handshake(infoHash [20]byte) []byte {
    var buf bytes.Buffer
    reserved := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}

    buf.WriteByte(byte(len(Protocolstr)))
    buf.WriteString(Protocolstr)
    buf.Write(reserved[:])
    buf.Write(infoHash[:])
    buf.Write(c.PeerId[:])
    
    return buf.Bytes()
}
