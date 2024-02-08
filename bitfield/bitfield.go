package bitfield

import (
	"fmt"
	"math"
)

type bitfield struct {
	field []byte
	size  int
}

type Bitfield interface {
	Replace([]byte) error
	Set(int) error
	Clear(int) error
	Size() int
	Get(int) (bool, error)
	Empty() bool
	Difference(Bitfield) (Bitfield, error)
}

func New(size int) Bitfield {
	length := int(math.Ceil(float64(size) / 8))

	return &bitfield{
		field: make([]byte, length),
		size:  size,
	}
}

func (bf *bitfield) Size() int {
	return bf.size
}

func (bf *bitfield) Replace(data []byte) error {
	if len(data) != len(bf.field) {
		return fmt.Errorf("invalid replace data size: %d", len(data))
	}

	spare := int(data[len(bf.field)-1] << (bf.size % 8))
	if spare != 0 {
		return fmt.Errorf("spare bits set")
	}

	bf.field = data
	return nil
}

func (bf *bitfield) Set(i int) error {
	if i >= bf.size || i < 0 {
		return fmt.Errorf("invalid index value: %d", i)
	}

	index := i / 8
	offset := i % 8
	bf.field[index] |= 0b10000000 >> offset
	return nil
}

func (bf *bitfield) Clear(i int) error {
	if i >= bf.size || i < 0 {
		return fmt.Errorf("invalid index value: %d", i)
	}

	index := i / 8
	offset := i % 8
	bf.field[index] &= 0b01111111 >> offset
	return nil
}

func (bf *bitfield) Empty() bool {
	for _, byte := range bf.field {
		if byte > 0 {
			return false
		}
	}

	return true
}

func (bf *bitfield) Get(i int) (bool, error) {
	if i >= bf.size || i < 0 {
		return false, fmt.Errorf("invalid index value: %d", i)
	}

	index := i / 8
	offset := i % 8
	bit := bf.field[index] & (0b10000000 >> offset)
	return bit != 0, nil
}

// PERF: iterating over bytes and using bitwise AND
func (bf *bitfield) Difference(x Bitfield) (Bitfield, error) {
	if bf.size != x.Size() {
		return nil, fmt.Errorf("invalid intersect bitfield size: %d", x.Size())
	}

	inter := New(bf.size)
	for i := range bf.field {
		v1, _ := bf.Get(i)
		v2, _ := x.Get(i)
		if v1 && !v2 {
			inter.Set(i)
		}
	}

	return inter, nil
}
