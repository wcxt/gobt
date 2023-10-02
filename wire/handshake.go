package wire

import (
	"bytes"
	"fmt"
	"io"
)

const Pstr = "BitTorrent protocol"

type Handshake struct {
    Pstr string
    Reserved [8]byte
    InfoHash [20]byte
    PeerId [20]byte
}

func NewHandshake(infoHash [20]byte, peerId [20]byte) *Handshake {
    return &Handshake{
        Pstr: Pstr,
        Reserved: [8]byte{0, 0, 0, 0, 0, 0, 0, 0}, 
        InfoHash: infoHash,
        PeerId: peerId,
    }
}

func (hs *Handshake) PstrLen() uint8 {
    return uint8(len(hs.Pstr))
}

func (hs *Handshake) Marshal() []byte {
    var buf bytes.Buffer

    buf.WriteByte(byte(hs.PstrLen()))
    buf.WriteString(hs.Pstr)
    buf.Write(hs.Reserved[:])
    buf.Write(hs.InfoHash[:])
    buf.Write(hs.PeerId[:])
    
    return buf.Bytes()
}

func ReadHandshake(r io.Reader) (*Handshake, error) {
    buf := make([]byte, 49 + len(Pstr))
    _, err := io.ReadFull(r, buf)
    if err != nil {
        return nil, err
    }

    pstrlen := uint8(buf[0])
    if pstrlen != uint8(len(Pstr)) {
        return nil, fmt.Errorf("pstrlen unexpected value: %d", pstrlen)
    }

    pstr := string(buf[1:pstrlen+1])
    if pstr != Pstr {
        return nil, fmt.Errorf("pstr unexpected value: %s", pstr)
    }
    
    reserved := [8]byte(buf[pstrlen+1:pstrlen+9])
    infoHash := [20]byte(buf[pstrlen+9:pstrlen+29])
    peerId := [20]byte(buf[pstrlen+29:pstrlen+49])

    return &Handshake{
        Pstr: pstr,
        Reserved: reserved,
        InfoHash: infoHash,
        PeerId: peerId,
    }, nil
}


