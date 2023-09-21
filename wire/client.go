package wire

import (
	"bytes"
	"errors"
	"io"
	"net"
	"time"
)

const (
    DefaultTimeout = time.Second * 5
    Protocolstr = "BitTorrent protocol"
)

type Handshake struct {
    Pstrlen uint8
    Pstr string
    Reserved [8]byte
    InfoHash [20]byte
    PeerId [20]byte
}

func NewHandshake(infoHash [20]byte, peerId [20]byte) *Handshake {
    return &Handshake{
        Pstrlen: uint8(len(Protocolstr)),
        Pstr: Protocolstr,
        Reserved: [8]byte{0, 0, 0, 0, 0, 0, 0, 0}, 
        InfoHash: infoHash,
        PeerId: peerId,
    }
}

func MarshalHandshake(hs *Handshake) []byte {
    var buf bytes.Buffer

    buf.WriteByte(byte(hs.Pstrlen))
    buf.WriteString(hs.Pstr)
    buf.Write(hs.Reserved[:])
    buf.Write(hs.InfoHash[:])
    buf.Write(hs.PeerId[:])
    
    return buf.Bytes()
}

func UnmarshalHandshake(r io.Reader) (*Handshake, error) {
    buf := make([]byte, 49 + len(Protocolstr))

    _, err := io.ReadFull(r, buf)
    if err != nil && err != io.EOF {
        return nil, err
    }

    pstrlen := uint8(buf[0])
    if pstrlen != uint8(len(Protocolstr)) {
        return nil, errors.New("invalid field: Pstrlen")
    }

    pstr := string(buf[1:pstrlen+1])
    if pstr != Protocolstr {
        return nil, errors.New("invalid field: Pstr")
    }
    
    var reserved [8]byte
    var infoHash, peerId [20]byte

    copy(reserved[:], buf[pstrlen+1:pstrlen+9])
    copy(infoHash[:], buf[pstrlen+9:pstrlen+29])
    copy(peerId[:], buf[pstrlen+29:pstrlen+49])

    return &Handshake{
        Pstrlen: pstrlen,
        Pstr: pstr,
        Reserved: reserved,
        InfoHash: infoHash,
        PeerId: peerId,
    }, nil
}


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
