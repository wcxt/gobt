package wire

import (
	"bytes"
	"encoding/binary"
	"io"
)

type MessageID uint8

const (
    MessageChoke MessageID = iota
    MessageUnchoke
    MessageInterested
    MessageUninterested
    MessageHave
    MessageBitfield
    MessageRequest
    MessagePiece
    MessageCancel
    MessagePort
)

type Message struct {
    KeepAlive bool
    ID MessageID
    Payload []byte
}

func MarshalMessage(msg *Message) []byte {
    var buf bytes.Buffer

    if msg.KeepAlive {
        binary.Write(&buf, binary.BigEndian, uint64(0))
        return buf.Bytes() 
    }
    
    binary.Write(&buf, binary.BigEndian, 1 + len(msg.Payload))
    buf.WriteByte(byte(msg.ID))
    buf.Write(msg.Payload)

    return buf.Bytes()
}

func UnmarshalMessage(r io.Reader) (*Message, error) {
    var buf []byte

    _, err := io.ReadFull(r, buf)
    if err != nil && err != io.EOF {
        return nil, err
    }

    msgLen := binary.BigEndian.Uint32(buf[0:4])
    if msgLen == 0 {
        return &Message{KeepAlive: true}, nil
    }

    id := uint8(buf[4])
    payload := buf[5:]

    return &Message{ID: MessageID(id), Payload: payload}, nil
}
