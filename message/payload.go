package message

import (
	"bytes"
	"encoding/binary"
)

type Request struct {
	Index  uint32
	Offset uint32
	Length uint32
}

func (r *Request) Marshal() []byte {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, r.Index)
	binary.Write(&buf, binary.BigEndian, r.Offset)
	binary.Write(&buf, binary.BigEndian, r.Length)

	return buf.Bytes()
}

type Block struct {
	Index  uint32
	Offset uint32
	Block  []byte
}

func (b *Block) Marshal() []byte {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, b.Index)
	binary.Write(&buf, binary.BigEndian, b.Offset)
	buf.Write(b.Block)

	return buf.Bytes()
}

type Have uint32

func (h Have) Marshal() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, h)

	return buf.Bytes()
}

type Payload []byte

func NewHavePayload(index uint32) Payload {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, index)

	return Payload(buf.Bytes())
}

func (p Payload) Request() Request {
	index := binary.BigEndian.Uint32(p[0:4])
	offset := binary.BigEndian.Uint32(p[4:8])
	length := binary.BigEndian.Uint32(p[8:12])

	return Request{Index: index, Offset: offset, Length: length}
}

func (p Payload) Block() Block {
	index := binary.BigEndian.Uint32(p[0:4])
	offset := binary.BigEndian.Uint32(p[4:8])
	block := p[8:]

	return Block{Index: index, Offset: offset, Block: block}
}

func (p Payload) Have() uint32 {
	return binary.BigEndian.Uint32(p[0:4])
}
