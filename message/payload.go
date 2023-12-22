package message

import (
	"bytes"
	"encoding/binary"
)

type Bitfield []byte

func (bf Bitfield) Set(i int) {
	byteI := i / 8
	bitI := i % 8
	bf[byteI] |= 0b10000000 >> bitI
}

func (bf Bitfield) Get(i int) bool {
	byteI := i / 8
	bitI := i % 8
	val := int(bf[byteI] & (0b10000000 >> bitI))
	return val != 0
}

type Request struct {
	Index  uint32
	Offset uint32
	Length uint32
}

type Block struct {
	Index  uint32
	Offset uint32
	Block  []byte
}

type Payload []byte

func NewBitfieldPayload(bitfield Bitfield) Payload {
	return Payload(bitfield)
}

func NewRequestPayload(request Request) Payload {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, request.Index)
	binary.Write(&buf, binary.BigEndian, request.Offset)
	binary.Write(&buf, binary.BigEndian, request.Length)

	return Payload(buf.Bytes())
}

func NewBlockPayload(block Block) Payload {
    var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, block.Index)
	binary.Write(&buf, binary.BigEndian, block.Offset)
    buf.Write(block.Block)

	return Payload(buf.Bytes())
}

func NewHavePayload(index uint32) Payload {
    var buf bytes.Buffer
    binary.Write(&buf, binary.BigEndian, index)

    return Payload(buf.Bytes())
}

func (p Payload) Bitfield() Bitfield {
	return Bitfield(p)
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

