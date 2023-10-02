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
	ID        MessageID
	Payload   []byte
}

func (msg *Message) Len() uint32 {
    if msg.KeepAlive {
        return 0
    } 
    return uint32(1 + len(msg.Payload))
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

func ReadMessage(r io.Reader) (*Message, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	msgLen := binary.BigEndian.Uint32(buf[0:4])
	if msgLen == 0 {
		return &Message{KeepAlive: true}, nil
	}

	buf = make([]byte, msgLen)
	_, err = io.ReadFull(r, buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	id := buf[0]
	payload := buf[1:]

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
