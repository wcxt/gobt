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
	Get(int) (bool, error)
}

func New(size int) Bitfield {
	length := int(math.Ceil(float64(size) / 8))

	return &bitfield{
		field: make([]byte, length),
		size:  size,
	}
}

func (bf *bitfield) Replace(data []byte) error {
	if len(data) != len(bf.field) {
		return fmt.Errorf("invalid replace data size: %d", len(data))
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

func (bf *bitfield) Get(i int) (bool, error) {
	if i >= bf.size || i < 0 {
		return false, fmt.Errorf("invalid index value: %d", i)
	}

	index := i / 8
	offset := i % 8
	bit := bf.field[index] & (0b10000000 >> offset)
	return bit != 0, nil
}
