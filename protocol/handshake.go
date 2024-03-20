package protocol

import (
	"bytes"
	"fmt"
	"io"
)

const (
	HandshakeDefaultPstr = "BitTorrent protocol"
	HandshakeConstSize   = 49
)

type Handshake struct {
	Pstr     string
	Reserved [8]byte
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewHandshake(infoHash [20]byte, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     HandshakeDefaultPstr,
		Reserved: [8]byte{0, 0, 0, 0, 0, 0, 0, 0},
		InfoHash: infoHash,
		PeerID:   peerID,
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
	buf.Write(hs.PeerID[:])

	return buf.Bytes()
}

func UnmarshalHandshake(r io.Reader) (*Handshake, error) {
	buf := make([]byte, HandshakeConstSize+len(HandshakeDefaultPstr))

	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	pstrlen := uint8(buf[0])
	if pstrlen != uint8(len(HandshakeDefaultPstr)) {
		return nil, fmt.Errorf("pstrlen unexpected value: %d", pstrlen)
	}

	pstr := string(buf[1 : pstrlen+1])
	if pstr != HandshakeDefaultPstr {
		return nil, fmt.Errorf("pstr unexpected value: %s", pstr)
	}

	reserved := [8]byte(buf[pstrlen+1 : pstrlen+9])
	infoHash := [20]byte(buf[pstrlen+9 : pstrlen+29])
	peerId := [20]byte(buf[pstrlen+29 : pstrlen+49])

	return &Handshake{
		Pstr:     pstr,
		Reserved: reserved,
		InfoHash: infoHash,
		PeerID:   peerId,
	}, nil
}
