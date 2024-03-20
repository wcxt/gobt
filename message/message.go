package message

import (
	"bytes"
	"encoding/binary"
	"io"
)

type ID uint8

const (
	IDChoke ID = iota
	IDUnchoke
	IDInterested
	IDNotInterested
	IDHave
	IDBitfield
	IDRequest
	IDPiece
	IDCancel
	IDPort
)

var stringMap = map[ID]string{
	IDChoke:         "CHOKE",
	IDUnchoke:       "UNCHOKE",
	IDInterested:    "INTERESTED",
	IDNotInterested: "NOTINTERESTED",
	IDHave:          "HAVE",
	IDBitfield:      "BITFIELD",
	IDRequest:       "REQUEST",
	IDPiece:         "PIECE",
	IDCancel:        "CANCEL",
	IDPort:          "PORT",
}

type Message struct {
	KeepAlive bool
	ID        ID
	Payload   Payload
}

func (msg *Message) Len() uint32 {
	if msg.KeepAlive {
		return 0
	}
	return uint32(1 + len(msg.Payload))
}

func (msg *Message) String() string {
	if msg.KeepAlive {
		return "KEEPALIVE"
	}

	return stringMap[msg.ID]
}

func (msg *Message) Marshal() []byte {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, msg.Len())

	if msg.KeepAlive {
		return buf.Bytes()
	}

	buf.WriteByte(byte(msg.ID))
	buf.Write(msg.Payload)
	return buf.Bytes()

}

func UnmarshalMessage(r io.Reader) (*Message, error) {
	buf := make([]byte, 4)

	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	msgLength := binary.BigEndian.Uint32(buf[0:4])
	if msgLength == 0 {
		return &Message{KeepAlive: true}, nil
	}

	buf = make([]byte, msgLength)

	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	id := ID(buf[0])
	payload := Payload(buf[1:])

	return &Message{ID: id, Payload: payload}, nil
}
