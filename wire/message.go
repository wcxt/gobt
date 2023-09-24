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

func NewBitfieldMessage(bf Bitfield) *Message {
    return &Message{
        ID: MessageBitfield,
        Payload: bf,
    }
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
    lenBuf := make([]byte, 4)
    _, err := io.ReadFull(r, lenBuf)
    if err != nil && err != io.EOF {
        return nil, err
    }

    msgLen := binary.BigEndian.Uint32(lenBuf[0:4])
    if msgLen == 0 {
        return &Message{KeepAlive: true}, nil
    }

    restBuf := make([]byte, msgLen)
    _, err = io.ReadFull(r, restBuf)
    if err != nil && err != io.EOF {
        return nil, err
    }

    id := uint8(restBuf[0])
    payload := restBuf[1:]

    return &Message{ID: MessageID(id), Payload: payload}, nil
}

type Bitfield []byte

func (bf Bitfield) Set(i int) {
    byteI := i / 8
    bitI := i % 8
    bf[byteI] |= 0b00000001 << bitI
}

func (bf Bitfield) Get(i int) bool {
    byteI := i / 8
    bitI := i % 8
    val := int(bf[byteI] & (0b00000001 << bitI))
    return val != 0
}
